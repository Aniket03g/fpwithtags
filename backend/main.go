package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/routes"
	"github.com/jilio/sqlitefs"

	"encoding/json"
	"html/template"

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
	c.HTML(http.StatusOK, "project-list.html", gin.H{})
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

// Handler for dashboard status fragment
func DashboardStatusFragment(c *gin.Context) {
	// Example: check if user is logged in (replace with real auth logic)
	_, loggedIn := c.Get("user_id")
	c.HTML(http.StatusOK, "dashboard-status.html", gin.H{
		"LoggedIn": loggedIn,
	})
}

// Handler for login page
func LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
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
	if err := db.Migrate(
		&models.User{}, &models.Project{}, &models.Feature{}, &models.SubFeature{},
		&models.Task{}, &models.FeatureTag{}, &models.TaskAttachment{}, &models.Comment{},
		&models.PullRequest{}, // <-- add this line
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
		api.POST("/tasks/:task_id/attachments", attachmentHandler.UploadAttachment)
		// Change file serving route to avoid conflict
		api.GET("/attachments/file/:filename", attachmentHandler.ServeAttachment)
		api.DELETE("/attachments/:id", attachmentHandler.DeleteAttachment)
		api.POST("/tasks/:task_id/comments", commentHandler.CreateComment)
		api.GET("/tasks/:task_id/comments", commentHandler.GetTaskComments)
		api.GET("/attachments/:attachment_id/comments", commentHandler.GetAttachmentComments)
		api.PUT("/comments/:comment_id", commentHandler.UpdateComment)
		api.DELETE("/comments/:comment_id", commentHandler.DeleteComment)
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
		web.GET("/projects/:id/features/:featureid", featureHandler.GetFeature)
		web.GET("/projects/:id/features/content", featureHandler.FeaturesContentHandler)
		web.GET("/projects/:id/features/new", featureHandler.NewFeatureForm)
		web.POST("/projects/:id/features", featureHandler.CreateFeatureForProject)
		web.GET("/features/:id/edit-inline", featureHandler.EditFeatureInline)

		// New fragment route for project list only
		web.GET("/projects-fragment", ProjectsFragmentHandler)

		// HTMX fragment and API routes (do not change)
		web.GET("/features/:id/tasks", taskHandler.GetTasksByFeature)
		web.GET("/features/:id/tasks/new", taskHandler.NewTaskForm)
		web.POST("/features/:id/tasks", taskHandler.CreateTaskForFeature)
		web.GET("/tasks/cancel", taskHandler.CancelTaskForm)
		web.GET("/features/cancel", func(c *gin.Context) { c.String(200, "") })
	}

	// Register the PR modal route for tasks (for HTMX modal)
	router.GET("/tasks/:id/prs/new", taskHandler.NewPullRequestModal)

	// Register PR routes
	router.POST("/tasks/:id/prs", prHandler.AddPullRequest)
	router.POST("/prs/:id/mark-tested", prHandler.MarkTested)

	// New fragment group for all HTMX fragments
	fragments := router.Group("/web/fragments", FragmentHTMXGuard())
	{
		fragments.GET("/projects", projectHandler.GetAllProjects)
		fragments.GET("/projects/:id", projectHandler.GetProject)
		fragments.GET("/projects/:id/features", featureHandler.GetProjectFeatures)
		fragments.GET("/projects/:id/features/new", featureHandler.NewFeatureForm)
		fragments.GET("/features/:id", featureHandler.GetFeature)
		fragments.GET("/features/:id/tasks/:task_id/edit", taskHandler.EditTaskForm)
		fragments.POST("/features/:id/tasks/:task_id/edit", taskHandler.UpdateTaskInline)
		fragments.GET("/features/:id/tasks/:task_id/view", taskHandler.ViewTaskCard)
		fragments.GET("/features/:id/view", featureHandler.ViewFeatureCard)
		fragments.GET("/features/cancel", func(c *gin.Context) { c.String(200, "") })
		// Add inline feature edit POST route
		fragments.POST("/features/:id/edit-inline", featureHandler.UpdateFeatureInline)
		fragments.GET("/tags/autocomplete", featureHandler.TagAutocomplete)
	}
	// ==========================================================

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Serve the minimal test upload page for debugging
	router.StaticFile("/test-upload", "./templates/task-attachments-test.html")

	// Register the dashboard status fragment
	router.GET("/web/fragments/dashboard-status", DashboardStatusFragment)

	// Register the login page route
	router.GET("/web/login", LoginPageHandler)

	// Start server
	log.Println("Server starting on :8080...")
	if err := router.Run(":8080"); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
