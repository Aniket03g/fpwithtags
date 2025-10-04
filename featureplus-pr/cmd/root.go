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
	// Always print version info
	fmt.Printf("FeaturePlus CLI v%s\n", rootCmd.Version)

	// Load configuration before executing any command
	fmt.Println("Loading configuration...")
	err := config.LoadConfig()
	if err != nil {
		// Just log the error but continue, as we can still use default values
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	} else {
		fmt.Println("Configuration loaded successfully")
	}

	// Always show basic config info
	fmt.Printf("API URL: %s\n", config.GetAPIURL())
	fmt.Printf("Auth token present: %v\n", config.GetAuthToken() != "")

	// More detailed debug output
	if os.Getenv("DEBUG") == "1" {
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
