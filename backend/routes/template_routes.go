package routes

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func RegisterTemplateRoutes(r *gin.Engine, db *gorm.DB) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARNING: Error loading .env file for templates: %v", err)
		log.Printf("INFO: Continuing with default values")
	} else {
		log.Printf("DEBUG: .env file loaded successfully for templates")
	}

	// Use environment variable for data path
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		log.Printf("INFO: DATA_PATH environment variable not set, using default path")
		dataPath = "./data" // fallback for local dev
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", dataPath, err)
		absPath = dataPath // Fallback to original path
	}
	log.Printf("DEBUG: RegisterTemplateRoutes called")
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)
	log.Printf("DEBUG: Final absPath = %s", absPath)



	log.Printf("INFO: Using absolute data path for templates: %s", absPath)

	// Ensure directory exists
	dirInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("WARNING: Data directory %s does not exist, creating it", absPath)
			if err := os.MkdirAll(absPath, 0755); err != nil {
				log.Printf("ERROR: Failed to create data directory: %v", err)
			} else {
				log.Printf("INFO: Successfully created data directory: %s", absPath)
			}
		} else {
			log.Printf("ERROR: Failed to stat data directory %s: %v", absPath, err)
		}
	} else if !dirInfo.IsDir() {
		log.Printf("ERROR: Path is not a directory: %s", absPath)
	} else {
		log.Printf("INFO: Data directory exists with permissions: %s", dirInfo.Mode())
	}

	// Check if templates.json exists
	templatesPath := filepath.Join(absPath, "templates.json")
	if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
		log.Printf("WARNING: Templates file %s does not exist", templatesPath)
	} else {
		log.Printf("INFO: Found templates file: %s", templatesPath)
	}

	templateHandler := handlers.NewTemplateHandler(db, absPath)
	log.Printf("INFO: Initialized template handler with data path: %s", absPath)

	// API routes
	api := r.Group("/api/templates")
	api.Use(middleware.AuthMiddleware())
	{
		// Get all templates
		api.GET("/", templateHandler.GetAllTemplates)
		log.Printf("INFO: Registered route GET /api/templates/")

		// Get template by ID
		api.GET("/:id", templateHandler.GetTemplateByID)
		log.Printf("INFO: Registered route GET /api/templates/:id")

		// Get template by stack
		api.GET("/stack/:stack", templateHandler.GetTemplateByStack)
		log.Printf("INFO: Registered route GET /api/templates/stack/:stack")

		// Get available stacks
		api.GET("/stacks", templateHandler.GetAvailableStacks)
		log.Printf("INFO: Registered route GET /api/templates/stacks")

		// Apply template to project
		api.POST("/apply", templateHandler.ApplyTemplate)
		log.Printf("INFO: Registered route POST /api/templates/apply")

		// Admin routes
		// TODO: Add proper admin role checking
		api.POST("/", templateHandler.AddTemplate)
		log.Printf("INFO: Registered route POST /api/templates/")

		api.DELETE("/:id", templateHandler.DeleteTemplate)
		log.Printf("INFO: Registered route DELETE /api/templates/:id")

		api.POST("/reload", templateHandler.ReloadTemplates)
		log.Printf("INFO: Registered route POST /api/templates/reload")
	}

	// Web routes for HTMX
	web := r.Group("/web/templates")
	web.Use(middleware.AuthMiddleware())
	{
		// Get template details fragment for HTMX
		web.GET("/:id/details", templateHandler.GetTemplateDetails)
		log.Printf("INFO: Registered route GET /web/templates/:id/details")

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
		log.Printf("INFO: Registered route GET /web/templates/details")

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
		log.Printf("INFO: Registered route GET /web/templates/${this.value}/details")
	}

	log.Printf("INFO: Successfully registered all template routes")
}
