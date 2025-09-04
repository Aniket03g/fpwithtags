package tests

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
	"featureplus-pr/internal/config"
)

func TestAuth() {
	// Load configuration
	config.LoadConfig()
	
	// Get auth token
	token := config.GetAuthToken()
	fmt.Printf("Auth token from config: %v\n", token != "")
	if token != "" {
		fmt.Printf("Token starts with: %s...\n", token[:10])
	}
	
	// Create client
	client := featureplus.NewClient(config.GetAPIURL(), &http.Client{})
	client.SetAuthToken(token)
	
	// Test API call
	fmt.Println("Making API call to list PRs...")
	prs, err := client.ListPRs(0)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Success! Found %d PRs\n", len(prs))
	for i, pr := range prs {
		if i < 3 { // Show only first 3 PRs
			fmt.Printf("  - PR #%d: %s\n", pr.ID, pr.Title)
		}
	}
	if len(prs) > 3 {
		fmt.Printf("  - ... and %d more\n", len(prs)-3)
	}
}
