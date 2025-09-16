package handlers

import (
	"log"
	"net/http"

	"github.com/FeaturePlus/backend/models"
	"github.com/gin-gonic/gin"
)

// GetMyPRsFragment returns PRs created by the current user
func (h *WebPRHandler) GetMyPRsFragment(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusOK, "all-prs-list.html", gin.H{
			"PRs":   []models.PullRequest{},
			"Title": "My Pull Requests",
			"Empty": "You haven't created any pull requests yet",
		})
		return
	}

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("ERROR: Failed to convert user_id to uint: %v", userID)
		c.HTML(http.StatusOK, "all-prs-list.html", gin.H{
			"PRs":   []models.PullRequest{},
			"Title": "My Pull Requests",
			"Empty": "Error retrieving your pull requests",
		})
		return
	}

	// Get PRs created by the user
	prs, err := h.prRepo.GetByCreatorID(userIDUint)
	if err != nil {
		log.Printf("ERROR: Failed to get PRs by creator: %v", err)
		c.HTML(http.StatusOK, "all-prs-list.html", gin.H{
			"PRs":   []models.PullRequest{},
			"Title": "My Pull Requests",
			"Empty": "Error retrieving your pull requests",
		})
		return
	}

	// Return the PRs as HTML fragment
	c.HTML(http.StatusOK, "all-prs-list.html", gin.H{
		"PRs":   prs,
		"Title": "My Pull Requests",
		"Empty": "You haven't created any pull requests yet",
	})
}
