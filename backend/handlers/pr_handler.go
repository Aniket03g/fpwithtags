package handlers

import (
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
)

type PullRequestHandler struct {
	prRepo repositories.PullRequestRepository
}

func NewPullRequestHandler(prRepo repositories.PullRequestRepository) *PullRequestHandler {
	return &PullRequestHandler{prRepo: prRepo}
}

// MarkTested marks a pull request as tested
func (h *PullRequestHandler) MarkTested(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR ID"})
		return
	}

	// We'll use the shared package types but not the client in this handler
	// since we're directly accessing the database

	// Get the PR from the database
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

	// Update PR tested status
	pr.Tested = true

	// Save the updated PR
	if err := h.prRepo.UpdatePR(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "PR marked as tested",
		"pr":      pr,
	})
}

// AddPullRequest handles the creation of a new pull request
func (h *PullRequestHandler) AddPullRequest(c *gin.Context) {
	// Parse the request into the shared package's UploadRequest type
	var uploadReq featureplus.UploadRequest
	if err := c.ShouldBindJSON(&uploadReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// We'll use the shared package types but not the client in this handler
	// since we're directly accessing the database

	// Convert to models.PullRequest for database storage
	pr := models.PullRequest{
		TaskID:      uint(uploadReq.TaskID),
		FeatureID:   uint(uploadReq.FeatureID),
		URL:         uploadReq.PRURL,
		Title:       uploadReq.Title,
		Branch:      uploadReq.Branch,
		Description: uploadReq.Description,
		Status:      "Open", // Always set status to Open on creation
		Tested:      uploadReq.IsTested,
		Version:     uploadReq.Version,
	}

	if err := h.prRepo.Create(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pull request"})
		return
	}

	c.JSON(http.StatusCreated, pr)
}
