package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/demo-markers/cli"
	"github.com/demo-markers/routes"
	"github.com/gin-gonic/gin"
)

// Configuration for the application
type Config struct {
	Port     string
	LogLevel string
	Mode     string
}

// Initialize the application with default configuration
func initApp() *Config {
	// MARKER:MAIN_INIT
	config := &Config{
		Port:     "8080",
		LogLevel: "info",
		Mode:     "development",
	}

	// Override with environment variables if present
	if port := os.Getenv("APP_PORT"); port != "" {
		config.Port = port
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	if mode := os.Getenv("APP_MODE"); mode != "" {
		config.Mode = mode
	}

	return config
}

func main() {
	// Initialize the application
	config := initApp()
	fmt.Printf("Starting application in %s mode on port %s\n", config.Mode, config.Port)

	// Handle CLI commands if arguments are provided
	if len(os.Args) > 1 {
		cmd := strings.Join(os.Args[1:], " ")
		cli.HandleCommand(cmd)
		return
	}

	// Otherwise, start the HTTP server
	r := gin.Default()
	routes.RegisterRoutes(r)
	r.Run(":" + config.Port)
}
