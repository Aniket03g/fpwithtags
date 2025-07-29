package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
)

func PRUploadAPIHandler(c *gin.Context) {
	var pr models.PullRequest
	if err := c.ShouldBindJSON(&pr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Received PR JSON from CLI: %+v\n", pr)

	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	prRepo := repositories.NewPullRequestRepository(db.DB)

	// Fetch the task to get its FeatureID
	var task models.Task
	if err := db.DB.First(&task, pr.TaskID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task_id or task not found"})
		return
	}
	pr.FeatureID = task.FeatureID

	pr.Status = "Open"      // Always set status to Open on upload
	pr.Version = pr.Version // Already set by JSON binding, but ensure it's not nil

	if err := prRepo.Create(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PR to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved", "data": pr})
}

func PRListAPIHandler(c *gin.Context) {
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	prRepo := repositories.NewPullRequestRepository(db.DB)

	// Check if feature_id query parameter is provided
	featureIDStr := c.Query("feature_id")
	if featureIDStr != "" {
		featureID, err := strconv.Atoi(featureIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature_id parameter"})
			return
		}

		prs, err := prRepo.GetByFeatureID(featureID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PRs"})
			return
		}
		c.JSON(http.StatusOK, prs)
		return
	}

	// If no feature_id, return all PRs
	prs, err := prRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PRs"})
		return
	}
	c.JSON(http.StatusOK, prs)
}
