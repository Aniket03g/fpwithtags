package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

var (
	featureID  int
	taskID     int
	markTested bool
	version    string
	verbose    bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload the current branch's PR info to FeaturePlus",
	Run: func(cmd *cobra.Command, args []string) {
		if featureID == 0 || taskID == 0 {
			fmt.Fprintln(os.Stderr, "--feature-id and --task-id are required")
			os.Exit(1)
		}

		prInfo, err := featureplus.GetPRInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get PR info: %v\n", err)
			os.Exit(1)
		}

		// Create client
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})

		// Create upload request
		req := &featureplus.UploadRequest{
			FeatureID:   featureID,
			TaskID:      taskID,
			PRURL:       prInfo.URL,
			Branch:      prInfo.HeadRefName,
			Title:       prInfo.Title,
			Description: prInfo.Body,
			IsTested:    markTested,
			Version:     version,
		}

		// Use the shared package to upload PR
		if verbose {
			fmt.Printf("Uploading PR to %s/api/pr\n", GetAPIURL())
		}

		if err := client.UploadPR(req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to upload PR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("PR uploaded successfully!")
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().IntVar(&featureID, "feature-id", 0, "Feature ID to link this PR to (required)")
	uploadCmd.Flags().IntVar(&taskID, "task-id", 0, "Task ID to link this PR to (required)")
	uploadCmd.Flags().BoolVar(&markTested, "mark-tested", false, "Mark this PR as tested")
	uploadCmd.Flags().StringVar(&version, "version", "", "Version or release to associate with this PR (optional)")
	uploadCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging")
	uploadCmd.MarkFlagRequired("feature-id")
	uploadCmd.MarkFlagRequired("task-id")
}

// These functions are no longer needed as they're provided by the shared package
