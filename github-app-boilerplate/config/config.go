package config

import (
	"os"
)

type Config struct {
	GitHubAppID          string
	GitHubPrivateKeyPEM  string
	GitHubWebhookSecret  string
	GitHubAppSlug        string
	ServerPort           string
}

func LoadConfig() *Config {
	return &Config{
		GitHubAppID:         getEnv("GITHUB_APP_ID", ""),
		GitHubPrivateKeyPEM: getEnv("GITHUB_PRIVATE_KEY_PEM", ""),
		GitHubWebhookSecret: getEnv("GITHUB_WEBHOOK_SECRET", ""),
		GitHubAppSlug:       getEnv("GITHUB_APP_SLUG", ""),
		ServerPort:          getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
