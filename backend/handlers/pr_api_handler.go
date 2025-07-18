package handlers

import (
	"fmt"
	"net/http"

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

	pr.Status = "Open"      // Always set status to Open on upload
	pr.Version = pr.Version // Already set by JSON binding, but ensure it's not nil

	if err := prRepo.Create(&pr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PR to database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "saved", "data": pr})
}
