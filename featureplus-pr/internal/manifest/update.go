package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// FeatureManifest represents the YAML structure for a feature
type FeatureManifest struct {
	ID                string    `yaml:"id"`
	Name              string    `yaml:"name"`
	Description       string    `yaml:"description"`
	Status            string    `yaml:"status"`
	Owner             string    `yaml:"owner,omitempty"`
	Files             []string  `yaml:"files"`
	PRs               []string  `yaml:"prs"`
	Commits           []string  `yaml:"commits"`
	CreatedAt         time.Time `yaml:"created_at,omitempty"`
	UpdatedAt         time.Time `yaml:"updated_at,omitempty"`
	LastSynced        time.Time `yaml:"last_synced,omitempty"`
	SyncedFromBackend bool      `yaml:"synced_from_backend"`
}

// UpdateManifest reads a feature manifest, merges new files and commits, and writes it back
func UpdateManifest(featureID string, files, commits []string) error {
	// Construct file path
	filePath := filepath.Join(".featureplus", "features", fmt.Sprintf("%s.yaml", featureID))
	
	// Read existing manifest
	manifest, err := readManifest(filePath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	
	// Merge files (avoid duplicates)
	manifest.Files = mergeUnique(manifest.Files, files)
	
	// Merge commits (avoid duplicates)
	manifest.Commits = mergeUnique(manifest.Commits, commits)
	
	// Update timestamp
	manifest.UpdatedAt = time.Now()
	
	// Write back to disk
	if err := writeManifest(filePath, manifest); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	
	return nil
}

// readManifest reads and parses a YAML manifest file
func readManifest(filePath string) (*FeatureManifest, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	
	var manifest FeatureManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	
	return &manifest, nil
}

// writeManifest writes a manifest to disk as YAML
func writeManifest(filePath string, manifest *FeatureManifest) error {
	// Marshal to YAML
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return err
	}
	
	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}
	
	return nil
}

// mergeUnique merges two string slices, removing duplicates
func mergeUnique(existing, new []string) []string {
	// Use map for deduplication
	seen := make(map[string]bool)
	
	// Add existing items
	for _, item := range existing {
		seen[item] = true
	}
	
	// Add new items
	for _, item := range new {
		seen[item] = true
	}
	
	// Convert back to slice
	result := make([]string, 0, len(seen))
	for item := range seen {
		result = append(result, item)
	}
	
	return result
}

// GetManifestStats returns the current file and commit counts for a manifest
func GetManifestStats(featureID string) (fileCount, commitCount int, err error) {
	filePath := filepath.Join(".featureplus", "features", fmt.Sprintf("%s.yaml", featureID))
	
	manifest, err := readManifest(filePath)
	if err != nil {
		return 0, 0, err
	}
	
	return len(manifest.Files), len(manifest.Commits), nil
}
