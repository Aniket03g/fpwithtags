package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"featureplus-pr/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration information",
	Long:  `Display current configuration settings and possible config file locations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("FeaturePlus PR Configuration")
		fmt.Println("============================")
		fmt.Println()
		
		// Show current configuration
		fmt.Println("Current Configuration:")
		fmt.Printf("  API URL: %s\n", config.GetAPIURL())
		fmt.Println()
		
		// Show possible config file locations
		fmt.Println("Config File Locations (checked in order):")
		showConfigLocations()
		fmt.Println()
		
		// Show usage instructions
		fmt.Println("Configuration Instructions:")
		fmt.Println("  1. Create a config.json file in any of the locations above")
		fmt.Println("  2. Use the following format:")
		fmt.Println("     {")
		fmt.Println("       \"api_url\": \"https://your-featureplus-instance.com\"")
		fmt.Println("     }")
		fmt.Println("  3. The tool will use the first valid config file it finds")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// showConfigLocations displays all possible config file locations
func showConfigLocations() {
	locations := []string{}
	
	// Current working directory
	if cwd, err := os.Getwd(); err == nil {
		locations = append(locations, filepath.Join(cwd, "config.json"))
	}
	
	// Executable directory
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		locations = append(locations, filepath.Join(exeDir, "config.json"))
	}
	
	// User home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(homeDir, ".featureplus-pr", "config.json"))
		locations = append(locations, filepath.Join(homeDir, "config.json"))
	}
	
	for i, location := range locations {
		exists := "❌"
		if _, err := os.Stat(location); err == nil {
			exists = "✅"
		}
		fmt.Printf("  %d. %s %s\n", i+1, exists, location)
	}
}
