package markers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MarkerCandidate represents a location where a marker should be inserted
type MarkerCandidate struct {
	Name       string `json:"name"`
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
	Type       string `json:"type"`
	StartLine  int    `json:"start_line,omitempty"`
	EndLine    int    `json:"end_line,omitempty"`
}

// MarkerCandidateInput represents the structure of the input JSON file
type MarkerCandidateInput struct {
	Candidates []MarkerCandidate `json:"candidates"`
}

// InserterConfig holds configuration for the marker inserter
type InserterConfig struct {
	DryRun       bool
	StartFormat  string
	EndFormat    string
	MarkerPrefix string
	RepairMode   bool
}

// InsertionResult holds statistics about the insertion operation
type InsertionResult struct {
	FilesProcessed  int
	MarkersInserted int
	MarkersSkipped  int
	MarkersRepaired int
	Errors          map[string]error
}

// MarkerInserter handles the insertion of markers into files
type MarkerInserter struct {
	config InserterConfig
}

// NewMarkerInserter creates a new MarkerInserter with the given configuration
func NewMarkerInserter(config InserterConfig) *MarkerInserter {
	return &MarkerInserter{
		config: config,
	}
}

// LoadMarkerCandidates loads marker candidates from a JSON file
func LoadMarkerCandidates(path string) ([]MarkerCandidate, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read marker candidates file: %w", err)
	}

	// Parse the JSON
	var input MarkerCandidateInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse marker candidates file: %w", err)
	}

	return input.Candidates, nil
}

// ConvertToCandidates converts regular markers to marker candidates
func ConvertToCandidates(markers []Marker) []MarkerCandidate {
	candidates := make([]MarkerCandidate, len(markers))
	for i, marker := range markers {
		// Extract line number from line hint if possible
		lineNumber := 0
		// Default to function type if not specified
		markerType := "function"

		candidates[i] = MarkerCandidate{
			Name:       marker.Name,
			File:       marker.File,
			LineNumber: lineNumber,
			Type:       markerType,
		}
	}
	return candidates
}

// InsertMarkers inserts markers into files based on the provided candidates
func (m *MarkerInserter) InsertMarkers(candidates []MarkerCandidate) (InsertionResult, error) {
	result := InsertionResult{
		Errors: make(map[string]error),
	}

	// Group candidates by file to avoid opening the same file multiple times
	fileMap := make(map[string][]MarkerCandidate)
	for _, candidate := range candidates {
		fileMap[candidate.File] = append(fileMap[candidate.File], candidate)
	}

	// Process each file
	for filePath, fileCandidates := range fileMap {
		result.FilesProcessed++

		// Process the file
		fileResult, err := m.processFile(filePath, fileCandidates)
		if err != nil {
			result.Errors[filePath] = err
			continue
		}

		// Update overall results
		result.MarkersInserted += fileResult.MarkersInserted
		result.MarkersSkipped += fileResult.MarkersSkipped
		result.MarkersRepaired += fileResult.MarkersRepaired
	}

	return result, nil
}

// processFile handles marker insertion for a single file
func (m *MarkerInserter) processFile(filePath string, candidates []MarkerCandidate) (InsertionResult, error) {
	result := InsertionResult{
		Errors: make(map[string]error),
	}

	// Resolve relative path to absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return result, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return result, fmt.Errorf("failed to read file: %w", err)
	}

	// Split content into lines
	lines := strings.Split(string(content), "\n")

	// Track modifications
	modified := false
	newLines := make([]string, 0, len(lines)+len(candidates)*2) // Allocate extra space for markers

	// Process each line
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for candidates at this line
		for _, candidate := range candidates {
			if candidate.LineNumber == i+1 { // Convert to 1-indexed
				// Check if markers already exist
				if m.markersExist(lines, i, candidate) {
					result.MarkersSkipped++
					continue
				}

				// Check if markers need repair
				if m.config.RepairMode && m.markersNeedRepair(lines, i, candidate) {
					if !m.config.DryRun {
						// Repair logic would go here
						// This is a simplified placeholder
					}
					result.MarkersRepaired++
					modified = true
					continue
				}

				// Insert start marker
				startMarker := fmt.Sprintf(m.config.StartFormat, m.getMarkerName(candidate))
				indentation := m.getIndentation(line)
				newLines = append(newLines, indentation+startMarker)

				// Add the original line
				newLines = append(newLines, line)

				// Insert end marker if this is a single-line marker
				if candidate.EndLine == 0 || candidate.EndLine == candidate.LineNumber {
					endMarker := fmt.Sprintf(m.config.EndFormat, m.getMarkerName(candidate))
					newLines = append(newLines, indentation+endMarker)
				}

				result.MarkersInserted++
				modified = true
				continue
			}

			// Check for end marker position
			if candidate.EndLine > 0 && candidate.EndLine == i+1 {
				// Add the original line
				newLines = append(newLines, line)

				// Insert end marker
				endMarker := fmt.Sprintf(m.config.EndFormat, m.getMarkerName(candidate))
				indentation := m.getIndentation(line)
				newLines = append(newLines, indentation+endMarker)

				modified = true
				continue
			}
		}

		// Add the original line if we haven't already
		if len(newLines) == 0 || newLines[len(newLines)-1] != line {
			newLines = append(newLines, line)
		}
	}

	// Write the modified content back to the file if changes were made
	if modified && !m.config.DryRun {
		err = os.WriteFile(absPath, []byte(strings.Join(newLines, "\n")), 0644)
		if err != nil {
			return result, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return result, nil
}

// markersExist checks if markers already exist for the given candidate
func (m *MarkerInserter) markersExist(lines []string, lineIndex int, candidate MarkerCandidate) bool {
	// Check for start marker
	startPattern := fmt.Sprintf(strings.ReplaceAll(m.config.StartFormat, "%s", ".*"), m.getMarkerName(candidate))
	startRegex := regexp.MustCompile(startPattern)

	// Look for start marker in the lines before the current line
	for i := max(0, lineIndex-5); i < lineIndex; i++ {
		if startRegex.MatchString(lines[i]) {
			return true
		}
	}

	// Check for end marker if applicable
	if candidate.EndLine > 0 {
		endPattern := fmt.Sprintf(strings.ReplaceAll(m.config.EndFormat, "%s", ".*"), m.getMarkerName(candidate))
		endRegex := regexp.MustCompile(endPattern)

		// Look for end marker in the lines after the end line
		endLineIndex := candidate.EndLine - 1
		for i := endLineIndex; i < min(len(lines), endLineIndex+5); i++ {
			if endRegex.MatchString(lines[i]) {
				return true
			}
		}
	}

	return false
}

// markersNeedRepair checks if markers exist but are malformed
func (m *MarkerInserter) markersNeedRepair(lines []string, lineIndex int, candidate MarkerCandidate) bool {
	// This is a simplified implementation
	// A real implementation would check for various malformations
	return false
}

// getMarkerName returns the marker name with prefix if configured
func (m *MarkerInserter) getMarkerName(candidate MarkerCandidate) string {
	if m.config.MarkerPrefix != "" {
		return m.config.MarkerPrefix + "-" + candidate.Name
	}
	return candidate.Name
}

// getIndentation extracts the leading whitespace from a line
func (m *MarkerInserter) getIndentation(line string) string {
	for i, c := range line {
		if c != ' ' && c != '\t' {
			return line[:i]
		}
	}
	return ""
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
