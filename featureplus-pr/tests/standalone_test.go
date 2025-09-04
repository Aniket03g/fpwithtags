package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Simple config structure
type Config struct {
	APIURL string `json:"api_url"`
	Auth   struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
		UserID    uint      `json:"user_id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		Role      string    `json:"role"`
	} `json:"auth"`
}

func StandaloneTest() {
	// Step 1: Create a test config file
	fmt.Println("Creating test config file...")
	
	config := Config{
		APIURL: "http://localhost:8080",
	}
	
	// Set auth info
	config.Auth.Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTcwMDY4NTEsInVzZXJfaWQiOjN9.J7pfi_mkynpcCSnVCdRuoWRuyC-7vJnMRCaRLmfMoyo"
	config.Auth.ExpiresAt = time.Now().Add(24 * time.Hour)
	config.Auth.UserID = 1
	config.Auth.Username = "testuser"
	config.Auth.Email = "test@example.com"
	config.Auth.Role = "admin"
	
	// Save config to file
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		os.Exit(1)
	}
	
	if err := ioutil.WriteFile("test_config.json", configData, 0644); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Config file created with token: %s...\n", config.Auth.Token[:20])
	
	// Step 2: Test reading the config file
	fmt.Println("\nReading config file...")
	
	data, err := ioutil.ReadFile("test_config.json")
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}
	
	var readConfig Config
	if err := json.Unmarshal(data, &readConfig); err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Read config with API URL: %s\n", readConfig.APIURL)
	fmt.Printf("Auth token present: %v\n", readConfig.Auth.Token != "")
	if readConfig.Auth.Token != "" {
		fmt.Printf("Token: %s...\n", readConfig.Auth.Token[:20])
	}
	
	// Step 3: Test making an API request with the token
	fmt.Println("\nMaking API request with token...")
	
	req, err := http.NewRequest("GET", readConfig.APIURL+"/api/pr", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	
	// Add auth header
	req.Header.Set("Authorization", "Bearer "+readConfig.Auth.Token)
	fmt.Printf("Request headers: %v\n", req.Header)
	
	// Make request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	
	fmt.Printf("Response status: %s\n", resp.Status)
	
	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Response body: %s\n", string(body))
}
