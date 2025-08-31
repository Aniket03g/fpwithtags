package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/pkg/featureplus"
	"github.com/gin-gonic/gin"
)

type WebReleaseHandler struct {
	releaseRepo repositories.ReleaseRepository
	prRepo      repositories.PullRequestRepository
}

func NewWebReleaseHandler(releaseRepo repositories.ReleaseRepository, prRepo repositories.PullRequestRepository) *WebReleaseHandler {
	return &WebReleaseHandler{
		releaseRepo: releaseRepo,
		prRepo:      prRepo,
	}
}

// RenderReleasesList renders the releases list page
func (h *WebReleaseHandler) RenderReleasesList(c *gin.Context) {
	// Get all releases
	releases, err := h.releaseRepo.GetAll()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to fetch releases: " + err.Error(),
		})
		return
	}

	// Render the releases list template
	c.HTML(http.StatusOK, "release-list.html", gin.H{
		"Releases":    releases,
		"CurrentUser": c.GetUint("user_id"),
	})
}

// RenderReleaseDetail renders the release detail page
func (h *WebReleaseHandler) RenderReleaseDetail(c *gin.Context) {
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid release ID",
		})
		return
	}

	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Release not found",
		})
		return
	}

	// Render the release detail template
	c.HTML(http.StatusOK, "release-detail.html", gin.H{
		"Release":     release,
		"CurrentUser": c.GetUint("user_id"),
	})
}

// NewReleaseModal renders the release creation modal with selected PR IDs
func (h *WebReleaseHandler) NewReleaseModal(c *gin.Context) {
	// Get PR IDs from query parameter
	prIDsStr := c.Query("pr_ids")
	var prIDs []string

	// If PR IDs are provided, validate them
	if prIDsStr != "" {
		// Split PR IDs string into slice
		prIDsStrSlice := strings.Split(prIDsStr, ",")

		// Validate each PR ID
		for _, idStr := range prIDsStrSlice {
			_, err := strconv.Atoi(idStr)
			if err != nil {
				c.HTML(http.StatusBadRequest, "error.html", gin.H{
					"error": "Invalid PR ID: " + idStr,
				})
				return
			}
			prIDs = append(prIDs, idStr)
		}
	}

	// Render the release modal template with PR IDs
	c.HTML(http.StatusOK, "release-modal.html", gin.H{
		"PRIDs":       prIDs,
		"CurrentUser": c.GetUint("user_id"),
	})
}

// RenderReleaseRow renders a single release row for HTMX updates
func (h *WebReleaseHandler) RenderReleaseRow(c *gin.Context) {
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Invalid release ID",
		})
		return
	}

	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Release not found",
		})
		return
	}

	// Render the release row template
	c.HTML(http.StatusOK, "_release_row.html", gin.H{
		"Release":     release,
		"CurrentUser": c.GetUint("user_id"),
	})
}

// CreateRelease handles the creation of a new release from the HTMX form
func (h *WebReleaseHandler) CreateRelease(c *gin.Context) {
	logPrefix := "[WebReleaseHandler.CreateRelease]"
	log.Printf("%s [%s] Received release creation request", logPrefix, time.Now().Format(time.RFC3339))
	
	// Log detailed request information
	log.Printf("%s Request method: %s, URL: %s", logPrefix, c.Request.Method, c.Request.URL.String())
	log.Printf("%s Content-Type: %s", logPrefix, c.ContentType())
	log.Printf("%s Is HTMX request: %v", logPrefix, c.GetHeader("HX-Request") == "true")
	log.Printf("%s All headers: %v", logPrefix, c.Request.Header)
	
	// Parse form data
	if err := c.Request.ParseForm(); err != nil {
		log.Printf("%s [ERROR] Failed to parse form: %v", logPrefix, err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Failed to parse form data",
		})
		return
	}
	
	// Log all form values for debugging
	log.Printf("%s All form values: %v", logPrefix, c.Request.Form)
	
	// Extract form values
	tag := c.PostForm("version")
	notes := c.PostForm("notes")
	prIDsStr := c.PostFormArray("prs[]")
	
	log.Printf("%s Extracted form data - Tag: %s, Notes length: %d, Notes content: %s", 
		logPrefix, tag, len(notes), notes)
	log.Printf("%s PR IDs from form (raw): %v", logPrefix, prIDsStr)
	
	// Validate required fields
	if tag == "" {
		log.Printf("%s [ERROR] Missing required field: tag/version", logPrefix)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "Version is required",
		})
		return
	}
	
	// Convert PR IDs from string to uint
	var prIDs []uint
	log.Printf("%s Converting PR IDs from string to uint", logPrefix)
	for i, idStr := range prIDsStr {
		log.Printf("%s Processing PR ID[%d]: '%s'", logPrefix, i, idStr)
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			log.Printf("%s [ERROR] Invalid PR ID: %s - %v", logPrefix, idStr, err)
			c.HTML(http.StatusBadRequest, "error.html", gin.H{
				"error": fmt.Sprintf("Invalid PR ID: %s", idStr),
			})
			return
		}
		prIDs = append(prIDs, uint(id))
		log.Printf("%s Converted PR ID[%d]: %s -> %d", logPrefix, i, idStr, id)
	}
	
	// Validate we have at least one PR ID
	if len(prIDs) == 0 {
		log.Printf("%s [ERROR] No PR IDs provided", logPrefix)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "At least one PR must be selected for a release",
		})
		return
	}
	
	// Create the release request
	req := &featureplus.CreateReleaseRequest{
		Tag:   tag,
		Notes: notes,
		PRIDs: prIDs,
	}
	
	log.Printf("%s Creating release request object: %+v", logPrefix, req)
	
	// Create a client to use the shared package
	client := featureplus.NewClient("http://localhost:8080", http.DefaultClient)
	log.Printf("%s Created client with base URL: %s", logPrefix, "http://localhost:8080")
	
	// Call the shared package to create the release
	log.Printf("%s BEFORE API CALL: Calling client.CreateRelease with Tag: %s, Notes length: %d, PR IDs: %v", 
		logPrefix, req.Tag, len(req.Notes), req.PRIDs)
	
	release, err := client.CreateRelease(req)
	
	log.Printf("%s AFTER API CALL: client.CreateRelease completed", logPrefix)
	
	if err != nil {
		log.Printf("%s [ERROR] Failed to create release: %v", logPrefix, err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to create release: " + err.Error(),
		})
		return
	}
	
	log.Printf("%s Release created successfully with ID: %d, Tag: %s", logPrefix, release.ID, release.Tag)
	log.Printf("%s Full release object: %+v", logPrefix, release)
	
	// Return success response for HTMX
	log.Printf("%s Rendering success template with Tag: %s", logPrefix, release.Tag)
	c.HTML(http.StatusOK, "_release_success.html", gin.H{
		"Tag": release.Tag,
	})
}
