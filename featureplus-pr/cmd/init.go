package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config represents the structure of config.yaml
type Config struct {
	APIURL    string `yaml:"api_url"`
	ProjectID string `yaml:"project_id"`
	LinkedAt  string `yaml:"linked_at"`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize FeaturePlus in the current directory",
	Long:  `Creates a .featureplus folder with config.yaml for project configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Define the hidden folder name
		folderName := ".featureplus"

		// Check if the folder already exists
		if _, err := os.Stat(folderName); err == nil {
			fmt.Println("FeaturePlus already initialized.")
			return
		}

		// Create the .featureplus folder
		if err := os.Mkdir(folderName, 0755); err != nil {
			fmt.Printf("Error creating %s folder: %v\n", folderName, err)
			os.Exit(1)
		}

		// Create the config structure
		config := Config{
			APIURL:    "http://localhost:8080",
			ProjectID: "",
			LinkedAt:  "",
		}

		// Marshal the config to YAML
		yamlData, err := yaml.Marshal(&config)
		if err != nil {
			fmt.Printf("Error creating config: %v\n", err)
			os.Exit(1)
		}

		// Write config.yaml inside .featureplus folder
		configPath := filepath.Join(folderName, "config.yaml")
		if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
			fmt.Printf("Error writing config file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Initialized FeaturePlus in this directory.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
