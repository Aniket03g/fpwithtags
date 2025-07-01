package main

import (
	"log"
	"net/http"

	"html/template"
	"strings"

	"github.com/FeaturePlus/backend/database"
	"github.com/FeaturePlus/backend/handlers"
	"github.com/FeaturePlus/backend/middleware"
	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/routes"
	"github.com/gin-gonic/gin"
	"github.com/jilio/sqlitefs"
)

func main() {
	// Initialize DB
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	sqliteFS, err := sqlitefs.NewSQLiteFS(sqlDB)
	if err != nil {
		log.Fatalf("failed to create sqlitefs: %v", err)
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
	taskRepo := repositories.NewTaskRepository(db.DB)
	commentRepo := repositories.NewCommentRepository(db.DB)
	tagRepo := repositories.NewTagRepository(db.DB)
	attachmentRepo := repositories.NewTaskAttachmentRepository(db.DB, sqliteFS)

	// Create handlers
	userHandler := handlers.NewUserHandler(userRepo)
	authHandler := handlers.NewAuthHandler(db.DB)
	projectHandler := handlers.NewProjectHandler(projectRepo)
	featureHandler := handlers.NewFeatureHandler(featureRepo, tagRepo, db.DB)
	taskHandler := handlers.NewTaskHandler(taskRepo, db.DB)
	commentHandler := handlers.NewCommentHandler(commentRepo, attachmentRepo)
	tagHandler := handlers.NewTagHandler(tagRepo, featureRepo, db.DB)
	taskAttachmentHandler := handlers.NewTaskAttachmentHandler(attachmentRepo, taskRepo, sqliteFS)

	// Initialize Gin router
	router := gin.Default()
	router.Use(middleware.LoggingMiddleware())

	// Add template functions
	router.SetFuncMap(template.FuncMap{
		"title": strings.Title,
	})

	// CORRECT WAY TO SERVE STATIC FILES
	// This serves files from `FeaturePlus/frontend/public` at the URL `/public`
	// The path is relative to the `backend` directory where `main.go` is run
	router.Static("/public", "../frontend/public")
	router.GET("/attachments/:filename/download", taskAttachmentHandler.DownloadAttachment)

	// Setup HTML rendering
	router.LoadHTMLFiles(
		"templates/layouts/base.html",
		"templates/layouts/_sidebar.html",
		"templates/login.html",
		"templates/signup.html",
		"templates/dashboard.html",
		"templates/projects.html",
		"templates/_project-card.html",
		"templates/_project-form.html",
	)

	// --- API Routes ---
	api := router.Group("/api")
	{
		routes.RegisterAuthRoutes(router, db.DB) // Original auth routes for API if needed

		userRoutes := api.Group("/users")
		{
			userRoutes.GET("", userHandler.GetAllUsers)
			userRoutes.POST("", userHandler.CreateUser)
			userRoutes.GET("/:id", userHandler.GetUser)
			userRoutes.PUT("/:id", userHandler.UpdateUser)
			userRoutes.DELETE("/:id", userHandler.DeleteUser)
		}

		protectedAPI := api.Group("")
		protectedAPI.Use(middleware.AuthMiddleware())
		{
			projectRoutes := protectedAPI.Group("/projects")
			{
				projectRoutes.POST("", projectHandler.CreateProject)
				projectRoutes.GET("", projectHandler.GetAllProjects)
				projectRoutes.GET("/:id", projectHandler.GetProject)
				projectRoutes.PUT("/:id", projectHandler.UpdateProject)
				projectRoutes.DELETE("/:id", projectHandler.DeleteProject)
				projectRoutes.GET("/user/:user_id", projectHandler.GetProjectsByUser)
			}

			featureRoutes := protectedAPI.Group("/features")
			{
				featureRoutes.POST("", featureHandler.CreateFeature)
				featureRoutes.GET("", featureHandler.GetAllFeatures)
				featureRoutes.GET("/:id", featureHandler.GetFeature)
				featureRoutes.PUT("/:id", featureHandler.UpdateFeature)
				featureRoutes.DELETE("/:id", featureHandler.DeleteFeature)
			}

			taskRoutes := protectedAPI.Group("/tasks")
			{
				taskRoutes.POST("", taskHandler.CreateTask)
				taskRoutes.GET("", taskHandler.GetAllTasks)
				taskRoutes.GET("/:task_id", taskHandler.GetTask)
				taskRoutes.PUT("/:task_id", taskHandler.UpdateTask)
				taskRoutes.DELETE("/:task_id", taskHandler.DeleteTask)
				taskRoutes.GET("/feature/:feature_id", taskHandler.GetTasksByFeature)
				taskRoutes.GET("/:task_id/attachments", taskAttachmentHandler.GetTaskAttachments)
				taskRoutes.POST("/:task_id/attachments", taskAttachmentHandler.UploadAttachment)
				taskRoutes.GET("/:task_id/comments", commentHandler.GetTaskComments)
				taskRoutes.POST("/:task_id/comments", commentHandler.CreateComment)
			}

			commentRoutes := protectedAPI.Group("/comments")
			{
				commentRoutes.PUT("/:comment_id", commentHandler.UpdateComment)
				commentRoutes.DELETE("/:comment_id", commentHandler.DeleteComment)
			}

			tagRoutes := protectedAPI.Group("/tags")
			{
				tagRoutes.GET("", tagHandler.GetAllTags)
			}

			attachmentRoutes := protectedAPI.Group("/attachments")
			{
				attachmentRoutes.DELETE("/:attachmentId", taskAttachmentHandler.DeleteAttachment)
			}
		}
		// Publicly accessible feature route
		api.GET("/features/project/:project_id", featureHandler.GetProjectFeatures)
	}

	// --- Web (HTMX) Routes ---
	web := router.Group("/web")
	{
		// Public routes
		web.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"title": "Login",
			})
		})
		web.GET("/signup", func(c *gin.Context) {
			c.HTML(http.StatusOK, "signup.html", gin.H{
				"title": "Sign Up",
			})
		})

		// Auth handlers
		web.POST("/login", authHandler.WebLogin)
		web.POST("/signup", authHandler.WebSignup)
		web.GET("/logout", authHandler.Logout)

		// Protected web routes
		protectedWeb := web.Group("")
		// protectedWeb.Use(middleware.RequireAuthWeb(userRepo)) // <--- Disabled for UI testing
		{
			protectedWeb.GET("/dashboard", projectHandler.ShowDashboard)
			protectedWeb.GET("/projects", projectHandler.GetProjectsPage)
			protectedWeb.GET("/projects/new-form", projectHandler.ShowNewProjectForm)
			// Add other protected web routes here
		}
	}

	// Redirect root to the web login page
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/web/login")
	})

	// Start server
	if err := router.Run(":8080"); err != nil {
		panic("failed to start server: " + err.Error())
	}
}
