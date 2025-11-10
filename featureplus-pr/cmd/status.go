package cmd

import (
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

// StatusResponse represents the API response from status endpoint
type StatusResponse struct {
	Status      string     `json:"status"`
	ProjectID   string     `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	Path        string     `json:"path,omitempty"`
	ConnectedAt *time.Time `json:"connected_at,omitempty"`
}

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show connection status to FeaturePlus project",
	Long:  `Displays the current connection status and project information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read config.yaml
		configPath := filepath.Join(".featureplus", "config.yaml")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			fmt.Printf("%s⚠️  Not connected to any FeaturePlus project.%s\n", colorYellow, colorReset)
			fmt.Printf("   Run 'featureplus-pr init' to initialize this directory.\n")
			os.Exit(0)
		}

		var config Config
		if err := yaml.Unmarshal(configData, &config); err != nil {
			fmt.Printf("%s❌ Error: Invalid config file format%s\n", colorRed, colorReset)
			os.Exit(1)
		}

		// Check if project_id is set
		if config.ProjectID == "" {
			fmt.Printf("%s⚠️  Not connected to any FeaturePlus project.%s\n", colorYellow, colorReset)
			fmt.Printf("   Run 'featureplus-pr connect <project-id>' to link this directory.\n")
			os.Exit(0)
		}

		// Call API to get status
		apiURL := config.APIURL
		if apiURL == "" {
			apiURL = "http://localhost:8080"
		}
		url := fmt.Sprintf("%s/api/projects/%s/status", apiURL, config.ProjectID)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("%s❌ Error: Could not reach server at %s%s\n", colorRed, apiURL, colorReset)
			fmt.Printf("   Details: %v\n", err)
			fmt.Printf("   Make sure the FeaturePlus server is running.\n")
			os.Exit(1)
		}
		defer resp.Body.Close()

		// Read response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s❌ Error: Could not read server response%s\n", colorRed, colorReset)
			os.Exit(1)
		}

		// Parse response
		var statusResp StatusResponse
		if err := json.Unmarshal(body, &statusResp); err != nil {
			fmt.Printf("%s❌ Error: Could not parse server response%s\n", colorRed, colorReset)
			fmt.Printf("   Details: %v\n", err)
			os.Exit(1)
		}

		// Pretty print status
		fmt.Println()
		fmt.Printf("%s╔════════════════════════════════════════════════╗%s\n", colorCyan, colorReset)
		fmt.Printf("%s║         FeaturePlus Connection Status         ║%s\n", colorCyan, colorReset)
		fmt.Printf("%s╚════════════════════════════════════════════════╝%s\n", colorCyan, colorReset)
		fmt.Println()

		// Project name or ID
		projectDisplay := config.ProjectID
		if statusResp.ProjectName != "" {
			projectDisplay = fmt.Sprintf("%s (%s)", statusResp.ProjectName, statusResp.ProjectID)
		}
		fmt.Printf("📦 %sProject:%s       %s\n", colorBlue, colorReset, projectDisplay)

		// Server URL
		fmt.Printf("🌐 %sServer:%s        %s\n", colorBlue, colorReset, apiURL)

		// Local path
		currentDir, _ := os.Getwd()
		displayPath := statusResp.Path
		if displayPath == "" {
			displayPath = currentDir
		}
		fmt.Printf("📂 %sPath:%s          %s\n", colorBlue, colorReset, displayPath)

		// Connection status
		connected := "no"
		connectedColor := colorRed
		if statusResp.Status == "linked" {
			connected = "yes"
			connectedColor = colorGreen
		}
		fmt.Printf("🔗 %sConnected:%s     %s%s%s\n", colorBlue, colorReset, connectedColor, connected, colorReset)

		// Linked timestamp
		linkedAtDisplay := "N/A"
		if statusResp.ConnectedAt != nil {
			linkedAtDisplay = statusResp.ConnectedAt.Format("2006-01-02 15:04:05")
		} else if config.LinkedAt != "" {
			// Try to parse from config
			if t, err := time.Parse(time.RFC3339, config.LinkedAt); err == nil {
				linkedAtDisplay = t.Format("2006-01-02 15:04:05")
			} else {
				linkedAtDisplay = config.LinkedAt
			}
		}
		fmt.Printf("🕓 %sLinked At:%s     %s\n", colorBlue, colorReset, linkedAtDisplay)

		fmt.Println()

		// Additional info if not connected
		if statusResp.Status != "linked" {
			fmt.Printf("%s⚠️  Project is not linked on the server.%s\n", colorYellow, colorReset)
			fmt.Printf("   Run 'featureplus-pr connect %s' to establish connection.\n", config.ProjectID)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
