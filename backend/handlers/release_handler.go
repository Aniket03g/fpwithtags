package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReleaseHandler struct {
	releaseRepo      repositories.ReleaseRepository
	releaseValidator *services.ReleaseValidator
}

// FinalizeRelease handles finalizing a release by creating a Git branch, tag, and cherry-picking commits
// @Summary Finalize a release
// @Description Verify release is in draft state, create Git branch and tag, cherry-pick commits, and update status
// @Tags releases
// @Accept json
// @Produce json
// @Param id path int true "Release ID"
// @Success 200 {object} map[string]string "Success response"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Release not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases/{id}/finalize [post]
func (h *ReleaseHandler) FinalizeRelease(c *gin.Context) {
	logPrefix := "[FinalizeRelease]"
	log.Printf("%s [%s] Start release finalization for request. URL Param: id=%s", logPrefix, time.Now().Format(time.RFC3339), c.Param("id"))
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	log.Printf("%s [%s] Parsed releaseID: %v", logPrefix, time.Now().Format(time.RFC3339), releaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release ID"})
		return
	}

	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	log.Printf("%s [%s] Fetching release by ID: %v", logPrefix, time.Now().Format(time.RFC3339), releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		return
	}

	// Verify release is in draft state
	log.Printf("%s [%s] Release status: %v", logPrefix, time.Now().Format(time.RFC3339), release.Status)
	if release.Status != models.ReleaseStatusDraft {
		log.Printf("%s [%s] Release not in draft state, cannot finalize", logPrefix, time.Now().Format(time.RFC3339))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft releases can be finalized"})
		return
	}

	// Extract PR IDs from the release
	var prIDs []int
	for _, pr := range release.PRs {
		prIDs = append(prIDs, int(pr.ID))
	}
	log.Printf("%s [%s] Extracted PR IDs for release: %v", logPrefix, time.Now().Format(time.RFC3339), prIDs)

	// Validate that all PRs belong to the same project
	log.Printf("%s [%s] Validating all PRs are from the same project", logPrefix, time.Now().Format(time.RFC3339))
	sameProject, err := h.releaseRepo.CheckPRsSameProject(prIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs: " + err.Error()})
		return
	}
	if !sameProject {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: PRs belong to different projects."})
		return
	}

	// Validate that no PR is already part of another release
	log.Printf("%s [%s] Checking for PRs already in other releases", logPrefix, time.Now().Format(time.RFC3339))
	prsAvailable, conflictingPRs, err := h.releaseRepo.CheckPRsNotInOtherReleases(releaseID, prIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs: " + err.Error()})
		return
	}
	if !prsAvailable {
		// Format the list of conflicting PR IDs
		conflictingPRsStr := fmt.Sprintf("%v", conflictingPRs)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: PR(s) " + conflictingPRsStr + " already belong to another release."})
		return
	}

	// Create a client to use the shared package
	client := featureplus.NewClient("http://localhost:8080", http.DefaultClient) // Local server URL

	// Use the shared package to finalize the release
	log.Printf("%s [%s] Calling shared package to finalize release: %d", logPrefix, time.Now().Format(time.RFC3339), releaseID)
	if err := client.FinalizeRelease(releaseID); err != nil {
		log.Printf("%s [%s] ERROR: Failed to finalize release: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize release: " + err.Error()})
		return
	}

	// Update release status to published
	log.Printf("%s [%s] Updating release status to published for releaseID: %d", logPrefix, time.Now().Format(time.RFC3339), releaseID)
	if err := h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusPublished); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update release status"})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Get the updated release for the template
		updatedRelease, err := h.releaseRepo.GetByID(releaseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Release finalized but failed to get updated data"})
			return
		}

		// Return the updated release row fragment for HTMX
		c.HTML(http.StatusOK, "_release_row.html", gin.H{
			"Release":     updatedRelease,
			"CurrentUser": c.MustGet("user"),
		})
		return
	}

	// For non-HTMX requests, return JSON
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"tag":    release.Tag,
	})
}

