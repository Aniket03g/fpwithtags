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

// PRDependencyHandler handles dependency-related operations for PRs
type PRDependencyHandler struct {
	db              *gorm.DB
	dependencyService *services.DependencyService
	prRepo          repositories.PullRequestRepository
}

// NewPRDependencyHandler creates a new PRDependencyHandler
func NewPRDependencyHandler(db *gorm.DB, dependencyService *services.DependencyService, prRepo repositories.PullRequestRepository) *PRDependencyHandler {
	return &PRDependencyHandler{
		db:              db,
		dependencyService: dependencyService,
		prRepo:          prRepo,
	}
}

// GetPRDependencyPanel handles GET /web/fragments/pr/:id/dependencies
// Returns HTML fragment for PR dependency panel
func (h *PRDependencyHandler) GetPRDependencyPanel(c *gin.Context) {
	// Parse PR ID from URL
	prIDStr := c.Param("id")
	prID, err := strconv.ParseUint(prIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid PR ID"})
		return
	}

	// Get PR from repository
	pr, err := h.prRepo.GetByID(uint(prID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "PR not found"})
		return
	}

	// Check if PR is blocked by dependencies
	isBlocked, _, err := h.dependencyService.CheckPRBlocked(uint(prID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check dependencies"})
		return
	}

	// Get detailed blocking dependency information
	blockingDependencies, err := h.dependencyService.GetPRBlockingDependencies(uint(prID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dependency details"})
		return
	}

	// Create PR response with dependency information
	prResponse := models.NewPRResponse(pr)
	prResponse.IsBlocked = isBlocked
	prResponse.BlockingDependencies = blockingDependencies
	
	// TODO: In a real implementation, we would fetch GitHub PR mergeable state
	// via GitHub API. For now, we'll set a default value.
	prResponse.MergeableState = "clean"
	if isBlocked {
		prResponse.MergeableState = "blocked"
	}

	// Render the PR dependencies panel
	c.HTML(http.StatusOK, "pr_dependencies.html", gin.H{
		"PR":                 pr,
		"IsBlocked":          isBlocked,
		"BlockingDependencies": blockingDependencies,
		"MergeableState":     prResponse.MergeableState,
	})
}

// RegisterPRDependencyRoutes registers routes for PR dependency handlers
func RegisterPRDependencyRoutes(router *gin.Engine, db *gorm.DB, dependencyService *services.DependencyService, prRepo repositories.PullRequestRepository) {
	handler := NewPRDependencyHandler(db, dependencyService, prRepo)
	
	// Register routes
	webFragments := router.Group("/web/fragments")
	{
		webFragments.GET("/pr/:id/dependencies", handler.GetPRDependencyPanel)
	}
}
