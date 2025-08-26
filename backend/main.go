package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/jilio/sqlitefs"
	"github.com/joho/godotenv" // For loading .env files
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
			log.Printf("       Raw query: %s", raw)
		}
	}
}

// AppHandler handles app shell rendering and common functionality
type AppHandler struct {
	db *database.Database
	projectRepo *repositories.ProjectRepository
}

// NewAppHandler creates a new app handler with database access
func NewAppHandler(db *database.Database) *AppHandler {
	return &AppHandler{
		db: db,
		projectRepo: repositories.NewProjectRepository(db.DB),
	}
}

// RenderAppShell renders the dashboard.html shell and passes the correct initial fragment URL
func (h *AppHandler) RenderAppShell(c *gin.Context) {
	path := c.Request.URL.Path
	fragment := ""

	// Remove "/web" prefix
	if strings.HasPrefix(path, "/web") {
		path = path[len("/web"):]
	}

	// Build fragment URL for new /web/fragments group
	switch {
	case path == "/dashboard":
		fragment = "/web/fragments/dashboard-status"
	case path == "/projects":
		fragment = "/web/fragments/projects"
	case strings.HasPrefix(path, "/projects/") && strings.HasSuffix(path, "/features"):
		// e.g. /projects/8/features -> /web/projects/8/features/content
		parts := strings.Split(path, "/")
		if len(parts) >= 4 {
			fragment = "/web/projects/" + parts[2] + "/features/content"
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

	// Get user ID and role from context (set by AuthMiddleware and RoleMiddleware)
	userID, _ := c.Get("user_id")
	userRole, _ := c.Get("user_role")
	
	// Create CurrentUser object for template
	currentUser := map[string]interface{}{
		"ID":   userID,
		"Role": userRole,
	}

	// Get projects for sidebar
	var projects []models.Project
	h.db.DB.Find(&projects)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"InitialURL":   fragment,
		"CurrentUser":  currentUser,
		"Projects":     projects,
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

// Handler for dashboard status fragment
func DashboardStatusFragment(c *gin.Context) {
	// Check if user is logged in
	userID, loggedIn := c.Get("user_id")
	
	// Get user role from context (set by RoleMiddleware)
	userRole, hasRole := c.Get("user_role")
	
	// Create CurrentUser object for template
	currentUser := map[string]interface{}{
		"ID":   userID,
		"Role": userRole,
	}
	
	c.HTML(http.StatusOK, "dashboard-status.html", gin.H{
		"LoggedIn":    loggedIn,
		"CurrentUser": currentUser,
		"HasRole":     hasRole,
	})
}

// Handler for login page
func LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// Handler for features by tag (HTML fragment)
func FeaturesByTagFragment(featureHandler *handlers.FeatureHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		tag := c.Query("tag")
		if tag == "" {
			c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": []interface{}{}, "TagFilter": tag})
			return
		}
		features, err := featureHandler.GetFeaturesByTagName(tag)
		if err != nil || len(features) == 0 {
			c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": []interface{}{}, "TagFilter": tag})
			return
		}
		c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": features, "TagFilter": tag})
	}
}

func main() {
	// --- Load environment variables from .env file at startup ---
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found (this is OK in production if env vars are set)")
	}

	// --- Check for required GitHub token ---
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("GITHUB_TOKEN is not set. Please add it to your .env file or environment.")
	}

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

	// Run upgrade to ensure releases have project-scoped unique tags BEFORE AutoMigrate
	if err := db.MigrateReleasesToProjectScoped(); err != nil {
		panic("failed to migrate releases to project-scoped tags: " + err.Error())
	}

	// Migrate all schemas
	if err := db.Migrate(
		&models.User{}, &models.Project{}, &models.Feature{}, &models.SubFeature{},
		&models.Task{}, &models.FeatureTag{}, &models.TaskAttachment{}, &models.Comment{},
		&models.PullRequest{},
		&models.ApprovedPR{},
	); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Create repositories
	userRepo := repositories.NewUserRepository(db.DB)
	projectRepo := repositories.NewProjectRepository(db.DB)
	featureRepo := repositories.NewFeatureRepository(db.DB)
	taskRepo := repositories.NewTaskRepository(db.DB)
	tagRepo := repositories.NewTagRepository(db.DB)
	attachmentRepo := repositories.NewTaskAttachmentRepository(db.DB, sqliteFS)
	commentRepo := repositories.NewCommentRepository(db.DB)
	prRepo := repositories.NewPullRequestRepository(db.DB)

	// Create handlers
	userHandler := handlers.NewUserHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo)
	featureHandler := handlers.NewFeatureHandler(featureRepo, tagRepo, taskRepo, db.DB)
	taskHandler := handlers.NewTaskHandler(taskRepo, db.DB)
	attachmentHandler := handlers.NewTaskAttachmentHandler(attachmentRepo, sqliteFS)
	commentHandler := handlers.NewCommentHandler(commentRepo, attachmentRepo)
	prHandler := handlers.NewPullRequestHandler(prRepo)

	router := gin.Default()

	// Add dict function to the template FuncMap BEFORE LoadHTMLGlob
	router.SetFuncMap(template.FuncMap{
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, _ := values[i].(string)
				dict[key] = values[i+1]
			}
			return dict
		},
		"toJson": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"float64": func(i interface{}) float64 {
			switch v := i.(type) {
			case int:
				return float64(v)
			case int64:
				return float64(v)
			case float64:
				return v
			default:
				return 0
			}
		},
		"hasPrefix": strings.HasPrefix,
	})

	router.LoadHTMLGlob("templates/*")

	// Serve static files
	router.Static("/static", "./static")

	// Configure multipart form handling
	router.MaxMultipartMemory = 8 << 20 // 8 MiB

	// Add logging middleware
	router.Use(LoggingMiddleware())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		// Get the origin from the request
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// Default to localhost if no origin is provided
			origin = "http://localhost:8080"
		}

		// Set the requesting origin instead of wildcard
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// Add Vary header to indicate that response varies based on Origin
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	// Register auth routes
	routes.RegisterAuthRoutes(router, db.DB)

	// Create role middleware with database instance
	roleMiddleware := middleware.CreateRoleMiddleware(db.DB)

	// --- EXISTING API ROUTES ---
	api := router.Group("/api")
	{
		api.Use(BodyLoggingMiddleware())

		// Public routes (no auth required)
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("", userHandler.GetAllUsers)
		}

		// Protected routes (auth required)
		authApi := api.Group("/", middleware.AuthMiddleware())
		authApi.Use(roleMiddleware()) // Apply role middleware to all authenticated routes

		projectRoutes := authApi.Group("/projects")
		{
			projectRoutes.GET("", projectHandler.GetAllProjects)
		}

		// Regular authenticated routes
		authApi.GET("/features/project/:project_id", featureHandler.GetProjectFeatures)
		authApi.POST("/tasks/:task_id/attachments", attachmentHandler.UploadAttachment)
		authApi.GET("/attachments/file/:filename", attachmentHandler.ServeAttachment)
		authApi.DELETE("/attachments/:id", attachmentHandler.DeleteAttachment)
		authApi.POST("/tasks/:task_id/comments", commentHandler.CreateComment)
		authApi.GET("/tasks/:task_id/comments", commentHandler.GetTaskComments)
		authApi.GET("/attachments/:attachment_id/comments", commentHandler.GetAttachmentComments)
		authApi.PUT("/comments/:comment_id", commentHandler.UpdateComment)
		authApi.DELETE("/comments/:comment_id", commentHandler.DeleteComment)
		authApi.POST("/pr", handlers.PRUploadAPIHandler)
		authApi.GET("/pr", handlers.PRListAPIHandler)
		authApi.GET("/prs/:id", handlers.PRGetByIDHandler)

		prReviewHandler := handlers.NewPRReviewHandler(prRepo)
		authApi.POST("/pr/:id/review", prReviewHandler.ReviewPR)
	}

	// Register release routes
	routes.RegisterReleaseRoutes(router, db.DB)

	// --- NEW WEB ROUTES FOR HTMX ---
	// Create app handler for web routes
	appHandler := NewAppHandler(db)

	web := router.Group("/web")
	{
		// Public web routes
		web.GET("/dashboard", appHandler.RenderAppShell)
		web.GET("/login", LoginPageHandler)

		// Protected web routes
		authWeb := web.Group("/", middleware.AuthMiddleware())
		authWeb.Use(roleMiddleware()) // Apply role middleware to all authenticated web routes
		{
			authWeb.GET("/projects", appHandler.RenderAppShell)
			authWeb.GET("/projects/:id", appHandler.RenderAppShell)
			authWeb.GET("/projects/:id/features", appHandler.RenderAppShell)
			authWeb.GET("/projects/:id/features/:featureid", featureHandler.GetFeature)
			authWeb.GET("/projects/:id/features/content", featureHandler.FeaturesContentHandler)
			authWeb.GET("/projects/:id/features/new", featureHandler.NewFeatureForm)
			authWeb.POST("/projects/:id/features", featureHandler.CreateFeatureForProject)
			authWeb.GET("/features/:id/edit-inline", featureHandler.EditFeatureInline)
			authWeb.GET("/features/:id/tasks", taskHandler.GetTasksByFeature)
			authWeb.GET("/features/:id/tasks/new", taskHandler.NewTaskForm)
			authWeb.POST("/features/:id/tasks", taskHandler.CreateTaskForFeature)
			authWeb.GET("/tasks/cancel", taskHandler.CancelTaskForm)
			authWeb.GET("/features/cancel", func(c *gin.Context) { c.String(200, "") })
		}
	}

	// Fragments routes - most require authentication
	fragments := web.Group("/fragments")
	fragments.Use(FragmentHTMXGuard())
	{
		// Public fragment routes
		fragments.GET("/dashboard-status", DashboardStatusFragment)

		// Protected fragment routes
		authFragments := fragments.Group("/", middleware.AuthMiddleware())
		authFragments.Use(roleMiddleware()) // Apply role middleware to all authenticated fragment routes
		{
			// THIS IS THE NEWLY ADDED ROUTE
			authFragments.GET("/projects", projectHandler.GetAllProjectsFragment)

			authFragments.GET("/projects/:id/features", featureHandler.GetProjectFeatures)
			authFragments.GET("/projects/:id/features/new", featureHandler.NewFeatureForm)
			authFragments.GET("/features/:id", featureHandler.GetFeature)
			authFragments.GET("/features/:id/tasks/:task_id/edit", taskHandler.EditTaskForm)
			authFragments.POST("/features/:id/tasks/:task_id/edit", taskHandler.UpdateTaskInline)
			authFragments.GET("/features/:id/tasks/:task_id/view", taskHandler.ViewTaskCard)
			authFragments.GET("/features/:id/view", featureHandler.ViewFeatureCard)
			authFragments.GET("/features/cancel", func(c *gin.Context) { c.String(200, "") })
			authFragments.POST("/features/:id/edit-inline", featureHandler.UpdateFeatureInline)
			authFragments.GET("/tags/autocomplete", featureHandler.TagAutocomplete)
			authFragments.GET("/features", FeaturesByTagFragment(featureHandler))
		}
	}

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Serve the minimal test upload page for debugging
	router.StaticFile("/test-upload", "./templates/task-attachments-test.html")

	// Register PR routes
	prGroup := router.Group("/prs", middleware.AuthMiddleware())
	prGroup.Use(roleMiddleware()) // Apply role middleware to all PR routes
	{
		prGroup.POST("/:id/test", prHandler.MarkAsTested)
		// Restrict approve action to managers only
		prGroup.POST("/:id/approve", roleMiddleware("manager"), prHandler.ApprovePR)
	}

	// Debug: List all registered routes
	log.Println("Listing all registered routes:")
	routesFile, err := os.Create("routes.txt")
	if err != nil {
		log.Printf("Error creating routes file: %v\n", err)
	} else {
		defer routesFile.Close()
		fmt.Fprintln(routesFile, "Registered Routes:")
		for _, route := range router.Routes() {
			fmt.Fprintf(routesFile, "Method: %s, Path: %s\n", route.Method, route.Path)
		}
	}

	// Start server
	log.Println("Server starting on :8080...")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		panic("failed to start server: " + err.Error())
	}
}

// BodyLoggingMiddleware logs the request body
func BodyLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				log.Printf("Error reading request body: %v", err)
			} else {
				log.Printf("Request Body: %s", string(bodyBytes))
			}
			// Restore the body for subsequent handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		c.Next()
	}
}
