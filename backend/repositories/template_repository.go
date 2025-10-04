package repositories

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Template represents a project template
type Template struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Stack                string            `json:"stack"`
	Description          string            `json:"description"`
	TechStack            string            `json:"tech_stack"`
	FeatureCategories    []string          `json:"feature_categories"`
	TaskTypes            []string          `json:"task_types"`
	Features             []TemplateFeature `json:"features"`
	Tasks                []TemplateTask    `json:"tasks"`
	Dependencies         []string          `json:"dependencies"`
	SetupSteps           []string          `json:"setup_steps"`
	EnvironmentVariables []string          `json:"environment_variables"`
	StarterRepo          string            `json:"starter_repo"`
	DocsLinks            []string          `json:"docs_links"`
}

// TemplateFeature represents a feature in a template
type TemplateFeature struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Context     string `json:"context"` // STAGE 2b: Project context (Development, Staging, Production, Testing)
}

// TemplateTask represents a task in a template
type TemplateTask struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Context     string `json:"context"` // STAGE 2b: Project context (Development, Staging, Production, Testing)
}

// TemplateData represents the entire template configuration
type TemplateData struct {
	Templates       []Template `json:"templates"`
	DefaultTemplate Template   `json:"default_template"`
}

// TemplateRepository handles template data operations
type TemplateRepository struct {
	data     *TemplateData
	dataPath string
	mu       sync.RWMutex
}

// FilePermissionInfo stores information about file permissions
type FilePermissionInfo struct {
	Exists      bool
	Readable    bool
	Writable    bool
	Permissions os.FileMode
	Size        int64
	Error       error
}

// checkFilePermissions checks if a file exists and has read/write permissions
func checkFilePermissions(path string) FilePermissionInfo {
	info := FilePermissionInfo{}
	
	// Check if file exists
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("INFO: File does not exist: %s", path)
			info.Error = fmt.Errorf("file does not exist: %w", err)
			return info
		}
		
		if os.IsPermission(err) {
			log.Printf("ERROR: Permission denied for file: %s", path)
			info.Error = fmt.Errorf("permission denied: %w", err)
			return info
		}
		
		log.Printf("ERROR: Failed to stat file %s: %v", path, err)
		info.Error = fmt.Errorf("failed to stat file: %w", err)
		return info
	}
	
	info.Exists = true
	info.Permissions = fileInfo.Mode()
	info.Size = fileInfo.Size()
	
	// Check read permission
	file, err := os.Open(path)
	if err != nil {
		log.Printf("ERROR: Cannot read file %s: %v", path, err)
		info.Error = fmt.Errorf("cannot read file: %w", err)
		return info
	}
	file.Close()
	info.Readable = true
	
	// Check write permission by trying to open for writing
	// Don't actually write anything
	file, err = os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			log.Printf("WARNING: File %s is not writable: %v", path, err)
			info.Writable = false
		} else {
			log.Printf("WARNING: Unknown error checking write permission for %s: %v", path, err)
			info.Writable = false
		}
	} else {
		file.Close()
		info.Writable = true
	}
	
	log.Printf("INFO: File permission check for %s: exists=%t, readable=%t, writable=%t, mode=%s, size=%d bytes",
		path, info.Exists, info.Readable, info.Writable, info.Permissions, info.Size)
	
	return info
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(dataPath string) (*TemplateRepository, error) {
	log.Printf("INFO: Creating template repository with initial data path: %s", dataPath)
	
	// Resolve absolute path
	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", dataPath, err)
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	
	log.Printf("INFO: Resolved absolute data path: %s", absPath)
	
	// Check if the directory exists
	dirInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("ERROR: Data directory does not exist: %s", absPath)
			return nil, fmt.Errorf("data directory does not exist: %s", absPath)
		}
		log.Printf("ERROR: Failed to stat data directory %s: %v", absPath, err)
		return nil, fmt.Errorf("failed to stat data directory: %w", err)
	}
	
	if !dirInfo.IsDir() {
		log.Printf("ERROR: Path is not a directory: %s", absPath)
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}
	
	log.Printf("INFO: Data directory exists with permissions: %s", dirInfo.Mode())
	
	// Check if the templates.json file exists
	templatesPath := filepath.Join(absPath, "templates.json")
	fileInfo := checkFilePermissions(templatesPath)
	if fileInfo.Error != nil {
		log.Printf("ERROR: Problem with templates.json file: %v", fileInfo.Error)
		return nil, fmt.Errorf("problem with templates.json file: %w", fileInfo.Error)
	}
	
	log.Printf("INFO: Templates file exists and is readable: %s", templatesPath)
	
	repo := &TemplateRepository{
		dataPath: absPath,
	}
	
	// Load initial data
	if err := repo.LoadData(); err != nil {
		log.Printf("ERROR: Failed to load template data: %v", err)
		return nil, fmt.Errorf("failed to load template data: %w", err)
	}
	
	return repo, nil
}

