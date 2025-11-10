package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"featureplus-pr/internal/config"
)

// MARKER:ROOT_CMD Root command definition
var rootCmd = &cobra.Command{
	Use:     "featureplus-pr",
	Short:   "A CLI to link GitHub PRs with FeaturePlus features and tasks",
	Long:    `featureplus-pr uploads PR info from the current branch to FeaturePlus, streamlining your workflow.`,
	Version: "0.1.0",
}

// MARKER:EXECUTE_FUNC CLI execution entry point
func Execute() {
	// Load configuration silently (for auth-based commands)
	config.LoadConfig()

	// Only show detailed info in DEBUG mode
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("FeaturePlus CLI v%s\n", rootCmd.Version)
		fmt.Printf("API URL: %s\n", config.GetAPIURL())
		fmt.Printf("Auth token present: %v\n", config.GetAuthToken() != "")
		token := config.GetAuthToken()
		if token != "" && len(token) > 10 {
			fmt.Printf("Token starts with: %s...\n", token[:10])
		}
		fmt.Printf("Auth info: %+v\n", config.GetAuthInfo())
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
