package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"featureplus-pr/internal/config"
	"github.com/spf13/cobra"
)

// Release represents a release from the API
type Release struct {
	ID      uint        `json:"id"`
	Tag     string      `json:"tag"`
	PRs     []PRDetails `json:"prs"`
	Status  string      `json:"status"`
	Notes   string      `json:"notes"`
	Created string      `json:"created_at"`
}

// PRDetails represents a PR in a release
type PRDetails struct {
	ID          int    `json:"id"`
	TaskID      int    `json:"task_id"`
	FeatureID   int    `json:"feature_id"`
	PRURL       string `json:"pr_url"`
	Title       string `json:"title"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IsTested    bool   `json:"is_tested"`
	Version     string `json:"version"`
}

var (
	releaseTag   string
	prsString   string
	releaseNotes string
)

var releaseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new release with the specified PRs",
	Long: `Create a new release with the specified pull requests.

Example:
  # Create a release with all required flags
  featureplus-pr release create --tag v1.2.0 --prs "12,15,18" --notes "Release notes here"

  # Interactive mode (will prompt for missing required fields)
  featureplus-pr release create`,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse PRs from comma-separated string
		prs := []int{}
		if prsString != "" {
			for _, idStr := range strings.Split(prsString, ",") {
				id, err := strconv.Atoi(strings.TrimSpace(idStr))
				if err != nil {
					return fmt.Errorf("invalid PR ID: %v", idStr)
				}
				prs = append(prs, id)
			}
		}

		// Make sure required fields are provided
		if releaseTag == "" {
			return fmt.Errorf("release tag is required (use --tag)")
		}
		if len(prs) == 0 {
			return fmt.Errorf("at least one PR ID is required (use --prs)")
		}

		// Get release notes
		notes := getReleaseNotes()

		// Create and send the release request
		if err := createRelease(releaseTag, prs, notes); err != nil {
			return fmt.Errorf("error creating release: %v", err)
		}

		// Print success message
		prList := make([]string, len(prs))
		for i, pr := range prs {
			prList[i] = strconv.Itoa(pr)
		}
		fmt.Printf("✅ Release %s created with PRs: %s\n", releaseTag, strings.Join(prList, ", "))
		return nil
	},
}

func getReleaseTag() string {
	if releaseTag == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter release tag (e.g., v1.2.0): ")
		scanner.Scan()
		releaseTag = strings.TrimSpace(scanner.Text())
		if releaseTag == "" {
			fmt.Fprintln(os.Stderr, "Error: Release tag is required")
			os.Exit(1)
		}
	}
	return releaseTag
}

func getReleasePRs() []int {
	if prsString == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter PR numbers (comma-separated, e.g., 12,15,18): ")
		scanner.Scan()
		prsString = strings.TrimSpace(scanner.Text())
		if prsString == "" {
			fmt.Fprintln(os.Stderr, "Error: At least one PR is required")
			os.Exit(1)
		}
	}

	// Parse PR numbers
	prs := []int{}
	for _, prStr := range strings.Split(prsString, ",") {
		pr, err := strconv.Atoi(strings.TrimSpace(prStr))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid PR number: %s\n", prStr)
			os.Exit(1)
		}
		prs = append(prs, pr)
	}

	return prs
}

func getReleaseNotes() string {
	if releaseNotes == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Enter release notes (press Enter twice to finish):\n> ")
		
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				// Check if previous line was also empty (double Enter)
				if len(lines) > 0 && lines[len(lines)-1] == "" {
					break
				}
			}
			lines = append(lines, line)
			fmt.Print("> ") // Continue prompt for next line
		}
		releaseNotes = strings.Join(lines, "\n")
	}
	return releaseNotes
}

type releaseRequest struct {
	Tag   string `json:"tag"`
	PRs   []int  `json:"prs"`
	Notes string `json:"notes"`
}

func createRelease(tag string, prs []int, notes string) error {
	// Prepare request payload
	reqBody := releaseRequest{
		Tag:   tag,
		PRs:   prs,
		Notes: notes,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Get API base URL from config
	baseURL, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create HTTP request
	// Ensure we don't have a trailing slash in the base URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/releases", baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// loadConfig loads the configuration and returns the API base URL
func loadConfig() (string, error) {
	// Use the config package to load configuration
	config.LoadConfig()
	baseURL := config.GetAPIURL()
	
	// Ensure the API path is included
	return baseURL + "/api", nil
}

// ReleaseCmd represents the release command
var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Manage releases",
}

func init() {
	// Add the release command to the root command
	rootCmd.AddCommand(ReleaseCmd)
	
	// Add flags to the existing releaseCreateCmd
	releaseCreateCmd.Flags().StringVarP(&releaseTag, "tag", "t", "", "Release tag (e.g., v1.2.0)")
	releaseCreateCmd.Flags().StringVarP(&prsString, "prs", "p", "", "Comma-separated list of PR IDs to include in the release")
	releaseCreateCmd.Flags().StringVarP(&releaseNotes, "notes", "n", "", "Release notes")

	// Add subcommands to the release command
	ReleaseCmd.AddCommand(releaseCreateCmd)
	ReleaseCmd.AddCommand(releaseListCmd)
}

// listReleases fetches and displays all releases
func listReleases() error {
	// Get API base URL from config
	baseURL, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create HTTP request
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/releases", baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check for error response
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Display releases in a table
	if len(releases) == 0 {
		fmt.Println("No releases found.")
		return nil
	}

	// Print table header
	fmt.Printf("%-10s %-12s %-10s %-50s\n", "TAG", "STATUS", "PRs", "NOTES")
	fmt.Println(strings.Repeat("-", 84))

	// Print each release
	for _, r := range releases {
		// Truncate notes if too long
		notes := r.Notes
		if len(notes) > 47 {
			notes = notes[:44] + "..."
		}

		// Format PRs as comma-separated string
		prs := make([]string, len(r.PRs))
		for i, pr := range r.PRs {
			prs[i] = strconv.Itoa(pr.ID)
		}
		prsStr := strings.Join(prs, ", ")
		if len(prsStr) > 10 {
			prsStr = prsStr[:7] + "..."
		}

		fmt.Printf("%-10s %-12s %-10s %-50s\n", 
			r.Tag, 
			r.Status, 
			prsStr,
			notes,
		)
	}

	return nil
}

var releaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := listReleases(); err != nil {
			return fmt.Errorf("error listing releases: %v", err)
		}
		return nil
	},
}