// LoadData loads template data from JSON file
func (r *TemplateRepository) LoadData() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Read the JSON file
	fullPath := filepath.Join(r.dataPath, "templates.json")
	log.Printf("DEBUG: Loading templates from: %s", fullPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		log.Printf("ERROR: Failed to read templates file: %v", err)
		return fmt.Errorf("failed to read templates file: %w", err)
	}
	
	// Log the first part of the file for debugging
	previewLen := min(200, len(data))
	log.Printf("DEBUG: File content preview: %s", string(data[:previewLen]))
	
	// Try to parse as direct array first
	var templates []Template
	err = json.Unmarshal(data, &templates)
	if err == nil {
		log.Printf("INFO: Successfully parsed JSON as direct array of %d templates", len(templates))
		r.data = &TemplateData{
			Templates: templates,
		}
		if len(templates) > 0 {
			r.data.DefaultTemplate = templates[0]
			log.Printf("INFO: Set first template as default template: %s", templates[0].ID)
		}
	} else {
		log.Printf("INFO: Failed to parse as direct array, trying as TemplateData: %v", err)
		
		// Try to parse as TemplateData
		var templateData TemplateData
		if err := json.Unmarshal(data, &templateData); err != nil {
			log.Printf("ERROR: Failed to parse templates JSON: %v", err)
			return fmt.Errorf("failed to parse templates JSON: %w", err)
		}
		r.data = &templateData
		log.Printf("INFO: Successfully parsed JSON as TemplateData with %d templates", len(templateData.Templates))
	}
	
	// Validate and log template details
	if r.data == nil || len(r.data.Templates) == 0 {
		log.Printf("WARNING: No templates found in the file")
		return nil
	}
	
	log.Printf("INFO: Successfully loaded %d templates", len(r.data.Templates))
	
	// Log template IDs for debugging
	var templateIDs []string
	totalFeatures := 0
	totalTasks := 0
	
	for i, template := range r.data.Templates {
		featureCount := len(template.Features)
		taskCount := len(template.Tasks)
		totalFeatures += featureCount
		totalTasks += taskCount
		
		log.Printf("INFO: Template %d: ID=%s, Name=%s, Stack=%s, Features=%d, Tasks=%d", 
			i+1, template.ID, template.Name, template.Stack, featureCount, taskCount)
		templateIDs = append(templateIDs, template.ID)
	}
	
	log.Printf("INFO: Available template IDs: %s", strings.Join(templateIDs, ", "))
	log.Printf("INFO: Total features across all templates: %d", totalFeatures)
	log.Printf("INFO: Total tasks across all templates: %d", totalTasks)
	
	
	return nil
}

// GetAllTemplates returns all available templates
func (r *TemplateRepository) GetAllTemplates() []Template {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return r.data.Templates
}

// GetTemplateByID retrieves a template by its ID
func (r *TemplateRepository) GetTemplateByID(id string) (*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	log.Printf("DEBUG: Searching for template with ID: %s", id)
	
	// Check if data is nil
	if r.data == nil {
		log.Printf("ERROR: Template data is nil. Repository may not be initialized properly.")
		return nil, fmt.Errorf("template data is nil")
	}
	
	// Log available templates for debugging
	log.Printf("DEBUG: Available templates: %d", len(r.data.Templates))
	
	for i := range r.data.Templates {
		template := &r.data.Templates[i] // Use pointer to avoid copy
		log.Printf("DEBUG: Comparing template %d: ID=%s with requested ID=%s", i, template.ID, id)
		if template.ID == id {
			log.Printf("DEBUG: Found matching template: %s with %d features and %d tasks", 
				template.Name, len(template.Features), len(template.Tasks))
			return template, nil
		}
	}
	
	log.Printf("DEBUG: Template with ID=%s not found, returning default template", id)
	// Return default template if not found
	return &r.data.DefaultTemplate, nil
}

// GetTemplateByStack retrieves a template by its stack name
func (r *TemplateRepository) GetTemplateByStack(stack string) (*Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, template := range r.data.Templates {
		if template.Stack == stack {
			return &template, nil
		}
	}
	
	return nil, fmt.Errorf("template not found for stack: %s", stack)
}

// GetAvailableStacks returns all unique stacks
func (r *TemplateRepository) GetAvailableStacks() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stackMap := make(map[string]bool)
	for _, template := range r.data.Templates {
		stackMap[template.Stack] = true
	}
	
	var stacks []string
	for stack := range stackMap {
		stacks = append(stacks, stack)
	}
	
	return stacks
}

// AddTemplate adds a new template or updates existing one
func (r *TemplateRepository) AddTemplate(template Template) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if template already exists
	for i, existing := range r.data.Templates {
		if existing.ID == template.ID {
			// Update existing
			r.data.Templates[i] = template
			return r.saveData()
		}
	}
	
	// Add new template
	r.data.Templates = append(r.data.Templates, template)
	return r.saveData()
}

// DeleteTemplate removes a template by ID
func (r *TemplateRepository) DeleteTemplate(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	for i, template := range r.data.Templates {
		if template.ID == id {
			// Remove the template
			r.data.Templates = append(r.data.Templates[:i], r.data.Templates[i+1:]...)
			return r.saveData()
		}
	}
	
	return fmt.Errorf("template not found with ID: %s", id)
}

// saveData writes the current data back to the JSON file
func (r *TemplateRepository) saveData() error {
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal template data: %w", err)
	}
	
	fullPath := filepath.Join(r.dataPath, "templates.json")
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write templates file: %w", err)
	}
	
	return nil
}

// ReloadData reloads template data from file
func (r *TemplateRepository) ReloadData() error {
	return r.LoadData()
}

// SetDefaultTemplate sets the default template for the repository
func (r *TemplateRepository) SetDefaultTemplate(template Template) {
	if r.data == nil {
		r.data = &TemplateData{
			Templates: []Template{},
		}
	}
	r.data.DefaultTemplate = template
}

// GetTemplatesByTechStack returns all templates for a specific tech stack
func (r *TemplateRepository) GetTemplatesByTechStack(techStack string) []Template {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []Template
	for _, template := range r.data.Templates {
		if template.TechStack == techStack {
			result = append(result, template)
		}
	}
	
	return result
}
