package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FeatureHandler struct {
	repo     *repositories.FeatureRepository
	tagRepo  *repositories.TagRepository
	taskRepo repositories.TaskRepository
	DB       *gorm.DB
}

func NewFeatureHandler(repo *repositories.FeatureRepository, tagRepo *repositories.TagRepository, taskRepo repositories.TaskRepository, db *gorm.DB) *FeatureHandler {
	return &FeatureHandler{repo: repo, tagRepo: tagRepo, taskRepo: taskRepo, DB: db}
}

type FeatureWithTags struct {
	models.Feature
	TagsInput string `json:"tags_input,omitempty"`
}

type FeatureFormInput struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Category    string `form:"category" binding:"required"`
}

func (h *FeatureHandler) CreateFeature(c *gin.Context) {
	var input FeatureFormInput
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID in URL"})
		return
	}

	feature := models.Feature{
		ProjectID:   projectID,
		Title:       input.Title,
		Description: input.Description,
		Category:    input.Category,
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	}

	// Strict validation for category against project config
	projectRepo := repositories.NewProjectRepository(h.DB)
	project, err := projectRepo.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID for category validation"})
		return
	}
	categories, ok := project.Config["feature_category"].([]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "project config missing feature_category"})
		return
	}
	validCategory := false
	for _, cat := range categories {
		if catStr, ok := cat.(string); ok && catStr == feature.Category {
			validCategory = true
			break
		}
	}
	if !validCategory {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category must be one of the allowed feature_category values in project config"})
		return
	}

	if err := h.repo.CreateFeature(&feature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the updated feature list (full block, not just inner fragment)
	categoriesIface2, ok2 := project.Config["feature_category"].([]interface{})
	categories2 := []string{}
	if ok2 {
		for _, cat := range categoriesIface2 {
			if catStr, ok := cat.(string); ok {
				categories2 = append(categories2, catStr)
			}
		}
	}
	features, _ := h.repo.GetFeaturesByProject(projectID)
	c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": features, "ProjectID": projectID, "FeatureCategories": categories2, "FilterCategory": "All"})
}

func (h *FeatureHandler) GetFeature(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		idStr = c.Param("featureid")
	}
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	feature, err := h.repo.GetFeatureByID(featureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}

	// Fetch tasks for this feature
	tasks, err := h.taskRepo.GetByFeatureID(uint(featureID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch tasks"})
		return
	}

	// Log the feature object with preloaded tags
	fmt.Printf("Fetched feature with tags in backend: %+v\n", feature)

	c.HTML(http.StatusOK, "feature-detail.html", gin.H{"Feature": feature, "Tasks": tasks})
}

