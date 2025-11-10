package api

import (
	"net/http"
	"time"

	internalModels "github.com/FeaturePlus/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConnectProjectRequest represents the request body for connecting a project
type ConnectProjectRequest struct {
	Path string `json:"path" binding:"required"`
}

// ConnectProjectResponse represents the response for a successful connection
type ConnectProjectResponse struct {
	Status      string    `json:"status"`
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Path        string    `json:"path"`
	ConnectedAt time.Time `json:"connected_at"`
}

// ProjectStatusResponse represents the response for project connection status
type ProjectStatusResponse struct {
	Status      string     `json:"status"`
	ProjectID   string     `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	Path        string     `json:"path,omitempty"`
	ConnectedAt *time.Time `json:"connected_at,omitempty"`
}

// ConnectProjectHandler handles POST /api/projects/:id/connect
func ConnectProjectHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
			return
		}

		// Parse request body
		var req ConnectProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path is required", "details": err.Error()})
			return
		}

		// Verify project exists and get project name
		var project struct {
			ID   int    `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		if err := db.Table("projects").Select("id, name").Where("id = ?", projectID).First(&project).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify project", "details": err.Error()})
			return
		}

		// Check if connection already exists
		var existingConnection internalModels.ProjectConnection
		err := db.Where("project_id = ?", projectID).First(&existingConnection).Error
		
		if err == nil {
			// Update existing connection
			existingConnection.Path = req.Path
			existingConnection.ConnectedAt = time.Now()
			if err := db.Save(&existingConnection).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update connection", "details": err.Error()})
				return
			}

			c.JSON(http.StatusOK, ConnectProjectResponse{
				Status:      "linked",
				ProjectID:   projectID,
				ProjectName: project.Name,
				Path:        existingConnection.Path,
				ConnectedAt: existingConnection.ConnectedAt,
			})
			return
		} else if err != gorm.ErrRecordNotFound {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "details": err.Error()})
			return
		}

		// Create new connection
		connection := internalModels.ProjectConnection{
			ProjectID:   projectID,
			Path:        req.Path,
			ConnectedAt: time.Now(),
		}

		if err := db.Create(&connection).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create connection", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, ConnectProjectResponse{
			Status:      "linked",
			ProjectID:   projectID,
			ProjectName: project.Name,
			Path:        connection.Path,
			ConnectedAt: connection.ConnectedAt,
		})
	}
}

// GetProjectConnectionStatus handles GET /api/projects/:id/status
func GetProjectConnectionStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
			return
		}

		// Check if connection exists
		var connection internalModels.ProjectConnection
		err := db.Where("project_id = ?", projectID).First(&connection).Error

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, ProjectStatusResponse{
				Status: "unlinked",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error", "details": err.Error()})
			return
		}

		// Get project name
		var project struct {
			Name string `gorm:"column:name"`
		}
		if err := db.Table("projects").Select("name").Where("id = ?", projectID).First(&project).Error; err != nil {
			// If project not found, still return connection info but without name
			c.JSON(http.StatusOK, ProjectStatusResponse{
				Status:      "linked",
				ProjectID:   connection.ProjectID,
				Path:        connection.Path,
				ConnectedAt: &connection.ConnectedAt,
			})
			return
		}

		c.JSON(http.StatusOK, ProjectStatusResponse{
			Status:      "linked",
			ProjectID:   connection.ProjectID,
			ProjectName: project.Name,
			Path:        connection.Path,
			ConnectedAt: &connection.ConnectedAt,
		})
	}
}
