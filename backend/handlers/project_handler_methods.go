package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
