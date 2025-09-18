package routes

import (
	"log"
	"os"
	"path/filepath"

	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func RegisterGuidanceRoutes(r *gin.Engine, db *gorm.DB) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARNING: Error loading .env file for guidance: %v", err)
		log.Printf("INFO: Continuing with default values")
	} else {
		log.Printf("DEBUG: .env file loaded successfully for guidance")
	}

	// Use environment variable for data path
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		log.Printf("INFO: DATA_PATH environment variable not set, using default path")
		dataPath = "./data" // fallback for local dev
	}
	
	log.Printf("DEBUG: DATA_PATH env = %s", os.Getenv("DATA_PATH"))
	log.Printf("DEBUG: Using data path: %s", dataPath)

	// Resolve absolute path
	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", dataPath, err)
		absPath = dataPath // Fallback to original path
	}
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
	guidancePath := filepath.Join(absPath, "guidance.json")
	if _, err := os.Stat(guidancePath); os.IsNotExist(err) {
		log.Printf("WARNING: Guidance file %s does not exist", guidancePath)
	} else {
		log.Printf("INFO: Found guidance file: %s", guidancePath)
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
