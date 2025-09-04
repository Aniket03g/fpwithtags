package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

var (
	listFeatureID int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests from FeaturePlus",
	Run: func(cmd *cobra.Command, args []string) {
		// Create authenticated client
		client := CreateAuthenticatedClient()

		// Use the shared package to list PRs
		fmt.Println("Debug: Calling client.ListPRs()")
		prs, err := client.ListPRs(uint(listFeatureID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get PR list: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Debug: Got %d PRs\n", len(prs))

		if len(prs) == 0 {
			fmt.Println("No pull requests found.")
			return
		}

		displayPRTable(prs)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVar(&listFeatureID, "feature-id", 0, "Filter PRs by feature ID (optional)")
}

// This function is no longer needed as it's provided by the shared package

func displayPRTable(prs []featureplus.PullRequest) {
	// Print header
	fmt.Printf("%-6s %-40s %-10s %-10s %-10s\n", "ID", "Title", "Status", "TaskID", "FeatureID")
	fmt.Println(strings.Repeat("-", 80))

	// Print each PR
	for _, pr := range prs {
		// Truncate title if too long
		title := pr.Title
		if len(title) > 38 {
			title = title[:35] + "..."
		}

		fmt.Printf("%-6d %-40s %-10s %-10d %-10d\n",
			pr.ID,
			title,
			pr.Status,
			pr.TaskID,
			int(pr.FeatureID))
	}

	fmt.Printf("\nTotal: %d pull request(s)\n", len(prs))
}
