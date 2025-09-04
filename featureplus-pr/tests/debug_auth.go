package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
	"featureplus-pr/internal/config"
)

func DebugAuth() {
	// Step 1: Load config and check if token exists
	fmt.Println("Step 1: Loading config...")
	err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
	}

	// Step 2: Check if config file exists and read it directly
	fmt.Println("\nStep 2: Reading config file directly...")
	configFile := "../config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Config file %s does not exist\n", configFile)
	} else {
		data, err := ioutil.ReadFile(configFile)
		if err != nil {
			fmt.Printf("Error reading config file: %v\n", err)
		} else {
			fmt.Printf("Config file content: %s\n", string(data))
			
			// Parse config
			var cfg struct {
				APIURL string `json:"api_url"`
				Auth   struct {
					Token     string `json:"token"`
					ExpiresAt int64  `json:"expires_at"`
					UserID    int    `json:"user_id"`
					Username  string `json:"username"`
					Email     string `json:"email"`
					Role      string `json:"role"`
				} `json:"auth"`
			}
			
			if err := json.Unmarshal(data, &cfg); err != nil {
				fmt.Printf("Error parsing config: %v\n", err)
			} else {
				fmt.Printf("API URL from config: %s\n", cfg.APIURL)
				fmt.Printf("Auth token present: %v\n", cfg.Auth.Token != "")
				if cfg.Auth.Token != "" {
					fmt.Printf("Token: %s\n", cfg.Auth.Token)
				}
			}
		}
	}

	// Step 3: Get token from config package
	fmt.Println("\nStep 3: Getting token from config package...")
	token := config.GetAuthToken()
	fmt.Printf("Token from config package: %v\n", token != "")
	if token != "" {
		fmt.Printf("Token: %s\n", token)
	}

	// Step 4: Test API call with token
	fmt.Println("\nStep 4: Testing API call with token...")
	client := featureplus.NewClient(config.GetAPIURL(), &http.Client{})
	client.SetAuthToken(token)
	
	// Make a simple API call
	fmt.Println("Making API call to list PRs...")
	prs, err := client.ListPRs(0)
	if err != nil {
		fmt.Printf("API call error: %v\n", err)
	} else {
		fmt.Printf("Success! Found %d PRs\n", len(prs))
	}
}
