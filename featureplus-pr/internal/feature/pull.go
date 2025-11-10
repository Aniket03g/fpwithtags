package feature

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/FeaturePlus/pkg/featureplus"
	"gopkg.in/yaml.v3"
)

// FeatureManifest represents the local YAML structure for a feature
type FeatureManifest struct {
	ID                 string    `yaml:"id"`
	Name               string    `yaml:"name"`
	Description        string    `yaml:"description"`
	Status             string    `yaml:"status"`
	Owner              string    `yaml:"owner,omitempty"`
	Files              []string  `yaml:"files"`
	PRs                []string  `yaml:"prs"`
	Commits            []string  `yaml:"commits"`
	CreatedAt          time.Time `yaml:"created_at,omitempty"`
	UpdatedAt          time.Time `yaml:"updated_at,omitempty"`
	LastSynced         time.Time `yaml:"last_synced,omitempty"`
	SyncedFromBackend  bool      `yaml:"synced_from_backend"`
}

// SaveFeatureManifest writes a feature to a YAML file in .featureplus/features/
func SaveFeatureManifest(feature *featureplus.Feature, owner string) (string, error) {
	// Ensure .featureplus/features directory exists
	featuresDir := filepath.Join(".featureplus", "features")
	if err := os.MkdirAll(featuresDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create features directory: %w", err)
	}

	// Create the manifest
	manifest := FeatureManifest{
		ID:                fmt.Sprintf("FTR-%03d", feature.ID),
		Name:              feature.Title,
		Description:       feature.Description,
		Status:            feature.Status,
		Owner:             owner,
		Files:             []string{},
		PRs:               []string{},
		Commits:           []string{},
		CreatedAt:         time.Now(),
		SyncedFromBackend: true,
	}

	// Marshal to YAML
	yamlData, err := yaml.Marshal(&manifest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal feature to YAML: %w", err)
	}

	// Write to file
	filename := fmt.Sprintf("FTR-%03d.yaml", feature.ID)
	filePath := filepath.Join(featuresDir, filename)
	
	if err := os.WriteFile(filePath, yamlData, 0644); err != nil {
		return "", fmt.Errorf("failed to write YAML file: %w", err)
	}

	return filePath, nil
}

// FeatureExists checks if a feature manifest already exists locally
func FeatureExists(featureID uint) bool {
	filename := fmt.Sprintf("FTR-%03d.yaml", featureID)
	filePath := filepath.Join(".featureplus", "features", filename)
	_, err := os.Stat(filePath)
	return err == nil
}

// LoadFeatureManifest reads a feature manifest from disk
func LoadFeatureManifest(featureID uint) (*FeatureManifest, error) {
	filename := fmt.Sprintf("FTR-%03d.yaml", featureID)
	filePath := filepath.Join(".featureplus", "features", filename)
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feature manifest: %w", err)
	}

	var manifest FeatureManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse feature manifest: %w", err)
	}

	return &manifest, nil
}
