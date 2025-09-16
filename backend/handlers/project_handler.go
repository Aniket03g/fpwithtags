package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	repo         *repositories.ProjectRepository
	db           *gorm.DB
	templateRepo *repositories.TemplateRepository
	featureRepo  *repositories.FeatureRepository
	taskRepo     repositories.TaskRepository
}

func NewProjectHandler(repo *repositories.ProjectRepository, db *gorm.DB) *ProjectHandler {
	templateRepo, err := repositories.NewTemplateRepository("./backend/data")
	if err != nil {
		log.Printf("Warning: Failed to load template repository: %v. Templates will not be available.", err)
	}

	return &ProjectHandler{
		repo:         repo,
		db:           db,
		templateRepo: templateRepo,
		featureRepo:  repositories.NewFeatureRepository(db),
		taskRepo:     repositories.NewTaskRepository(db),
	}
}

// ShowProjectCreateModal renders the create project modal for HTMX
func (h *ProjectHandler) ShowProjectCreateModal(c *gin.Context) {
	c.HTML(http.StatusOK, "create_project.html", gin.H{})
}

// CreateProjectFromForm handles the project creation from the HTMX modal form
func (h *ProjectHandler) CreateProjectFromForm(c *gin.Context) {
	var err error
	// Declare projectList at the beginning of the function
	var projectList []models.Project
	log.Println("DEBUG: Starting project creation from form")

	// Log all form data for debugging
	formData := c.Request.Form
	if formData == nil {
		if err := c.Request.ParseForm(); err != nil {
			log.Printf("ERROR: Failed to parse form data: %v", err)
		} else {
			formData = c.Request.Form
		}
	}
	log.Printf("DEBUG: Form data received: %+v", formData)
	// Get the current user ID from the authenticated session
	userID, exists := c.Get("user_id")
	if !exists {
		c.HTML(http.StatusUnauthorized, "create_project.html", gin.H{
			"error": "You must be logged in to create a project",
		})
		return
	}

	// Parse form data from the modal
	name := c.PostForm("name")
	description := c.PostForm("description")
	taskTypes := c.PostForm("task_types")
	featureCategories := c.PostForm("feature_categories")
	techStack := c.PostForm("tech_stack")
	templateID := c.PostForm("template_id")
	tagsEnabled := c.PostForm("tags_enabled") == "on" // Checkbox returns "on" when checked

	// Log all form values individually for debugging
	log.Printf("DEBUG: Form values - name: %s, description: %s, taskTypes: %s, featureCategories: %s, techStack: %s, templateID: %s, tagsEnabled: %v",
		name, description, taskTypes, featureCategories, techStack, templateID, tagsEnabled)

	// Basic validation
	if name == "" {
		c.HTML(http.StatusBadRequest, "create_project.html", gin.H{
			"error":       "Project name is required",
			"name":        name,
			"description": description,
		})
		return
	}

	// Create new project model
	// Convert userID from uint to int for the Project model
	userIDUint, ok := userID.(uint)
	if !ok {
		c.HTML(http.StatusInternalServerError, "create_project.html", gin.H{
			"error":       "Invalid user ID type",
			"name":        name,
			"description": description,
		})
		return
	}

	// Process task types and feature categories
	var taskTypesList []string
	var featureCategoriesList []string

	// Parse task types
	if taskTypes != "" {
		// Split by comma and trim spaces
		for _, t := range strings.Split(taskTypes, ",") {
			taskTypesList = append(taskTypesList, strings.TrimSpace(t))
		}
	} else {
		// Default task types
		taskTypesList = []string{"UI", "Dev", "DB", "Backend"}
	}

	// Parse feature categories
	if featureCategories != "" {
		// Split by comma and trim spaces
		for _, c := range strings.Split(featureCategories, ",") {
			featureCategoriesList = append(featureCategoriesList, strings.TrimSpace(c))
		}
	} else {
		// Default feature categories
		featureCategoriesList = []string{"Auth", "Payment", "Tags", "Tasks", "Features"}
	}

	// Validate tech stack and default to "Other" if empty
	if techStack == "" {
		techStack = "Other"
	}

	// Create custom config
	customConfig := models.JSONB{
		"task_types":       taskTypesList,
		"feature_category": featureCategoriesList,
		"tech_stack":       techStack,
		"tags_enabled":     tagsEnabled,
	}

	project := models.Project{
		Name:        name,
		Description: description,
		OwnerID:     int(userIDUint), // Convert uint to int
		Config:      customConfig,
	}

	// Save to database using your existing repository
	if err := h.repo.CreateProject(&project); err != nil {
		c.HTML(http.StatusInternalServerError, "create_project.html", gin.H{
			"error":       "Failed to create project",
			"name":        name,
			"description": description,
		})
		return
	}

	// Apply template if one was selected
	if templateID != "" && h.templateRepo != nil {
		log.Printf("DEBUG: Attempting to apply template %s to project %d", templateID, project.ID)

		// Check if templateRepo is properly initialized
		if h.templateRepo == nil {
			log.Printf("ERROR: Template repository is nil despite earlier check")
			return
		}

		// Get the template
		log.Printf("DEBUG: Fetching template with ID: %s", templateID)
		template, err := h.templateRepo.GetTemplateByID(templateID)
		if err != nil {
			log.Printf("ERROR: Failed to get template: %v", err)
			return
		} else if template == nil {
			log.Printf("ERROR: Template is nil but no error returned")
			return
		}

		// Log template details for debugging
		log.Printf("DEBUG: Applying template '%s' -> inserting %d features, %d tasks, %d dependencies", 
			template.Name, len(template.Features), len(template.Tasks), len(template.Dependencies))

		// Create features from template
		log.Printf("DEBUG: Starting to create %d features from template", len(template.Features))
		createdFeatures := []models.Feature{}
		featureMap := make(map[string]uint) // Map feature names to IDs for dependency creation

		for i, templateFeature := range template.Features {
			log.Printf("DEBUG: Creating feature %d/%d: %s", i+1, len(template.Features), templateFeature.Name)
			feature := models.Feature{
				Title:       templateFeature.Name,
				Description: templateFeature.Description,
				Category:    templateFeature.Category,
				ProjectID:   int(project.ID),
				Status:      models.StatusTodo,
				Priority:    models.PriorityMedium,
			}

			log.Printf("DEBUG: About to save feature to database: %+v", feature)
			if err := h.featureRepo.CreateFeature(&feature); err != nil {
				log.Printf("ERROR: Failed to create feature %s: %v", templateFeature.Name, err)
				continue
			}
			log.Printf("DEBUG: Successfully created feature with ID: %d", feature.ID)
			createdFeatures = append(createdFeatures, feature)
			
			// Store feature ID by name for dependency creation
			featureMap[templateFeature.Name] = feature.ID
		}
		log.Printf("DEBUG: Created %d/%d features successfully", len(createdFeatures), len(template.Features))

		// Create tasks from template with more intelligent feature assignment
		log.Printf("DEBUG: Starting to create %d tasks from template", len(template.Tasks))
		createdTasks := []models.Task{}
		taskMap := make(map[string]uint) // Map task names to IDs for dependency creation

		// Group tasks by type for better feature assignment
		tasksByType := make(map[string][]repositories.TemplateTask)
		for _, task := range template.Tasks {
			tasksByType[task.Type] = append(tasksByType[task.Type], task)
		}
		log.Printf("DEBUG: Grouped tasks into %d types: %v", len(tasksByType), getMapKeysGeneric(tasksByType))

		// Create a map of feature categories for quick lookup
		featureCategories := make(map[string][]models.Feature)
		for _, feature := range createdFeatures {
			featureCategories[feature.Category] = append(featureCategories[feature.Category], feature)
		}
		log.Printf("DEBUG: Available feature categories: %v", getMapKeysGeneric(featureCategories))

		for i, templateTask := range template.Tasks {
			log.Printf("DEBUG: Creating task %d/%d: %s (Type: %s)", i+1, len(template.Tasks), templateTask.Name, templateTask.Type)
			
			// Find the appropriate feature for this task based on type matching
			var featureID uint
			if len(createdFeatures) > 0 {
				// Try exact match first
				if features, exists := featureCategories[templateTask.Type]; exists && len(features) > 0 {
					featureID = features[0].ID
					log.Printf("DEBUG: Exact match - Task type '%s' with feature category '%s' (ID: %d)", 
						templateTask.Type, features[0].Category, features[0].ID)
				} else {
					// Try case-insensitive match
					for _, feature := range createdFeatures {
						if strings.EqualFold(feature.Category, templateTask.Type) {
							featureID = feature.ID
							log.Printf("DEBUG: Case-insensitive match - Task type '%s' with feature category '%s' (ID: %d)", 
								templateTask.Type, feature.Category, feature.ID)
							break
						}
					}
					
					// If still no match, try partial match
					if featureID == 0 {
						for _, feature := range createdFeatures {
							if strings.Contains(strings.ToLower(feature.Category), strings.ToLower(templateTask.Type)) ||
							   strings.Contains(strings.ToLower(templateTask.Type), strings.ToLower(feature.Category)) {
								featureID = feature.ID
								log.Printf("DEBUG: Partial match - Task type '%s' with feature category '%s' (ID: %d)", 
									templateTask.Type, feature.Category, feature.ID)
								break
							}
						}
					}
					
					// If still no match, use the first feature
					if featureID == 0 {
						featureID = createdFeatures[0].ID
						log.Printf("DEBUG: No match found, assigning task to first feature (ID: %d, Category: %s)", 
							featureID, createdFeatures[0].Category)
					}
				}
			} else {
				log.Printf("WARNING: No features available to assign task to")
			}

			task := models.Task{
				TaskName:    templateTask.Name,
				Description: templateTask.Description,
				TaskType:    templateTask.Type,
				FeatureID:   featureID,
			}

			log.Printf("DEBUG: About to save task to database: %+v", task)
			if err := h.taskRepo.Create(&task); err != nil {
				log.Printf("ERROR: Failed to create task %s: %v", templateTask.Name, err)
				continue
			}
			log.Printf("DEBUG: Successfully created task with ID: %d", task.ID)
			createdTasks = append(createdTasks, task)
			
			// Store task ID by name for dependency creation
			taskMap[templateTask.Name] = task.ID
		}
		log.Printf("DEBUG: Created %d/%d tasks successfully", len(createdTasks), len(template.Tasks))

		// Create dependencies from template
		if len(template.Dependencies) > 0 {
			log.Printf("DEBUG: Starting to create %d dependencies from template", len(template.Dependencies))
			
			// Get dependency service
			dependencyService := services.NewDependencyService(repositories.NewDependencyRepository(h.db))
			
			// Parse dependencies into a structured format
			dependencyPairs := parseDependencies(template.Dependencies)
			log.Printf("DEBUG: Parsed %d dependency pairs", len(dependencyPairs))
			
			// Create a map of feature titles to IDs for easier lookup
			featureTitleMap := make(map[string]uint)
			for _, feature := range createdFeatures {
				featureTitleMap[strings.ToLower(feature.Title)] = feature.ID
				// Also add category as a possible match
				featureTitleMap[strings.ToLower(feature.Category)] = feature.ID
			}
			
			// Create task name to ID map
			taskNameMap := make(map[string]uint)
			for _, task := range createdTasks {
				taskNameMap[strings.ToLower(task.TaskName)] = task.ID
			}
			
			createdDependencies := 0
			
			// Process each dependency pair
			for i, pair := range dependencyPairs {
				log.Printf("DEBUG: Processing dependency pair %d/%d: %s -> %s", 
					i+1, len(dependencyPairs), pair.Parent, pair.Child)
				
				// Try to find matching features or tasks for parent and child
				parentID, parentType := findEntityMatch(pair.Parent, featureTitleMap, taskNameMap)
				childID, childType := findEntityMatch(pair.Child, featureTitleMap, taskNameMap)
				
				// If we couldn't find exact matches, try partial matches
				if parentID == 0 {
					parentID, parentType = findPartialMatch(pair.Parent, createdFeatures, createdTasks)
				}
				
				if childID == 0 {
					childID, childType = findPartialMatch(pair.Child, createdFeatures, createdTasks)
				}
				
				// If we still don't have valid IDs, use default features
				if parentID == 0 && len(createdFeatures) > 0 {
					parentID = createdFeatures[0].ID
					parentType = models.EntityTypeFeature
					log.Printf("DEBUG: Using default feature (ID: %d) for parent dependency: %s", parentID, pair.Parent)
				}
				
				if childID == 0 && len(createdFeatures) > 1 {
					childID = createdFeatures[1].ID
					childType = models.EntityTypeFeature
					log.Printf("DEBUG: Using default feature (ID: %d) for child dependency: %s", childID, pair.Child)
				} else if childID == 0 && len(createdFeatures) > 0 {
					// If we only have one feature, use a task as the child
					if len(createdTasks) > 0 {
						childID = createdTasks[0].ID
						childType = models.EntityTypeTask
						log.Printf("DEBUG: Using default task (ID: %d) for child dependency: %s", childID, pair.Child)
					}
				}
				
				// Create the dependency if we have valid IDs
				if parentID > 0 && childID > 0 && parentID != childID {
					dependency := models.Dependency{
						Description: fmt.Sprintf("%s depends on %s", pair.Child, pair.Parent),
						ParentType:  parentType,
						ParentID:    parentID,
						ChildType:   childType,
						ChildID:     childID,
					}
					
					if err := dependencyService.CreateDependency(&dependency); err != nil {
						log.Printf("ERROR: Failed to create dependency %s -> %s: %v", pair.Parent, pair.Child, err)
						continue
					}
					
					log.Printf("DEBUG: Successfully created dependency with ID: %d (%s[%d] -> %s[%d])", 
						dependency.ID, parentType, parentID, childType, childID)
					createdDependencies++
				} else {
					log.Printf("WARNING: Could not create dependency %s -> %s - missing valid IDs or self-reference", 
						pair.Parent, pair.Child)
				}
			}
			
			log.Printf("INFO: Created %d/%d dependencies successfully", createdDependencies, len(dependencyPairs))
		}

		log.Printf("DEBUG: Successfully applied template %s to project %d", templateID, project.ID)
		log.Printf("INFO: Project %d created with template '%s': %d features, %d tasks, %d dependencies", 
			project.ID, template.Name, len(createdFeatures), len(createdTasks), len(template.Dependencies))
	}

	// Get all projects to refresh the list
	projectList, err = h.repo.GetAllProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "modals/create_project.html", gin.H{
			"error": "Project created but failed to refresh project list",
		})
		return
	}

	// Return two fragments:
	// 1. Close the modal
	// 2. Update the project list
	c.Header("HX-Trigger", "projectCreated")
	c.HTML(http.StatusOK, "project-list-fragment.html", gin.H{
		"Projects":   projectList,
		"NewProject": project,
	})
}

