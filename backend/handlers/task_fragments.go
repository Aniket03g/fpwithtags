package handlers

import (
	"log"
	"net/http"

	"github.com/FeaturePlus/backend/models"
	"github.com/gin-gonic/gin"
)

// GetAssignedTasksFragment returns tasks assigned to the current user
func (h *TaskHandler) GetAssignedTasksFragment(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusOK, "all-tasks-list.html", gin.H{
			"Tasks": []models.Task{},
			"Title": "My Assigned Tasks",
			"Empty": "No tasks are currently assigned to you",
		})
		return
	}

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("ERROR: Failed to convert user_id to uint: %v", userID)
		c.HTML(http.StatusOK, "all-tasks-list.html", gin.H{
			"Tasks": []models.Task{},
			"Title": "My Assigned Tasks",
			"Empty": "Error retrieving assigned tasks",
		})
		return
	}

	// Get tasks assigned to the user
	tasks, err := h.taskRepo.GetTasksByAssignee(userIDUint)
	if err != nil {
		log.Printf("ERROR: Failed to get tasks by assignee: %v", err)
		c.HTML(http.StatusOK, "all-tasks-list.html", gin.H{
			"Tasks": []models.Task{},
			"Title": "My Assigned Tasks",
			"Empty": "Error retrieving assigned tasks",
		})
		return
	}

	// Return the tasks as HTML fragment
	c.HTML(http.StatusOK, "all-tasks-list.html", gin.H{
		"Tasks": tasks,
		"Title": "My Assigned Tasks",
		"Empty": "No tasks are currently assigned to you",
	})
}
