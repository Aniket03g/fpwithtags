package routes

import (
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterGuidanceRoutes(r *gin.Engine, db *gorm.DB) {
	// Use the data directory for guidance JSON files
	guidanceHandler := handlers.NewGuidanceHandler(db, "./backend/data")

	// Web routes for HTMX
	web := r.Group("/web")
	web.Use(middleware.AuthMiddleware())
	{
		// Task guidance endpoint
		web.GET("/tasks/:id/guidance", guidanceHandler.GetTaskGuidance)
	}

	// API routes for admin management
	api := r.Group("/api/guidance")
	api.Use(middleware.AuthMiddleware())
	{
		// Get available stacks
		api.GET("/stacks", guidanceHandler.GetAvailableStacks)
		
		// Get guidances by stack
		api.GET("/stack/:stack", guidanceHandler.GetGuidancesByStack)
		
		// Admin only routes - these would require admin role middleware
		// For now, these are available to authenticated users
		// TODO: Add proper admin role checking
		
		// Add or update guidance
		api.POST("/", guidanceHandler.AddGuidance)
		
		// Delete guidance
		api.DELETE("/:stack/:task_type", guidanceHandler.DeleteGuidance)
		
		// Reload guidance data from file
		api.POST("/reload", guidanceHandler.ReloadGuidance)
	}
}
