package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

var checkoutPRID int

var prCheckoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Checkout the branch for a given PR",
	Run: func(cmd *cobra.Command, args []string) {
		if checkoutPRID == 0 {
			fmt.Fprintln(os.Stderr, "--id is required")
			os.Exit(1)
		}

		// Create client
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})

		// Get PR info from FeaturePlus
		pr, err := client.GetPR(checkoutPRID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch PR: %v\n", err)
			os.Exit(1)
		}
		if pr.Branch == "" {
			fmt.Fprintf(os.Stderr, "PR #%d does not have a branch name.\n", checkoutPRID)
			os.Exit(1)
		}

		if err := runGitFetch(pr.Branch); err != nil {
			fmt.Fprintf(os.Stderr, "git fetch failed: %v\n", err)
			os.Exit(1)
		}
		if err := runGitCheckout(pr.Branch); err != nil {
			fmt.Fprintf(os.Stderr, "git checkout failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Checked out branch %s (PR #%d)", pr.Branch, pr.ID)
		if pr.FeatureID != 0 || pr.TaskID != 0 {
			fmt.Printf(" [feature_id: %d, task_id: %d]", pr.FeatureID, pr.TaskID)
		}
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(prCheckoutCmd)
	prCheckoutCmd.Flags().IntVar(&checkoutPRID, "id", 0, "PR ID to checkout (required)")
	prCheckoutCmd.MarkFlagRequired("id")
}

// This function is no longer needed as we use the shared package

func runGitFetch(branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGitCheckout(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
