package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"featureplus-pr/pkg/markers"
	"github.com/spf13/cobra"
)

var (
	contextLines int
	patchFile    string
	outputFile   string
)

// markersInsertPatchCmd represents the enhanced insert command for creating patches
var markersInsertPatchCmd = &cobra.Command{
	Use:   "insert [marker]",
	Short: "Insert code at a marker location and create a patch file",
	Long: `Insert code at a marker location and create a patch file.
This command locates a marker in the code, shows context around it,
and prompts you to enter a code snippet. The snippet is saved to a
.fppatch file for later application.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		markerName := args[0]

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
		marker, err := markers.FindMarkerByName(markersList, markerName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			// List available markers
			fmt.Fprintf(os.Stderr, "\nAvailable markers:\n")
			for _, m := range markersList {
				fmt.Fprintf(os.Stderr, "  - %s\n", m.Name)
			}
			os.Exit(1)
		}

		// Get the absolute path to the file
		projectRoot := filepath.Dir(indexPath)
		filePath := filepath.Join(projectRoot, marker.File)

		// Get context around the marker
		beforeContext, afterContext, insertionPoint, err := markers.GetMarkerContext(filePath, markerName, contextLines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Display context
		fmt.Printf("\nContext for marker '%s' in file '%s':\n\n", markerName, marker.File)
		fmt.Println("--- Before ---")
		for _, line := range beforeContext {
			fmt.Println(line)
		}
		fmt.Println("\n--- Insertion Point (line", insertionPoint+1, ") ---")
		fmt.Println("\n--- After ---")
		for _, line := range afterContext {
			fmt.Println(line)
		}
		fmt.Println()

		// Prompt for snippet
		snippet, err := markers.PromptForSnippet(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Create patch
		patch, err := markers.CreatePatch(markerName, filePath, snippet, beforeContext, afterContext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Generate output filename if not provided
		if outputFile == "" {
			outputFile = markers.GeneratePatchFilename(markerName)
		}

		// Save patch
		err = markers.SavePatch(patch, outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nPatch saved to %s\n", outputFile)
		fmt.Printf("\nTo apply this patch, run:\n  featureplus markers apply %s\n", outputFile)
	},
}

// markersApplyCmd represents the apply command
var markersApplyCmd = &cobra.Command{
	Use:   "apply [patchfile]",
	Short: "Apply a patch file to insert code at a marker location",
	Long: `Apply a patch file to insert code at a marker location.
This command reads a .fppatch file, locates the marker in the code,
and inserts the snippet at the marker location. For Go files, it also
runs go fmt to ensure proper formatting.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchFile := args[0]

		// Ensure the file has the .fppatch extension
		if !strings.HasSuffix(patchFile, ".fppatch") {
			patchFile += ".fppatch"
		}

		// Load the patch
		patch, err := markers.LoadPatch(patchFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Applying patch to marker '%s' in file '%s'...\n", patch.Marker, patch.TargetFile)

		// Apply the patch
		err = markers.ApplyPatch(patch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Patch applied successfully!\n")
	},
}

func init() {
	// Override the existing insert command with our enhanced version
	markersCmd.AddCommand(markersInsertPatchCmd)
	markersCmd.AddCommand(markersApplyCmd)

	// Add flags
	markersInsertPatchCmd.Flags().IntVar(&contextLines, "context", 5, "Number of context lines to show before and after the marker")
	markersInsertPatchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file name (default: auto-generated)")

	// Add aliases
	markersInsertPatchCmd.Aliases = []string{"patch"}
	markersApplyCmd.Aliases = []string{"apply-patch"}
}
