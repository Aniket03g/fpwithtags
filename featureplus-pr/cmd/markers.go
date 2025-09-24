package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"featureplus-pr/pkg/markers"
	"github.com/spf13/cobra"
)

// Variables for flags
var (
	verboseMarkers bool
	markerFormat   string
)

// markersCmd represents the markers command
var markersCmd = &cobra.Command{
	Use:   "markers",
	Short: "Manage code markers in the project",
	Long: `The markers command allows you to list and find code markers in the project.
These markers indicate where future code snippets can be inserted automatically.`,
}

// markersListCmd represents the list command
var markersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all markers in the project",
	Long:  `List all markers defined in the markers_index.json file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Find markers_index.json
		if verboseMarkers {
			fmt.Println("Searching for markers_index.json...")
		}

		indexPath, err := markers.FindMarkersIndexPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nMake sure you're running this command from a directory that has access to markers_index.json\n")
			fmt.Fprintf(os.Stderr, "The file should be in the project root directory.\n")
			os.Exit(1)
		}

		// Load markers
		if verboseMarkers {
			fmt.Printf("Loading markers from: %s\n", indexPath)
		}

		markersList, err := markers.LoadMarkers(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Print markers in a table
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tFILE\tLINE_HINT")
		for _, marker := range markersList {
			fmt.Fprintf(w, "%s\t%s\t%s\n", marker.Name, marker.File, marker.LineHint)
		}
		w.Flush()
		
		// Print summary
		fmt.Printf("\nFound %d markers in %s\n", len(markersList), indexPath)
	},
}

// markersFindCmd represents the find command
var markersFindCmd = &cobra.Command{
	Use:   "find [name]",
	Short: "Find a marker by name",
	Long:  `Find a marker by its name and display its details.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Find markers_index.json
		if verboseMarkers {
			fmt.Println("Searching for markers_index.json...")
		}

		indexPath, err := markers.FindMarkersIndexPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			fmt.Fprintf(os.Stderr, "\nMake sure you're running this command from a directory that has access to markers_index.json\n")
			fmt.Fprintf(os.Stderr, "The file should be in the project root directory.\n")
			os.Exit(1)
		}

		// Load markers
		if verboseMarkers {
			fmt.Printf("Loading markers from: %s\n", indexPath)
		}

		markersList, err := markers.LoadMarkers(indexPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Find marker by name
		marker, err := markers.FindMarkerByName(markersList, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			// List available markers
			fmt.Fprintf(os.Stderr, "\nAvailable markers:\n")
			for _, m := range markersList {
				fmt.Fprintf(os.Stderr, "  - %s\n", m.Name)
			}
			os.Exit(1)
		}

		// Print marker details based on format
		switch markerFormat {
		case "json":
			fmt.Printf("{\"name\": \"%s\", \"file\": \"%s\", \"line_hint\": \"%s\"}\n", 
				marker.Name, marker.File, marker.LineHint)
		default:
			fmt.Printf("Name: %s\n", marker.Name)
			fmt.Printf("File: %s\n", marker.File)
			fmt.Printf("Line Hint: %s\n", marker.LineHint)
			if verboseMarkers {
				fmt.Printf("\nLoaded from: %s\n", indexPath)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(markersCmd)
	markersCmd.AddCommand(markersListCmd)
	markersCmd.AddCommand(markersFindCmd)
	
	// Add flags similar to other commands
	markersCmd.PersistentFlags().BoolVarP(&verboseMarkers, "verbose", "v", false, "Enable verbose output")
	markersFindCmd.Flags().StringVarP(&markerFormat, "format", "f", "text", "Output format (text or json)")
}
