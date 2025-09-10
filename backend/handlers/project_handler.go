package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	repo *repositories.ProjectRepository
}

func NewProjectHandler(repo *repositories.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

// ShowProjectCreateModal renders the create project modal for HTMX
func (h *ProjectHandler) ShowProjectCreateModal(c *gin.Context) {
	c.HTML(http.StatusOK, "create_project.html", gin.H{})
}

// CreateProjectFromForm handles the project creation from the HTMX modal form
func (h *ProjectHandler) CreateProjectFromForm(c *gin.Context) {
	// Get the current user ID from the authenticated session
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusUnauthorized, "create_project.html", gin.H{
			"error": "You must be logged in to create a project",
		})
		return
	}

	// Parse form data from the modal
	name := c.PostForm("name")
	description := c.PostForm("description")
	taskTypes := c.PostForm("task_types")
	featureCategories := c.PostForm("feature_categories")
	tagsEnabled := c.PostForm("tags_enabled") == "on" // Checkbox returns "on" when checked

	// Basic validation
	if name == "" {
		c.HTML(http.StatusBadRequest, "create_project.html", gin.H{
			"error":       "Project name is required",
			"name":        name,
			"description": description,
		})
		return
	}

	// Create new project model
	// Convert userID from uint to int for the Project model
	userIDUint, ok := userID.(uint)
	if !ok {
		c.HTML(http.StatusInternalServerError, "create_project.html", gin.H{
			"error":       "Invalid user ID type",
			"name":        name,
			"description": description,
		})
		return
	}

	// Process task types and feature categories
	var taskTypesList []string
	var featureCategoriesList []string

	// Parse task types
	if taskTypes != "" {
		// Split by comma and trim spaces
		for _, t := range strings.Split(taskTypes, ",") {
			taskTypesList = append(taskTypesList, strings.TrimSpace(t))
		}
	} else {
		// Default task types
		taskTypesList = []string{"UI", "Dev", "DB", "Backend"}
	}

	// Parse feature categories
	if featureCategories != "" {
		// Split by comma and trim spaces
		for _, c := range strings.Split(featureCategories, ",") {
			featureCategoriesList = append(featureCategoriesList, strings.TrimSpace(c))
		}
	} else {
		// Default feature categories
		featureCategoriesList = []string{"Auth", "Payment", "Tags", "Tasks", "Features"}
	}

	// Create custom config
	customConfig := models.JSONB{
		"task_types":       taskTypesList,
		"feature_category": featureCategoriesList,
		"tags_enabled":     tagsEnabled,
	}

	project := models.Project{
		Name:        name,
		Description: description,
		OwnerID:     int(userIDUint), // Convert uint to int
		Config:      customConfig,
	}

	// Save to database using your existing repository
	if err := h.repo.CreateProject(&project); err != nil {
		c.HTML(http.StatusInternalServerError, "create_project.html", gin.H{
			"error":       "Failed to create project",
			"name":        name,
			"description": description,
		})
		return
	}

	// Get all projects to refresh the list
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "modals/create_project.html", gin.H{
			"error": "Project created but failed to refresh project list",
		})
		return
	}

	// Return two fragments: 
	// 1. Close the modal
	// 2. Update the project list
	c.Header("HX-Trigger", "projectCreated")
	c.HTML(http.StatusOK, "project-list-fragment.html", gin.H{
		"Projects": projects,
		"NewProject": project,
	})
}

// CreateProject handles project creation (for APIs expecting JSON)
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if project.OwnerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}

	if err := h.repo.CreateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetAllProjects handles getting all projects
func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	log.Println("DEBUG: Getting all projects from repository")
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		log.Printf("ERROR: Failed to get projects: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d projects from repository\n", len(projects))
	for i, p := range projects {
		log.Printf("DEBUG: Project %d: ID=%d, Name=%s, OwnerID=%d\n", i+1, p.ID, p.Name, p.OwnerID)
	}
	c.HTML(http.StatusOK, "project-list.html", gin.H{"Projects": projects})
}

// GetProject handles getting a single project
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.repo.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.HTML(http.StatusOK, "project-detail.html", gin.H{"Project": project})
}

// UpdateProject handles project updates
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = projectID
	if err := h.repo.UpdateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles project deletion
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	if err := h.repo.DeleteProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectsByUser handles getting projects for a specific user
func (h *ProjectHandler) GetProjectsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	projects, err := h.repo.GetProjectsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) ShowDashboard(c *gin.Context) {
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"Projects": projects})
}

// GetAllProjectsFragment handles getting all projects for HTMX fragment requests
func (h *ProjectHandler) GetAllProjectsFragment(c *gin.Context) {
	log.Println("DEBUG: Getting all projects for fragment")

	// Get the user's role from the context (set by the AuthMiddleware)
	userRole, _ := c.Get("user_role")
	isManager := userRole == "manager"

	projects, err := h.repo.GetAllProjects()
	if err != nil {
		log.Printf("ERROR: Failed to get projects for fragment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d projects for fragment\n", len(projects))

	// Pass both the Projects and the CurrentUser's role to the template
	c.HTML(http.StatusOK, "project-list.html", gin.H{
		"Projects": projects,
		"CurrentUser": gin.H{
			"IsManager": isManager,
		},
	})
}
