package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type PRInfo struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Body        string `json:"body"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
}

type UploadRequest struct {
	FeatureID   int    `json:"feature_id"`
	TaskID      int    `json:"task_id"`
	PRURL       string `json:"pr_url"`
	Branch      string `json:"branch"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsTested    bool   `json:"is_tested"`
	Version     string `json:"version"`
}

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

		prInfo, err := getPRInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get PR info: %v\n", err)
			os.Exit(1)
		}

		req := UploadRequest{
			FeatureID:   featureID,
			TaskID:      taskID,
			PRURL:       prInfo.URL,
			Branch:      prInfo.HeadRefName,
			Title:       prInfo.Title,
			Description: prInfo.Body,
			IsTested:    markTested,
			Version:     version,
		}

		if err := sendToFeaturePlus(req); err != nil {
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

func getPRInfo() (*PRInfo, error) {
	cmd := exec.Command("gh", "pr", "view", "--json", "title,url,body,baseRefName,headRefName")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var pr PRInfo
	if err := json.Unmarshal(out, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

func sendToFeaturePlus(req UploadRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	apiURL := GetAPIURL()

	if verbose {
		fmt.Printf("Sending request to: %s/api/pr\n", apiURL)
		fmt.Printf("Request payload: %s\n", string(jsonData))
	}

	resp, err := http.Post(apiURL+"/api/pr", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if verbose {
		responseBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("Server response status: %s\n", resp.Status)
		fmt.Printf("Server response body: %s\n", string(responseBody))
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody)) // Restore body for further reading if needed
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status: %s", resp.Status)
	}
	return nil
}