func NewReleaseHandler(releaseRepo repositories.ReleaseRepository, releaseValidator *services.ReleaseValidator) *ReleaseHandler {
	return &ReleaseHandler{
		releaseRepo:      releaseRepo,
		releaseValidator: releaseValidator,
	}
}

// getDB attempts to extract the underlying *gorm.DB from the repository
func getDB(repo repositories.ReleaseRepository) *gorm.DB {
	// Use type assertion to our concrete type which exposes DB()
	type dbProvider interface{ DB() *gorm.DB }
	if p, ok := repo.(dbProvider); ok {
		return p.DB()
	}
	return nil
}

// parseUintParam parses a string parameter from the URL into a uint
func parseUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

type CreateReleaseRequest struct {
	Tag   string `json:"tag" binding:"required"`
	PRs   []int  `json:"prs,omitempty"`
	PRIDs []uint `json:"pr_ids,omitempty" binding:"omitempty"`
	Notes string `json:"notes"`
}

// GetReleases handles retrieving all releases or releases for a specific project
// @Summary Get releases
// @Description Get all releases or releases for a specific project
// @Tags releases
// @Accept json
// @Produce json
// @Param project_id query int false "Project ID"
// @Success 200 {array} models.Release "List of releases"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases [get]
func (h *ReleaseHandler) GetReleases(c *gin.Context) {
	// Create a client to use the shared package
	client := featureplus.NewClient("http://localhost:8080", http.DefaultClient) // Local server URL

	// Check if project_id is provided as a query parameter
	projectIDStr := c.Query("project_id")

	if projectIDStr != "" {
		// Validate project_id is a valid uint
		_, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		// Call the shared package to list all releases
		// We'll filter by project ID later in our handler
		_, err = client.ListReleases()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch releases: " + err.Error()})
			return
		}
	}

	// Get all releases from the database
	// We still use the database directly since we need the full release details with PRs
	dbReleases, err := h.releaseRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch releases"})
		return
	}

	// Filter by project ID if needed
	var filteredReleases []models.Release
	if projectIDStr != "" {
		projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)
		for _, release := range dbReleases {
			if uint(release.ProjectID) == uint(projectID) {
				filteredReleases = append(filteredReleases, release)
			}
		}
		c.JSON(http.StatusOK, filteredReleases)
	} else {
		// Return all releases
		c.JSON(http.StatusOK, dbReleases)
	}
}

