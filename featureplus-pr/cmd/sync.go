package cmd

import (
	"fmt"
	"os"

	"featureplus-pr/internal/config"
	"featureplus-pr/internal/sync"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync local code mappings to FeaturePlus backend",
	Long: `Syncs local feature manifests with the FeaturePlus backend.

This command:
- Reads all feature manifests from .featureplus/features/
- Identifies features with new data (files/commits not empty)
- Sends data to backend via POST /api/features/sync
- Updates last_synced timestamp on success

Example:
  featureplus sync`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if project is initialized
		if _, err := os.Stat(".featureplus"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ FeaturePlus not initialized. Run 'featureplus init' first.")
			os.Exit(1)
		}

		// Get API URL and auth token (auth is optional for sync)
		apiURL := config.GetAPIURL()
		authToken := config.GetAuthToken()

		fmt.Println("☁️  Syncing features to FeaturePlus backend...")

		// Sync all features
		results, err := sync.SyncAll(apiURL, authToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to sync: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("ℹ️  No features need syncing. All up to date!")
			return
		}

		// Display results
		successCount := 0
		failCount := 0

		fmt.Println()
		for _, result := range results {
			if result.Success {
				fmt.Printf("   ✅ Synced %s (%d files, %d commits)\n", 
					result.FeatureID, result.FileCount, result.CommitCount)
				successCount++
			} else {
				fmt.Printf("   ❌ Failed to sync %s: %v\n", result.FeatureID, result.Error)
				failCount++
			}
		}

		// Summary
		fmt.Printf("\n☁️  Synced %d feature(s) with FeaturePlus backend.\n", successCount)
		if failCount > 0 {
			fmt.Printf("   ⚠️  %d feature(s) failed to sync.\n", failCount)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
