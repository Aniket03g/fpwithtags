package repositories

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
)

// GuidanceEntry represents a single guidance item
type GuidanceEntry struct {
	Stack       string   `json:"stack"`
	TaskType    string   `json:"task_type"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Snippet     string   `json:"snippet"`
	Language    string   `json:"language"`
	Commands    []string `json:"commands"`
	SetupSteps  []string `json:"setup_steps"`
	DocsLink    string   `json:"docs_link"`
	StarterRepo string   `json:"starter_repo"`
}

// GuidanceData represents the entire guidance configuration
type GuidanceData struct {
	Guidances       []GuidanceEntry `json:"guidances"`
	DefaultGuidance GuidanceEntry   `json:"default_guidance"`
}

// GuidanceRepository handles guidance data operations
type GuidanceRepository struct {
	data     *GuidanceData
	dataPath string
	mu       sync.RWMutex
}

// NewGuidanceRepository creates a new guidance repository
func NewGuidanceRepository(dataPath string) (*GuidanceRepository, error) {
	repo := &GuidanceRepository{
		dataPath: dataPath,
	}
	
	// Load initial data
	if err := repo.LoadData(); err != nil {
		return nil, err
	}
	
	return repo, nil
}

// LoadData loads guidance data from JSON file
func (r *GuidanceRepository) LoadData() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Read the JSON file
	fullPath := filepath.Join(r.dataPath, "guidance.json")
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("failed to read guidance file: %w", err)
	}
	
	// Parse JSON
	var guidanceData GuidanceData
	if err := json.Unmarshal(data, &guidanceData); err != nil {
		return fmt.Errorf("failed to parse guidance JSON: %w", err)
	}
	
	r.data = &guidanceData
	return nil
}

// GetGuidance retrieves guidance for a specific stack and task type
func (r *GuidanceRepository) GetGuidance(stack, taskType string) GuidanceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Search for matching guidance
	for _, guidance := range r.data.Guidances {
		if guidance.Stack == stack && guidance.TaskType == taskType {
			return guidance
		}
	}
	
	// Return default guidance if no match found
	return r.data.DefaultGuidance
}

// GetAllGuidances returns all available guidances
func (r *GuidanceRepository) GetAllGuidances() []GuidanceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.data.Guidances
}

// GetGuidancesByStack returns all guidances for a specific stack
func (r *GuidanceRepository) GetGuidancesByStack(stack string) []GuidanceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []GuidanceEntry
	for _, guidance := range r.data.Guidances {
		if guidance.Stack == stack {
			result = append(result, guidance)
		}
	}
	
	return result
}

// GetAvailableStacks returns all unique stacks
func (r *GuidanceRepository) GetAvailableStacks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stackMap := make(map[string]bool)
	for _, guidance := range r.data.Guidances {
		stackMap[guidance.Stack] = true
	}
	
	var stacks []string
	for stack := range stackMap {
		stacks = append(stacks, stack)
	}
	
	return stacks
}

// AddGuidance adds a new guidance entry (for admin functionality)
func (r *GuidanceRepository) AddGuidance(guidance GuidanceEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if guidance already exists
	for i, existing := range r.data.Guidances {
		if existing.Stack == guidance.Stack && existing.TaskType == guidance.TaskType {
			// Update existing
			r.data.Guidances[i] = guidance
			return r.saveData()
		}
	}
	
	// Add new guidance
	r.data.Guidances = append(r.data.Guidances, guidance)
	return r.saveData()
}

// DeleteGuidance removes a guidance entry
func (r *GuidanceRepository) DeleteGuidance(stack, taskType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i, guidance := range r.data.Guidances {
		if guidance.Stack == stack && guidance.TaskType == taskType {
			// Remove the guidance
			r.data.Guidances = append(r.data.Guidances[:i], r.data.Guidances[i+1:]...)
			return r.saveData()
		}
	}
	
	return fmt.Errorf("guidance not found for stack: %s, task type: %s", stack, taskType)
}

// saveData writes the current data back to the JSON file
func (r *GuidanceRepository) saveData() error {
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal guidance data: %w", err)
	}
	
	fullPath := filepath.Join(r.dataPath, "guidance.json")
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write guidance file: %w", err)
	}
	
	return nil
}

// ReloadData reloads guidance data from file (useful for hot-reloading)
func (r *GuidanceRepository) ReloadData() error {
	return r.LoadData()
}
