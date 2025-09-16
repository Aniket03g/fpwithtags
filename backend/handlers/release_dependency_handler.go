package handlers

import (
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReleaseDependencyHandler handles dependency-related operations for releases
type ReleaseDependencyHandler struct {
	db                *gorm.DB
	dependencyService *services.DependencyService
	releaseValidator  *services.ReleaseValidator
	prRepo            repositories.PullRequestRepository
}

// NewReleaseDependencyHandler creates a new ReleaseDependencyHandler
func NewReleaseDependencyHandler(
	db *gorm.DB,
	dependencyService *services.DependencyService,
	releaseValidator *services.ReleaseValidator,
	prRepo repositories.PullRequestRepository,
) *ReleaseDependencyHandler {
	return &ReleaseDependencyHandler{
		db:                db,
		dependencyService: dependencyService,
		releaseValidator:  releaseValidator,
		prRepo:            prRepo,
	}
}

// CheckDependenciesForRelease handles POST /releases/check-dependencies
// Checks dependencies for selected PRs in a release
func (h *ReleaseDependencyHandler) CheckDependenciesForRelease(c *gin.Context) {
	// Parse PR IDs from form
	prIDStrs := c.PostFormArray("pr_ids")
	if len(prIDStrs) == 0 {
		c.HTML(http.StatusOK, "release-dependency-check.html", gin.H{
			"CheckComplete": true,
			"HasDependencyIssues": false,
		})
		return
	}

	// Convert PR IDs to integers
	var prIDs []int
	for _, idStr := range prIDStrs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR ID"})
			return
		}
		prIDs = append(prIDs, id)
	}

	// Get PRs from database
	var prs []models.PullRequest
	for _, prID := range prIDs {
		pr, err := h.prRepo.GetByID(uint(prID))
		if err != nil {
			continue
		}
		prs = append(prs, *pr)
	}

	// Check for blocked PRs
	var blockedPRs []models.PullRequest
	for _, pr := range prs {
		isBlocked, _, err := h.dependencyService.CheckPRBlocked(pr.ID)
		if err != nil {
			continue
		}
		if isBlocked {
			blockedPRs = append(blockedPRs, pr)
		}
	}

	// Create a temporary release ID for dependency validation
	// In a real implementation, this would be a draft release ID
	tempReleaseID := uint(0)

	// Check for missing dependencies within the release
	allSatisfied, missingDeps, err := h.dependencyService.ValidateDependenciesInRelease(tempReleaseID)
	
	// Prepare missing dependency details
	var missingDependencies []struct {
		FeatureID    uint
		FeatureTitle string
		MissingDeps  []uint
	}

	if !allSatisfied && err == nil {
		for featureID, deps := range missingDeps {
			var feature models.Feature
			if err := h.db.First(&feature, featureID).Error; err == nil {
				missingDependencies = append(missingDependencies, struct {
					FeatureID    uint
					FeatureTitle string
					MissingDeps  []uint
				}{
					FeatureID:    featureID,
					FeatureTitle: feature.Title,
					MissingDeps:  deps,
				})
			}
		}
	}

	// Render dependency check results
	c.HTML(http.StatusOK, "release-dependency-check.html", gin.H{
		"ReleaseID":           tempReleaseID,
		"CheckComplete":       true,
		"HasDependencyIssues": len(blockedPRs) > 0 || !allSatisfied,
		"BlockedPRs":          blockedPRs,
		"MissingDependencies": missingDependencies,
	})
}

// RegisterReleaseDependencyRoutes registers routes for release dependency handlers
func RegisterReleaseDependencyRoutes(
	router *gin.Engine,
	db *gorm.DB,
	dependencyService *services.DependencyService,
	releaseValidator *services.ReleaseValidator,
	prRepo repositories.PullRequestRepository,
) {
	handler := NewReleaseDependencyHandler(db, dependencyService, releaseValidator, prRepo)
	
	// Register routes
	router.POST("/releases/check-dependencies", handler.CheckDependenciesForRelease)
}
