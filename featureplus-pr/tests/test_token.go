package tests

import (
	"fmt"
	"io/ioutil"
	"os"
)

func TestToken() {
	// Check if config file exists
	configFile := "config.json"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Config file %s does not exist\n", configFile)
		return
	}
	
	// Read config file directly
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		return
	}
	
	// Print config file content
	fmt.Printf("Config file content:\n%s\n", string(data))
}
