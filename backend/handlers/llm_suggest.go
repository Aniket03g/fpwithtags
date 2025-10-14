package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LLMSuggestHandler handles LLM-based feature suggestion requests
type LLMSuggestHandler struct {
	projectRepo *repositories.ProjectRepository
	llmService  *services.LLMService
}

// NewLLMSuggestHandler creates a new LLM suggest handler
func NewLLMSuggestHandler(db *gorm.DB) *LLMSuggestHandler {
	return &LLMSuggestHandler{
		projectRepo: repositories.NewProjectRepository(db),
		llmService:  services.NewLLMService(),
	}
}

// SuggestFeatures generates feature suggestions based on project context
// POST /api/projects/:id/suggest
func (h *LLMSuggestHandler) SuggestFeatures(c *gin.Context) {
	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid project ID: %s", projectIDStr)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	log.Printf("INFO: Generating feature suggestions for project ID: %d", projectID)

	// Fetch project from database
	project, err := h.projectRepo.GetProjectByID(projectID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch project %d: %v", projectID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Project not found",
		})
		return
	}

	// Extract project context and tech stack from config
	projectContext := ""
	techStack := "General"

	if project.Config != nil {
		// Extract project_context (Stage 3 field)
		if ctx, ok := project.Config["project_context"].(string); ok && ctx != "" {
			projectContext = ctx
		}

		// Extract tech_stack
		if stack, ok := project.Config["tech_stack"].(string); ok && stack != "" {
			techStack = stack
		}
	}

	// Validate that project context exists
	if projectContext == "" {
		log.Printf("WARNING: Project %d has no context. Using project description as fallback.", projectID)
		projectContext = project.Description
		if projectContext == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Project has no context or description. Please add a project context to generate suggestions.",
			})
			return
		}
	}

	log.Printf("INFO: Using project context (length: %d) and tech stack: %s", len(projectContext), techStack)

	// Generate suggestions using LLM service
	suggestions, err := h.llmService.GenerateFeatureSuggestions(projectContext, techStack)
	if err != nil {
		log.Printf("ERROR: Failed to generate suggestions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate feature suggestions",
			"details": err.Error(),
		})
		return
	}

	// Return response
	response := gin.H{
		"project_id":   projectID,
		"project_name": project.Name,
		"context_used": projectContext,
		"tech_stack":   techStack,
		"suggestions":  suggestions,
		"count":        len(suggestions),
	}

	log.Printf("INFO: Successfully generated %d suggestions for project %d", len(suggestions), projectID)

	c.JSON(http.StatusOK, response)
}

// SuggestFeaturesWeb generates feature suggestions and returns HTML (for HTMX)
// POST /web/projects/:id/suggest
func (h *LLMSuggestHandler) SuggestFeaturesWeb(c *gin.Context) {
	// Get project ID from URL parameter
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid project ID: %s", projectIDStr)
		c.HTML(http.StatusBadRequest, "suggestions-error.html", gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	log.Printf("INFO: Generating feature suggestions (web) for project ID: %d", projectID)

	// Fetch project from database
	project, err := h.projectRepo.GetProjectByID(projectID)
	if err != nil {
		log.Printf("ERROR: Failed to fetch project %d: %v", projectID, err)
		c.HTML(http.StatusNotFound, "suggestions-error.html", gin.H{
			"error": "Project not found",
		})
		return
	}

	// Extract project context and tech stack from config
	projectContext := ""
	techStack := "General"

	if project.Config != nil {
		if ctx, ok := project.Config["project_context"].(string); ok && ctx != "" {
			projectContext = ctx
		}
		if stack, ok := project.Config["tech_stack"].(string); ok && stack != "" {
			techStack = stack
		}
	}

	// Validate that project context exists
	if projectContext == "" {
		log.Printf("WARNING: Project %d has no context", projectID)
		c.HTML(http.StatusBadRequest, "suggestions-error.html", gin.H{
			"error": "Project context missing. Please add a project context to generate suggestions.",
		})
		return
	}

	// Generate suggestions using LLM service
	suggestions, err := h.llmService.GenerateFeatureSuggestions(projectContext, techStack)
	if err != nil {
		log.Printf("ERROR: Failed to generate suggestions: %v", err)
		c.HTML(http.StatusInternalServerError, "suggestions-error.html", gin.H{
			"error": "Could not fetch suggestions. Try again later.",
		})
		return
	}

	// Render suggestions HTML
	c.HTML(http.StatusOK, "suggestions-list.html", gin.H{
		"ProjectID":   projectID,
		"Suggestions": suggestions,
		"Count":       len(suggestions),
	})
}
