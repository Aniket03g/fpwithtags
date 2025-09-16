package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateHandler handles template-related requests
type TemplateHandler struct {
	DB               *gorm.DB
	templateRepo     *repositories.TemplateRepository
	guidanceRepo     *repositories.GuidanceRepository
	projectRepo      *repositories.ProjectRepository
	featureRepo      *repositories.FeatureRepository
	taskRepo         repositories.TaskRepository
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(db *gorm.DB, dataPath string) *TemplateHandler {
	// Convert to absolute path if needed
	if !filepath.IsAbs(dataPath) {
		cwd, err := os.Getwd()
		if err == nil {
			dataPath = filepath.Join(cwd, dataPath)
			log.Printf("INFO: Using absolute data path: %s", dataPath)
		}
	}
	templateRepo, err := repositories.NewTemplateRepository(dataPath)
	if err != nil {
		log.Printf("Warning: Failed to load template repository: %v. Creating fallback repository.", err)
		// Create a fallback repository with empty data
		templateRepo = createFallbackTemplateRepo()
	}
	
	guidanceRepo, err := repositories.NewGuidanceRepository(dataPath)
	if err != nil {
		log.Printf("Warning: Failed to load guidance repository: %v. Creating fallback repository.", err)
		// Create a fallback repository with empty data
		guidanceRepo = createTemplateGuidanceRepo()
	}
	
	return &TemplateHandler{
		DB:           db,
		templateRepo: templateRepo,
		guidanceRepo: guidanceRepo,
		projectRepo:  repositories.NewProjectRepository(db),
		featureRepo:  repositories.NewFeatureRepository(db),
		taskRepo:     repositories.NewTaskRepository(db),
	}
}

// GetAllTemplates returns all available templates
func (h *TemplateHandler) GetAllTemplates(c *gin.Context) {
	templates := h.templateRepo.GetAllTemplates()
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// GetTemplateByID returns a specific template by ID
func (h *TemplateHandler) GetTemplateByID(c *gin.Context) {
	templateID := c.Param("id")
	
	template, err := h.templateRepo.GetTemplateByID(templateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	// Enrich tasks with guidance from the guidance repository
	enrichedTasks := h.enrichTasksWithGuidance(template)
	
	c.JSON(http.StatusOK, gin.H{
		"template": template,
		"enriched_tasks": enrichedTasks,
	})
}

// GetTemplateByStack returns a template for a specific stack
func (h *TemplateHandler) GetTemplateByStack(c *gin.Context) {
	stack := c.Param("stack")
	
	template, err := h.templateRepo.GetTemplateByStack(stack)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found for stack"})
		return
	}
	
	c.JSON(http.StatusOK, template)
}

// GetTemplateDetails returns template details for HTMX
func (h *TemplateHandler) GetTemplateDetails(c *gin.Context) {
	templateID := c.Param("id")
	
	template, err := h.templateRepo.GetTemplateByID(templateID)
	if err != nil {
		c.String(http.StatusNotFound, "Template not found")
		return
	}
	
	// Enrich tasks with guidance
	enrichedTasks := h.enrichTasksWithGuidance(template)
	
	// Render the template details fragment
	c.HTML(http.StatusOK, "template-details-fragment.html", gin.H{
		"Template":      template,
		"EnrichedTasks": enrichedTasks,
	})
}

// ApplyTemplate applies a template to a project
func (h *TemplateHandler) ApplyTemplate(c *gin.Context) {
	var request struct {
		ProjectID  int    `json:"project_id"`
		TemplateID string `json:"template_id"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get the template
	template, err := h.templateRepo.GetTemplateByID(request.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}
	
	// Get the project
	project, err := h.projectRepo.GetProjectByID(request.ProjectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Update project tech stack
	if project.Config == nil {
		project.Config = models.JSONB{}
	}
	project.Config["tech_stack"] = template.TechStack
	project.Config["template_id"] = template.ID
	
	// Save project updates
	if err := h.projectRepo.UpdateProject(project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}
	
	// Create features from template
	createdFeatures := []models.Feature{}
	for _, templateFeature := range template.Features {
		feature := models.Feature{
			Title:       templateFeature.Name,
			Description: templateFeature.Description,
			Category:    templateFeature.Category,
			ProjectID:   request.ProjectID,
			Status:      "pending",
			Priority:    models.PriorityMedium,
		}
		
		if err := h.featureRepo.CreateFeature(&feature); err != nil {
			log.Printf("Failed to create feature %s: %v", templateFeature.Name, err)
			continue
		}
		createdFeatures = append(createdFeatures, feature)
	}
	
	// Create tasks from template
	createdTasks := []models.Task{}
	for _, templateTask := range template.Tasks {
		// Find the appropriate feature for this task
		var featureID uint
		if len(createdFeatures) > 0 {
			// Simple assignment - you might want more sophisticated logic
			featureID = createdFeatures[0].ID
		}
		
		task := models.Task{
			TaskName:    templateTask.Name,
			Description: templateTask.Description,
			TaskType:    templateTask.Type,
			FeatureID:   featureID,
		}
		
		if err := h.taskRepo.Create(&task); err != nil {
			log.Printf("Failed to create task %s: %v", templateTask.Name, err)
			continue
		}
		createdTasks = append(createdTasks, task)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":          "Template applied successfully",
		"project":          project,
		"created_features": len(createdFeatures),
		"created_tasks":    len(createdTasks),
	})
}

// AddTemplate adds or updates a template (admin only)
func (h *TemplateHandler) AddTemplate(c *gin.Context) {
	var template repositories.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.templateRepo.AddTemplate(template); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add template"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Template added successfully"})
}

// DeleteTemplate removes a template (admin only)
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	templateID := c.Param("id")
	
	if err := h.templateRepo.DeleteTemplate(templateID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

// ReloadTemplates reloads template data from file
func (h *TemplateHandler) ReloadTemplates(c *gin.Context) {
	if err := h.templateRepo.ReloadData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload template data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Template data reloaded successfully"})
}

// GetAvailableStacks returns all available stacks
func (h *TemplateHandler) GetAvailableStacks(c *gin.Context) {
	stacks := h.templateRepo.GetAvailableStacks()
	c.JSON(http.StatusOK, gin.H{"stacks": stacks})
}

// enrichTasksWithGuidance enriches template tasks with guidance from the guidance repository
func (h *TemplateHandler) enrichTasksWithGuidance(template *repositories.Template) []map[string]interface{} {
	enrichedTasks := []map[string]interface{}{}
	
	for _, task := range template.Tasks {
		// Get guidance for this task type and tech stack
		guidance := h.guidanceRepo.GetGuidance(template.TechStack, task.Type)
		
		enrichedTask := map[string]interface{}{
			"task":     task,
			"guidance": guidance,
		}
		
		enrichedTasks = append(enrichedTasks, enrichedTask)
	}
	
	return enrichedTasks
}
