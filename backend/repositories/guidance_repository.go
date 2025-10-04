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
	Context     string   `json:"context"`      // NEW: Project context (Development, Staging, Production, Testing)
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
	
	// STAGE 2a: Set default context for backwards compatibility
	// If guidance entries don't have a context field, default to "Development"
	for i := range guidanceData.Guidances {
		if guidanceData.Guidances[i].Context == "" {
			guidanceData.Guidances[i].Context = "Development"
		}
	}
	
	// Set default context for default guidance if not specified
	if guidanceData.DefaultGuidance.Context == "" {
		guidanceData.DefaultGuidance.Context = "Development"
	}
	
	r.data = &guidanceData
	return nil
}

// GetGuidance retrieves guidance for a specific stack and task type
// Deprecated: Use GetGuidanceWithContext for context-aware guidance
func (r *GuidanceRepository) GetGuidance(stack, taskType string) GuidanceEntry {
	// Default to Development context for backwards compatibility
	return r.GetGuidanceWithContext(stack, taskType, "Development")
}

// GetGuidanceWithContext retrieves guidance for a specific stack, task type, and context
// STAGE 2a: New method that includes context filtering
func (r *GuidanceRepository) GetGuidanceWithContext(stack, taskType, context string) GuidanceEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Default context to "Development" if empty
	if context == "" {
		context = "Development"
	}
	
	// Search for exact match (stack + task type + context)
	for _, guidance := range r.data.Guidances {
		if guidance.Stack == stack && guidance.TaskType == taskType && guidance.Context == context {
			return guidance
		}
	}
	
	// Fallback: Try to find guidance with same stack and task type but Development context
	if context != "Development" {
		for _, guidance := range r.data.Guidances {
			if guidance.Stack == stack && guidance.TaskType == taskType && guidance.Context == "Development" {
				return guidance
			}
		}
	}
	
	// Fallback: Try to find guidance with same stack and task type (any context)
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
	
	// STAGE 2a: Set default context if not provided
	if guidance.Context == "" {
		guidance.Context = "Development"
	}
	
	// Check if guidance already exists (stack + task type + context)
	for i, existing := range r.data.Guidances {
		if existing.Stack == guidance.Stack && existing.TaskType == guidance.TaskType && existing.Context == guidance.Context {
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
// Note: This deletes all contexts for the given stack and task type
func (r *GuidanceRepository) DeleteGuidance(stack, taskType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	found := false
	// Remove all guidances matching stack and task type (all contexts)
	for i := len(r.data.Guidances) - 1; i >= 0; i-- {
		guidance := r.data.Guidances[i]
		if guidance.Stack == stack && guidance.TaskType == taskType {
			r.data.Guidances = append(r.data.Guidances[:i], r.data.Guidances[i+1:]...)
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("guidance not found for stack: %s, task type: %s", stack, taskType)
	}
	
	return r.saveData()
}

// DeleteGuidanceWithContext removes a specific guidance entry by stack, task type, and context
// STAGE 2a: New method for context-specific deletion
func (r *GuidanceRepository) DeleteGuidanceWithContext(stack, taskType, context string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i, guidance := range r.data.Guidances {
		if guidance.Stack == stack && guidance.TaskType == taskType && guidance.Context == context {
			// Remove the guidance
			r.data.Guidances = append(r.data.Guidances[:i], r.data.Guidances[i+1:]...)
			return r.saveData()
		}
	}
	
	return fmt.Errorf("guidance not found for stack: %s, task type: %s, context: %s", stack, taskType, context)
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
