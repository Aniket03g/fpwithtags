package mcpbridge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client represents the MCP Bridge client
type Client struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewClient creates a new MCP Bridge client
func NewClient() *Client {
	// Read configuration from environment
	baseURL := os.Getenv("MCP_BRIDGE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8089"
	}

	apiKey := os.Getenv("MCP_BRIDGE_API_KEY")
	if apiKey == "" {
		// Default for local development
		apiKey = "featureplus-local"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 100 * time.Second, // Slightly longer than bridge timeout
		},
	}
}

// AnalyzeRequest represents the request to analyze a repository
type AnalyzeRequest struct {
	RepoURL string `json:"repo_url"`
	Format  string `json:"format"`
}

// CallLocalMCPBridge calls the local MCP Bridge to analyze a GitHub repository
func CallLocalMCPBridge(repoURL string) ([]byte, error) {
	client := NewClient()
	return client.Analyze(repoURL)
}

// Analyze sends an analysis request to the MCP Bridge
func (c *Client) Analyze(repoURL string) ([]byte, error) {
	// Prepare request payload
	request := AnalyzeRequest{
		RepoURL: repoURL,
		Format:  "featureplus",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/mcp/analyze", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to MCP Bridge: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("unauthorized: invalid MCP Bridge API key")
	case http.StatusBadRequest:
		var errorResp map[string]string
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["error"]; ok {
				return nil, fmt.Errorf("bad request: %s", msg)
			}
		}
		return nil, fmt.Errorf("bad request: invalid repository URL or format")
	case http.StatusTooManyRequests:
		var errorResp map[string]string
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["error"]; ok {
				return nil, fmt.Errorf("rate limited: %s", msg)
			}
		}
		return nil, fmt.Errorf("rate limited: another analysis is in progress")
	case http.StatusGatewayTimeout:
		return nil, fmt.Errorf("analysis timeout: repository analysis took too long")
	case http.StatusInternalServerError:
		var errorResp map[string]string
		if err := json.Unmarshal(body, &errorResp); err == nil {
			if msg, ok := errorResp["error"]; ok {
				return nil, fmt.Errorf("bridge error: %s", msg)
			}
		}
		return nil, fmt.Errorf("internal server error in MCP Bridge")
	default:
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}
}

// CheckHealth checks if the MCP Bridge is healthy
func (c *Client) CheckHealth() error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("MCP Bridge unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("MCP Bridge unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
