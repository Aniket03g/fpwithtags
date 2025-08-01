package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"featureplus-pr/internal/config"
)

var rootCmd = &cobra.Command{
	Use:     "featureplus-pr",
	Short:   "A CLI to link GitHub PRs with FeaturePlus features and tasks",
	Long:    `featureplus-pr uploads PR info from the current branch to FeaturePlus, streamlining your workflow.`,
	Version: "0.1.0",
	Run: func(cmd *cobra.Command, args []string) {
		// Show current configuration when no subcommand is provided
		fmt.Println("FeaturePlus PR CLI")
		fmt.Println("==================")
		fmt.Println(config.GetConfigInfo())
		fmt.Println("\nAvailable commands:")
		fmt.Println("  upload     - Upload current branch's PR info to FeaturePlus")
		fmt.Println("  list       - List features and tasks from FeaturePlus")
		fmt.Println("  checkout   - Checkout a feature or task")
		fmt.Println("  config     - Show configuration information")
		fmt.Println("\nUse --help with any command for more information.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
