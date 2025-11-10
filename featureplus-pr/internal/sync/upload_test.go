package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNeedsSyncing(t *testing.T) {
	tests := []struct {
		name     string
		manifest FeatureManifest
		expected bool
	}{
		{
			name: "New manifest with files",
			manifest: FeatureManifest{
				Files:      []string{"file1.go"},
				LastSynced: time.Time{}, // Zero time (never synced)
			},
			expected: true,
		},
		{
			name: "New manifest with commits",
			manifest: FeatureManifest{
				Commits:    []string{"abc123"},
				LastSynced: time.Time{},
			},
			expected: true,
		},
		{
			name: "Empty manifest",
			manifest: FeatureManifest{
				Files:      []string{},
				Commits:    []string{},
				LastSynced: time.Time{},
			},
			expected: false,
		},
		{
			name: "Already synced, no updates",
			manifest: FeatureManifest{
				Files:      []string{"file1.go"},
				UpdatedAt:  time.Now().Add(-2 * time.Hour),
				LastSynced: time.Now().Add(-1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "Updated after last sync",
			manifest: FeatureManifest{
				Files:      []string{"file1.go"},
				UpdatedAt:  time.Now(),
				LastSynced: time.Now().Add(-1 * time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NeedsSyncing(tt.manifest)
			if result != tt.expected {
				t.Errorf("NeedsSyncing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadAllManifests(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	featuresDir := filepath.Join(tempDir, ".featureplus", "features")
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Change to temp directory
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create test manifest files
	manifest1 := `id: FTR-001
name: Test Feature 1
files:
  - file1.go
commits:
  - abc123
`
	manifest2 := `id: FTR-002
name: Test Feature 2
files:
  - file2.go
`

	os.WriteFile(filepath.Join(featuresDir, "FTR-001.yaml"), []byte(manifest1), 0644)
	os.WriteFile(filepath.Join(featuresDir, "FTR-002.yaml"), []byte(manifest2), 0644)

	// Load manifests
	manifests, err := LoadAllManifests()
	if err != nil {
		t.Fatalf("LoadAllManifests() error = %v", err)
	}

	if len(manifests) != 2 {
		t.Errorf("LoadAllManifests() returned %d manifests, want 2", len(manifests))
	}
}

func TestUpdateManifestSyncTime(t *testing.T) {
	// Create temporary directory
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
	manifestYAML := `id: FTR-001
name: Test Feature
files:
  - file1.go
`
	manifestPath := filepath.Join(featuresDir, "FTR-001.yaml")
	os.WriteFile(manifestPath, []byte(manifestYAML), 0644)

	// Update sync time
	beforeUpdate := time.Now()
	err := UpdateManifestSyncTime("FTR-001")
	if err != nil {
		t.Fatalf("UpdateManifestSyncTime() error = %v", err)
	}

	// Read updated manifest
	data, _ := os.ReadFile(manifestPath)
	
	// Check that last_synced was added
	if !contains(string(data), "last_synced:") {
		t.Error("last_synced field was not added to manifest")
	}

	// Load and verify timestamp
	manifests, _ := LoadAllManifests()
	if len(manifests) == 0 {
		t.Fatal("Failed to load updated manifest")
	}

	if manifests[0].LastSynced.Before(beforeUpdate) {
		t.Error("LastSynced timestamp is before update time")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
