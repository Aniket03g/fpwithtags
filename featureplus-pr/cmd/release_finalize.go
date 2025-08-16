package cmd

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

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

		// Convert release ID to uint
		releaseIDInt, err := strconv.ParseUint(releaseID, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid release ID: %v", err)
		}

		// Create client
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})

		// Use the shared package to finalize release
		if err := client.FinalizeRelease(uint(releaseIDInt)); err != nil {
			return fmt.Errorf("error finalizing release: %v", err)
		}

		// Print success message
		fmt.Printf("✅ Release %s finalized successfully\n", releaseID)

		return nil
	},
}

// This function is no longer needed as we use the shared package

func init() {
	// Add the finalize command to the release command
	ReleaseCmd.AddCommand(releaseFinalizeCmd)
}