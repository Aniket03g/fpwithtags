package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/internal/mcpbridge"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ImportHandler handles GitHub MCP-imported project requests
type ImportHandler struct {
	db          *gorm.DB
	importRepo  *repositories.ImportRepository
	projectRepo *repositories.ProjectRepository
	featureRepo *repositories.FeatureRepository
	taskRepo    repositories.TaskRepository
	mcpService  *services.GitHubMCPService
}

// NewImportHandler creates a new import handler
func NewImportHandler(db *gorm.DB, dataPath string) *ImportHandler {
	return &ImportHandler{
		db:          db,
		importRepo:  repositories.NewImportRepository(dataPath),
		projectRepo: repositories.NewProjectRepository(db),
		featureRepo: repositories.NewFeatureRepository(db),
		taskRepo:    repositories.NewTaskRepository(db),
		mcpService:  services.NewGitHubMCPService(),
	}
}

// ImportProject handles dynamic import from GitHub MCP-generated JSON
// POST /api/imports/import
func (h *ImportHandler) ImportProject(c *gin.Context) {
	var request struct {
		ProjectID   string `json:"project_id" form:"project_id" binding:"required"`
		ProjectName string `json:"project_name" form:"project_name" binding:"required"`
		Description string `json:"description" form:"description"`
	}

	// Try to bind as form data first (HTMX), then JSON (API)
	contentType := c.GetHeader("Content-Type")
	var err error
	
	if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" || c.GetHeader("HX-Request") == "true" {
		// HTMX form submission
		err = c.ShouldBind(&request)
	} else {
		// JSON API request
		err = c.ShouldBindJSON(&request)
	}
	
	if err != nil {
		log.Printf("ERROR: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Importing project from template: %s (name: %s)", request.ProjectID, request.ProjectName)

	// Dynamically load template (NOT from cache)
	template, err := h.importRepo.LoadImportTemplate(request.ProjectID)
	if err != nil {
		log.Printf("ERROR: Failed to load import template: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Import template not found: %s", request.ProjectID),
			"error":   err.Error(),
		})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("ERROR: User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Authentication required",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("ERROR: Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Invalid user ID type",
		})
		return
	}

	// Create project with imported template data
	project := models.Project{
		Name:        request.ProjectName,
		Description: request.Description,
		OwnerID:     int(userIDUint),
		Config: models.JSONB{
			"tech_stack":       template.TechStack,
			"feature_category": template.FeatureCategories,
			"task_types":       template.TaskTypes,
			"imported_from":    request.ProjectID,
			"import_source":    "github_mcp",
			"template_id":      template.ID,
		},
	}

	log.Printf("INFO: Creating project: %s (owner: %d)", project.Name, project.OwnerID)

	if err := h.projectRepo.CreateProject(&project); err != nil {
		log.Printf("ERROR: Failed to create project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create project",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Project created successfully with ID: %d", project.ID)

	// Apply imported template (create features and tasks)
	featuresCreated, tasksCreated, err := h.applyImportedTemplate(&project, template)
	if err != nil {
		log.Printf("ERROR: Failed to apply template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Project created but failed to apply template",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Successfully imported project %d: %d features, %d tasks", 
		project.ID, featuresCreated, tasksCreated)

	// Check if this is an HTMX request (from web UI)
	if c.GetHeader("HX-Request") == "true" {
		// Return updated project list for HTMX
		projectList, err := h.projectRepo.GetAllProjects()
		if err != nil {
			log.Printf("ERROR: Failed to get updated project list: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Project imported but failed to refresh list",
			})
			return
		}

		c.HTML(http.StatusOK, "project-list-fragment.html", gin.H{
			"Projects":   projectList,
			"NewProject": project,
		})
		return
	}

	// Return JSON for API requests
	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          "Project imported successfully",
		"project_id":       project.ID,
		"project_name":     project.Name,
		"features_created": featuresCreated,
		"tasks_created":    tasksCreated,
	})
}

