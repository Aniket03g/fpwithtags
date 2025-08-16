package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

var (
	approvePRID   int
	approveComment string
)

var approveCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a pull request in FeaturePlus",
	Long: `Approve a pull request in FeaturePlus with an optional comment.

Example:
  featureplus-pr approve --id=123 --comment="LGTM!"`,
	Run: func(cmd *cobra.Command, args []string) {
		if approvePRID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --id is required")
			os.Exit(1)
		}

		// Get the current git branch to find the PR number if not provided
		prNumber, err := getCurrentPRNumber()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not determine PR number from git: %v\n", err)
			fmt.Fprintf(os.Stderr, "Continuing with FeaturePlus approval only...\n")
		} else if prNumber != "" {
			// Run GitHub CLI to approve the PR
			if err := runGitHubApprove(prNumber, approveComment); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: GitHub approval failed: %v\n", err)
				fmt.Fprintf(os.Stderr, "Continuing with FeaturePlus approval only...\n")
			} else {
				fmt.Printf("✅ Approved PR #%s on GitHub\n", prNumber)
			}
		}

		// Create client
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})

		// Create review request
		reqBody := &featureplus.ReviewRequest{
			Status:     "approved",
			Comment:    approveComment,
			ApprovedAt: time.Now().Unix(),
			Version:    getGitVersion(),
		}

		// Use the shared package to approve PR
		if err := client.ApprovePR(approvePRID, reqBody); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Successfully approved PR #%d\n", approvePRID)
		if approveComment != "" {
			fmt.Printf("   Comment: %s\n", approveComment)
		}
	},
}

// This function is no longer needed as we use the shared package

func getCurrentPRNumber() (string, error) {
	// Get the current branch name
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchOut, err := branchCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %v", err)
	}
	branchName := strings.TrimSpace(string(branchOut))

	// Get PR number from branch name using GitHub CLI
	prCmd := exec.Command("gh", "pr", "view", branchName, "--json", "number")
	prOut, err := prCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get PR info: %v", err)
	}

	// Parse the JSON output to get the PR number
	var result struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(prOut, &result); err != nil {
		return "", fmt.Errorf("failed to parse PR info: %v", err)
	}

	return fmt.Sprintf("%d", result.Number), nil
}

func runGitHubApprove(prNumber, comment string) error {
	args := []string{"pr", "review", prNumber, "--approve"}
	if comment != "" {
		args = append(args, "--body", fmt.Sprintf(`"%s"`, comment))
	}

	cmd := exec.Command("gh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

func getGitVersion() string {
	// Get the current git commit hash
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

func init() {
	rootCmd.AddCommand(approveCmd)
	approveCmd.Flags().IntVar(&approvePRID, "id", 0, "PR ID to approve (required)")
	approveCmd.Flags().StringVar(&approveComment, "comment", "", "Optional approval comment")
	approveCmd.MarkFlagRequired("id")
}
