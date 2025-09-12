package routes

import (
	"net/http"

	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterDependencyRoutes registers all routes for the dependency module
func RegisterDependencyRoutes(r *gin.Engine, db *gorm.DB) {
	// Create repository, service, and handler
	dependencyRepo := repositories.NewDependencyRepository(db)
	dependencyService := services.NewDependencyService(dependencyRepo)
	dependencyHandler := handlers.NewDependencyHandler(dependencyService)

	// API routes (protected by auth middleware)
	dependencyAPI := r.Group("/api/dependencies", middleware.AuthMiddleware())
	{
		// CRUD operations
		dependencyAPI.POST("", dependencyHandler.CreateDependency)
		dependencyAPI.GET("", dependencyHandler.ListDependencies)
		dependencyAPI.DELETE("/:id", dependencyHandler.DeleteDependency)
		
		// HTML fragments for HTMX
		dependencyAPI.GET("/panels", dependencyHandler.GetDependencyPanels)
		dependencyAPI.GET("/modal", dependencyHandler.ShowDependencyModal)
	}
	
	// Add route to the web group for the shell page
	webGroup := r.Group("/web", middleware.AuthMiddleware())
	{
		webGroup.GET("/dependencies", func(c *gin.Context) {
			// Get projects for sidebar
			var projects []models.Project
			db.Find(&projects)

			// Get user ID and role from context
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("user_role")

			// Create CurrentUser object for template
			currentUser := map[string]interface{}{
				"ID":        userID,
				"Role":      userRole,
				"IsManager": userRole == "manager",
			}

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"InitialURL":  "/web/fragments/dependencies",
				"CurrentUser": currentUser,
				"Projects":    projects,
			})
		})
	}

	// Register the fragment routes
	fragments := r.Group("/web/fragments", middleware.AuthMiddleware())
	{
		// Dependencies list fragment
		fragments.GET("/dependencies", dependencyHandler.GetDependenciesListFragment)
		
		// Dependencies panels and modal fragments
		fragments.GET("/dependencies/panels", dependencyHandler.GetDependencyPanelsFragment)
		fragments.GET("/dependencies/modal", dependencyHandler.ShowDependencyModalFragment)
	}
}
