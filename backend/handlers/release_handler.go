package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	client := featureplus.NewClient("", http.DefaultClient) // Empty base URL for local operations

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

	// Return success response
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
	PRs   []int  `json:"prs" binding:"required,gt=0"`
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
	client := featureplus.NewClient("", http.DefaultClient) // Empty base URL for local operations

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
// @Param input body CreateReleaseRequest true "Release information"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases [post]
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	var req CreateReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate tag format (vX.Y.Z)
	if !repositories.ValidateTag(req.Tag) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag format. Must be in format vX.Y.Z"})
		return
	}

	// Create a client to use the shared package
	client := featureplus.NewClient("", http.DefaultClient) // Empty base URL for local operations

	// Convert to shared package CreateReleaseRequest
	sharedReq := &featureplus.CreateReleaseRequest{
		Tag:   req.Tag,
		Notes: req.Notes,
	}

	// Convert PR IDs to uint
	for _, prID := range req.PRs {
		sharedReq.PRIDs = append(sharedReq.PRIDs, uint(prID))
	}

	// Perform validations using the repository directly
	// Derive ProjectID from PRs' features to scope the release per project
	sameProject, err := h.releaseRepo.CheckPRsSameProject(req.PRs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs: " + err.Error()})
		return
	}
	if !sameProject {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All PRs in a release must belong to the same project"})
		return
	}

	// Fetch one PR to compute the ProjectID via its Feature
	var firstPRID = req.PRs[0]
	var pr models.PullRequest
	db := getDB(h.releaseRepo)
	if db == nil {
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
	prsExist, err := h.releaseRepo.PRsExist(req.PRs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate PRs"})
		return
	}
	if !prsExist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more PRs do not exist"})
		return
	}

	// Validate the PRs before creating the release
	if err := h.releaseValidator.ValidatePRsForRelease(req.PRs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the shared package to validate the request
	// We don't use the returned release since we're managing our own database
	_, err = client.CreateRelease(sharedReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release: " + err.Error()})
		return
	}

	// Create the release in our database
	release := &models.Release{
		ProjectID: projectID,
		Tag:       req.Tag,
		Status:    models.ReleaseStatusDraft,
		Notes:     req.Notes,
	}

	// Create the release in the database
	if err := h.releaseRepo.Create(release, req.PRs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save release to database: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"release_id": release.ID,
		"tag":        release.Tag,
		"prs":        req.PRs,
	})
}