func (h *FeatureHandler) GetProjectFeatures(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Fetch project to get categories
	projectRepo := repositories.NewProjectRepository(h.DB)
	project, err := projectRepo.GetProjectByID(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}
	categoriesIface, ok := project.Config["feature_category"].([]interface{})
	categories := []string{}
	if ok {
		for _, cat := range categoriesIface {
			if catStr, ok := cat.(string); ok {
				categories = append(categories, catStr)
			}
		}
	}

	// Get filter from query param
	filterCategory := c.Query("category")
	if filterCategory == "" || filterCategory == "All" {
		features, err := h.repo.GetFeaturesByProject(projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": features, "ProjectID": projectID, "FeatureCategories": categories, "FilterCategory": "All"})
		return
	}

	features, err := h.repo.GetFeaturesByProject(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filtered := []models.Feature{}
	for _, f := range features {
		if f.Category == filterCategory {
			filtered = append(filtered, f)
		}
	}
	c.HTML(http.StatusOK, "feature-list.html", gin.H{"Features": filtered, "ProjectID": projectID, "FeatureCategories": categories, "FilterCategory": filterCategory})
}

// GetSubfeatures returns all subfeatures for a given parent feature
func (h *FeatureHandler) GetSubfeatures(c *gin.Context) {
	parentIDStr := c.Param("id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent feature ID"})
		return
	}

	features, err := h.repo.GetSubfeaturesByParentID(uint(parentID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, features)
}

func (h *FeatureHandler) UpdateFeature(c *gin.Context) {
	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	var featureWithTags FeatureWithTags
	if err := c.ShouldBindJSON(&featureWithTags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract feature from the combined structure
	feature := featureWithTags.Feature

	// Convert string values to proper types
	feature.Status = models.FeatureStatus(feature.Status)
	feature.Priority = models.FeaturePriority(feature.Priority)

	if !isValidStatus(feature.Status) || !isValidPriority(feature.Priority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status or priority"})
		return
	}

	existingFeature, err := h.repo.GetFeatureByID(featureID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}

	// Update fields
	existingFeature.Title = feature.Title
	existingFeature.Description = feature.Description
	existingFeature.Status = feature.Status
	existingFeature.Priority = feature.Priority
	existingFeature.AssigneeID = feature.AssigneeID
	existingFeature.Category = feature.Category

	// Update parent feature ID if provided
	if feature.ParentFeatureID != nil {
		existingFeature.ParentFeatureID = feature.ParentFeatureID
	}

	if err := h.repo.UpdateFeature(existingFeature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Handle tags if provided
	if featureWithTags.TagsInput != "" {
		var createdByUser uint = 1 // Default to admin if not available
		if userID, exists := c.Get("user_id"); exists {
			createdByUser = userID.(uint)
		}

		err := h.tagRepo.UpdateFeatureTags(existingFeature.ID, createdByUser, featureWithTags.TagsInput)
		if err != nil {
			// Log the error but don't fail the whole request
			// We already updated the feature successfully
			c.JSON(http.StatusOK, gin.H{
				"feature": existingFeature,
				"warning": "Feature updated but failed to save tags",
			})
			return
		}

		// Fetch the feature again with its updated tags
		updatedFeature, err := h.repo.GetFeatureByID(featureID)
		if err == nil {
			c.JSON(http.StatusOK, updatedFeature)
			return
		}
	}

	c.JSON(http.StatusOK, existingFeature)
}

func (h *FeatureHandler) DeleteFeature(c *gin.Context) {
	idStr := c.Param("id")
	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	// Delete associated tags first
	h.tagRepo.DeleteTagsByFeatureID(uint(featureID))

	if err := h.repo.DeleteFeature(featureID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/features?tag=p0
func (h *FeatureHandler) GetAllFeatures(c *gin.Context) {
	var features []models.Feature
	var err error

	features, err = h.repo.GetAllFeatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, features)
}

// UpdateFeatureField updates a single field of a feature
func (h *FeatureHandler) UpdateFeatureField(c *gin.Context) {
	fmt.Printf("Received PATCH request to update feature field\n")
	fmt.Printf("Headers: %+v\n", c.Request.Header)
	fmt.Printf("Method: %s\n", c.Request.Method)

	idStr := c.Param("id")
	fmt.Printf("Feature ID: %s\n", idStr)

	featureID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Printf("Error converting ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feature ID"})
		return
	}

	var updateData struct {
		Field string      `json:"field" binding:"required"`
		Value interface{} `json:"value"`
	}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		fmt.Printf("Request body: %+v\n", c.Request.Body)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Update data: %+v\n", updateData)

	existingFeature, err := h.repo.GetFeatureByID(featureID)
	if err != nil {
		fmt.Printf("Error getting feature: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "feature not found"})
		return
	}

	fmt.Printf("Existing feature: %+v\n", existingFeature)

	// Validate and update the specific field
	switch updateData.Field {
	case "title":
		if title, ok := updateData.Value.(string); ok {
			existingFeature.Title = title
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title value"})
			return
		}
	case "description":
		if desc, ok := updateData.Value.(string); ok {
			existingFeature.Description = desc
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid description value"})
			return
		}
	case "status":
		if status, ok := updateData.Value.(string); ok {
			newStatus := models.FeatureStatus(status)
			if !isValidStatus(newStatus) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
				return
			}
			existingFeature.Status = newStatus
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
	case "priority":
		if priority, ok := updateData.Value.(string); ok {
			newPriority := models.FeaturePriority(priority)
			if !isValidPriority(newPriority) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
				return
			}
			existingFeature.Priority = newPriority
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
			return
		}
	case "category":
		if category, ok := updateData.Value.(string); ok {
			// Validate category against project config
			projectRepo := repositories.NewProjectRepository(h.DB)
			project, err := projectRepo.GetProjectByID(existingFeature.ProjectID)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID for category validation"})
				return
			}
			categories, ok := project.Config["feature_category"].([]interface{})
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "project config missing feature_category"})
				return
			}
			validCategory := false
			for _, cat := range categories {
				if catStr, ok := cat.(string); ok && catStr == category {
					validCategory = true
					break
				}
			}
			if !validCategory {
				c.JSON(http.StatusBadRequest, gin.H{"error": "category must be one of the allowed feature_category values in project config"})
				return
			}
			existingFeature.Category = category
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category value"})
			return
		}
	case "tags":
		if tags, ok := updateData.Value.(string); ok {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}
			if err := h.tagRepo.UpdateFeatureTags(existingFeature.ID, userID.(uint), tags); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tags"})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tags value"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported field"})
		return
	}

	if err := h.repo.UpdateFeature(existingFeature); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If we updated tags, fetch the feature again to include updated tags
	if updateData.Field == "tags" {
		existingFeature, err = h.repo.GetFeatureByID(featureID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch updated feature"})
			return
		}
	}

	c.JSON(http.StatusOK, existingFeature)
}

// NewFeatureForm renders the inline form for creating a new feature
func (h *FeatureHandler) NewFeatureForm(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	projectRepo := repositories.NewProjectRepository(h.DB)
	project, err := projectRepo.GetProjectByID(projectID)
	if err != nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}

	categoriesIface, ok := project.Config["feature_category"].([]interface{})
	if !ok {
		c.String(http.StatusInternalServerError, "Project config missing feature_category")
		return
	}

	// Convert []interface{} to []string
	categories := make([]string, 0, len(categoriesIface))
	for _, cat := range categoriesIface {
		if catStr, ok := cat.(string); ok {
			categories = append(categories, catStr)
		}
	}

	c.HTML(http.StatusOK, "feature-form.html", gin.H{
		"ProjectID":         projectID,
		"FeatureCategories": categories,
	})
}

// Helper functions
func isValidStatus(status models.FeatureStatus) bool {
	switch status {
	case models.StatusTodo, models.StatusInProgress, models.StatusDone:
		return true
	}
	return false
}

func isValidPriority(priority models.FeaturePriority) bool {
	switch priority {
	case models.PriorityLow, models.PriorityMedium, models.PriorityHigh:
		return true
	}
	return false
}
