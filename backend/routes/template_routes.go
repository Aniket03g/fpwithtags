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
	// Get the absolute path to the project root
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("WARNING: Failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	log.Printf("DEBUG: Executable directory: %s", execDir)
	
	// Try to load .env from various locations
	// First try project root (where the executable is)
	envPaths := []string{
		"../../.env",              // Project root from backend/routes
		"../.env",                 // Backend directory from routes
		"./.env",                  // Current directory
		"../../../.env",           // One level up from project root
		filepath.Join(execDir, ".env"), // Executable directory
	}
	
	envLoaded := false
	for _, path := range envPaths {
		log.Printf("DEBUG: Trying to load .env from %s", path)
		err := godotenv.Load(path)
		if err == nil {
			log.Printf("DEBUG: .env file loaded successfully from %s for templates", path)
			envLoaded = true
			break
		}
	}
	
	if !envLoaded {
		log.Printf("WARNING: Could not load .env file from any location for templates")
		log.Printf("INFO: Continuing with default values")
	}

	// Use environment variable for data path
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		log.Printf("INFO: DATA_PATH environment variable not set, using default path")
		dataPath = "./data" // fallback for local dev
	}
	
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)

	// Resolve absolute path to templates.json directly (like the standalone code)
	templatesPath := filepath.Join(dataPath, "templates.json")
	absPath, err := filepath.Abs(dataPath)
	absTemplatesPath, err := filepath.Abs(templatesPath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", templatesPath, err)
		absTemplatesPath = templatesPath // Fallback to original path
	}
	log.Printf("DEBUG: RegisterTemplateRoutes called")
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)
	log.Printf("DEBUG: Templates file path: %s", templatesPath)
	log.Printf("DEBUG: Absolute templates file path: %s", absTemplatesPath)
	log.Printf("DEBUG: Final absPath for data directory: %s", absPath)



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
	log.Printf("DEBUG: Looking for templates file at: %s", absTemplatesPath)
	
	// Create sample template data if it doesn't exist
	sampleTemplateData := []byte(`[
	{
		"id": "nodejs-express-mongodb",
		"name": "Node.js Express MongoDB",
		"stack": "Node.js",
		"description": "A template for Node.js applications with Express and MongoDB",
		"tech_stack": "Node.js",
		"features": [
			{
				"name": "User Authentication",
				"category": "Auth",
				"description": "User registration and login functionality"
			}
		],
		"tasks": [
			{
				"name": "Set up Express server",
				"type": "Setup",
				"description": "Initialize Express application",
				"priority": "high"
			}
		]
	},
	{
		"id": "go-postgresql",
		"name": "Go with PostgreSQL",
		"stack": "Go",
		"description": "A template for Go applications with PostgreSQL",
		"tech_stack": "Go",
		"features": [
			{
				"name": "API Endpoints",
				"category": "API",
				"description": "RESTful API endpoints"
			}
		],
		"tasks": [
			{
				"name": "Set up Go server",
				"type": "Setup",
				"description": "Initialize Go application",
				"priority": "high"
			}
		]
	}
]`)
	
	if _, err := os.Stat(absTemplatesPath); os.IsNotExist(err) {
		log.Printf("WARNING: Templates file %s does not exist", templatesPath)
		// Create a sample templates file
		if writeErr := os.WriteFile(templatesPath, sampleTemplateData, 0644); writeErr != nil {
			log.Printf("ERROR: Failed to create sample templates file: %v", writeErr)
		} else {
			log.Printf("INFO: Created sample templates file at %s", templatesPath)
		}
	} else if err != nil {
		log.Printf("ERROR: Failed to check templates file: %v", err)
	} else {
		log.Printf("INFO: Found templates file: %s", templatesPath)
		// Read file contents for debugging
		content, readErr := os.ReadFile(templatesPath)
		if readErr != nil {
			log.Printf("ERROR: Failed to read templates file: %v", readErr)
		} else {
			log.Printf("DEBUG: Templates file size: %d bytes", len(content))
			previewLen := 200
			if len(content) < previewLen {
				previewLen = len(content)
			}
			log.Printf("DEBUG: Templates file content preview: %s", string(content[:previewLen]))
		}
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
