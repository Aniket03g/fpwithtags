package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"

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

// CreateProject handles project creation
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project models.Project

	// Check if it's a form submission (HTMX) or JSON (API)
	contentType := c.GetHeader("Content-Type")
	if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
		// Form data for HTMX
		if err := c.ShouldBind(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		// JSON for API
		if err := c.ShouldBindJSON(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Get user ID from context (for web requests) or from request body (for API)
	if project.OwnerID == 0 {
		if userID, exists := c.Get("user_id"); exists {
			project.OwnerID = userID.(int)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
			return
		}
	}

	if err := h.repo.CreateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// We need to fetch the project again to get all the details, like the owner's username
		fullProject, err := h.repo.GetProjectByID(project.ID)
		if err != nil {
			log.Printf("could not fetch project after creation: %v", err)
			// Fallback or error
			c.String(http.StatusInternalServerError, "Could not render project card.")
			return
		}

		// Return HTML for HTMX
		tmpl, err := template.ParseFiles("templates/_project-card.html")
		if err != nil {
			log.Printf("template parse error: %v", err)
			c.String(http.StatusInternalServerError, "Template error")
			return
		}
		err = tmpl.ExecuteTemplate(c.Writer, "project-card", fullProject)
		if err != nil {
			log.Printf("template execute error: %v", err)
			c.String(http.StatusInternalServerError, "Render error")
		}
	} else {
		// Return JSON for API
		c.JSON(http.StatusCreated, project)
	}
}

// GetAllProjects handles getting all projects
func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, projects)
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

	c.JSON(http.StatusOK, project)
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
	username, _ := c.Get("username")
	projects, _ := h.repo.GetAllProjects() // You may want to handle errors here

	// TODO: Replace with real counts if you have methods for these
	totalProjects := len(projects)
	activeFeatures := 0
	pendingTasks := 0

	data := gin.H{
		"username":       username,
		"ActivePage":     "dashboard",
		"Projects":       projects,
		"TotalProjects":  totalProjects,
		"ActiveFeatures": activeFeatures,
		"PendingTasks":   pendingTasks,
	}
	c.HTML(http.StatusOK, "dashboard.html", data)
}

func (h *ProjectHandler) GetProjectsPage(c *gin.Context) {
	projects, err := h.repo.GetAllProjects() // Fetch all projects, not filtered by user
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to fetch projects")
		return
	}
	log.Printf("[DEBUG] Number of projects fetched: %d", len(projects))
	data := gin.H{
		"Projects":   projects,
		"ActivePage": "projects",
		"username":   "TestUser", // For now, hardcoded
	}
	c.HTML(http.StatusOK, "projects.html", data)
}

func (h *ProjectHandler) ShowNewProjectForm(c *gin.Context) {
	c.HTML(http.StatusOK, "_project-form.html", gin.H{})
}
