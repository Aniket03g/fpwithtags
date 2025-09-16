package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GuidanceHandler handles task guidance requests
type GuidanceHandler struct {
	DB           *gorm.DB
	taskRepo     repositories.TaskRepository
	guidanceRepo *repositories.GuidanceRepository
}

// NewGuidanceHandler creates a new guidance handler
func NewGuidanceHandler(db *gorm.DB, dataPath string) *GuidanceHandler {
	guidanceRepo, err := repositories.NewGuidanceRepository(dataPath)
	if err != nil {
		log.Printf("Warning: Failed to load guidance repository: %v. Using fallback guidance.", err)
		// Create a fallback repository with default data
		guidanceRepo = createFallbackGuidanceRepo()
	}
	
	return &GuidanceHandler{
		DB:           db,
		taskRepo:     repositories.NewTaskRepository(db),
		guidanceRepo: guidanceRepo,
	}
}

// GetTaskGuidance returns guidance for a specific task and tech stack
func (h *GuidanceHandler) GetTaskGuidance(c *gin.Context) {
	taskIDStr := c.Param("id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid task ID")
		return
	}

	// Get the tech stack from query parameter
	techStack := c.Query("stack")
	if techStack == "" {
		techStack = "Other"
	}

	// Get the task to determine its type
	task, err := h.taskRepo.GetByID(uint(taskID))
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}

	// Get guidance from repository
	guidance := h.guidanceRepo.GetGuidance(techStack, task.TaskType)

	// Render the guidance fragment
	c.HTML(http.StatusOK, "task-guidance-fragment.html", gin.H{
		"TaskID":   taskID,
		"Task":     task,
		"Guidance": guidance,
	})
}

// GetAvailableStacks returns all available tech stacks
func (h *GuidanceHandler) GetAvailableStacks(c *gin.Context) {
	stacks := h.guidanceRepo.GetAvailableStacks()
	c.JSON(http.StatusOK, gin.H{"stacks": stacks})
}

// GetGuidancesByStack returns all guidances for a specific stack
func (h *GuidanceHandler) GetGuidancesByStack(c *gin.Context) {
	stack := c.Param("stack")
	guidances := h.guidanceRepo.GetGuidancesByStack(stack)
	c.JSON(http.StatusOK, gin.H{"guidances": guidances})
}

// ReloadGuidance reloads guidance data from file
func (h *GuidanceHandler) ReloadGuidance(c *gin.Context) {
	if err := h.guidanceRepo.ReloadData(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload guidance data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Guidance data reloaded successfully"})
}

// AddGuidance adds or updates a guidance entry (admin only)
func (h *GuidanceHandler) AddGuidance(c *gin.Context) {
	var guidance repositories.GuidanceEntry
	if err := c.ShouldBindJSON(&guidance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.guidanceRepo.AddGuidance(guidance); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add guidance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Guidance added successfully"})
}

// DeleteGuidance removes a guidance entry (admin only)
func (h *GuidanceHandler) DeleteGuidance(c *gin.Context) {
	stack := c.Param("stack")
	taskType := c.Param("task_type")

	if err := h.guidanceRepo.DeleteGuidance(stack, taskType); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Guidance deleted successfully"})
}

// createFallbackGuidanceRepo creates a fallback repository with minimal default data
func createFallbackGuidanceRepo() *repositories.GuidanceRepository {
	// Create an in-memory fallback repository
	// This would be used if the JSON file cannot be loaded
	repo := &repositories.GuidanceRepository{}
	// Initialize with empty data structure
	return repo
}
