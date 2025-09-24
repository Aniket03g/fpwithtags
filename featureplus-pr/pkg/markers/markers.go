package markers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Marker represents a code marker in the project
type Marker struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	LineHint string `json:"line_hint"`
}

// MarkersIndex represents the structure of markers_index.json
type MarkersIndex struct {
	Markers []Marker `json:"markers"`
}

// LoadMarkers reads markers_index.json from the given path and returns a slice of Markers
func LoadMarkers(path string) ([]Marker, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read markers index file: %w", err)
	}

	// Parse the JSON
	var index MarkersIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse markers index file: %w", err)
	}

	return index.Markers, nil
}

// FindMarkerByName searches for a marker with the given name
func FindMarkerByName(markers []Marker, name string) (*Marker, error) {
	for _, marker := range markers {
		if marker.Name == name {
			return &marker, nil
		}
	}
	return nil, fmt.Errorf("marker '%s' not found", name)
}

// FindMarkersIndexPath attempts to locate the markers_index.json file
func FindMarkersIndexPath() (string, error) {
	// Try current directory first
	if _, err := os.Stat("markers_index.json"); err == nil {
		return "markers_index.json", nil
	}

	// Try parent directory
	parentPath := filepath.Join("..", "markers_index.json")
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath, nil
	}

	// Try project root (assuming we might be in a subdirectory)
	rootPath := filepath.Join("..", "..", "markers_index.json")
	if _, err := os.Stat(rootPath); err == nil {
		return rootPath, nil
	}

	// Try going up one more level
	grandparentPath := filepath.Join("..", "..", "..", "markers_index.json")
	if _, err := os.Stat(grandparentPath); err == nil {
		return grandparentPath, nil
	}

	// Try going up to the fourth level
	greatGrandparentPath := filepath.Join("..", "..", "..", "..", "markers_index.json")
	if _, err := os.Stat(greatGrandparentPath); err == nil {
		return greatGrandparentPath, nil
	}

	// Print current working directory for debugging
	cwd, _ := os.Getwd()
	return "", fmt.Errorf("markers_index.json not found in current directory or parent directories (searched from %s)", cwd)
}
