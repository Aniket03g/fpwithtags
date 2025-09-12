package handlers

import (
	"net/http"
	"strconv"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/services"
	"github.com/gin-gonic/gin"
)

// DependencyHandler handles HTTP requests for dependencies
type DependencyHandler struct {
	service *services.DependencyService
}

// NewDependencyHandler creates a new dependency handler
func NewDependencyHandler(service *services.DependencyService) *DependencyHandler {
	return &DependencyHandler{service: service}
}

// CreateDependency handles POST /api/dependencies
func (h *DependencyHandler) CreateDependency(c *gin.Context) {
	var dependency models.Dependency
	if err := c.ShouldBindJSON(&dependency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		dependency.CreatedByID = userID.(uint)
	}

	if err := h.service.CreateDependency(&dependency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	// Render the dependency panels HTML fragment
	c.HTML(http.StatusOK, "dependency_panels.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"Blocking":   blocking,
		"BlockedBy":  blockedBy,
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

	// Render the dependency panels HTML fragment
	c.HTML(http.StatusOK, "dependency_panels.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"Blocking":   blocking,
		"BlockedBy":  blockedBy,
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

	// Render the dependency modal HTML fragment
	c.HTML(http.StatusOK, "dependency_modal.html", gin.H{
		"EntityType": entityType,
		"EntityID":   entityID,
		"ModalType":  modalType,
	})
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
