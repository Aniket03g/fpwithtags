package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/spf13/cobra"
)

// No need to define types here as they're defined in the shared package

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

		// Create client
		client := featureplus.NewClient(GetAPIURL(), &http.Client{})

		// Convert PR IDs to uint
		prIDs := make([]uint, len(prs))
		for i, id := range prs {
			prIDs[i] = uint(id)
		}

		// Create release request
		req := &featureplus.CreateReleaseRequest{
			Tag:   releaseTag,
			PRIDs: prIDs,
			Notes: notes,
		}

		// Use the shared package to create release
		_, err := client.CreateRelease(req)
		if err != nil {
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

// These functions are no longer needed as they're provided by the shared package

// This function is no longer needed as we use GetAPIURL from cmd package

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
	// Create client
	client := featureplus.NewClient(GetAPIURL(), &http.Client{})

	// Use the shared package to list releases
	releases, err := client.ListReleases()
	if err != nil {
		return fmt.Errorf("failed to list releases: %v", err)
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
		prs := make([]string, len(r.PRIDs))
		for i, prID := range r.PRIDs {
			prs[i] = strconv.Itoa(int(prID))
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
