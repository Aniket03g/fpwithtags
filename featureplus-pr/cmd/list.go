package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"featureplus-pr/internal/config"
)

type PRListItem struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	FeatureID int    `json:"feature_id"`
	TaskID    int    `json:"task_id"`
	URL       string `json:"pr_url"`
	Branch    string `json:"branch"`
	Tested    bool   `json:"is_tested"`
	Version   string `json:"version"`
	CreatedAt int64  `json:"created_at"`
}

var (
	listFeatureID int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pull requests from FeaturePlus",
	Run: func(cmd *cobra.Command, args []string) {
		prs, err := getPRList()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get PR list: %v\n", err)
			os.Exit(1)
		}

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

func getPRList() ([]PRListItem, error) {
	apiURL := config.GetAPIURL()
	url := apiURL + "/api/pr"
	if listFeatureID > 0 {
		url += "?feature_id=" + strconv.Itoa(listFeatureID)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prs []PRListItem
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, err
	}

	return prs, nil
}

func displayPRTable(prs []PRListItem) {
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
			pr.FeatureID)
	}

	fmt.Printf("\nTotal: %d pull request(s)\n", len(prs))
}
