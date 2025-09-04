package tests

import (
	"fmt"
	"net/http"
	"os"

	"github.com/FeaturePlus/pkg/featureplus"
	"featureplus-pr/internal/config"
)

func TestLogin() {
	// Step 1: Initialize config
	fmt.Println("Initializing config...")
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Config not found, will create a new one: %v\n", err)
	}
	
	// Set API URL
	config.SetAPIURL("http://localhost:8080")
	
	// Save the config to ensure the file exists
	if err := config.SaveConfig(); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}
	
	// Step 2: Perform login
	fmt.Println("Performing login...")
	client := featureplus.NewClient("http://localhost:8080", &http.Client{})
	
	// Hard-coded credentials for testing
	email := "admin@example.com"
	password := "admin123"
	
	loginResp, err := client.Login(email, password)
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Login successful! Token: %s\n", loginResp.Token)
	fmt.Printf("User: %s (%s)\n", loginResp.AuthInfo.Username, loginResp.AuthInfo.Role)
	
	// Step 3: Save auth info to config
	fmt.Println("Saving auth info to config...")
	if err := config.SaveAuthInfo(
		loginResp.Token,
		loginResp.AuthInfo.ID,
		loginResp.AuthInfo.Username,
		loginResp.AuthInfo.Email,
		loginResp.AuthInfo.Role,
	); err != nil {
		fmt.Printf("Error saving auth info: %v\n", err)
		os.Exit(1)
	}
	
	// Step 4: Test API call with token
	fmt.Println("Testing API call with token...")
	client.SetAuthToken(loginResp.Token)
	
	prs, err := client.ListPRs(0)
	if err != nil {
		fmt.Printf("API call failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("API call successful! Found %d PRs\n", len(prs))
}
