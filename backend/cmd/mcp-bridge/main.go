package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds the bridge configuration
type Config struct {
	Port                  string
	MCPBridgeAPIKey      string
	GitHubToken          string
	AnalysisTimeout      time.Duration
	MCPContainerImage    string
}

// AnalyzeRequest represents the incoming request payload
type AnalyzeRequest struct {
	RepoURL string `json:"repo_url"`
	Format  string `json:"format"`
}

// MCPBridge handles MCP analysis requests
type MCPBridge struct {
	config Config
	mu     sync.Mutex // Ensures single-concurrency for MCP analysis
}

func main() {
	// Load configuration from environment
	config := Config{
		Port:              getEnv("MCP_BRIDGE_PORT", "8089"),
		MCPBridgeAPIKey:   getEnv("MCP_BRIDGE_API_KEY", ""),
		GitHubToken:       getEnv("GITHUB_PERSONAL_ACCESS_TOKEN", ""),
		AnalysisTimeout:   parseDuration(getEnv("MCP_ANALYSIS_TIMEOUT", "90s"), 90*time.Second),
		MCPContainerImage: getEnv("MCP_CONTAINER_IMAGE", "ghcr.io/github/github-mcp-server"),
	}

	// Validate required configuration
	if config.MCPBridgeAPIKey == "" {
		log.Fatal("ERROR: MCP_BRIDGE_API_KEY environment variable is required")
	}
	if config.GitHubToken == "" {
		log.Fatal("ERROR: GITHUB_PERSONAL_ACCESS_TOKEN environment variable is required")
	}

	// Create bridge instance
	bridge := &MCPBridge{
		config: config,
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp/analyze", bridge.handleAnalyze)
	mux.HandleFunc("/health", handleHealth)

	// Bind to localhost only for security
	addr := fmt.Sprintf("127.0.0.1:%s", config.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: config.AnalysisTimeout + 10*time.Second, // Allow extra time for response
	}

	log.Printf("INFO: MCP Bridge starting on %s", addr)
	log.Printf("INFO: Using MCP container image: %s", config.MCPContainerImage)
	log.Printf("INFO: Analysis timeout: %v", config.AnalysisTimeout)
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("ERROR: Failed to start server: %v", err)
	}
}

// handleAnalyze processes MCP analysis requests
func (b *MCPBridge) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Authorization header
	authHeader := r.Header.Get("Authorization")
	expectedAuth := fmt.Sprintf("Bearer %s", b.config.MCPBridgeAPIKey)
	if authHeader != expectedAuth {
		log.Printf("WARN: Invalid authorization attempt from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to parse request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default format if not specified
	if req.Format == "" {
		req.Format = "featureplus"
	}

	// Validate repo URL (basic validation)
	if !isValidGitHubURL(req.RepoURL) {
		log.Printf("WARN: Invalid GitHub URL: %s", req.RepoURL)
		http.Error(w, "Invalid GitHub repository URL", http.StatusBadRequest)
		return
	}

	// Try to acquire lock for single-concurrency
	if !b.mu.TryLock() {
		log.Printf("INFO: Concurrent request rejected - analysis already in progress")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Analysis already in progress. Please try again later.",
		})
		return
	}
	defer b.mu.Unlock()

	log.Printf("INFO: Starting MCP analysis for repo: %s (format: %s)", req.RepoURL, req.Format)

	// Run the analysis with timeout
	ctx, cancel := context.WithTimeout(r.Context(), b.config.AnalysisTimeout)
	defer cancel()

	result, err := b.runAnalysis(ctx, req.RepoURL, req.Format)
	if err != nil {
		log.Printf("ERROR: MCP analysis failed: %v", err)
		
		// Determine appropriate error code
		statusCode := http.StatusInternalServerError
		if ctx.Err() == context.DeadlineExceeded {
			statusCode = http.StatusGatewayTimeout
		}
		
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Analysis failed: %v", err),
		})
		return
	}

	log.Printf("INFO: MCP analysis completed successfully for %s", req.RepoURL)

	// Return the raw MCP response
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

// runAnalysis spawns the MCP container and performs the analysis
func (b *MCPBridge) runAnalysis(ctx context.Context, repoURL, format string) ([]byte, error) {
	// Prepare the MCP request payload
	mcpRequest := map[string]interface{}{
		"repo_url": repoURL,
		"format":   format,
		"prompt": `Analyze this GitHub repository and generate a FeaturePlus-compatible JSON structure with features and tasks.
Return a JSON object with: id, name, description, tech_stack, features (array), tasks (array), dependencies, setup_steps, environment_variables.
Each feature should have: name, category, description, context (set to "Development").
Each task should have: name, type, description, priority (high/medium/low), context (set to "Development").`,
	}

	requestJSON, err := json.Marshal(mcpRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Spawn the Docker container
	// Note: Using stdio communication with the MCP server
	cmd := exec.CommandContext(ctx, "docker", "run",
		"--rm",                                    // Remove container after exit
		"-i",                                       // Interactive (for stdio)
		"--network=none",                          // No network access for security
		"-e", fmt.Sprintf("GITHUB_TOKEN=%s", b.config.GitHubToken), // Pass GitHub token
		b.config.MCPContainerImage,
	)

	// Setup pipes for stdio communication
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	defer stderr.Close()

	// Start the container
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP container: %w", err)
	}

	// Ensure the process is killed on timeout or error
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Log stderr in a separate goroutine for debugging
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("MCP STDERR: %s", scanner.Text())
		}
	}()

	// Write the framed request to stdin
	if err := writeFramedMessage(stdin, requestJSON); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Read the framed response from stdout
	response, err := readFramedResponse(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Wait for the container to exit
	if err := cmd.Wait(); err != nil {
		// Log but don't fail if container exits with non-zero code
		// as we may have already received the response
		log.Printf("WARN: MCP container exited with error: %v", err)
	}

	return response, nil
}

// writeFramedMessage writes a Content-Length framed message to the writer
func writeFramedMessage(w io.Writer, data []byte) error {
	// MCP uses Content-Length framing: "Content-Length: N\r\n\r\n<payload>"
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	
	// Write header
	if _, err := w.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	
	// Write payload
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}
	
	return nil
}

// readFramedResponse reads a Content-Length framed response from the reader
func readFramedResponse(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)
	
	// Read headers until we find Content-Length
	var contentLength int
	foundContentLength := false
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}
		
		line = strings.TrimSpace(line)
		
		// Empty line indicates end of headers
		if line == "" {
			if !foundContentLength {
				return nil, fmt.Errorf("Content-Length header not found")
			}
			break
		}
		
		// Parse Content-Length header
		if strings.HasPrefix(line, "Content-Length:") {
			lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
			length, err := strconv.Atoi(lengthStr)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length: %s", lengthStr)
			}
			contentLength = length
			foundContentLength = true
		}
	}
	
	// Read the payload based on Content-Length
	payload := make([]byte, contentLength)
	_, err := io.ReadFull(reader, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to read payload: %w", err)
	}
	
	return payload, nil
}

// isValidGitHubURL performs basic validation of GitHub repository URLs
func isValidGitHubURL(url string) bool {
	// Basic validation: must contain github.com and have a path
	if !strings.Contains(url, "github.com") {
		return false
	}
	
	// Should have format: https://github.com/owner/repo
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return false
	}
	
	// Check protocol
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return false
	}
	
	return true
}

// handleHealth provides a simple health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "mcp-bridge",
	})
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string, defaultValue time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return defaultValue
	}
	return d
}
