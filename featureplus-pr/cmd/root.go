package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "featureplus-pr",
	Short:   "A CLI to link GitHub PRs with FeaturePlus features and tasks",
	Long:    `featureplus-pr uploads PR info from the current branch to FeaturePlus, streamlining your workflow.`,
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
