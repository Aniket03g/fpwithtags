package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	APIURL string `json:"api_url"`
	Auth   AuthConfig `json:"auth"`
}

type AuthConfig struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
}

var appConfig Config

func LoadConfig() error {
	// Default fallback
	appConfig = Config{
		APIURL: "http://localhost:8080",
		Auth:   AuthConfig{},
	}

	// Try to read config.json from current working directory
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		// Config file not found or unreadable, use default
		return fmt.Errorf("config file not found: %w", err)
	}

	// Parse the config file
	if err := json.Unmarshal(configData, &appConfig); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid config.json format, using default API URL\n")
		return fmt.Errorf("invalid config format: %w", err)
	}

	// Debug output
	fmt.Printf("DEBUG CONFIG: Loaded config with API URL: %s\n", appConfig.APIURL)
	fmt.Printf("DEBUG CONFIG: Auth token present: %v\n", appConfig.Auth.Token != "")
	if appConfig.Auth.Token != "" {
		fmt.Printf("DEBUG CONFIG: Token starts with: %s...\n", appConfig.Auth.Token[:10])
	}

	return nil
}

func GetAPIURL() string {
	return appConfig.APIURL
}

func SetAPIURL(url string) {
	appConfig.APIURL = url
}

func GetAuthToken() string {
	return appConfig.Auth.Token
}

func GetAuthInfo() AuthConfig {
	return appConfig.Auth
}

func IsAuthenticated() bool {
	return appConfig.Auth.Token != "" && time.Now().Unix() < appConfig.Auth.ExpiresAt
}

func SaveAuthInfo(token string, userID uint, username, email, role string) error {
	// Update the auth config
	appConfig.Auth.Token = token
	appConfig.Auth.ExpiresAt = time.Now().Add(72 * time.Hour).Unix() // 72 hours from now (matching backend)
	appConfig.Auth.UserID = userID
	appConfig.Auth.Username = username
	appConfig.Auth.Email = email
	appConfig.Auth.Role = role

	// Save to config file
	return SaveConfig()
}

func ClearAuthInfo() error {
	appConfig.Auth = AuthConfig{}
	return SaveConfig()
}

func SaveConfig() error {
	configData, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile("config.json", configData, 0644)
}
