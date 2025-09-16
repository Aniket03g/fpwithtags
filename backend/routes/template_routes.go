package routes

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterTemplateRoutes(r *gin.Engine, db *gorm.DB) {
	// Use the data directory for template JSON files with absolute path
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("ERROR: Failed to get current working directory: %v", err)
		cwd = "."
	}
	// Use the absolute path to the data directory
	dataPath := filepath.Join(cwd, "data")
	// Check if the directory exists
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		log.Printf("WARNING: Data directory %s does not exist, creating it", dataPath)
		if err := os.MkdirAll(dataPath, 0755); err != nil {
			log.Printf("ERROR: Failed to create data directory: %v", err)
		}
	}
	log.Printf("INFO: Using data path for templates: %s", dataPath)
	templateHandler := handlers.NewTemplateHandler(db, dataPath)

	// API routes
	api := r.Group("/api/templates")
	api.Use(middleware.AuthMiddleware())
	{
		// Get all templates
		api.GET("/", templateHandler.GetAllTemplates)
		
		// Get template by ID
		api.GET("/:id", templateHandler.GetTemplateByID)
		
		// Get template by stack
		api.GET("/stack/:stack", templateHandler.GetTemplateByStack)
		
		// Get available stacks
		api.GET("/stacks", templateHandler.GetAvailableStacks)
		
		// Apply template to project
		api.POST("/apply", templateHandler.ApplyTemplate)
		
		// Admin routes
		// TODO: Add proper admin role checking
		api.POST("/", templateHandler.AddTemplate)
		api.DELETE("/:id", templateHandler.DeleteTemplate)
		api.POST("/reload", templateHandler.ReloadTemplates)
	}

	// Web routes for HTMX
	web := r.Group("/web/templates")
	web.Use(middleware.AuthMiddleware())
	{
		// Get template details fragment for HTMX
		web.GET("/:id/details", templateHandler.GetTemplateDetails)
		
		// New route that handles template_id as a query parameter
		web.GET("/details", func(c *gin.Context) {
			// Extract template_id from query parameter
			templateID := c.Query("template_id")
			if templateID != "" {
				// Set the ID parameter and forward to the regular handler
				c.Params = append(c.Params, gin.Param{Key: "id", Value: templateID})
				templateHandler.GetTemplateDetails(c)
				return
			}
			c.String(http.StatusOK, "<div class='text-sm text-gray-500 p-4'>Select a template to see details</div>")
		})
		
		// Keep the old route for backward compatibility
		web.GET("/${this.value}/details", func(c *gin.Context) {
			// Extract template_id from query parameter
			templateID := c.Query("template_id")
			if templateID != "" {
				// Set the ID parameter and forward to the regular handler
				c.Params = append(c.Params, gin.Param{Key: "id", Value: templateID})
				templateHandler.GetTemplateDetails(c)
				return
			}
			c.String(http.StatusOK, "<div class='text-sm text-gray-500 p-4'>Select a template to see details</div>")
		})
	}
}
