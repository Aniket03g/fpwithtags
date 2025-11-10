package manifest

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMergeUnique(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		new      []string
		expected int // Expected length after merge
	}{
		{
			name:     "No duplicates",
			existing: []string{"a", "b"},
			new:      []string{"c", "d"},
			expected: 4,
		},
		{
			name:     "With duplicates",
			existing: []string{"a", "b", "c"},
			new:      []string{"b", "c", "d"},
			expected: 4, // a, b, c, d
		},
		{
			name:     "All duplicates",
			existing: []string{"a", "b"},
			new:      []string{"a", "b"},
			expected: 2,
		},
		{
			name:     "Empty existing",
			existing: []string{},
			new:      []string{"a", "b"},
			expected: 2,
		},
		{
			name:     "Empty new",
			existing: []string{"a", "b"},
			new:      []string{},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeUnique(tt.existing, tt.new)
			if len(result) != tt.expected {
				t.Errorf("mergeUnique() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestUpdateManifest(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	featuresDir := filepath.Join(tempDir, ".featureplus", "features")
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create initial manifest
	initialManifest := &FeatureManifest{
		ID:          "FTR-001",
		Name:        "Test Feature",
		Description: "Test description",
		Status:      "in-progress",
		Files:       []string{"file1.go", "file2.go"},
		Commits:     []string{"abc123", "def456"},
		CreatedAt:   time.Now(),
	}

	// Write initial manifest
	manifestPath := filepath.Join(featuresDir, "FTR-001.yaml")
	if err := writeManifest(manifestPath, initialManifest); err != nil {
		t.Fatalf("Failed to write initial manifest: %v", err)
	}

	// Update manifest with new files and commits
	newFiles := []string{"file2.go", "file3.go"} // file2.go is duplicate
	newCommits := []string{"def456", "ghi789"}   // def456 is duplicate

	err := UpdateManifest("FTR-001", newFiles, newCommits)
	if err != nil {
		t.Fatalf("UpdateManifest() error = %v", err)
	}

	// Read updated manifest
	updated, err := readManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read updated manifest: %v", err)
	}

	// Check that files were merged (should have 3 unique files)
	if len(updated.Files) != 3 {
		t.Errorf("Updated manifest has %d files, want 3", len(updated.Files))
	}

	// Check that commits were merged (should have 3 unique commits)
	if len(updated.Commits) != 3 {
		t.Errorf("Updated manifest has %d commits, want 3", len(updated.Commits))
	}

	// Check that UpdatedAt was set
	if updated.UpdatedAt.IsZero() {
		t.Error("UpdatedAt was not set")
	}

	// Check that other fields were preserved
	if updated.Name != "Test Feature" {
		t.Errorf("Name was changed to %q, want %q", updated.Name, "Test Feature")
	}
}

func TestGetManifestStats(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	featuresDir := filepath.Join(tempDir, ".featureplus", "features")
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create test manifest
	manifest := &FeatureManifest{
		ID:      "FTR-001",
		Name:    "Test Feature",
		Files:   []string{"file1.go", "file2.go", "file3.go"},
		Commits: []string{"abc123", "def456"},
	}

	manifestPath := filepath.Join(featuresDir, "FTR-001.yaml")
	if err := writeManifest(manifestPath, manifest); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Get stats
	fileCount, commitCount, err := GetManifestStats("FTR-001")
	if err != nil {
		t.Fatalf("GetManifestStats() error = %v", err)
	}

	if fileCount != 3 {
		t.Errorf("GetManifestStats() fileCount = %d, want 3", fileCount)
	}

	if commitCount != 2 {
		t.Errorf("GetManifestStats() commitCount = %d, want 2", commitCount)
	}
}
