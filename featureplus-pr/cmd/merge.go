package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

		// Get the current git branch to find the PR number if not provided
		prNumber, err := getCurrentPRNumber()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Run GitHub CLI to merge the PR
		if err := runGitHubMerge(prNumber, mergeMethod, mergeDelete, mergeComment); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to merge PR: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully merged PR #%s\n", prNumber)
	},
}

func runGitHubMerge(prNumber, method string, deleteBranch bool, comment string) error {
	args := []string{"pr", "merge", prNumber, "--merge"}

	// Add merge method if specified
	switch strings.ToLower(method) {
	case "merge", "":
		// Default is merge commit, no flag needed
	case "squash":
		args = append(args, "--squash")
	case "rebase":
		args = append(args, "--rebase")
	default:
		return fmt.Errorf("invalid merge method: %s. Must be one of: merge, squash, rebase", method)
	}

	// Add delete branch flag if specified
	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	// Add comment if provided
	if comment != "" {
		args = append(args, "--body", fmt.Sprintf(`"%s"`, comment))
	}

	// Add auto-confirm flag
	args = append(args, "--auto")

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().IntVarP(&mergePRID, "id", "i", 0, "PR ID to merge (required)")
	mergeCmd.Flags().StringVarP(&mergeMethod, "method", "m", "merge", "Merge method: merge, squash, or rebase")
	mergeCmd.Flags().BoolVarP(&mergeDelete, "delete-branch", "d", false, "Delete the branch after merge")
	mergeCmd.Flags().StringVarP(&mergeComment, "comment", "c", "", "Optional merge comment")
	mergeCmd.MarkFlagRequired("id")
}
