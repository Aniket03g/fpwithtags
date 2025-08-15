package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReviewRequest struct {
	Status  string `json:"status" binding:"required,oneof=approved rejected changes_requested"`
	Comment string `json:"comment"`
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
	var req ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Get the PR from the database
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

	// Update PR status and comment
	pr.Status = req.Status
	if req.Comment != "" {
		pr.Description = req.Comment
	}

	// Save the updated PR
	if err := h.prRepo.UpdatePR(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update PR"})
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
