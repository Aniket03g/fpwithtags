package main

import (
	"featureplus-pr/cmd"
	"featureplus-pr/internal/config"
)

func main() {
	// Load configuration before executing commands
	config.LoadConfig()
	cmd.Execute()
}
