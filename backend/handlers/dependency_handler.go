package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DependencyHandler handles HTTP requests for dependencies
type DependencyHandler struct {
	service *services.DependencyService
	db      *gorm.DB
}

// NewDependencyHandler creates a new dependency handler
func NewDependencyHandler(service *services.DependencyService, db *gorm.DB) *DependencyHandler {
	return &DependencyHandler{service: service, db: db}
}

// CreateDependency handles POST /api/dependencies
func (h *DependencyHandler) CreateDependency(c *gin.Context) {
	var dependency models.Dependency
	
	// Log request content type and method
	fmt.Printf("[DEBUG] Content-Type: %s, Method: %s\n", c.ContentType(), c.Request.Method)
	
	// Get form data directly
	parentType := c.PostForm("parent_type")
	parentIDStr := c.PostForm("parent_id")
	childType := c.PostForm("child_type")
	childIDStr := c.PostForm("child_id")
	description := c.PostForm("description")
	
	// Log form data
	fmt.Printf("[DEBUG] Form Data - parentType: %s, parentID: %s, childType: %s, childID: %s\n", 
		parentType, parentIDStr, childType, childIDStr)
	
	// If form data is empty, try to bind JSON
	if parentType == "" && childType == "" {
		fmt.Println("[DEBUG] Form data empty, trying JSON binding")
		if err := c.ShouldBindJSON(&dependency); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Validate required fields
		if parentType == "" || parentIDStr == "" || childType == "" || childIDStr == "" {
			missing := []string{}
			if parentType == "" { missing = append(missing, "parent_type") }
			if parentIDStr == "" { missing = append(missing, "parent_id") }
			if childType == "" { missing = append(missing, "child_type") }
			if childIDStr == "" { missing = append(missing, "child_id") }
			
			fmt.Printf("[DEBUG] Missing fields: %v\n", missing)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields: " + strings.Join(missing, ", ")})
			return
		}
		
		// Parse IDs
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			fmt.Printf("[DEBUG] Invalid parent_id: %s, error: %v\n", parentIDStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent_id"})
			return
		}
		
		childID, err := strconv.ParseUint(childIDStr, 10, 32)
		if err != nil {
			fmt.Printf("[DEBUG] Invalid child_id: %s, error: %v\n", childIDStr, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid child_id"})
			return
		}
		
		// Create dependency from form data
		dependency = models.Dependency{
			ParentType:  models.EntityType(parentType),
			ParentID:    uint(parentID),
			ChildType:   models.EntityType(childType),
			ChildID:     uint(childID),
			Description: description,
		}
		fmt.Printf("[DEBUG] Created dependency: %+v\n", dependency)
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		dependency.CreatedByID = userID.(uint)
	}

	if err := h.service.CreateDependency(&dependency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// If this is an HTMX request, return the updated dependency panels
	if c.GetHeader("HX-Request") == "true" {
		// Get dependencies where this entity is the parent (blocking others)
		blocking, _ := h.service.ListDependenciesByParent(dependency.ChildType, dependency.ChildID)

		// Get dependencies where this entity is the child (blocked by others)
		blockedBy, _ := h.service.ListDependenciesByChild(dependency.ChildType, dependency.ChildID)

		// Check if entity is blocked
		isBlocked, _, _ := h.service.CheckBlocked(dependency.ChildType, dependency.ChildID)

		// Render the dependency panels HTML fragment
		c.HTML(http.StatusOK, "dependency_panels.html", gin.H{
			"EntityType": dependency.ChildType,
			"EntityID":   dependency.ChildID,
			"Blocking":   blocking,
			"BlockedBy":  blockedBy,
			"IsBlocked":  isBlocked,
		})
		return
	}

	c.JSON(http.StatusCreated, dependency)
}

// ListDependencies handles GET /api/dependencies with query parameters
func (h *DependencyHandler) ListDependencies(c *gin.Context) {
	// Check if we're listing by parent or by child
	parentType := c.Query("parent_type")
	parentIDStr := c.Query("parent_id")
	childType := c.Query("child_type")
	childIDStr := c.Query("child_id")

	// We need either parent or child parameters, but not both
	if (parentType != "" && parentIDStr != "") && (childType != "" && childIDStr != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provide either parent or child parameters, not both"})
		return
	}

	// List by parent (what this entity is blocking)
	if parentType != "" && parentIDStr != "" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent_id"})
			return
		}

		dependencies, err := h.service.ListDependenciesByParent(models.EntityType(parentType), uint(parentID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, dependencies)
		return
	}

	// List by child (what is blocking this entity)
	if childType != "" && childIDStr != "" {
		childID, err := strconv.ParseUint(childIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid child_id"})
			return
		}

		dependencies, err := h.service.ListDependenciesByChild(models.EntityType(childType), uint(childID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, dependencies)
		return
	}

	// If we get here, the request is missing required parameters
	c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required parameters"})
}

// DeleteDependency handles DELETE /api/dependencies/:id
func (h *DependencyHandler) DeleteDependency(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dependency ID"})
		return
	}

	if err := h.service.DeleteDependency(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dependency deleted successfully"})
}

// GetDependencyPanels handles GET /api/dependencies/panels
// Returns HTML fragments for dependency panels
func (h *DependencyHandler) GetDependencyPanels(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_type or entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Get dependencies where this entity is the parent (blocking others)
	blocking, err := h.service.ListDependenciesByParent(models.EntityType(entityType), uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get dependencies where this entity is the child (blocked by others)
	blockedBy, err := h.service.ListDependenciesByChild(models.EntityType(entityType), uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if entity is blocked
	isBlocked, _, _ := h.service.CheckBlocked(models.EntityType(entityType), uint(entityID))

	// Enhance dependencies with feature names
	enhancedBlockedBy := h.enhanceDependenciesWithNames(blockedBy)
	enhancedBlocking := h.enhanceDependenciesWithNames(blocking)

	// Render the dependency panels HTML fragment
	c.HTML(http.StatusOK, "dependency_panels.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"Blocking":   enhancedBlocking,
		"BlockedBy":  enhancedBlockedBy,
		"IsBlocked":  isBlocked,
	})
}

// ShowDependencyModal handles GET /api/dependencies/modal
// Returns HTML fragment for the add dependency modal
func (h *DependencyHandler) ShowDependencyModal(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	modalType := c.DefaultQuery("modal_type", "depends_on") // "depends_on" or "blocking"

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_type or entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Render the dependency modal HTML fragment
	c.HTML(http.StatusOK, "dependency_modal.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"ModalType":  modalType,
	})
}

// GetDependencyPanelsFragment handles GET /web/fragments/dependencies/panels
// Returns HTML fragment for dependency panels (no shell layout)
func (h *DependencyHandler) GetDependencyPanelsFragment(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_type or entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Get dependencies where this entity is the parent (blocking others)
	blocking, err := h.service.ListDependenciesByParent(models.EntityType(entityType), uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get dependencies where this entity is the child (blocked by others)
	blockedBy, err := h.service.ListDependenciesByChild(models.EntityType(entityType), uint(entityID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if entity is blocked
	isBlocked, _, _ := h.service.CheckBlocked(models.EntityType(entityType), uint(entityID))

	// Enhance dependencies with feature names
	enhancedBlockedBy := h.enhanceDependenciesWithNames(blockedBy)
	enhancedBlocking := h.enhanceDependenciesWithNames(blocking)

	// Render the dependency panels HTML fragment
	c.HTML(http.StatusOK, "dependency_panels.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"Blocking":   enhancedBlocking,
		"BlockedBy":  enhancedBlockedBy,
		"IsBlocked":  isBlocked,
	})
}

// ShowDependencyModalFragment handles GET /web/fragments/dependencies/modal
// Returns HTML fragment for the add dependency modal (no shell layout)
func (h *DependencyHandler) ShowDependencyModalFragment(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")
	modalType := c.DefaultQuery("modal_type", "depends_on") // "depends_on" or "blocking"

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_type or entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Get project features for the dropdown
	var projectID int
	var features []models.Feature

	// If the current entity is a feature, get its project ID
	if entityType == "feature" {
		var currentFeature models.Feature
		if err := h.db.First(&currentFeature, entityID).Error; err == nil {
			projectID = currentFeature.ProjectID
			
			// Get all features from the same project
			featureRepo := repositories.NewFeatureRepository(h.db)
			features, _ = featureRepo.GetFeaturesByProject(projectID)
		}
	}

	// Render the dependency modal HTML fragment
	c.HTML(http.StatusOK, "dependency_modal.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"ModalType":  modalType,
		"ProjectID":  projectID,
		"Features":   features,
	})
}

// ShowDependencyTypeSelector handles GET /web/fragments/dependencies/type_selector
// Returns HTML fragment for the dependency type selector modal
func (h *DependencyHandler) ShowDependencyTypeSelector(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity_type or entity_id"})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
		return
	}

	// Get project ID if entity is a feature
	var projectID int
	if entityType == "feature" {
		var currentFeature models.Feature
		if err := h.db.First(&currentFeature, entityID).Error; err == nil {
			projectID = currentFeature.ProjectID
		}
	}

	// Render the dependency type selector HTML fragment
	c.HTML(http.StatusOK, "dependency_type_selector.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"ProjectID":  projectID,
	})
}

// enhanceDependenciesWithNames adds entity names to dependencies
func (h *DependencyHandler) enhanceDependenciesWithNames(dependencies []models.Dependency) []map[string]interface{} {
	enhanced := make([]map[string]interface{}, 0, len(dependencies))
	
	for _, dep := range dependencies {
		item := map[string]interface{}{
			"ID":          dep.ID,
			"ParentType":  dep.ParentType,
			"ParentID":    dep.ParentID,
			"ChildType":   dep.ChildType,
			"ChildID":     dep.ChildID,
			"Description": dep.Description,
			"CreatedByID": dep.CreatedByID,
			"CreatedAt":   dep.CreatedAt,
		}
		
		// Add feature names for better display
		if dep.ParentType == "feature" {
			var feature models.Feature
			if err := h.db.First(&feature, dep.ParentID).Error; err == nil {
				item["ParentName"] = feature.Title
			}
		}
		
		if dep.ChildType == "feature" {
			var feature models.Feature
			if err := h.db.First(&feature, dep.ChildID).Error; err == nil {
				item["ChildName"] = feature.Title
			}
		}
		
		enhanced = append(enhanced, item)
	}
	
	return enhanced
}

// GetDependenciesListFragment handles GET /web/fragments/dependencies
// Returns HTML fragment for the dependencies list page
func (h *DependencyHandler) GetDependenciesListFragment(c *gin.Context) {
	filter := c.DefaultQuery("filter", "all")

	var dependencies []models.Dependency
	var err error

	// Apply filters if specified
	switch filter {
	case "feature":
		dependencies, err = h.service.ListDependenciesByType("feature")
	case "task":
		dependencies, err = h.service.ListDependenciesByType("task")
	case "pr":
		dependencies, err = h.service.ListDependenciesByType("pr")
	default: // "all"
		dependencies, err = h.service.ListAllDependencies()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Render the dependencies list HTML fragment
	c.HTML(http.StatusOK, "dependencies-list.html", gin.H{
		"Dependencies": dependencies,
		"Filter":       filter,
	})
}
