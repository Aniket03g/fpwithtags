package markers

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// PatchFile represents the structure of a .fppatch file
type PatchFile struct {
	Marker     string `yaml:"marker"`
	Snippet    string `yaml:"snippet"`
	TargetFile string `yaml:"target_file"`
	Context    struct {
		Before []string `yaml:"before"`
		After  []string `yaml:"after"`
	} `yaml:"context"`
}

// CreatePatch creates a new patch file with the given marker, snippet, and context
func CreatePatch(markerName, targetFile, snippet string, beforeContext, afterContext []string) (*PatchFile, error) {
	patch := &PatchFile{
		Marker:     markerName,
		Snippet:    snippet,
		TargetFile: targetFile,
	}
	patch.Context.Before = beforeContext
	patch.Context.After = afterContext

	return patch, nil
}

// SavePatch saves a patch to a file with the given name
func SavePatch(patch *PatchFile, filename string) error {
	// Ensure the filename has the .fppatch extension
	if !strings.HasSuffix(filename, ".fppatch") {
		filename += ".fppatch"
	}

	// Marshal the patch to YAML
	data, err := yaml.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch to YAML: %w", err)
	}

	// Write the YAML to the file
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write patch file: %w", err)
	}

	return nil
}

// LoadPatch loads a patch from a file
func LoadPatch(filename string) (*PatchFile, error) {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read patch file: %w", err)
	}

	// Unmarshal the YAML
	var patch PatchFile
	err = yaml.Unmarshal(data, &patch)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal patch file: %w", err)
	}

	return &patch, nil
}

// GetMarkerContext retrieves the context around a marker in a file
func GetMarkerContext(filePath, markerName string, contextLines int) ([]string, []string, int, error) {
	// Read the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, fmt.Errorf("error reading file: %w", err)
	}

	// Find the marker
	markerLine := -1
	markerEndLine := -1
	startPattern := fmt.Sprintf("// @fp:marker-start:%s", markerName)
	endPattern := fmt.Sprintf("// @fp:marker-end:%s", markerName)

	for i, line := range lines {
		if strings.Contains(line, startPattern) {
			markerLine = i
		} else if markerLine != -1 && strings.Contains(line, endPattern) {
			markerEndLine = i
			break
		}
	}

	if markerLine == -1 {
		return nil, nil, 0, fmt.Errorf("marker '%s' not found in file", markerName)
	}

	// Get context before the marker
	beforeStart := max(0, markerLine-contextLines)
	beforeContext := lines[beforeStart:markerLine]

	// Get context after the marker
	afterEnd := min(len(lines), markerEndLine+contextLines+1)
	afterContext := lines[markerEndLine+1:afterEnd]

	// Return the insertion point (line after the start marker)
	insertionPoint := markerLine + 1

	return beforeContext, afterContext, insertionPoint, nil
}

// ApplyPatch applies a patch to a file
func ApplyPatch(patch *PatchFile) error {
	// Read the file
	file, err := os.Open(patch.TargetFile)
	if err != nil {
		return fmt.Errorf("failed to open target file: %w", err)
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Find the marker
	markerLine := -1
	markerEndLine := -1
	startPattern := fmt.Sprintf("// @fp:marker-start:%s", patch.Marker)
	endPattern := fmt.Sprintf("// @fp:marker-end:%s", patch.Marker)

	for i, line := range lines {
		if strings.Contains(line, startPattern) {
			markerLine = i
		} else if markerLine != -1 && strings.Contains(line, endPattern) {
			markerEndLine = i
			break
		}
	}

	if markerLine == -1 {
		return fmt.Errorf("marker '%s' not found in file", patch.Marker)
	}

	// Validate context to ensure the file hasn't changed
	beforeStart := max(0, markerLine-len(patch.Context.Before))
	beforeContext := lines[beforeStart:markerLine]
	
	afterEnd := min(len(lines), markerEndLine+len(patch.Context.After)+1)
	afterContext := lines[markerEndLine+1:afterEnd]

	// Simple context validation (could be more sophisticated)
	if !contextMatches(beforeContext, patch.Context.Before) || !contextMatches(afterContext, patch.Context.After) {
		return fmt.Errorf("context mismatch: file has changed since patch was created")
	}

	// Insert the snippet after the marker start line
	insertionPoint := markerLine + 1
	indentation := getIndentation(lines[insertionPoint])
	
	// Split the snippet into lines and apply indentation
	snippetLines := strings.Split(patch.Snippet, "\n")
	for i, line := range snippetLines {
		if line != "" {
			snippetLines[i] = indentation + line
		}
	}
	
	// Add comment above the snippet
	comment := indentation + "// inserted by featureplus"
	
	// Create the new content
	newLines := make([]string, 0, len(lines)+len(snippetLines)+1)
	newLines = append(newLines, lines[:insertionPoint]...)
	newLines = append(newLines, comment)
	newLines = append(newLines, snippetLines...)
	newLines = append(newLines, lines[insertionPoint:]...)
	
	// Write the modified content back to a temporary file
	tempFile := patch.TargetFile + ".tmp"
	err = os.WriteFile(tempFile, []byte(strings.Join(newLines, "\n")), 0644)
	if err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	
	// Run go fmt on the temporary file if it's a Go file
	if strings.HasSuffix(patch.TargetFile, ".go") {
		cmd := exec.Command("go", "fmt", tempFile)
		err = cmd.Run()
		if err != nil {
			// Clean up the temporary file
			os.Remove(tempFile)
			return fmt.Errorf("failed to format file: %w", err)
		}
	}
	
	// Rename the temporary file to the original file
	err = os.Rename(tempFile, patch.TargetFile)
	if err != nil {
		// Clean up the temporary file
		os.Remove(tempFile)
		return fmt.Errorf("failed to replace original file: %w", err)
	}
	
	return nil
}

// contextMatches checks if the actual context matches the expected context
func contextMatches(actual, expected []string) bool {
	// If the expected context is empty, don't validate
	if len(expected) == 0 {
		return true
	}
	
	// If the actual context is shorter than expected, it can't match
	if len(actual) < len(expected) {
		return false
	}
	
	// Check the last N lines of actual against expected
	actualOffset := len(actual) - len(expected)
	for i := 0; i < len(expected); i++ {
		if !strings.Contains(actual[actualOffset+i], expected[i]) {
			return false
		}
	}
	
	return true
}

// getIndentation extracts the leading whitespace from a line
func getIndentation(line string) string {
	for i, c := range line {
		if c != ' ' && c != '\t' {
			return line[:i]
		}
	}
	return ""
}

// PromptForSnippet prompts the user to enter a code snippet
func PromptForSnippet(reader io.Reader) (string, error) {
	fmt.Println("Enter your code snippet (type 'EOF' on a new line to finish):")
	
	scanner := bufio.NewScanner(reader)
	var lines []string
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "EOF" {
			break
		}
		lines = append(lines, line)
	}
	
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	
	return strings.Join(lines, "\n"), nil
}

// GeneratePatchFilename generates a filename for a patch based on the marker name
func GeneratePatchFilename(markerName string) string {
	// Convert to lowercase and replace spaces with hyphens
	name := strings.ToLower(markerName)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	
	// Add a prefix based on the marker name
	prefix := ""
	if strings.Contains(name, "add") {
		prefix = "add"
	} else if strings.Contains(name, "update") {
		prefix = "update"
	} else if strings.Contains(name, "fix") {
		prefix = "fix"
	} else {
		prefix = "patch"
	}
	
	return fmt.Sprintf("%s-%s.fppatch", prefix, name)
}
