package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	APIURL string `json:"api_url"`
}

var appConfig Config

func LoadConfig() {
	// Default fallback
	appConfig = Config{
		APIURL: "http://localhost:8080",
	}

	// Try to read config.json from current working directory
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		// Config file not found or unreadable, use default
		return
	}

	// Parse the config file
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid config.json format, using default API URL\n")
		return
	}

	// Use the loaded config if API URL is provided
	if config.APIURL != "" {
		appConfig.APIURL = config.APIURL
	}
}

func GetAPIURL() string {
	return appConfig.APIURL
}
