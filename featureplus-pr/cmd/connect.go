package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ConnectResponse represents the API response from connect endpoint
type ConnectResponse struct {
	Status      string    `json:"status"`
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Path        string    `json:"path"`
	ConnectedAt time.Time `json:"connected_at"`
}

// connectCmd represents the connect command
var connectCmd = &cobra.Command{
	Use:   "connect <project-id>",
	Short: "Connect this directory to a FeaturePlus project",
	Long:  `Links the current directory to a FeaturePlus project by its ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectID := args[0]

		// Read config.yaml
		configPath := filepath.Join(".featureplus", "config.yaml")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("❌ Error: Could not read config file. Have you run 'featureplus-pr init'?\n")
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		var config Config
		if err := yaml.Unmarshal(configData, &config); err != nil {
			fmt.Printf("❌ Error: Invalid config file format\n")
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		// Get current directory
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("❌ Error: Could not get current directory\n")
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		// Prepare request body
		requestBody := map[string]string{
			"path": currentDir,
		}
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Printf("❌ Error: Could not prepare request\n")
			os.Exit(1)
		}

		// Call API
		apiURL := config.APIURL
		if apiURL == "" {
			apiURL = "http://localhost:8080"
		}
		url := fmt.Sprintf("%s/api/projects/%s/connect", apiURL, projectID)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			fmt.Printf("❌ Error: Could not create request\n")
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("❌ Error: Could not reach server at %s\n", apiURL)
			fmt.Printf("   Details: %v\n", err)
			fmt.Printf("   Make sure the FeaturePlus server is running.\n")
			os.Exit(1)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("❌ Error: Could not read server response\n")
			os.Exit(1)
		}

		// Check for errors
		if resp.StatusCode != http.StatusOK {
			var errorResp map[string]interface{}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				if errMsg, ok := errorResp["error"].(string); ok {
					fmt.Printf("❌ Error: %s\n", errMsg)
					if details, ok := errorResp["details"].(string); ok {
						fmt.Printf("   Details: %s\n", details)
					}
				} else {
					fmt.Printf("❌ Error: Server returned status %d\n", resp.StatusCode)
				}
			} else {
				fmt.Printf("❌ Error: Server returned status %d\n", resp.StatusCode)
			}
			os.Exit(1)
		}

		// Parse success response
		var connectResp ConnectResponse
		if err := json.Unmarshal(body, &connectResp); err != nil {
			fmt.Printf("❌ Error: Could not parse server response\n")
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		// Update config.yaml
		config.ProjectID = connectResp.ProjectID
		config.LinkedAt = connectResp.ConnectedAt.Format(time.RFC3339)

		updatedYAML, err := yaml.Marshal(&config)
		if err != nil {
			fmt.Printf("⚠️  Warning: Could not update config file\n")
			fmt.Printf("   Details: %v\n", err)
		} else {
			if err := os.WriteFile(configPath, updatedYAML, 0644); err != nil {
				fmt.Printf("⚠️  Warning: Could not save config file\n")
				fmt.Printf("   Details: %v\n", err)
			}
		}

		// Print success message
		fmt.Printf("✅ Linked this folder to FeaturePlus project %s (%s) at %s\n",
			connectResp.ProjectName,
			connectResp.ProjectID,
			connectResp.ConnectedAt.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
