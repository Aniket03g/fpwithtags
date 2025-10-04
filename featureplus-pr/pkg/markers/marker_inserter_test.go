package markers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkerInserter(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "marker-inserter-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.go")
	testContent := `package test

import (
	"fmt"
)

func main() {
	fmt.Println("Hello, world!")
}

func add(a, b int) int {
	return a + b
}
`
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create marker candidates
	candidates := []MarkerCandidate{
		{
			Name:       "MAIN_FUNC",
			File:       testFilePath,
			LineNumber: 6,
			Type:       "function",
		},
		{
			Name:       "ADD_FUNC",
			File:       testFilePath,
			LineNumber: 10,
			Type:       "function",
		},
	}

	// Create inserter with default config
	inserter := NewMarkerInserter(InserterConfig{
		DryRun:      false,
		StartFormat: "// @fp:marker-start:%s",
		EndFormat:   "// @fp:marker-end:%s",
	})

	// Insert markers
	result, err := inserter.InsertMarkers(candidates)
	if err != nil {
		t.Fatalf("Failed to insert markers: %v", err)
	}

	// Check results
	if result.MarkersInserted != 2 {
		t.Errorf("Expected 2 markers inserted, got %d", result.MarkersInserted)
	}

	// Read the modified file
	modifiedContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	// Check that markers were inserted
	modifiedStr := string(modifiedContent)
	if !strings.Contains(modifiedStr, "// @fp:marker-start:MAIN_FUNC") {
		t.Error("Start marker for MAIN_FUNC not found")
	}
	if !strings.Contains(modifiedStr, "// @fp:marker-end:MAIN_FUNC") {
		t.Error("End marker for MAIN_FUNC not found")
	}
	if !strings.Contains(modifiedStr, "// @fp:marker-start:ADD_FUNC") {
		t.Error("Start marker for ADD_FUNC not found")
	}
	if !strings.Contains(modifiedStr, "// @fp:marker-end:ADD_FUNC") {
		t.Error("End marker for ADD_FUNC not found")
	}

	// Test idempotency - running again should not insert duplicate markers
	result, err = inserter.InsertMarkers(candidates)
	if err != nil {
		t.Fatalf("Failed to insert markers (second run): %v", err)
	}

	// Check that no new markers were inserted
	if result.MarkersSkipped != 2 {
		t.Errorf("Expected 2 markers skipped, got %d", result.MarkersSkipped)
	}
	if result.MarkersInserted != 0 {
		t.Errorf("Expected 0 markers inserted, got %d", result.MarkersInserted)
	}
}

func TestDryRun(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "marker-inserter-dryrun-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.go")
	testContent := `package test

func main() {
	println("Hello, world!")
}
`
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create marker candidates
	candidates := []MarkerCandidate{
		{
			Name:       "MAIN_FUNC",
			File:       testFilePath,
			LineNumber: 3,
			Type:       "function",
		},
	}

	// Create inserter with dry run enabled
	inserter := NewMarkerInserter(InserterConfig{
		DryRun:      true,
		StartFormat: "// @fp:marker-start:%s",
		EndFormat:   "// @fp:marker-end:%s",
	})

	// Insert markers
	result, err := inserter.InsertMarkers(candidates)
	if err != nil {
		t.Fatalf("Failed to insert markers: %v", err)
	}

	// Check results
	if result.MarkersInserted != 1 {
		t.Errorf("Expected 1 marker inserted (in dry run), got %d", result.MarkersInserted)
	}

	// Read the file and verify it wasn't modified
	modifiedContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check that the file wasn't modified
	if string(modifiedContent) != testContent {
		t.Error("File was modified despite dry run mode")
	}
}

func TestFilteredInsertion(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "marker-inserter-filter-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files in different directories
	dirs := []string{
		filepath.Join(tempDir, "src", "handlers"),
		filepath.Join(tempDir, "src", "models"),
	}
	
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create test files
	files := map[string]string{
		filepath.Join(dirs[0], "user_handler.go"): `package handlers

func GetUser() {
	// Get user logic
}`,
		filepath.Join(dirs[1], "user.go"): `package models

type User struct {
	ID   int
	Name string
}`,
	}

	for path, content := range files {
		err = os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}

	// Create marker candidates for all files
	var candidates []MarkerCandidate
	for path := range files {
		candidates = append(candidates, MarkerCandidate{
			Name:       filepath.Base(path),
			File:       path,
			LineNumber: 3,
			Type:       strings.Contains(path, "handler") ? "handler" : "model",
		})
	}

	// Filter candidates to only include handlers
	var handlerCandidates []MarkerCandidate
	for _, c := range candidates {
		if c.Type == "handler" {
			handlerCandidates = append(handlerCandidates, c)
		}
	}

	// Create inserter
	inserter := NewMarkerInserter(InserterConfig{
		DryRun:      false,
		StartFormat: "// @fp:marker-start:%s",
		EndFormat:   "// @fp:marker-end:%s",
	})

	// Insert markers for handlers only
	result, err := inserter.InsertMarkers(handlerCandidates)
	if err != nil {
		t.Fatalf("Failed to insert markers: %v", err)
	}

	// Check results
	if result.MarkersInserted != 1 {
		t.Errorf("Expected 1 marker inserted, got %d", result.MarkersInserted)
	}

	// Verify that only the handler file was modified
	for path, originalContent := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", path, err)
		}

		if strings.Contains(path, "handler") {
			// Handler file should have markers
			if !strings.Contains(string(content), "// @fp:marker-start:") {
				t.Errorf("Handler file %s should have markers", path)
			}
		} else {
			// Model file should not have markers
			if string(content) != originalContent {
				t.Errorf("Model file %s should not be modified", path)
			}
		}
	}
}
