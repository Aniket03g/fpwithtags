package tests

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
)

func DebugCLI() {
	fmt.Println("=== FeaturePlus CLI Debug Tool ===")
	
	// Check if config file exists
	configFile := "config.json"
	fmt.Printf("Checking for config file: %s\n", configFile)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Config file does not exist: %v\n", err)
		os.Exit(1)
	}
	
	// Read config file
	fmt.Println("Reading config file...")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}
	
	// Print raw config content
	fmt.Println("\nRaw config content:")
	fmt.Println(string(data))
	
	// Parse config
	var cfg struct {
		APIURL string `json:"api_url"`
		Auth   struct {
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expires_at"`
			UserID    uint   `json:"user_id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			Role      string `json:"role"`
		} `json:"auth"`
	}
	
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		os.Exit(1)
	}
	
	// Print parsed config
	fmt.Println("\nParsed config:")
	fmt.Printf("API URL: %s\n", cfg.APIURL)
	fmt.Printf("Auth token present: %v\n", cfg.Auth.Token != "")
	if cfg.Auth.Token != "" {
		fmt.Printf("Token: %s\n", cfg.Auth.Token)
	}
	
	fmt.Println("\nDebug complete!")
}
