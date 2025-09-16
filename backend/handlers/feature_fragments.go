package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/gin-gonic/gin"
)

// GetFeaturesProgressFragment returns progress statistics for features
func (h *FeatureHandler) GetFeaturesProgressFragment(c *gin.Context) {
	// Check if project_id is provided
	projectIDStr := c.Query("project_id")
	var features []models.Feature
	var err error

	if projectIDStr != "" {
		// Get features for specific project
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			log.Printf("ERROR: Invalid project_id: %s", projectIDStr)
			c.HTML(http.StatusOK, "feature-progress.html", gin.H{
				"Error": "Invalid project ID",
			})
			return
		}
		features, err = h.repo.GetFeaturesByProject(projectID)
	} else {
		// Get all features
		features, err = h.repo.GetAllFeatures()
	}

	if err != nil {
		log.Printf("ERROR: Failed to get features: %v", err)
		c.HTML(http.StatusOK, "feature-progress.html", gin.H{
			"Error": "Failed to retrieve features",
		})
		return
	}

	// Calculate progress statistics
	var todoCount, inProgressCount, doneCount int
	for _, feature := range features {
		switch feature.Status {
		case models.StatusTodo:
			todoCount++
		case models.StatusInProgress:
			inProgressCount++
		case models.StatusDone:
			doneCount++
		}
	}

	totalCount := len(features)
	var todoPercent, inProgressPercent, donePercent float64
	if totalCount > 0 {
		todoPercent = float64(todoCount) / float64(totalCount) * 100
		inProgressPercent = float64(inProgressCount) / float64(totalCount) * 100
		donePercent = float64(doneCount) / float64(totalCount) * 100
	}

	// Return the progress statistics as HTML fragment
	c.HTML(http.StatusOK, "feature-progress.html", gin.H{
		"TodoCount":         todoCount,
		"InProgressCount":   inProgressCount,
		"DoneCount":         doneCount,
		"TotalCount":        totalCount,
		"TodoPercent":       todoPercent,
		"InProgressPercent": inProgressPercent,
		"DonePercent":       donePercent,
		"ProjectID":         projectIDStr,
	})
}
