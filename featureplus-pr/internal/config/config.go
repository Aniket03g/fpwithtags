package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	APIURL string `json:"api_url"`
}

var appConfig Config

// LoadConfig loads configuration from multiple possible locations
func LoadConfig() {
	// Default fallback
	appConfig = Config{
		APIURL: "http://localhost:8080",
	}

	// Try to load config from multiple locations in order of preference
	configLocations := getConfigLocations()
	
	for _, location := range configLocations {
		if loadConfigFromFile(location) {
			fmt.Printf("Loaded config from: %s\n", location)
			return
		}
	}
	
	fmt.Println("No config file found, using default API URL: http://localhost:8080")
}

// getConfigLocations returns a list of possible config file locations in order of preference
func getConfigLocations() []string {
	locations := []string{}
	
	// 1. Current working directory
	if cwd, err := os.Getwd(); err == nil {
		locations = append(locations, filepath.Join(cwd, "config.json"))
	}
	
	// 2. Executable directory
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		locations = append(locations, filepath.Join(exeDir, "config.json"))
	}
	
	// 3. User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(homeDir, ".featureplus-pr", "config.json"))
		locations = append(locations, filepath.Join(homeDir, "config.json"))
	}
	
	return locations
}

// loadConfigFromFile attempts to load config from a specific file location
func loadConfigFromFile(filePath string) bool {
	configData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false
	}

	// Parse the config file
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid config format in %s, skipping\n", filePath)
		return false
	}

	// Use the loaded config if API URL is provided
	if config.APIURL != "" {
		appConfig.APIURL = config.APIURL
		return true
	}
	
	return false
}

// GetAPIURL returns the configured API URL
func GetAPIURL() string {
	return appConfig.APIURL
}

// GetConfigInfo returns information about the current configuration
func GetConfigInfo() string {
	return fmt.Sprintf("API URL: %s", appConfig.APIURL)
}
