package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
)

// WebPRHandler handles web UI PR-related requests
type WebPRHandler struct {
	prRepo repositories.PullRequestRepository
}

// NewWebPRHandler creates a new WebPRHandler
func NewWebPRHandler(prRepo repositories.PullRequestRepository) *WebPRHandler {
	return &WebPRHandler{prRepo: prRepo}
}

// GetPRForm returns the PR form modal for a specific task
func (h *WebPRHandler) GetPRForm(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Render the PR form modal template
	c.HTML(http.StatusOK, "pr-modal.html", gin.H{
		"TaskID": taskID,
	})
}

// AddTaskPR handles adding a PR to a task from the web UI
func (h *WebPRHandler) AddTaskPR(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}
	
	// Get the task to get its FeatureID
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	
	var task models.Task
	if err := db.DB.First(&task, taskID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task not found"})
		return
	}
	
	// Parse form data
	url := c.PostForm("url")
	title := c.PostForm("title")
	branch := c.PostForm("branch")
	description := c.PostForm("description")
	
	// Create the PR
	pr := models.PullRequest{
		TaskID:      uint(taskID),
		FeatureID:   task.FeatureID,
		URL:         url,
		Title:       title,
		Branch:      branch,
		Description: description,
		Status:      string(featureplus.StatusOpen),
		Tested:      false,
		Version:     "", // Can be empty for web UI submissions
	}
	
	if err := h.prRepo.Create(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pull request"})
		return
	}
	
	// Get user role from context (set by RoleMiddleware)
	userRole, exists := c.Get("user_role")
	var isManager bool
	
	if exists {
		isManager = userRole == "manager"
		log.Printf("[AddTaskPR] user_role from context: '%v', isManager: %v", userRole, isManager)
	}
	
	// Get all PRs for this task to refresh the table
	prs, err := h.prRepo.GetByTaskID(uint(taskID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PRs"})
		return
	}
	
	// Return an empty div to close the modal and the updated PR table
	c.HTML(http.StatusOK, "_pr_form_response.html", gin.H{
		"TaskID":      taskID,
		"PullRequests": prs,
		"CurrentUser":  gin.H{"Role": userRole, "IsManager": isManager},
	})
}

// GetAllPRsFragment renders all pull requests as an HTMX fragment for the dashboard
func (h *WebPRHandler) GetAllPRsFragment(c *gin.Context) {
	log.Println("DEBUG: Getting all PRs for fragment")
	
	// Get the user's role from the context (set by the AuthMiddleware)
	userRole, _ := c.Get("user_role")
	isManager := userRole == "manager"
	
	// Get all PRs
	prs, err := h.prRepo.GetAll()
	if err != nil {
		log.Printf("ERROR: Failed to get PRs for fragment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d PRs for fragment\n", len(prs))
	
	// Pass both the PRs and the CurrentUser's role to the template
	c.HTML(http.StatusOK, "all-prs-list.html", gin.H{
		"PullRequests": prs,
		"CurrentUser": gin.H{
			"IsManager": isManager,
			"Role": userRole,
		},
	})
}