// CreateProject handles project creation (for APIs expecting JSON)
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if project.OwnerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_id is required"})
		return
	}

	if err := h.repo.CreateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetAllProjects handles getting all projects
func (h *ProjectHandler) GetAllProjects(c *gin.Context) {
	log.Println("DEBUG: Getting all projects from repository")
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		log.Printf("ERROR: Failed to get projects: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d projects from repository\n", len(projects))
	for i, p := range projects {
		log.Printf("DEBUG: Project %d: ID=%d, Name=%s, OwnerID=%d\n", i+1, p.ID, p.Name, p.OwnerID)
	}
	c.HTML(http.StatusOK, "project-list.html", gin.H{"Projects": projects})
}

// GetProject handles getting a single project
func (h *ProjectHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.repo.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.HTML(http.StatusOK, "project-detail.html", gin.H{"Project": project})
}

// UpdateProject handles project updates
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = projectID
	if err := h.repo.UpdateProject(&project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject handles project deletion
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	if err := h.repo.DeleteProject(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectsByUser handles getting projects for a specific user
func (h *ProjectHandler) GetProjectsByUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	projects, err := h.repo.GetProjectsByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) ShowDashboard(c *gin.Context) {
	projects, err := h.repo.GetAllProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "dashboard.html", gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"Projects": projects})
}

// GetAllProjectsFragment handles getting all projects for HTMX fragment requests
func (h *ProjectHandler) GetAllProjectsFragment(c *gin.Context) {
	log.Println("DEBUG: Getting all projects for fragment")

	// Get the user's role from the context (set by the AuthMiddleware)
	userRole, _ := c.Get("user_role")
	isManager := userRole == "manager"

	projects, err := h.repo.GetAllProjects()
	if err != nil {
		log.Printf("ERROR: Failed to get projects for fragment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("DEBUG: Retrieved %d projects for fragment\n", len(projects))

	// Pass both the Projects and the CurrentUser's role to the template
	c.HTML(http.StatusOK, "project-list.html", gin.H{
		"Projects": projects,
		"CurrentUser": gin.H{
			"IsManager": isManager,
		},
	})
}
