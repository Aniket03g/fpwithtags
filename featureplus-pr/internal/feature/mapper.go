package feature

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"featureplus-pr/internal/git"
	"gopkg.in/yaml.v3"
)

// LocalFeature represents a feature loaded from local manifest
type LocalFeature struct {
	FilePath string
	Manifest *FeatureManifest
}

// LoadAllFeatureManifests reads all feature manifests from .featureplus/features/
func LoadAllFeatureManifests() ([]LocalFeature, error) {
	featuresDir := filepath.Join(".featureplus", "features")
	
	// Check if features directory exists
	if _, err := os.Stat(featuresDir); os.IsNotExist(err) {
		return []LocalFeature{}, nil // No features yet, return empty list
	}
	
	// Read all YAML files in the directory
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read features directory: %w", err)
	}
	
	var features []LocalFeature
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// Only process .yaml and .yml files
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}
		
		filePath := filepath.Join(featuresDir, entry.Name())
		
		// Read the manifest
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Failed to read %s: %v\n", filePath, err)
			continue
		}
		
		// Parse YAML
		var manifest FeatureManifest
		if err := yaml.Unmarshal(data, &manifest); err != nil {
			fmt.Printf("Warning: Failed to parse %s: %v\n", filePath, err)
			continue
		}
		
		features = append(features, LocalFeature{
			FilePath: filePath,
			Manifest: &manifest,
		})
	}
	
	return features, nil
}

// MapResult represents the result of mapping git history to features
type MapResult struct {
	FeatureID    string
	FeatureName  string
	Files        []string
	CommitCount  int
	CommitHashes []string
}

// MapGitHistoryToFeatures scans git history and maps commits to local features
func MapGitHistoryToFeatures(commitLimit int) ([]MapResult, error) {
	// Load all local feature manifests
	localFeatures, err := LoadAllFeatureManifests()
	if err != nil {
		return nil, fmt.Errorf("failed to load feature manifests: %w", err)
	}
	
	if len(localFeatures) == 0 {
		return nil, fmt.Errorf("no feature manifests found in .featureplus/features/")
	}
	
	// Scan git history
	commits, err := git.ScanGitHistory(commitLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to scan git history: %w", err)
	}
	
	// Map commits to features
	featureMap := git.MapCommitsToFeatures(commits)
	
	// Build results
	var results []MapResult
	
	// Create a map of local features by ID for quick lookup
	localFeatureMap := make(map[string]*FeatureManifest)
	for _, lf := range localFeatures {
		localFeatureMap[lf.Manifest.ID] = lf.Manifest
	}
	
	// Process each feature found in git history
	for featureID, mapping := range featureMap {
		// Check if this feature exists locally
		manifest, exists := localFeatureMap[featureID]
		
		featureName := ""
		if exists {
			featureName = manifest.Name
		}
		
		// Get commit hashes
		commitHashes := make([]string, 0, len(mapping.Commits))
		for _, commit := range mapping.Commits {
			commitHashes = append(commitHashes, commit.Hash)
		}
		
		results = append(results, MapResult{
			FeatureID:    featureID,
			FeatureName:  featureName,
			Files:        mapping.GetUniqueFiles(),
			CommitCount:  mapping.GetCommitCount(),
			CommitHashes: commitHashes,
		})
	}
	
	return results, nil
}