// applyImportedTemplate creates features and tasks from the imported template
// Reuses logic similar to handlers/apply_template.go
func (h *ImportHandler) applyImportedTemplate(project *models.Project, template *repositories.Template) (int, int, error) {
	log.Printf("INFO: Applying imported template '%s' to project %d", template.Name, project.ID)

	// Get project context for filtering (default to Development)
	projectContext := "Development"
	if project.Config != nil {
		if ctx, ok := project.Config["context"].(string); ok && ctx != "" {
			projectContext = ctx
		}
	}

	// Filter features by context
	var filteredFeatures []repositories.TemplateFeature
	for _, feature := range template.Features {
		featureContext := feature.Context
		if featureContext == "" {
			featureContext = "Development"
		}

		// Include feature if it matches project context
		if featureContext == projectContext {
			filteredFeatures = append(filteredFeatures, feature)
		}
	}

	// Fallback: If no features match, use Development context features
	if len(filteredFeatures) == 0 && projectContext != "Development" {
		log.Printf("INFO: No features found for context '%s', falling back to Development", projectContext)
		for _, feature := range template.Features {
			featureContext := feature.Context
			if featureContext == "" || featureContext == "Development" {
				filteredFeatures = append(filteredFeatures, feature)
			}
		}
	}

	// If still no features, use all features
	if len(filteredFeatures) == 0 {
		log.Printf("WARNING: No context-specific features found, using all features")
		filteredFeatures = template.Features
	}

	log.Printf("INFO: Creating %d/%d features (context: %s)", len(filteredFeatures), len(template.Features), projectContext)

	// Create features from template
	createdFeatures := []models.Feature{}
	featureMap := make(map[string]uint) // Map feature names to IDs

	for i, templateFeature := range filteredFeatures {
		log.Printf("DEBUG: Creating feature %d/%d: %s", i+1, len(filteredFeatures), templateFeature.Name)

		feature := models.Feature{
			Title:       templateFeature.Name,
			Description: templateFeature.Description,
			Category:    templateFeature.Category,
			ProjectID:   int(project.ID),
			Status:      models.StatusTodo,
			Priority:    models.PriorityMedium,
		}

		if err := h.featureRepo.CreateFeature(&feature); err != nil {
			log.Printf("ERROR: Failed to create feature %s: %v", templateFeature.Name, err)
			continue
		}

		log.Printf("DEBUG: Successfully created feature with ID: %d", feature.ID)
		createdFeatures = append(createdFeatures, feature)
		featureMap[templateFeature.Name] = feature.ID
	}

	log.Printf("INFO: Created %d/%d features successfully", len(createdFeatures), len(filteredFeatures))

	// Filter tasks by context
	var filteredTasks []repositories.TemplateTask
	for _, task := range template.Tasks {
		taskContext := task.Context
		if taskContext == "" {
			taskContext = "Development"
		}

		// Include task if it matches project context
		if taskContext == projectContext {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Fallback: If no tasks match, use Development context tasks
	if len(filteredTasks) == 0 && projectContext != "Development" {
		log.Printf("INFO: No tasks found for context '%s', falling back to Development", projectContext)
		for _, task := range template.Tasks {
			taskContext := task.Context
			if taskContext == "" || taskContext == "Development" {
				filteredTasks = append(filteredTasks, task)
			}
		}
	}

	// If still no tasks, use all tasks
	if len(filteredTasks) == 0 {
		log.Printf("WARNING: No context-specific tasks found, using all tasks")
		filteredTasks = template.Tasks
	}

	log.Printf("INFO: Creating %d/%d tasks (context: %s)", len(filteredTasks), len(template.Tasks), projectContext)

	// Create tasks from template with feature assignment
	createdTasks := []models.Task{}

	// Create a map of feature categories for quick lookup
	featureCategories := make(map[string][]models.Feature)
	for _, feature := range createdFeatures {
		featureCategories[feature.Category] = append(featureCategories[feature.Category], feature)
	}

	for i, templateTask := range filteredTasks {
		log.Printf("DEBUG: Creating task %d/%d: %s (Type: %s)", i+1, len(filteredTasks), templateTask.Name, templateTask.Type)

		// Find the appropriate feature for this task based on type matching
		var featureID uint
		if len(createdFeatures) > 0 {
			// Try exact match first
			if features, exists := featureCategories[templateTask.Type]; exists && len(features) > 0 {
				featureID = features[0].ID
				log.Printf("DEBUG: Exact match - Task type '%s' with feature category '%s'", templateTask.Type, features[0].Category)
			} else {
				// Try case-insensitive match
				for _, feature := range createdFeatures {
					if strings.EqualFold(feature.Category, templateTask.Type) {
						featureID = feature.ID
						log.Printf("DEBUG: Case-insensitive match - Task type '%s' with feature category '%s'", templateTask.Type, feature.Category)
						break
					}
				}

				// If still no match, use the first feature
				if featureID == 0 {
					featureID = createdFeatures[0].ID
					log.Printf("DEBUG: No match found, assigning task to first feature (ID: %d)", featureID)
				}
			}
		} else {
			log.Printf("WARNING: No features available to assign task to")
		}

		task := models.Task{
			TaskName:    templateTask.Name,
			Description: templateTask.Description,
			TaskType:    templateTask.Type,
			FeatureID:   featureID,
		}

		if err := h.taskRepo.Create(&task); err != nil {
			log.Printf("ERROR: Failed to create task %s: %v", templateTask.Name, err)
			continue
		}

		log.Printf("DEBUG: Successfully created task with ID: %d", task.ID)
		createdTasks = append(createdTasks, task)
	}

	log.Printf("INFO: Created %d/%d tasks successfully", len(createdTasks), len(filteredTasks))

	// Create dependencies if we have a dependency repository
	if len(template.Dependencies) > 0 && h.db != nil {
		dependencyService := services.NewDependencyService(repositories.NewDependencyRepository(h.db))
		createdDependencies := 0

		for _, dep := range template.Dependencies {
			// Create a simple dependency record
			dependency := models.Dependency{
				Description: fmt.Sprintf("Template dependency: %s", dep),
				ParentType:  models.EntityTypeFeature,
				ChildType:   models.EntityTypeFeature,
			}

			// Try to find matching features
			if len(createdFeatures) > 0 {
				dependency.ParentID = createdFeatures[0].ID

				if len(createdFeatures) > 1 {
					dependency.ChildID = createdFeatures[1].ID
				} else {
					dependency.ChildID = createdFeatures[0].ID
				}

				if err := dependencyService.CreateDependency(&dependency); err != nil {
					log.Printf("ERROR: Failed to create dependency: %v", err)
					continue
				}

				createdDependencies++
			}
		}

		log.Printf("INFO: Created %d dependencies", createdDependencies)
	}

	return len(createdFeatures), len(createdTasks), nil
}

// ListAvailableImports returns all available GitHub import templates
// GET /api/imports
func (h *ImportHandler) ListAvailableImports(c *gin.Context) {
	imports, err := h.importRepo.ListAvailableImports()
	if err != nil {
		log.Printf("ERROR: Failed to list imports: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list imports",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"imports": imports,
		"count":   len(imports),
	})
}

// SaveImportTemplate saves a new import template
// POST /api/imports/save
func (h *ImportHandler) SaveImportTemplate(c *gin.Context) {
	var request struct {
		ProjectID string                    `json:"project_id" binding:"required"`
		Template  repositories.Template     `json:"template" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("ERROR: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if err := h.importRepo.SaveImportTemplate(request.ProjectID, &request.Template); err != nil {
		log.Printf("ERROR: Failed to save import template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to save import template",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Import template saved successfully",
	})
}

// DeleteImportTemplate deletes an import template
// DELETE /api/imports/:id
func (h *ImportHandler) DeleteImportTemplate(c *gin.Context) {
	projectID := c.Param("id")

	if err := h.importRepo.DeleteImportTemplate(projectID); err != nil {
		log.Printf("ERROR: Failed to delete import template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete import template",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Import template deleted successfully",
	})
}

// ImportFromGitHub handles automatic import from GitHub repository using MCP
// POST /api/imports/github
func (h *ImportHandler) ImportFromGitHub(c *gin.Context) {
	// Step 1: Parse the request body to get the GitHub repository URL
	var request struct {
		RepoURL     string `json:"repo_url" form:"repo_url" binding:"required"`
		ProjectName string `json:"project_name" form:"project_name"` // Optional: override auto-generated name
	}

	// Try to bind as form data first (HTMX), then JSON (API)
	contentType := c.GetHeader("Content-Type")
	var err error
	
	if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" || c.GetHeader("HX-Request") == "true" {
		// HTMX form submission
		err = c.ShouldBind(&request)
	} else {
		// JSON API request
		err = c.ShouldBindJSON(&request)
	}
	
	if err != nil {
		log.Printf("ERROR: Invalid request body for GitHub import: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request body. Please provide a valid GitHub repository URL",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Starting GitHub MCP import for repository: %s", request.RepoURL)

	// Step 2: Validate the repository URL format
	if !strings.Contains(request.RepoURL, "github.com") {
		log.Printf("ERROR: Invalid GitHub URL: %s", request.RepoURL)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid GitHub URL. Must be a github.com repository",
		})
		return
	}

	// Step 3: Call local MCP Bridge to analyze the repository
	// REMOVED: direct call to GitHubMCPService.AnalyzeRepository — replaced with local MCP Bridge call
	log.Printf("INFO: Calling local MCP Bridge to analyze repository...")
	
	// Call the MCP Bridge
	responseJSON, err := mcpbridge.CallLocalMCPBridge(request.RepoURL)
	if err != nil {
		log.Printf("ERROR: MCP Bridge analysis failed: %v", err)
		
		// Determine appropriate error response
		statusCode := http.StatusInternalServerError
		message := "Failed to analyze GitHub repository"
		
		// Check for specific error types
		if strings.Contains(err.Error(), "unavailable") {
			statusCode = http.StatusBadGateway
			message = "MCP Bridge unavailable. Please ensure the bridge service is running."
		} else if strings.Contains(err.Error(), "rate limited") {
			statusCode = http.StatusTooManyRequests
			message = "Analysis already in progress. Please try again later."
		} else if strings.Contains(err.Error(), "timeout") {
			statusCode = http.StatusGatewayTimeout
			message = "Repository analysis timed out. Please try again."
		}
		
		c.JSON(statusCode, gin.H{
			"status":  "error",
			"message": message,
			"error":   err.Error(),
		})
		return
	}
	
	// Parse the response JSON into a Template
	var template repositories.Template
	if err := json.Unmarshal(responseJSON, &template); err != nil {
		log.Printf("ERROR: Failed to parse MCP Bridge response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to parse analysis results",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: MCP Bridge analysis successful - Template ID: %s, Features: %d, Tasks: %d", 
		template.ID, len(template.Features), len(template.Tasks))

	// Step 4: Generate a unique project ID from the template
	projectID := template.ID
	if projectID == "" {
		// Fallback: extract from URL
		parts := strings.Split(strings.TrimSuffix(request.RepoURL, "/"), "/")
		projectID = parts[len(parts)-1]
	}
	
	// Make it unique by adding timestamp suffix
	projectID = fmt.Sprintf("%s_%d", projectID, time.Now().Unix())
	log.Printf("INFO: Generated project ID: %s", projectID)

	// Step 5: Save the MCP-generated template to the imports directory
	// This allows for inspection and reuse of the generated template
	log.Printf("INFO: Saving MCP template to imports directory...")
	if err := h.importRepo.SaveImportTemplate(projectID, &template); err != nil {
		log.Printf("ERROR: Failed to save MCP template: %v", err)
		// Continue anyway - we can still import without saving
	} else {
		log.Printf("INFO: MCP template saved successfully: %s.json", projectID)
	}

	// Step 6: Get user ID from authentication context
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("ERROR: User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "Authentication required",
		})
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("ERROR: Invalid user ID type")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Invalid user ID type",
		})
		return
	}

	// Step 7: Determine the project name (use provided name or template name)
	finalProjectName := request.ProjectName
	if finalProjectName == "" {
		finalProjectName = template.Name
	}
	log.Printf("INFO: Creating project with name: %s", finalProjectName)

	// Step 8: Create the project in the database
	project := models.Project{
		Name:        finalProjectName,
		Description: template.Description,
		OwnerID:     int(userIDUint),
		Config: models.JSONB{
			"tech_stack":       template.TechStack,
			"feature_category": template.FeatureCategories,
			"task_types":       template.TaskTypes,
			"imported_from":    request.RepoURL,
			"import_source":    "github_mcp",
			"template_id":      template.ID,
			"mcp_generated":    true,
		},
	}

	if err := h.projectRepo.CreateProject(&project); err != nil {
		log.Printf("ERROR: Failed to create project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create project from GitHub repository",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Project created successfully with ID: %d", project.ID)

	// Step 9: Apply the template to create features and tasks
	// Reuse existing logic from ImportProject
	featuresCreated, tasksCreated, err := h.applyImportedTemplate(&project, &template)
	if err != nil {
		log.Printf("ERROR: Failed to apply MCP template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Project created but failed to apply template",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("INFO: Successfully imported GitHub project %d: %d features, %d tasks", 
		project.ID, featuresCreated, tasksCreated)

	// Step 10: Return response based on request type (HTMX or API)
	// Check if this is an HTMX request (from web UI)
	if c.GetHeader("HX-Request") == "true" {
		// Return updated project list for HTMX
		projectList, err := h.projectRepo.GetAllProjects()
		if err != nil {
			log.Printf("ERROR: Failed to get updated project list: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Project imported but failed to refresh list",
			})
			return
		}

		c.HTML(http.StatusOK, "project-list-fragment.html", gin.H{
			"Projects":   projectList,
			"NewProject": project,
		})
		return
	}

	// Return JSON for API requests
	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          "GitHub repository imported successfully via MCP",
		"project_id":       project.ID,
		"project_name":     project.Name,
		"features_created": featuresCreated,
		"tasks_created":    tasksCreated,
		"repo_url":         request.RepoURL,
		"template_saved":   projectID,
	})
}
