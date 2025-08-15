package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"featureplus-pr/internal/config"
	"github.com/spf13/cobra"
)

// FinalizeResponse represents the API response for release finalization
type FinalizeResponse struct {
	Status string `json:"status"`
	Tag    string `json:"tag"`
}

var releaseFinalizeCmd = &cobra.Command{
	Use:   "finalize <release-id>",
	Short: "Finalize a release",
	Long: `Finalize a release by triggering the backend API.

This will create a Git branch and tag, cherry-pick commits from linked PRs,
push the branch and tag to remote, and update the release status in the database.

Example:
  featureplus-pr release finalize 12`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the release ID from arguments
		releaseID := args[0]

		// Finalize the release
		if err := finalizeRelease(releaseID); err != nil {
			return fmt.Errorf("error finalizing release: %v", err)
		}

		return nil
	},
}

func finalizeRelease(releaseID string) error {
	// Get API base URL from config
	config.LoadConfig()
	baseURL := config.GetAPIURL()

	// Create HTTP request
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/api/releases/%s/finalize", baseURL, releaseID)
	
	// Create an empty JSON body ({})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte("{}"))) 
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response FinalizeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Print success message
	fmt.Printf("✅ Release %s finalized successfully with tag: %s\n", releaseID, response.Tag)
	return nil
}

func init() {
	// Add the finalize command to the release command
	ReleaseCmd.AddCommand(releaseFinalizeCmd)
}