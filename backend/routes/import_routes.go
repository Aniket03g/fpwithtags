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

// RegisterImportRoutes registers routes for GitHub MCP import functionality
// These routes handle dynamic loading of project templates from /backend/data/imports/
func RegisterImportRoutes(r *gin.Engine, db *gorm.DB) {
	log.Printf("INFO: Registering import routes")

	// Get data path from environment
	dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		dataPath = "./data"
		log.Printf("INFO: DATA_PATH not set, using default: %s", dataPath)
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(dataPath)
	if err != nil {
		log.Printf("ERROR: Failed to resolve absolute path for %s: %v", dataPath, err)
		absPath = dataPath
	}

	log.Printf("INFO: Using data path for imports: %s", absPath)

	// Create import handler (does NOT preload files)
	importHandler := handlers.NewImportHandler(db, absPath)

	// API routes - all require authentication
	api := r.Group("/api/imports")
	api.Use(middleware.AuthMiddleware())
	{
		// List available import templates
		// GET /api/imports
		api.GET("", importHandler.ListAvailableImports)
		log.Printf("INFO: Registered route GET /api/imports")

		// Import a project from GitHub MCP JSON
		// POST /api/imports/import
		// Body: { "project_id": "github_project_demo", "project_name": "Imported Demo", "description": "..." }
		api.POST("/import", importHandler.ImportProject)
		log.Printf("INFO: Registered route POST /api/imports/import")

		// Save a new import template (for MCP to upload)
		// POST /api/imports/save
		// Body: { "project_id": "...", "template": {...} }
		api.POST("/save", importHandler.SaveImportTemplate)
		log.Printf("INFO: Registered route POST /api/imports/save")

		// Delete an import template
		// DELETE /api/imports/:id
		api.DELETE("/:id", importHandler.DeleteImportTemplate)
		log.Printf("INFO: Registered route DELETE /api/imports/:id")

		// Import directly from GitHub repository using MCP
		// POST /api/imports/github
		// Body: { "repo_url": "https://github.com/user/repo", "project_name": "Optional Name" }
		api.POST("/github", importHandler.ImportFromGitHub)
		log.Printf("INFO: Registered route POST /api/imports/github")
	}

	log.Printf("INFO: Successfully registered all import routes")
}
