package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BackendReviewRequest is the internal review request structure
// We keep this separate from the shared package's ReviewRequest to maintain API compatibility
type BackendReviewRequest struct {
	Status  string `json:"status" binding:"required,oneof=approved rejected changes_requested"`
	Comment string `json:"comment"`
}

// GetPRStatus converts a string status to PRStatus type
func GetPRStatus(status string) featureplus.PRStatus {
	switch status {
	case "approved":
		return featureplus.StatusApproved
	case "rejected":
		return featureplus.StatusRejected
	case "changes_requested":
		return featureplus.StatusChangesRequested
	default:
		return featureplus.StatusOpen
	}
}

// PRReviewHandler handles PR review operations
type PRReviewHandler struct {
	prRepo repositories.PullRequestRepository
}

// NewPRReviewHandler creates a new PRReviewHandler
func NewPRReviewHandler(prRepo repositories.PullRequestRepository) *PRReviewHandler {
	return &PRReviewHandler{prRepo: prRepo}
}

// ReviewPR handles POST /api/pr/:id/review
func (h *PRReviewHandler) ReviewPR(c *gin.Context) {
	// Parse PR ID from URL
	prID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR ID"})
		return
	}

	// Parse request body
	var backendReq BackendReviewRequest
	if err := c.ShouldBindJSON(&backendReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Create a client to use the shared package
	client := featureplus.NewClient("", http.DefaultClient) // Empty base URL for local operations

	// Convert to shared package ReviewRequest
	reviewReq := &featureplus.ReviewRequest{
		Status:     GetPRStatus(backendReq.Status),
		Comment:    backendReq.Comment,
		ApprovedAt: time.Now().Unix(),
	}

	// Get the PR from the database first to verify it exists
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}

	var pr models.PullRequest
	if err := db.DB.First(&pr, prID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PR"})
		return
	}

	// Use the shared package's ApprovePR function
	if err := client.ApprovePR(int(prID), reviewReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PR: " + err.Error()})
		return
	}

	// Update PR status and comment in our database
	pr.Status = string(GetPRStatus(backendReq.Status))
	if backendReq.Comment != "" {
		pr.Description = backendReq.Comment
	}

	// Save the updated PR
	if err := h.prRepo.UpdatePR(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PR in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "PR review submitted successfully",
		"pr":       pr,
	})
}

// RegisterPRReviewRoutes registers the PR review routes
func RegisterPRReviewRoutes(router *gin.Engine, prRepo repositories.PullRequestRepository) {
	handler := NewPRReviewHandler(prRepo)
	api := router.Group("/api")
	{
		api.POST("/pr/:id/review", handler.ReviewPR)
	}
}
