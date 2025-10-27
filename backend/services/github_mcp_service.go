package services

import (
	// "bytes"  // REMOVED: was used for direct MCP calls, now using MCP Bridge
	// "encoding/json" // REMOVED: was used for direct MCP calls, now using MCP Bridge
	"fmt"
	// "io"     // REMOVED: was used for direct MCP calls, now using MCP Bridge
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/repositories"
)

// GitHubMCPService handles interactions with GitHub MCP API
type GitHubMCPService struct {
	apiURL      string
	githubToken string
	client      *http.Client
}

// NewGitHubMCPService creates a new GitHub MCP service
func NewGitHubMCPService() *GitHubMCPService {
	// Read GitHub token from environment variable
	// Try GITHUB_TOKEN first (existing), then fall back to GITHUB_PAT
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_PAT")
	}
	if githubToken == "" {
		log.Println("WARNING: GITHUB_TOKEN environment variable not set. GitHub MCP integration will not work.")
	}

	return &GitHubMCPService{
		apiURL:      "https://api.githubcopilot.com/mcp/",
		githubToken: githubToken,
		client: &http.Client{
			Timeout: 60 * time.Second, // MCP analysis can take time
		},
	}
}

// MCPRequest represents the request payload to GitHub MCP
type MCPRequest struct {
	RepoURL string `json:"repo_url"`
	Prompt  string `json:"prompt"`
}

// MCPResponse represents the response from GitHub MCP
type MCPResponse struct {
	Analysis string                  `json:"analysis"`
	Template repositories.Template   `json:"template,omitempty"`
	Error    string                  `json:"error,omitempty"`
}

// AnalyzeRepository calls GitHub MCP to analyze a repository and return structured project data
func (s *GitHubMCPService) AnalyzeRepository(repoURL string) (*repositories.Template, error) {
	log.Printf("INFO: Starting GitHub MCP analysis for repository: %s", repoURL)

	// Step 1: Validate GitHub token
	if s.githubToken == "" {
		log.Printf("ERROR: GitHub token not configured")
		return nil, fmt.Errorf("GitHub token not configured. Please set GITHUB_TOKEN environment variable")
	}

	// Step 2: Extract repo owner and name from URL
	repoInfo, err := s.extractRepoInfo(repoURL)
	if err != nil {
		log.Printf("ERROR: Failed to extract repo info: %v", err)
		return nil, err
	}
	log.Printf("INFO: Extracted repo info - Owner: %s, Name: %s", repoInfo.Owner, repoInfo.Name)

	// REMOVED: direct call to https://api.githubcopilot.com/mcp/ — replaced with local MCP Bridge call (CallLocalMCPBridge)
	// TODO: The code below attempted to call the remote MCP endpoint directly but got "Invalid session ID"
	// This has been replaced with the local MCP Bridge approach in import_handler.go
	
	/*
	// Step 3: Construct the MCP prompt with specific instructions for FeaturePlus format
	prompt := s.buildMCPPrompt(repoInfo)
	log.Printf("DEBUG: MCP prompt constructed (%d chars)", len(prompt))

	// Step 4: Create the MCP API request
	requestBody := MCPRequest{
		RepoURL: repoURL,
		Prompt:  prompt,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("ERROR: Failed to marshal MCP request: %v", err)
		return nil, fmt.Errorf("failed to create MCP request: %w", err)
	}

	// Step 5: Send request to GitHub MCP API
	log.Printf("INFO: Sending request to GitHub MCP API: %s", s.apiURL)
	req, err := http.NewRequest("POST", s.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ERROR: Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Step 6: Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.githubToken))
	req.Header.Set("User-Agent", "FeaturePlus/1.0")

	// Step 7: Execute the request
	log.Printf("INFO: Executing MCP API request...")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("ERROR: MCP API request failed: %v", err)
		return nil, fmt.Errorf("MCP API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Step 8: Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read MCP response: %v", err)
		return nil, fmt.Errorf("failed to read MCP response: %w", err)
	}

	log.Printf("INFO: MCP API response status: %d, body length: %d bytes", resp.StatusCode, len(body))

	// Step 9: Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: MCP API returned error status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("MCP API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Step 10: Parse the MCP response
	var mcpResponse MCPResponse
	if err := json.Unmarshal(body, &mcpResponse); err != nil {
		log.Printf("ERROR: Failed to parse MCP response JSON: %v", err)
		log.Printf("DEBUG: Response body: %s", string(body))
		return nil, fmt.Errorf("failed to parse MCP response: %w", err)
	}

	// Step 11: Check for errors in the response
	if mcpResponse.Error != "" {
		log.Printf("ERROR: MCP returned error: %s", mcpResponse.Error)
		return nil, fmt.Errorf("MCP analysis error: %s", mcpResponse.Error)
	}
	*/
	
	// This service is now deprecated in favor of the MCP Bridge
	return nil, fmt.Errorf("direct MCP calls are disabled - use MCP Bridge instead")

	/*
	// Step 12: Extract the template from the response
	// If MCP returns a direct template, use it
	if mcpResponse.Template.ID != "" {
		log.Printf("INFO: MCP returned structured template with ID: %s", mcpResponse.Template.ID)
		return &mcpResponse.Template, nil
	}

	// Step 13: If MCP returns analysis text, try to parse it as JSON
	if mcpResponse.Analysis != "" {
		log.Printf("INFO: MCP returned analysis text, attempting to parse as JSON template")
		var template repositories.Template
		if err := json.Unmarshal([]byte(mcpResponse.Analysis), &template); err != nil {
			log.Printf("ERROR: Failed to parse analysis as template: %v", err)
			return nil, fmt.Errorf("MCP analysis could not be parsed as template: %w", err)
		}
		
		// Validate the parsed template has required fields
		if template.ID == "" {
			template.ID = repoInfo.Name
		}
		if template.Name == "" {
			template.Name = repoInfo.Name
		}
		
		log.Printf("INFO: Successfully parsed template from analysis - %d features, %d tasks", 
			len(template.Features), len(template.Tasks))
		return &template, nil
	}

	// Step 14: No valid template found
	log.Printf("ERROR: MCP response did not contain valid template data")
	return nil, fmt.Errorf("MCP response did not contain valid template data")
	*/
}

