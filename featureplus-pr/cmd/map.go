package cmd

import (
	"fmt"
	"os"
	"sort"

	"featureplus-pr/internal/feature"
	"featureplus-pr/internal/manifest"
	"github.com/spf13/cobra"
)

var (
	mapCommitLimit int
)

// mapCmd represents the map command
var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map git commits to features",
	Long: `Scans git history and maps commits to local feature manifests.

This command:
- Reads all local feature manifests from .featureplus/features/
- Scans git log for commits containing feature IDs (e.g., FTR-001)
- Aggregates files and commits per feature
- Displays a summary of the mapping

Example:
  featureplus map
  featureplus map --commits 100`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if project is initialized
		if _, err := os.Stat(".featureplus"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ FeaturePlus not initialized. Run 'featureplus init' first.")
			os.Exit(1)
		}

		// Check if we're in a git repository
		if _, err := os.Stat(".git"); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "❌ Not a git repository. Initialize git first.")
			os.Exit(1)
		}

		fmt.Printf("🔍 Scanning git history (last %d commits)...\n", mapCommitLimit)

		// Map git history to features
		results, err := feature.MapGitHistoryToFeatures(mapCommitLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Failed to map git history: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("ℹ️  No features found in git history.")
			fmt.Println("   Make sure your commit messages include feature IDs (e.g., FTR-001)")
			return
		}

		// Sort results by feature ID
		sort.Slice(results, func(i, j int) bool {
			return results[i].FeatureID < results[j].FeatureID
		})

		// Display results
		fmt.Printf("\n📊 Found %d feature(s) in git history:\n\n", len(results))

		for _, result := range results {
			displayFeatureMapping(result)
		}

		// Update manifests
		fmt.Println("\n📝 Updating feature manifests...")
		updatedCount := 0
		for _, result := range results {
			// Only update if feature has a local manifest
			if result.FeatureName != "" {
				err := manifest.UpdateManifest(result.FeatureID, result.Files, result.CommitHashes)
				if err != nil {
					fmt.Printf("   ⚠️  Failed to update %s: %v\n", result.FeatureID, err)
					continue
				}
				
				// Get updated stats
				fileCount, commitCount, err := manifest.GetManifestStats(result.FeatureID)
				if err != nil {
					fileCount = len(result.Files)
					commitCount = len(result.CommitHashes)
				}
				
				fmt.Printf("   ✅ Updated .featureplus/features/%s.yaml (%d files, %d commits)\n", 
					result.FeatureID, fileCount, commitCount)
				updatedCount++
			}
		}

		fmt.Printf("\n✅ Mapping complete! Found %d feature(s) with %d total commits.\n", 
			len(results), getTotalCommits(results))
		if updatedCount > 0 {
			fmt.Printf("   Updated %d manifest(s).\n", updatedCount)
		}
	},
}

func init() {
	rootCmd.AddCommand(mapCmd)
	mapCmd.Flags().IntVar(&mapCommitLimit, "commits", 50, "Number of commits to scan")
}

// displayFeatureMapping displays a single feature mapping result
func displayFeatureMapping(result feature.MapResult) {
	// Feature header
	if result.FeatureName != "" {
		fmt.Printf("📦 %s → %d file(s), %d commit(s)\n", 
			result.FeatureID, len(result.Files), result.CommitCount)
		fmt.Printf("   Name: %s\n", result.FeatureName)
	} else {
		fmt.Printf("📦 %s → %d file(s), %d commit(s)\n", 
			result.FeatureID, len(result.Files), result.CommitCount)
		fmt.Printf("   ⚠️  No local manifest found\n")
	}

	// Display files
	if len(result.Files) > 0 {
		// Sort files for consistent display
		sortedFiles := make([]string, len(result.Files))
		copy(sortedFiles, result.Files)
		sort.Strings(sortedFiles)

		// Limit display to first 10 files
		displayLimit := 10
		if len(sortedFiles) > displayLimit {
			for i := 0; i < displayLimit; i++ {
				fmt.Printf("   - %s\n", sortedFiles[i])
			}
			fmt.Printf("   ... and %d more file(s)\n", len(sortedFiles)-displayLimit)
		} else {
			for _, file := range sortedFiles {
				fmt.Printf("   - %s\n", file)
			}
		}
	}

	// Display commit hashes (first 3)
	if len(result.CommitHashes) > 0 {
		fmt.Printf("   Commits: ")
		displayLimit := 3
		if len(result.CommitHashes) > displayLimit {
			for i := 0; i < displayLimit; i++ {
				fmt.Printf("%s ", result.CommitHashes[i][:7])
			}
			fmt.Printf("(+%d more)", len(result.CommitHashes)-displayLimit)
		} else {
			for _, hash := range result.CommitHashes {
				fmt.Printf("%s ", hash[:7])
			}
		}
		fmt.Println()
	}

	fmt.Println()
}

// getTotalCommits calculates total commits across all features
func getTotalCommits(results []feature.MapResult) int {
	total := 0
	for _, result := range results {
		total += result.CommitCount
	}
	return total
}
