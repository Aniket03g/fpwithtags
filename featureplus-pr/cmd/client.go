package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
	"featureplus-pr/internal/config"
)

// CreateAuthenticatedClient creates a new client with authentication token set
func CreateAuthenticatedClient() *featureplus.Client {
	// Create client with API URL from config
	apiURL := config.GetAPIURL()
	client := featureplus.NewClient(apiURL, &http.Client{})
	
	// Get auth token from config
	token := config.GetAuthToken()
	
	// Set auth token on client
	client.SetAuthToken(token)
	
	// Debug output
	if os.Getenv("DEBUG") == "1" {
		fmt.Printf("DEBUG CLIENT: Using API URL: %s\n", apiURL)
		fmt.Printf("DEBUG CLIENT: Auth token present: %v\n", token != "")
		if token != "" && len(token) > 10 {
			fmt.Printf("DEBUG CLIENT: Token starts with: %s...\n", token[:10])
		}
	}
	
	return client
}