// RepoInfo holds extracted repository information
type RepoInfo struct {
	Owner string
	Name  string
	URL   string
}

// extractRepoInfo extracts owner and repo name from GitHub URL
func (s *GitHubMCPService) extractRepoInfo(repoURL string) (*RepoInfo, error) {
	// Remove trailing slashes and .git suffix
	repoURL = strings.TrimSuffix(strings.TrimSpace(repoURL), "/")
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Expected format: https://github.com/owner/repo
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid GitHub URL format: %s", repoURL)
	}

	// Extract owner and repo name (last two parts)
	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]

	if owner == "" || name == "" {
		return nil, fmt.Errorf("could not extract owner/repo from URL: %s", repoURL)
	}

	return &RepoInfo{
		Owner: owner,
		Name:  name,
		URL:   repoURL,
	}, nil
}

// buildMCPPrompt constructs a detailed prompt for GitHub MCP to generate FeaturePlus-compatible JSON
func (s *GitHubMCPService) buildMCPPrompt(repoInfo *RepoInfo) string {
	return fmt.Sprintf(`Analyze the GitHub repository "%s/%s" and generate a structured JSON output for a project management system.

The JSON must follow this exact structure:

{
  "id": "repository-name",
  "name": "Human-readable Project Name",
  "stack": "Technology stack description",
  "description": "Detailed project description",
  "tech_stack": "Primary technology (e.g., React, Go, Python, Node.js)",
  "feature_categories": ["Category1", "Category2", ...],
  "task_types": ["Type1", "Type2", ...],
  "features": [
    {
      "name": "Feature name",
      "category": "Category from feature_categories",
      "description": "Detailed feature description",
      "context": "Development"
    }
  ],
  "tasks": [
    {
      "name": "Task name",
      "type": "Type from task_types",
      "description": "Detailed task description",
      "priority": "high|medium|low",
      "context": "Development"
    }
  ],
  "dependencies": ["List of dependencies"],
  "setup_steps": ["Step 1", "Step 2", ...],
  "environment_variables": ["VAR1", "VAR2", ...],
  "starter_repo": "%s",
  "docs_links": ["Link1", "Link2", ...]
}

Instructions:
1. Analyze the repository's README, package files, and code structure
2. Identify main features and break them into logical categories
3. Create actionable tasks for setting up and developing the project
4. Set all feature and task "context" to "Development" by default
5. Prioritize tasks: high (critical setup), medium (important features), low (nice-to-have)
6. Include actual dependencies from package.json, requirements.txt, go.mod, etc.
7. Extract setup steps from README or infer from project structure
8. List environment variables found in .env.example or documentation

Return ONLY the JSON object, no additional text or markdown formatting.`, 
		repoInfo.Owner, repoInfo.Name, repoInfo.URL)
}
