package cmd

import (
	"fmt"
	"os"

	"featureplus-pr/internal/config"
	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug the CLI configuration and authentication",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== FeaturePlus CLI Debug Information ===")
		
		// Check if config file exists
		configFile := "config.json"
		fmt.Printf("Checking for config file: %s\n", configFile)
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			fmt.Printf("Config file does not exist: %v\n", err)
		} else {
			fmt.Println("Config file exists")
		}
		
		// Print config info
		fmt.Println("\n=== Config Information ===")
		fmt.Printf("API URL: %s\n", config.GetAPIURL())
		
		// Auth info
		fmt.Println("\n=== Authentication Information ===")
		token := config.GetAuthToken()
		fmt.Printf("Auth token present: %v\n", token != "")
		if token != "" && len(token) > 10 {
			fmt.Printf("Token starts with: %s...\n", token[:10])
		}
		
		authInfo := config.GetAuthInfo()
		fmt.Printf("User ID: %d\n", authInfo.UserID)
		fmt.Printf("Username: %s\n", authInfo.Username)
		fmt.Printf("Email: %s\n", authInfo.Email)
		fmt.Printf("Role: %s\n", authInfo.Role)
		fmt.Printf("Token expiration: %d\n", authInfo.ExpiresAt)
		
		// Check if authenticated
		fmt.Printf("Is authenticated: %v\n", config.IsAuthenticated())
		
		// Create client and test
		fmt.Println("\n=== Testing Client ===")
		client := CreateAuthenticatedClient()
		fmt.Println("Client created successfully")
		
		// Test API call
		fmt.Println("\n=== Testing API Call ===")
		fmt.Println("Attempting to list PRs...")
		prs, err := client.ListPRs(0)
		if err != nil {
			fmt.Printf("API call failed: %v\n", err)
		} else {
			fmt.Printf("API call successful! Found %d PRs\n", len(prs))
		}
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)
}
