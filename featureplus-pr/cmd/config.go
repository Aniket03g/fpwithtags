package cmd

import "featureplus-pr/internal/config"

// GetAPIURL returns the configured API URL
func GetAPIURL() string {
	return config.GetAPIURL()
}

// GetAuthToken returns the authentication token
func GetAuthToken() string {
	return config.GetAuthToken()
}
