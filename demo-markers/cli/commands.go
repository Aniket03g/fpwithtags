package cli

import (
	"fmt"
	"os"
	"strings"
)

// CommandHandler is a function that handles a CLI command
type CommandHandler func(args []string) error

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     CommandHandler
}

// Available commands
var commands = map[string]Command{
	"help": {
		Name:        "help",
		Description: "Show help information",
		Handler:     handleHelp,
	},
	"version": {
		Name:        "version",
		Description: "Show version information",
		Handler:     handleVersion,
	},
}

// HandleCommand processes a CLI command
func HandleCommand(cmdLine string) {
	args := strings.Fields(cmdLine)
	if len(args) == 0 {
		fmt.Println("No command provided")
		handleHelp(nil)
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "help":
		handleHelp(cmdArgs)
	case "version":
		handleVersion(cmdArgs)
	// MARKER:CLI_COMMANDS
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		handleHelp(nil)
		os.Exit(1)
	}
}

// handleHelp shows help information
func handleHelp(args []string) error {
	fmt.Println("Available commands:")
	for name, cmd := range commands {
		fmt.Printf("  %-10s %s\n", name, cmd.Description)
	}
	return nil
}

// handleVersion shows version information
func handleVersion(args []string) error {
	fmt.Println("Demo Markers v1.0.0")
	return nil
}
