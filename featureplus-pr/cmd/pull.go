package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"featureplus-pr/internal/feature"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull <feature-id>",
	Short: "Pull a feature from FeaturePlus backend",
	Long: `Fetches a feature's metadata from the FeaturePlus backend and saves it locally as a YAML manifest.

Example:
  featureplus pull FTR-001
  featureplus pull 1

The feature will be saved to .featureplus/features/FTR-001.yaml`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse feature ID from argument
		featureIDStr := args[0]
		
		// Strip "FTR-" prefix if present
		featureIDStr = strings.TrimPrefix(featureIDStr, "FTR-")
		featureIDStr = strings.TrimPrefix(featureIDStr, "ftr-")
		
		// Convert to uint
		featureIDInt, err := strconv.Atoi(featureIDStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid feature ID: %s\n", args[0])
			os.Exit(1)
		}
		featureID := uint(featureIDInt)

		// Check if project is initialized
		if _, err := os.Stat(".featureplus"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ FeaturePlus not initialized. Run 'featureplus init' first.")
			os.Exit(1)
		}

		// Check if config.yaml exists
		if _, err := os.Stat(".featureplus/config.yaml"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ Configuration file not found. Run 'featureplus init' first.")
			os.Exit(1)
		}

		// Check if feature already exists locally
		if feature.FeatureExists(featureID) {
			// Ask user if they want to overwrite
			fmt.Printf("⚠️  Feature FTR-%03d already exists locally. Overwrite? (y/N): ", featureID)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Error reading input: %v\n", err)
				os.Exit(1)
			}
			
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("⏭️  Skipped.")
				return
			}
		}

		// Create authenticated client
		client := CreateAuthenticatedClient()

		// Fetch feature from backend
		fmt.Printf("⬇️  Fetching feature FTR-%03d from backend...\n", featureID)
		fetchedFeature, err := client.GetFeature(featureID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to fetch feature: %v\n", err)
			os.Exit(1)
		}

		// Get owner information (you might want to add this to the API response)
		// For now, we'll use a placeholder or get it from auth info
		owner := GetAuthToken() // You might want to get username instead
		
		// Save feature manifest locally
		filePath, err := feature.SaveFeatureManifest(fetchedFeature, owner)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to save feature manifest: %v\n", err)
			os.Exit(1)
		}

		// Print success message
		fmt.Printf("✅ Pulled feature FTR-%03d (%s)\n", featureID, fetchedFeature.Title)
		fmt.Printf("   Saved to %s\n", filePath)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
