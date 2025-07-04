package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/routes"
	"github.com/jilio/sqlitefs"

	"github.com/gin-gonic/gin"
)

// Simple logging middleware to print requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		log.Printf("[GIN] %v | %3d | %13v | %15s | %-7s %s",
			end.Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
		if raw != "" {
			log.Printf("      Raw query: %s", raw)
		}
	}
}

func main() {
	// Initialize DB
	db, err := database.InitDB()
	if err != nil {
		panic("failed to connect database")
	}

	// Get the underlying *sql.DB from GORM
	sqlDB, err := db.DB.DB()
	if err != nil {
		panic("failed to get database connection: " + err.Error())
	}

	// Initialize SQLiteFS with the *sql.DB
	sqliteFS, err := sqlitefs.NewSQLiteFS(sqlDB)
	if err != nil {
		panic("failed to initialize SQLiteFS: " + err.Error())
	}
	defer sqliteFS.Close()

	// Migrate all schemas
	if err := db.Migrate(&models.User{}, &models.Project{}, &models.Feature{}, &models.SubFeature{}, &models.Task{}, &models.FeatureTag{}, &models.TaskAttachment{}, &models.Comment{}); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	projectRepo := repositories.NewProjectRepository(db.DB)
	featureRepo := repositories.NewFeatureRepository(db.DB)
	// taskRepo := repositories.NewTaskRepository(db.DB) // Comment out for now
	// tagRepo := repositories.NewTagRepository(db.DB) // Comment out for now
	// attachmentRepo := repositories.NewTaskAttachmentRepository(db.DB, sqliteFS) // Comment out for now
	// commentRepo := repositories.NewCommentRepository(db.DB) // Comment out for now

	// Create handlers
	userHandler := handlers.NewUserHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo)
	featureHandler := handlers.NewFeatureHandler(featureRepo, nil, db.DB) // Passing nil for tagRepo for now
	// taskHandler := handlers.NewTaskHandler(taskRepo, db.DB) // Comment out for now
	// tagHandler := handlers.NewTagHandler(tagRepo, featureRepo, db.DB) // Comment out for now
	// attachmentHandler := handlers.NewTaskAttachmentHandler(attachmentRepo, sqliteFS) // Comment out for now
	// commentHandler := handlers.NewCommentHandler(commentRepo, attachmentRepo) // Comment out for now

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Configure multipart form handling
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Add logging middleware
	router.Use(LoggingMiddleware())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Register auth routes
	routes.RegisterAuthRoutes(router, db.DB)

	// --- EXISTING API ROUTES FOR NEXT.JS (NO CHANGES) ---
	api := router.Group("/api")
	{
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("", userHandler.GetAllUsers)
			// ... etc
		}
		projectRoutes := api.Group("/projects", middleware.AuthMiddleware())
		{
			// Note: This GET is the original JSON route for Next.js
			projectRoutes.GET("", projectHandler.GetAllProjects)
			// ... etc
		}
		api.GET("/features/project/:project_id", featureHandler.GetProjectFeatures)
	}

	// ==========================================================
	// ===== NEW WEB ROUTES FOR HTMX TESTING =====
	// ==========================================================
	web := router.Group("/web")
	{
		web.GET("/dashboard", projectHandler.ShowDashboard)
		web.GET("/projects", projectHandler.GetAllProjects)
		web.GET("/projects/:id", projectHandler.GetProject)

		// *** FIX WAS HERE: Renamed :projectid to :id to match the route above ***
		web.GET("/projects/:id/features", featureHandler.GetProjectFeatures)

		web.GET("/projects/:id/features/:featureid", featureHandler.GetFeature)
	}
	// ==========================================================

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Serving static frontend files
	staticDir := "../frontend/.next"
	router.Static("/static", filepath.Join(staticDir, "static"))
	router.Static("/_next", filepath.Join(staticDir, "static"))

	// Handle unmatched routes
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") || strings.HasPrefix(c.Request.URL.Path, "/web/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}
		path := filepath.Join(staticDir, c.Request.URL.Path)
		if _, err := os.Stat(path); err == nil {
			c.File(path)
			return
		}
		indexPath := filepath.Join(staticDir, "server", "app", "index.html")
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Frontend build not found"})
			return
		}
		c.File(indexPath)
	})

	// Start server
	log.Println("Server starting on :8080...")
	if err := router.Run(":8080"); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
