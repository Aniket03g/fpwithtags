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

// RenderDashboardShell renders the main dashboard.html shell for all main web pages
func RenderDashboardShell(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{})
}

// ProjectsFragmentHandler returns the projects list fragment (placeholder for now)
func ProjectsFragmentHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "projects-list.html", gin.H{})
}

// RenderAppShell renders the dashboard.html shell and passes the correct initial fragment URL
func RenderAppShell(c *gin.Context) {
	path := c.Request.URL.Path
	fragment := ""

	// Remove "/web" prefix
	if strings.HasPrefix(path, "/web") {
		path = path[len("/web"):]
	}

	// Build fragment URL for new /web/fragments group
	switch {
	case path == "/dashboard" || path == "/projects":
		fragment = "/web/fragments/projects"
	case strings.HasPrefix(path, "/projects/") && strings.HasSuffix(path, "/features"):
		// e.g. /projects/8/features -> /web/fragments/projects/8/features
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			fragment = "/web/fragments/projects/" + parts[2] + "/features"
		}
	case strings.HasPrefix(path, "/projects/") && strings.Contains(path, "/features/"):
		// e.g. /projects/8/features/33 -> /web/fragments/features/33
		parts := strings.Split(path, "/")
		if len(parts) >= 5 {
			fragment = "/web/fragments/features/" + parts[4]
		}
	case strings.HasPrefix(path, "/projects/"):
		// e.g. /projects/8 -> /web/fragments/projects/8
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			fragment = "/web/fragments/projects/" + parts[2]
		}
	default:
		fragment = "/web/fragments/projects"
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"InitialURL": fragment,
	})
}

// Middleware to redirect non-HTMX requests for fragments to the app shell route
func FragmentHTMXGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		hx := c.GetHeader("HX-Request")
		if hx == "true" {
			c.Next()
			return
		}

		// Not an HTMX request: redirect to the corresponding app shell route
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/web/fragments/projects/") && strings.HasSuffix(path, "/features") {
			// /web/fragments/projects/:id/features -> /web/projects/:id/features
			parts := strings.Split(path, "/")
			if len(parts) >= 6 {
				redirect := "/web/projects/" + parts[4] + "/features"
				c.Redirect(http.StatusFound, redirect)
				c.Abort()
				return
			}
		} else if strings.HasPrefix(path, "/web/fragments/features/") {
			// /web/fragments/features/:id -> /web/projects/:pid/features/:id (try to guess parent, fallback)
			parts := strings.Split(path, "/")
			if len(parts) >= 5 {
				// You may want to look up the parent project, but fallback to /web/features/:id
				redirect := "/web/features/" + parts[4]
				c.Redirect(http.StatusFound, redirect)
				c.Abort()
				return
			}
		} else if strings.HasPrefix(path, "/web/fragments/projects/") {
			// /web/fragments/projects/:id -> /web/projects/:id
			parts := strings.Split(path, "/")
			if len(parts) >= 5 {
				redirect := "/web/projects/" + parts[4]
				c.Redirect(http.StatusFound, redirect)
				c.Abort()
				return
			}
		} else if path == "/web/fragments/projects" {
			c.Redirect(http.StatusFound, "/web/projects")
			c.Abort()
			return
		}
		// Default fallback: redirect to dashboard
		c.Redirect(http.StatusFound, "/web/dashboard")
		c.Abort()
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
	taskRepo := repositories.NewTaskRepository(db.DB) // Uncommented
	// tagRepo := repositories.NewTagRepository(db.DB) // Comment out for now
	// attachmentRepo := repositories.NewTaskAttachmentRepository(db.DB, sqliteFS) // Comment out for now
	// commentRepo := repositories.NewCommentRepository(db.DB) // Comment out for now

	// Create handlers
	userHandler := handlers.NewUserHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo)
	featureHandler := handlers.NewFeatureHandler(featureRepo, nil, taskRepo, db.DB) // Added taskRepo
	taskHandler := handlers.NewTaskHandler(taskRepo, db.DB)                         // Uncommented
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
		// Main web page routes: always render dashboard.html shell
		web.GET("/dashboard", RenderAppShell)
		web.GET("/projects", RenderAppShell)
		web.GET("/projects/:id", RenderAppShell)
		web.GET("/projects/:id/features", RenderAppShell)
		web.GET("/projects/:id/features/:featureid", RenderAppShell)

		// HTMX fragment and API routes (do not change)
		web.GET("/features/:id/tasks", taskHandler.GetTasksByFeature)
		web.GET("/features/:id/tasks/new", taskHandler.NewTaskForm)
		web.POST("/features/:id/tasks", taskHandler.CreateTaskForFeature)
		web.GET("/tasks/cancel", taskHandler.CancelTaskForm)
	}

	// New fragment group for all HTMX fragments
	fragments := router.Group("/web/fragments", FragmentHTMXGuard())
	{
		fragments.GET("/projects", projectHandler.GetAllProjects)
		fragments.GET("/projects/:id", projectHandler.GetProject)
		fragments.GET("/projects/:id/features", featureHandler.GetProjectFeatures)
		fragments.GET("/features/:id", featureHandler.GetFeature)
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
