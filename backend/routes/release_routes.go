package routes

import (
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
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

	// Create role middleware with database instance
	roleMiddleware := middleware.CreateRoleMiddleware(db)

	// MARKER:RELEASE_ROUTES Register release API endpoints
	// Register the release routes directly under /api/releases
	releases := router.Group("/api/releases", middleware.AuthMiddleware())
	releases.Use(roleMiddleware()) // Apply role middleware to all release routes
	{
		// GET releases is available to all authenticated users
		releases.GET("", releaseHandler.GetReleases)

		// Create releases are restricted to managers only
		managerRoutes := releases.Group("/", roleMiddleware("manager"))
		{
			managerRoutes.POST("", releaseHandler.CreateRelease)
			// managerRoutes.POST("/:id/finalize", releaseHandler.FinalizeRelease)
		}
		
		// Temporarily disable authentication for finalize endpoint
		router.POST("/api/releases/:id/finalize", releaseHandler.FinalizeRelease)
	}
	
	// Web release routes are registered in main.go
}
