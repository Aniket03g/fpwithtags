package routes

import (
	"log"
	"os"
	"path/filepath"

	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)


func RegisterGuidanceRoutes(r *gin.Engine, db *gorm.DB) {
	// Get the absolute path to the project root
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("WARNING: Failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	log.Printf("DEBUG: Executable directory: %s", execDir)
	
	// Environment variables are loaded in main.go
	log.Printf("DEBUG: Using environment variables loaded in main.go")

	// Use environment variable for data path
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		log.Printf("INFO: DATA_PATH environment variable not set, using default path")
		dataPath = "./data" // fallback for local dev
	}
	
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)

	// Resolve absolute path to guidance.json directly (like the standalone code)
	guidancePath := filepath.Join(dataPath, "guidance.json")
	absPath, err := filepath.Abs(dataPath)
	absGuidancePath, err := filepath.Abs(guidancePath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", guidancePath, err)
		absGuidancePath = guidancePath // Fallback to original path
	}
	log.Printf("DEBUG: RegisterGuidanceRoutes called")
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)
	log.Printf("DEBUG: Guidance file path: %s", guidancePath)
	log.Printf("DEBUG: Absolute guidance file path: %s", absGuidancePath)
	log.Printf("DEBUG: Final absPath for data directory: %s", absPath)
	log.Printf("DEBUG: Final absPath = %s", absPath)
	log.Printf("INFO: Using absolute data path for guidance: %s", absPath)

	// Ensure directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		log.Printf("WARNING: Data directory %s does not exist, creating it", absPath)
		if err := os.MkdirAll(absPath, 0755); err != nil {
			log.Printf("ERROR: Failed to create data directory: %v", err)
		}
	}

	// Check if guidance.json exists
	log.Printf("DEBUG: Looking for guidance file at: %s", absGuidancePath)
	
	// Create sample guidance data if it doesn't exist
	sampleGuidanceData := []byte(`[
	{
		"stack": "Node.js",
		"task_type": "Setup",
		"guidance": "Install dependencies with npm install. Set up your Express server in server.js."
	},
	{
		"stack": "Go",
		"task_type": "Setup",
		"guidance": "Initialize your Go module with go mod init. Create a main.go file for your application entry point."
	}
]`)
	
	if _, err := os.Stat(absGuidancePath); os.IsNotExist(err) {
		log.Printf("WARNING: Guidance file %s does not exist", guidancePath)
		// Create a sample guidance file
		if writeErr := os.WriteFile(guidancePath, sampleGuidanceData, 0644); writeErr != nil {
			log.Printf("ERROR: Failed to create sample guidance file: %v", writeErr)
		} else {
			log.Printf("INFO: Created sample guidance file at %s", guidancePath)
		}
	} else if err != nil {
		log.Printf("ERROR: Failed to check guidance file: %v", err)
	} else {
		log.Printf("INFO: Found guidance file: %s", guidancePath)
		// Read file contents for debugging
		content, readErr := os.ReadFile(guidancePath)
		if readErr != nil {
			log.Printf("ERROR: Failed to read guidance file: %v", readErr)
		} else {
			log.Printf("DEBUG: Guidance file size: %d bytes", len(content))
			previewLen := 200
			if len(content) < previewLen {
				previewLen = len(content)
			}
			log.Printf("DEBUG: Guidance file content preview: %s", string(content[:previewLen]))
		}
	}

	guidanceHandler := handlers.NewGuidanceHandler(db, absPath)
	log.Printf("INFO: Initialized guidance handler with data path: %s", absPath)

	// Web routes for HTMX
	web := r.Group("/web")
	web.Use(middleware.AuthMiddleware())
	{
		// Task guidance endpoint
		web.GET("/tasks/:id/guidance", guidanceHandler.GetTaskGuidance)
		log.Printf("INFO: Registered route GET /web/tasks/:id/guidance")
	}

	// API routes for admin management
	api := r.Group("/api/guidance")
	api.Use(middleware.AuthMiddleware())
	{
		// Get available stacks
		api.GET("/stacks", guidanceHandler.GetAvailableStacks)
		log.Printf("INFO: Registered route GET /api/guidance/stacks")
		
		// Get guidances by stack
		api.GET("/stack/:stack", guidanceHandler.GetGuidancesByStack)
		log.Printf("INFO: Registered route GET /api/guidance/stack/:stack")
		
		// Admin only routes - these would require admin role middleware
		// For now, these are available to authenticated users
		// TODO: Add proper admin role checking
		
		// Add or update guidance
		api.POST("/", guidanceHandler.AddGuidance)
		log.Printf("INFO: Registered route POST /api/guidance/")
		
		// Delete guidance
		api.DELETE("/:stack/:task_type", guidanceHandler.DeleteGuidance)
		log.Printf("INFO: Registered route DELETE /api/guidance/:stack/:task_type")
		
		// Reload guidance data from file
		api.POST("/reload", guidanceHandler.ReloadGuidance)
		log.Printf("INFO: Registered route POST /api/guidance/reload")
	}

	log.Printf("INFO: Successfully registered all guidance routes")
}
