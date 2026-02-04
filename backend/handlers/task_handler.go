package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type TaskHandler struct {
	taskRepo  repositories.TaskRepository
	DB        *gorm.DB
	validator *validator.Validate
}

func NewTaskHandler(taskRepo repositories.TaskRepository, db *gorm.DB) *TaskHandler {
	return &TaskHandler{
		taskRepo:  taskRepo,
		DB:        db,
		validator: validator.New(),
	}
}

// CreateTask creates a standalone task not tied to a specific feature
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate task fields
	if err := h.validator.Struct(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	task.CreatedByUser = userID.(uint)

	// Strict validation for task_type against project config
	var projectID int
	if task.FeatureID != 0 {
		// Fetch the feature to get the project ID
		featureRepo := repositories.NewFeatureRepository(h.DB)
		feature, err := featureRepo.GetFeatureByID(int(task.FeatureID))
		if err == nil {
			projectID = feature.ProjectID
		}
	}
	if projectID != 0 {
		projectRepo := repositories.NewProjectRepository(h.DB)
		project, err := projectRepo.GetProjectByID(projectID)
		if err == nil {
			// Use helper function to safely get task types
			ensureTaskProjectConfig(project)
			taskTypes := getTaskTypes(project)

			// Validate task type
			validType := false
			for _, tStr := range taskTypes {
				if tStr == task.TaskType {
					validType = true
					break
				}
			}

			if !validType {
				c.JSON(http.StatusBadRequest, gin.H{"error": "task_type must be one of the allowed task_types values in project config"})
				return
			}
		}
	}

	if err := h.taskRepo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates a standalone task by JSON input
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskRepo.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask deletes a standalone task by ID
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.taskRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete task"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// GetTask retrieves a task by its ID
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	task, err := h.taskRepo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// GetTasksByFeature lists all tasks under a specific feature
func (h *TaskHandler) GetTasksByFeature(c *gin.Context) {
	featureID, _ := strconv.Atoi(c.Param("id"))

	// Get filter parameter from query string, default to "All"
	filterType := c.Query("type")
	if filterType == "" {
		filterType = "All"
	}

	// Get the feature to access its project ID
	featureRepo := repositories.NewFeatureRepository(h.DB)
	feature, err := featureRepo.GetFeatureByID(featureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Get the project to access its config for task types
	projectRepo := repositories.NewProjectRepository(h.DB)
	project, err := projectRepo.GetProjectByID(feature.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch project configuration"})
		return
	}

	// Extract task types from project config using helper function
	ensureTaskProjectConfig(project)
	taskTypes := getTaskTypes(project)

	var tasks []models.Task

	if filterType != "All" {
		err = h.DB.Preload("Attachments").Where("feature_id = ? AND task_type = ?", featureID, filterType).Find(&tasks).Error
	} else {
		err = h.DB.Preload("Attachments").Where("feature_id = ?", featureID).Find(&tasks).Error
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks"})
		return
	}

	// Preload comments for each task and each attachment
	commentRepo := repositories.NewCommentRepository(h.DB)
	prRepo := repositories.NewPullRequestRepository(h.DB)
	for i := range tasks {
		comments, _ := commentRepo.GetByTaskID(tasks[i].ID)
		// Only general comments (AttachmentID == nil)
		var generalComments []models.Comment
		for _, cm := range comments {
			if cm.AttachmentID == nil {
				generalComments = append(generalComments, cm)
			}
		}
		tasks[i].Comments = generalComments
		// Preload comments for each attachment
		for j := range tasks[i].Attachments {
			attComments, _ := commentRepo.GetByAttachmentID(tasks[i].Attachments[j].ID)
			tasks[i].Attachments[j].Comments = attComments
		}
		// Load PullRequests for this task
		prs, _ := prRepo.GetByTaskID(tasks[i].ID)
		fmt.Printf("[DEBUG] Task %d PRs: %+v\n", tasks[i].ID, prs)
		tasks[i].PullRequests = prs
	}

	// Get user role from context (set by RoleMiddleware)
	userRole, exists := c.Get("user_role")
	var isManager bool
	if exists {
		isManager = userRole == "manager"
		log.Printf("[GetTasksByFeature] user_role from context: '%v', isManager: %v", userRole, isManager)
	}

	c.HTML(http.StatusOK, "task-list.html", gin.H{
		"Tasks":       tasks,
		"Feature":     gin.H{"ID": featureID},
		"FilterType":  filterType,
		"TaskTypes":   taskTypes,
		"CurrentUser": gin.H{"Role": userRole, "IsManager": isManager},
	})
}

// NewTaskForm serves the empty task creation form
func (h *TaskHandler) NewTaskForm(c *gin.Context) {
	// Get feature ID from URL parameter
	featureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// Get the feature to access its project ID
	featureRepo := repositories.NewFeatureRepository(h.DB)
	feature, err := featureRepo.GetFeatureByID(featureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Feature not found"})
		return
	}

	// Get the project to access its config
	projectRepo := repositories.NewProjectRepository(h.DB)
	project, err := projectRepo.GetProjectByID(feature.ProjectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch project configuration"})
		return
	}

	// Extract task types from project config using helper function
	ensureTaskProjectConfig(project)
	taskTypes := getTaskTypes(project)

	// Render the task form template
	c.HTML(http.StatusOK, "task-form.html", gin.H{
		"FeatureID": featureID,
		"TaskTypes": taskTypes,
	})
}

// CreateTaskForFeature creates a task and links it to a feature
func (h *TaskHandler) CreateTaskForFeature(c *gin.Context) {
	featureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	// userID, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	// 	return
	// }

	var task models.Task
	// Accept both JSON and form submissions
	if err := c.ShouldBind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.FeatureID = uint(featureID)
	// For testing: hardcode CreatedByUser to 1
	task.CreatedByUser = 1

	// Normalize task type to 'DB' if user submitted 'Db'
	if task.TaskType == "Db" {
		task.TaskType = "DB"
	}

	if err := h.taskRepo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create task"})
		return
	}

	// Fetch the updated list of tasks for this feature
	tasks, err := h.taskRepo.GetByFeatureID(uint(featureID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks"})
		return
	}

	// Preload comments for each task and each attachment (same as GetTasksByFeature)
	commentRepo := repositories.NewCommentRepository(h.DB)
	for i := range tasks {
		comments, _ := commentRepo.GetByTaskID(tasks[i].ID)
		// Only general comments (AttachmentID == nil)
		var generalComments []models.Comment
		for _, cm := range comments {
			if cm.AttachmentID == nil {
				generalComments = append(generalComments, cm)
			}
		}
		tasks[i].Comments = generalComments
		// Preload comments for each attachment
		for j := range tasks[i].Attachments {
			attComments, _ := commentRepo.GetByAttachmentID(tasks[i].Attachments[j].ID)
			tasks[i].Attachments[j].Comments = attComments
		}
		// Always set PullRequests to an empty slice for template rendering
		tasks[i].PullRequests = []models.PullRequest{}
	}

	// Render the updated task list
	c.HTML(http.StatusOK, "task-list.html", gin.H{
		"Tasks":      tasks,
		"Feature":    gin.H{"ID": featureID},
		"FilterType": c.DefaultQuery("type", "All"),
	})
}

// UpdateTaskForFeature updates a task that belongs to a feature
func (h *TaskHandler) UpdateTaskForFeature(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get feature ID
	featureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feature ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = uint(taskID)
	task.FeatureID = uint(featureID)

	if err := h.taskRepo.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTaskForFeature deletes a task under a feature by ID
func (h *TaskHandler) DeleteTaskForFeature(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.taskRepo.Delete(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// GetTasksBySubFeature lists all tasks under a specific sub-feature
func (h *TaskHandler) GetTasksBySubFeature(c *gin.Context) {
	subFeatureID, _ := strconv.Atoi(c.Param("id"))
	tasks, err := h.taskRepo.GetBySubFeatureID(uint(subFeatureID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// CreateTaskForSubFeature creates a task and links it to a sub-feature
func (h *TaskHandler) CreateTaskForSubFeature(c *gin.Context) {
	subFeatureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sub-feature ID"})
		return
	}

	// Get user ID from context first
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.SubFeatureID = uint(subFeatureID)
	task.CreatedByUser = userID.(uint)

	if err := h.taskRepo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTaskForSubFeature updates a task that belongs to a sub-feature
func (h *TaskHandler) UpdateTaskForSubFeature(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get sub-feature ID
	subFeatureID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sub-feature ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = uint(taskID)
	task.SubFeatureID = uint(subFeatureID)

	if err := h.taskRepo.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTaskForSubFeature deletes a task under a sub-feature by ID
func (h *TaskHandler) DeleteTaskForSubFeature(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.taskRepo.Delete(uint(taskID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

// GetAllTasks retrieves all tasks with their feature title
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.taskRepo.GetAllWithFeatureTitle()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTasksByProject retrieves all tasks for a given project ID
func (h *TaskHandler) GetTasksByProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	tasks, err := h.taskRepo.GetByProjectID(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks for project"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// CancelTaskForm returns an empty response for cancelling the inline task form
func (h *TaskHandler) CancelTaskForm(c *gin.Context) {
	c.Status(200)
}

// EditTaskForm serves the inline edit form for a task
func (h *TaskHandler) EditTaskForm(c *gin.Context) {
	featureID, _ := strconv.Atoi(c.Param("id"))
	taskID, _ := strconv.Atoi(c.Param("task_id"))
	task, err := h.taskRepo.GetByID(uint(taskID))
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	// Get task types from project config using helper function
	featureRepo := repositories.NewFeatureRepository(h.DB)
	feature, _ := featureRepo.GetFeatureByID(featureID)
	projectRepo := repositories.NewProjectRepository(h.DB)
	project, _ := projectRepo.GetProjectByID(feature.ProjectID)

	// Use helper function to safely get task types
	ensureTaskProjectConfig(project)
	taskTypes := getTaskTypes(project)
	c.HTML(http.StatusOK, "task-edit-form.html", gin.H{
		"FeatureID": featureID,
		"Task":      task,
		"TaskTypes": taskTypes,
	})
}

// UpdateTaskInline handles inline task updates
func (h *TaskHandler) UpdateTaskInline(c *gin.Context) {
	featureID, _ := strconv.Atoi(c.Param("id"))
	taskID, _ := strconv.Atoi(c.Param("task_id"))
	var task models.Task
	if err := c.ShouldBind(&task); err != nil {
		c.String(http.StatusBadRequest, "Invalid input")
		return
	}
	task.ID = uint(taskID)
	task.FeatureID = uint(featureID)
	if task.TaskType == "Db" {
		task.TaskType = "DB"
	}
	if err := h.taskRepo.Update(&task); err != nil {
		c.String(http.StatusInternalServerError, "Could not update task")
		return
	}
	// Get user role from context (set by RoleMiddleware)
	userRole, exists := c.Get("user_role")
	var isManager bool
	if exists {
		isManager = userRole == "manager"
		log.Printf("[UpdateTaskInline] user_role from context: '%v', isManager: %v", userRole, isManager)
	}

	// Get project tech stack for guidance
	var projectTechStack string = "Other"
	if task.FeatureID != 0 {
		featureRepo := repositories.NewFeatureRepository(h.DB)
		feature, err := featureRepo.GetFeatureByID(int(task.FeatureID))
		if err == nil && feature.ProjectID != 0 {
			projectRepo := repositories.NewProjectRepository(h.DB)
			project, err := projectRepo.GetProjectByID(feature.ProjectID)
			if err == nil && project.Config != nil {
				if stack, ok := project.Config["tech_stack"].(string); ok {
					projectTechStack = stack
				}
			}
		}
	}

	// Re-render the updated card (reuse the card HTML from task-list.html)
	c.HTML(http.StatusOK, "task-card.html", gin.H{
		"Task":             task,
		"FeatureID":        featureID,
		"ProjectTechStack": projectTechStack,
		"CurrentUser":      gin.H{"Role": userRole, "IsManager": isManager},
	})
}

// ViewTaskCard serves the card partial for a single task (for cancel)
func (h *TaskHandler) ViewTaskCard(c *gin.Context) {
	featureID, _ := strconv.Atoi(c.Param("id"))
	taskID, _ := strconv.Atoi(c.Param("task_id"))
	task, err := h.taskRepo.GetByID(uint(taskID))
	if err != nil {
		c.String(http.StatusNotFound, "Task not found")
		return
	}
	// Get user role from context (set by RoleMiddleware)
	userRole, exists := c.Get("user_role")
	var isManager bool
	if exists {
		isManager = userRole == "manager"
		log.Printf("[ViewTaskCard] user_role from context: '%v', isManager: %v", userRole, isManager)
	}

	// Get project tech stack for guidance
	var projectTechStack string = "Other"
	if task.FeatureID != 0 {
		featureRepo := repositories.NewFeatureRepository(h.DB)
		feature, err := featureRepo.GetFeatureByID(int(task.FeatureID))
		if err == nil && feature.ProjectID != 0 {
			projectRepo := repositories.NewProjectRepository(h.DB)
			project, err := projectRepo.GetProjectByID(feature.ProjectID)
			if err == nil && project.Config != nil {
				if stack, ok := project.Config["tech_stack"].(string); ok {
					projectTechStack = stack
				}
			}
		}
	}

	c.HTML(http.StatusOK, "task-card.html", gin.H{
		"Task":             task,
		"FeatureID":        featureID,
		"ProjectTechStack": projectTechStack,
		"CurrentUser":      gin.H{"Role": userRole, "IsManager": isManager},
	})
}

// Serve the PR modal for a task
func (h *TaskHandler) NewPullRequestModal(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid task ID")
		return
	}
	c.HTML(http.StatusOK, "pr-modal.html", gin.H{"TaskID": taskID})
}

// GetAllTasksFragment renders all tasks as an HTMX fragment for the dashboard
func (h *TaskHandler) GetAllTasksFragment(c *gin.Context) {
	log.Println("DEBUG: Getting all tasks for fragment")

	// Get the user's role from the context (set by the AuthMiddleware)
	userRole, _ := c.Get("user_role")
	isManager := userRole == "manager"

	tasks, err := h.taskRepo.GetAllWithFeatureTitle()
	if err != nil {
		log.Printf("ERROR: Failed to get tasks for fragment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d tasks for fragment\n", len(tasks))

	// Pass both the Tasks and the CurrentUser's role to the template
	c.HTML(http.StatusOK, "all-tasks-list.html", gin.H{
		"Tasks": tasks,
		"CurrentUser": gin.H{
			"IsManager": isManager,
			"Role":      userRole,
		},
	})
}
