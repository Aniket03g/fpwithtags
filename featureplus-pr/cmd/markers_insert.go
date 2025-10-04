package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"featureplus-pr/pkg/markers"
	"github.com/spf13/cobra"
)

// Variables for flags
var (
	dryRun        bool
	fileFilter    string
	dirFilter     string
	typeFilter    string
	inputFile     string
	startMarker   string
	endMarker     string
	markerPrefix  string
	repairMarkers bool
)

// markersInsertCmd represents the insert command
var markersInsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert code markers into files",
	Long: `Insert code markers into files based on input from Phase 1 analysis.
This command reads a JSON file with marker locations and inserts markers
into the specified files. It can operate in dry-run mode to show what
would be changed without actually modifying files.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set default marker format if not provided
		if startMarker == "" {
			startMarker = "// @fp:marker-start:%s"
		}
		if endMarker == "" {
			endMarker = "// @fp:marker-end:%s"
		}

		// Determine input source
		var candidateMarkers []markers.MarkerCandidate
		var err error

		if inputFile != "" {
			// Load from specified input file
			candidateMarkers, err = markers.LoadMarkerCandidates(inputFile)
		} else {
			// Load from default markers_index.json
			indexPath, err := markers.FindMarkersIndexPath()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				fmt.Fprintf(os.Stderr, "\nMake sure you're running this command from a directory that has access to markers_index.json\n")
				fmt.Fprintf(os.Stderr, "The file should be in the project root directory.\n")
				os.Exit(1)
			}

			// Convert existing markers to candidates
			existingMarkers, err := markers.LoadMarkers(indexPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading markers: %v\n", err)
				os.Exit(1)
			}

			candidateMarkers = markers.ConvertToCandidates(existingMarkers)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading marker candidates: %v\n", err)
			os.Exit(1)
		}

		// Apply filters
		if fileFilter != "" || dirFilter != "" || typeFilter != "" {
			candidateMarkers = filterCandidates(candidateMarkers, fileFilter, dirFilter, typeFilter)
		}

		// Create inserter with configuration
		inserter := markers.NewMarkerInserter(markers.InserterConfig{
			DryRun:       dryRun,
			StartFormat:  startMarker,
			EndFormat:    endMarker,
			MarkerPrefix: markerPrefix,
			RepairMode:   repairMarkers,
		})

		// Perform insertion
		result, err := inserter.InsertMarkers(candidateMarkers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error inserting markers: %v\n", err)
			os.Exit(1)
		}

		// Print summary
		printInsertionSummary(result, dryRun)
	},
}

// filterCandidates applies filters to the candidate markers
func filterCandidates(candidates []markers.MarkerCandidate, fileFilter, dirFilter, typeFilter string) []markers.MarkerCandidate {
	var filtered []markers.MarkerCandidate

	for _, candidate := range candidates {
		// Apply file filter
		if fileFilter != "" && !strings.Contains(candidate.File, fileFilter) {
			continue
		}

		// Apply directory filter
		if dirFilter != "" {
			dir := filepath.Dir(candidate.File)
			if !strings.Contains(dir, dirFilter) {
				continue
			}
		}

		// Apply type filter
		if typeFilter != "" && !strings.Contains(candidate.Type, typeFilter) {
			continue
		}

		filtered = append(filtered, candidate)
	}

	return filtered
}

// printInsertionSummary prints a summary of the insertion operation
func printInsertionSummary(result markers.InsertionResult, dryRun bool) {
	if dryRun {
		fmt.Println("\n[DRY RUN] Summary of changes that would be made:")
	} else {
		fmt.Println("\nSummary of changes:")
	}

	fmt.Printf("- Files processed: %d\n", result.FilesProcessed)
	fmt.Printf("- Markers inserted: %d\n", result.MarkersInserted)
	fmt.Printf("- Markers skipped (already exist): %d\n", result.MarkersSkipped)
	fmt.Printf("- Markers repaired: %d\n", result.MarkersRepaired)

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors encountered:")
		for file, err := range result.Errors {
			fmt.Printf("- %s: %v\n", file, err)
		}
	}
}

func init() {
	markersCmd.AddCommand(markersInsertCmd)

	// Add flags
	markersInsertCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without modifying files")
	markersInsertCmd.Flags().StringVar(&fileFilter, "file", "", "Filter by file name (substring match)")
	markersInsertCmd.Flags().StringVar(&dirFilter, "dir", "", "Filter by directory name (substring match)")
	markersInsertCmd.Flags().StringVar(&typeFilter, "type", "", "Filter by marker type (substring match)")
	markersInsertCmd.Flags().StringVar(&inputFile, "input", "", "Path to input JSON file with marker candidates (defaults to markers_index.json)")
	markersInsertCmd.Flags().StringVar(&startMarker, "start-format", "", "Format string for start markers (default: '// @fp:marker-start:%s')")
	markersInsertCmd.Flags().StringVar(&endMarker, "end-format", "", "Format string for end markers (default: '// @fp:marker-end:%s')")
	markersInsertCmd.Flags().StringVar(&markerPrefix, "prefix", "", "Prefix to add to marker names")
	markersInsertCmd.Flags().BoolVar(&repairMarkers, "repair", false, "Repair malformed markers")
}
