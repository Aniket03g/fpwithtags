package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// FeatureManifest represents the local YAML structure for a feature
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

// SyncPayload represents the data sent to the backend
type SyncPayload struct {
	FeatureID string   `json:"feature_id"`
	Files     []string `json:"files"`
	Commits   []string `json:"commits"`
	Status    string   `json:"status"`
}

// SyncResult represents the result of syncing a single feature
type SyncResult struct {
	FeatureID string
	Success   bool
	Error     error
	FileCount int
	CommitCount int
}

// Client interface for backend communication
type Client interface {
	SyncFeature(payload SyncPayload) error
}

// LoadAllManifests reads all feature manifests from .featureplus/features/
func LoadAllManifests() ([]FeatureManifest, error) {
	featuresDir := filepath.Join(".featureplus", "features")
	
	// Check if features directory exists
	if _, err := os.Stat(featuresDir); os.IsNotExist(err) {
		return []FeatureManifest{}, nil
	}
	
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read features directory: %w", err)
	}
	
	var manifests []FeatureManifest
	
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
		
		manifests = append(manifests, manifest)
	}
	
	return manifests, nil
}

// NeedsSyncing checks if a manifest has new data to sync
func NeedsSyncing(manifest FeatureManifest) bool {
	// Has files or commits
	hasData := len(manifest.Files) > 0 || len(manifest.Commits) > 0
	
	// Not synced yet, or updated after last sync
	notSynced := manifest.LastSynced.IsZero() || 
		(!manifest.UpdatedAt.IsZero() && manifest.UpdatedAt.After(manifest.LastSynced))
	
	return hasData && notSynced
}

// SyncFeatureToBackend syncs a single feature to the backend
func SyncFeatureToBackend(manifest FeatureManifest, apiURL, authToken string) error {
	payload := SyncPayload{
		FeatureID: manifest.ID,
		Files:     manifest.Files,
		Commits:   manifest.Commits,
		Status:    manifest.Status,
	}
	
	// Marshal payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// Create request
	url := fmt.Sprintf("%s/api/features/sync", apiURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}
	
	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication required: please login first")
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	return nil
}

// UpdateManifestSyncTime updates the last_synced timestamp in the manifest file
func UpdateManifestSyncTime(featureID string) error {
	filePath := filepath.Join(".featureplus", "features", fmt.Sprintf("%s.yaml", featureID))
	
	// Read manifest
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var manifest FeatureManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	// Update last_synced
	manifest.LastSynced = time.Now()
	
	// Write back
	updatedData, err := yaml.Marshal(&manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	
	return nil
}

// SyncAll syncs all manifests that need syncing
func SyncAll(apiURL, authToken string) ([]SyncResult, error) {
	// Load all manifests
	manifests, err := LoadAllManifests()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifests: %w", err)
	}
	
	if len(manifests) == 0 {
		return nil, fmt.Errorf("no feature manifests found")
	}
	
	var results []SyncResult
	
	for _, manifest := range manifests {
		// Skip if doesn't need syncing
		if !NeedsSyncing(manifest) {
			continue
		}
		
		result := SyncResult{
			FeatureID:   manifest.ID,
			FileCount:   len(manifest.Files),
			CommitCount: len(manifest.Commits),
		}
		
		// Sync to backend
		err := SyncFeatureToBackend(manifest, apiURL, authToken)
		if err != nil {
			result.Success = false
			result.Error = err
			results = append(results, result)
			continue
		}
		
		// Update sync timestamp
		if err := UpdateManifestSyncTime(manifest.ID); err != nil {
			result.Success = false
			result.Error = fmt.Errorf("synced but failed to update timestamp: %w", err)
			results = append(results, result)
			continue
		}
		
		result.Success = true
		results = append(results, result)
	}
	
	return results, nil
}
