package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

var (
	mergePRID    int
	mergeMethod  string
	mergeDelete  bool
	mergeComment string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge a pull request using GitHub CLI",
	Long: `Merge a pull request using GitHub CLI with options for merge method and branch cleanup.

Example:
  # Merge with default options (merge commit)
  featureplus-pr merge --id=123

  # Merge with squash and delete branch
  featureplus-pr merge --id=123 --method=squash --delete-branch

  # Merge with a comment
  featureplus-pr merge --id=123 --comment="Merging feature"`,

	Run: func(cmd *cobra.Command, args []string) {
		if mergePRID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --id is required")
			os.Exit(1)
		}

		// Get PR info from FeaturePlus
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})
		pr, err := client.GetPR(mergePRID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting PR info: %v\n", err)
			os.Exit(1)
		}

		// Extract PR number from URL
		prNumber := extractPRNumber(pr.URL)
		if prNumber == "" {
			fmt.Fprintf(os.Stderr, "Error: Could not extract PR number from URL: %s\n", pr.URL)
			os.Exit(1)
		}

		// Use the shared package to merge the PR
		if err := client.MergePR(prNumber, mergeMethod, mergeDelete, mergeComment); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to merge PR: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully merged PR #%s\n", prNumber)
	},
}

// extractPRNumber extracts the PR number from a GitHub PR URL
func extractPRNumber(url string) string {
	// URL format: https://github.com/org/repo/pull/123
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().IntVarP(&mergePRID, "id", "i", 0, "PR ID to merge (required)")
	mergeCmd.Flags().StringVarP(&mergeMethod, "method", "m", "merge", "Merge method: merge, squash, or rebase")
	mergeCmd.Flags().BoolVarP(&mergeDelete, "delete-branch", "d", false, "Delete the branch after merge")
	mergeCmd.Flags().StringVarP(&mergeComment, "comment", "c", "", "Optional merge comment")
	mergeCmd.MarkFlagRequired("id")
}
