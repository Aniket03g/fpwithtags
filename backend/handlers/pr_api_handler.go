package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
)

func PRUploadAPIHandler(c *gin.Context) {
	// Parse the request into the shared package's UploadRequest type
	var uploadReq featureplus.UploadRequest
	if err := c.ShouldBindJSON(&uploadReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Received PR JSON from CLI: %+v\n", uploadReq)

	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}

	// Fetch the task to get its FeatureID if not provided
	if uploadReq.FeatureID == 0 {
		var task models.Task
		if err := db.DB.First(&task, uploadReq.TaskID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task_id or task not found"})
			return
		}
		uploadReq.FeatureID = int(task.FeatureID)
	}

	// Convert to models.PullRequest for database storage
	pr := models.PullRequest{
		TaskID:      uint(uploadReq.TaskID),
		FeatureID:   uint(uploadReq.FeatureID),
		URL:         uploadReq.PRURL,
		Title:       uploadReq.Title,
		Branch:      uploadReq.Branch,
		Description: uploadReq.Description,
		Status:      string(featureplus.StatusOpen), // Always set status to Open on upload
		Tested:      uploadReq.IsTested,
		Version:     uploadReq.Version,
	}

	// Save to database
	prRepo := repositories.NewPullRequestRepository(db.DB)
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

		// Use repository directly since we're in the backend
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

func PRGetByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR id"})
		return
	}

	// Use repository directly since we're in the backend
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}

	var pr models.PullRequest
	if err := db.DB.First(&pr, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Convert to shared package type if needed
	sharedPR := featureplus.PullRequest{
		ID:          pr.ID,
		TaskID:      pr.TaskID,
		FeatureID:   pr.FeatureID,
		URL:         pr.URL,
		Title:       pr.Title,
		Branch:      pr.Branch,
		Description: pr.Description,
		Status:      featureplus.PRStatus(pr.Status),
		Tested:      pr.Tested,
		Version:     pr.Version,
		CreatedAt:   pr.CreatedAt,
	}

	c.JSON(http.StatusOK, sharedPR)
}
