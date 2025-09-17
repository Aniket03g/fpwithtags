package repositories

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
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
}

// TemplateTask represents a task in a template
type TemplateTask struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
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

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(dataPath string) (*TemplateRepository, error) {
	// Log the data path for debugging
	log.Printf("INFO: Creating template repository with data path: %s", dataPath)
	
	repo := &TemplateRepository{
		dataPath: dataPath,
	}
	
	// Load initial data
	if err := repo.LoadData(); err != nil {
		log.Printf("ERROR: Failed to load template data: %v", err)
		return nil, err
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
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("ERROR: Failed to read templates file: %v", err)
		return fmt.Errorf("failed to read templates file: %w", err)
	}
	
	// Parse JSON
	var templateData TemplateData
	if err := json.Unmarshal(data, &templateData); err != nil {
		log.Printf("ERROR: Failed to parse templates JSON: %v", err)
		return fmt.Errorf("failed to parse templates JSON: %w", err)
	}
	
	r.data = &templateData
	log.Printf("DEBUG: Successfully loaded %d templates", len(templateData.Templates))
	
	// Log template IDs for debugging
	for i, template := range templateData.Templates {
		log.Printf("DEBUG: Template %d: ID=%s, Name=%s, Features=%d, Tasks=%d", 
			i+1, template.ID, template.Name, len(template.Features), len(template.Tasks))
	}
	
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
	if err := ioutil.WriteFile(fullPath, data, 0644); err != nil {
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