// CreateRelease handles the creation of a new release
// @Summary Create a new release
// @Description Create a new draft release with the specified tag, PRs, and notes
// @Tags releases
// @Accept json
// @Produce json
// @Param request body CreateReleaseRequest true "Release creation request"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases [post]
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	logPrefix := "[CreateRelease]"
	log.Printf("%s Request received with content type: %s", logPrefix, c.ContentType())
	log.Printf("%s Headers: %v", logPrefix, c.Request.Header)
	
	// Create a request struct
	var req CreateReleaseRequest
	
	// Try to bind based on content type
	contentType := c.ContentType()
	var err error
	
	if strings.Contains(contentType, "application/json") {
		// Parse as JSON
		log.Printf("%s Parsing as JSON", logPrefix)
		err = c.ShouldBindJSON(&req)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Parse as form data
		log.Printf("%s Parsing as form data", logPrefix)
		
		// First bind the form data
		if err = c.ShouldBind(&req); err != nil {
			log.Printf("%s Form binding error: %v", logPrefix, err)
			
			// Try to manually extract form values
			log.Printf("%s Attempting manual form extraction", logPrefix)
			req.Tag = c.PostForm("tag")
			if req.Tag == "" {
				req.Tag = c.PostForm("version") // Try alternate field name
			}
			req.Notes = c.PostForm("notes")
			
			// Try to extract PR IDs
			prIdsStr := c.PostFormArray("prs[]")
			if len(prIdsStr) == 0 {
				prIdsStr = c.PostFormArray("prs")
			}
			
			for _, idStr := range prIdsStr {
				id, convErr := strconv.Atoi(idStr)
				if convErr == nil {
					req.PRs = append(req.PRs, id)
				}
			}
			
			// Check if we have the minimum required data
			if req.Tag == "" || len(req.PRs) == 0 {
				log.Printf("%s Manual extraction failed to get required fields", logPrefix)
				err = errors.New("missing required fields: tag and prs")
			} else {
				// We successfully extracted the data manually
				err = nil
			}
		}
	} else {
		// Unknown content type
		log.Printf("%s Unsupported content type: %s", logPrefix, contentType)
		err = errors.New("unsupported content type")
	}
	
	if err != nil {
		log.Printf("%s Binding error: %v", logPrefix, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	
	// Log the parsed request
	log.Printf("%s Parsed request: Tag=%s, Notes=%s", logPrefix, req.Tag, req.Notes)
	
	// Handle both PR field formats (prs from UI, pr_ids from CLI)
	prIDs := make([]int, 0)
	if len(req.PRs) > 0 {
		log.Printf("%s Using PRs field: %v", logPrefix, req.PRs)
		prIDs = req.PRs
	} else if len(req.PRIDs) > 0 {
		log.Printf("%s Using PRIDs field: %v", logPrefix, req.PRIDs)
		// Convert uint to int for database operations
		for _, id := range req.PRIDs {
			prIDs = append(prIDs, int(id))
		}
	} else {
		log.Printf("%s No PR IDs provided in request", logPrefix)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No PR IDs provided"})
		return
	}
	
	// Perform validations using the repository directly
	// Derive ProjectID from PRs' features to scope the release per project
	sameProject, err := h.releaseRepo.CheckPRsSameProject(prIDs)
	if err != nil {
		log.Printf("%s Failed to validate PRs: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs: " + err.Error()})
		return
	}
	if !sameProject {
		log.Printf("%s PRs belong to different projects", logPrefix)
		c.JSON(http.StatusBadRequest, gin.H{"error": "All PRs in a release must belong to the same project"})
		return
	}

	// Fetch one PR to compute the ProjectID via its Feature
	if len(prIDs) == 0 {
		log.Printf("%s No PR IDs available for validation", logPrefix)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid PR IDs provided"})
		return
	}
	
	var firstPRID = prIDs[0]
	log.Printf("%s Using first PR ID for project validation: %d", logPrefix, firstPRID)
	
	var pr models.PullRequest
	db := getDB(h.releaseRepo)
	if db == nil {
		log.Printf("%s Failed to get database connection", logPrefix)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal repository error"})
		return
	}
	if err := db.Where("id = ?", firstPRID).First(&pr).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load PR: " + err.Error()})
		return
	}
	var feature models.Feature
	if err := db.Where("id = ?", pr.FeatureID).First(&feature).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load Feature: " + err.Error()})
		return
	}
	projectID := feature.ProjectID

	// Check if tag already exists within this project
	existingRelease, err := h.releaseRepo.GetByTag(uint(projectID), req.Tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing releases"})
		return
	}
	if existingRelease != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A release with this tag already exists"})
		return
	}

	// Verify all PRs exist
	prsExist, err := h.releaseRepo.PRsExist(prIDs)
	if err != nil {
		log.Printf("%s Failed to check if PRs exist: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs"})
		return
	}
	if !prsExist {
		log.Printf("%s One or more PRs do not exist", logPrefix)
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more PRs do not exist"})
		return
	}

	// Validate the PRs before creating the release
	if err := h.releaseValidator.ValidatePRsForRelease(prIDs); err != nil {
		log.Printf("%s PR validation failed: %v", logPrefix, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the release directly in our database
	release := &models.Release{
		ProjectID: projectID,
		Tag:       req.Tag,
		Notes:     req.Notes,
		Status:    models.ReleaseStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Save the release
	err = h.releaseRepo.Create(release, prIDs)
	if err != nil {
		log.Printf("%s Failed to create release: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"release_id": release.ID,
		"tag":        release.Tag,
		"prs":        prIDs,
	})
}
