package routes

import (
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterReleaseRoutes registers release-related routes
func RegisterReleaseRoutes(router *gin.Engine, db *gorm.DB) {
	releaseRepo := repositories.NewReleaseRepository(db)
	releaseValidator := services.NewReleaseValidator(db)
	releaseHandler := handlers.NewReleaseHandler(releaseRepo, releaseValidator)

	// Register the release routes directly under /api/releases
	releases := router.Group("/api/releases")
	{
		releases.POST("", releaseHandler.CreateRelease)
		releases.GET("", releaseHandler.GetReleases)
		releases.POST("/:id/finalize", releaseHandler.FinalizeRelease)
	}
}
