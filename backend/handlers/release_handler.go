package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
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

	// Create a temporary directory for Git operations
	log.Printf("%s [%s] Creating temp dir for git operations...", logPrefix, time.Now().Format(time.RFC3339))
	tempDir, err := os.MkdirTemp("", "release-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temporary directory"})
		return
	}
	defer os.RemoveAll(tempDir)
	log.Printf("%s [%s] Temp dir created: %s", logPrefix, time.Now().Format(time.RFC3339), tempDir)

	// --- Begin GitHub and Git logic ---
	if len(release.PRs) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No PRs found for this release"})
		return
	}
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GITHUB_TOKEN is required"})
		return
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	client := github.NewClient(oauth2.NewClient(ctx, ts))

	// Parse owner and repo from the first PR URL to determine the base repository
	firstPRURL := release.PRs[0].URL
	parts := strings.Split(firstPRURL, "/")
	if len(parts) < 5 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed PR URL: " + firstPRURL})
		return
	}
	owner := parts[3]
	repo := parts[4]

	// Clone the base repository
	baseRepoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	log.Printf("%s [%s] Cloning base repo: %s", logPrefix, time.Now().Format(time.RFC3339), baseRepoURL)
	cloneCmd := exec.Command("git", "clone", baseRepoURL, tempDir)
	if err := cloneCmd.Run(); err != nil {
		log.Printf("%s [%s] ERROR: Failed to clone repository. RepoURL: %s, TempDir: %s, Error: %v", logPrefix, time.Now().Format(time.RFC3339), baseRepoURL, tempDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone repository"})
		return
	}

	// **FIXED:** Create and checkout the new release branch BEFORE cherry-picking
	branchName := "release/" + release.Tag
	log.Printf("%s [%s] Creating and checking out branch: %s", logPrefix, time.Now().Format(time.RFC3339), branchName)
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName, "main")
	checkoutCmd.Dir = tempDir
	if err := checkoutCmd.Run(); err != nil {
		log.Printf("%s [%s] ERROR: Failed to create branch. Branch: %s, Error: %v", logPrefix, time.Now().Format(time.RFC3339), branchName, err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch"})
		return
	}

	// Loop through PRs, fetch details, and cherry-pick real commits
	for _, prModel := range release.PRs {
		prURL := prModel.URL
		parts := strings.Split(prURL, "/")
		if len(parts) < 7 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed PR URL: " + prURL})
			return
		}
		prNumStr := parts[6]
		prNum, err := strconv.Atoi(prNumStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PR number in URL: " + prURL})
			return
		}
		log.Printf("%s [%s] Processing PR: %s/%s #%d", logPrefix, time.Now().Format(time.RFC3339), owner, repo, prNum)

		pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNum)
		if err != nil {
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PR details for PR #" + prNumStr + ": " + err.Error()})
			return
		}

		isMerged := pr.GetMerged()
		log.Printf("%s [%s] PR #%d merged: %v", logPrefix, time.Now().Format(time.RFC3339), prNum, isMerged)
		var shas []string

		if isMerged {
			mergeSHA := pr.GetMergeCommitSHA()
			shas = []string{mergeSHA}
			log.Printf("%s [%s] PR #%d merge commit: %s", logPrefix, time.Now().Format(time.RFC3339), prNum, mergeSHA[:8])
		} else {
			// Logic for unmerged PRs
			baseSHA := pr.GetBase().GetSHA()
			headSHA := pr.GetHead().GetSHA()
			log.Printf("%s [%s] PR #%d base SHA: %s, head SHA: %s", logPrefix, time.Now().Format(time.RFC3339), prNum, baseSHA, headSHA)

			// Fetch the head branch to make sure commits are available locally
			headRepoOwner := pr.GetHead().GetRepo().GetOwner().GetLogin()
			headRepoName := pr.GetHead().GetRepo().GetName()
			headRef := pr.GetHead().GetRef()

			fetchRemoteName := "origin"
			if headRepoOwner != owner || headRepoName != repo {
				fetchRemoteName = fmt.Sprintf("pr-%d-fork", prNum)
				remoteURL := pr.GetHead().GetRepo().GetCloneURL()
				addRemoteCmd := exec.Command("git", "remote", "add", fetchRemoteName, remoteURL)
				addRemoteCmd.Dir = tempDir
				if err := addRemoteCmd.Run(); err != nil {
					log.Printf("%s [%s] WARN: Could not add remote %s. It might already exist. Error: %v", logPrefix, time.Now().Format(time.RFC3339), fetchRemoteName, err)
				}
			}

			log.Printf("%s [%s] Fetching %s from %s", logPrefix, time.Now().Format(time.RFC3339), headRef, fetchRemoteName)
			fetchCmd := exec.Command("git", "fetch", fetchRemoteName, headRef)
			fetchCmd.Dir = tempDir
			if err := fetchCmd.Run(); err != nil {
				_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch head branch for PR #%d: %v", prNum, err)})
				return
			}

			// Use git rev-list to get commits instead of API compare
			revListCmd := exec.Command("git", "rev-list", fmt.Sprintf("%s..%s", baseSHA, headSHA))
			revListCmd.Dir = tempDir
			out, err := revListCmd.Output()
			if err != nil {
				_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to list commits for PR #%d: %v", prNum, err)})
				return
			}

			commitList := strings.Split(strings.TrimSpace(string(out)), "\n")
			// Reverse the list so commits are cherry-picked in chronological order
			for i := len(commitList) - 1; i >= 0; i-- {
				shas = append(shas, commitList[i])
			}

			if len(shas) == 0 {
				log.Printf("%s [%s] PR #%d has no new commits to cherry-pick", logPrefix, time.Now().Format(time.RFC3339), prNum)
				continue
			}
		}

		// Cherry-pick each SHA
		for _, sha := range shas {
			log.Printf("%s [%s] Cherry-picking SHA: %s", logPrefix, time.Now().Format(time.RFC3339), sha[:8])
			cherryPickCmd := exec.Command("git", "cherry-pick", sha)
			cherryPickCmd.Dir = tempDir
			if err := cherryPickCmd.Run(); err != nil {
				log.Printf("%s [%s] ERROR: Cherry-pick failed on SHA %s", logPrefix, time.Now().Format(time.RFC3339), sha[:8])
				_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("cherry-pick failed on SHA %s. Please resolve conflicts manually.", sha[:8])})
				return
			}
		}
	}

	// **FIXED:** The old, incorrect cherry-picking loop has been REMOVED.

	// Create a tag on the new release branch
	log.Printf("%s [%s] Creating tag: %s", logPrefix, time.Now().Format(time.RFC3339), release.Tag)
	tagCmd := exec.Command("git", "tag", release.Tag)
	tagCmd.Dir = tempDir
	if err := tagCmd.Run(); err != nil {
		log.Printf("%s [%s] ERROR: Failed to create tag. Tag: %s, Error: %v", logPrefix, time.Now().Format(time.RFC3339), release.Tag, err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	// Push the new branch to remote
	log.Printf("%s [%s] Pushing branch to remote: %s", logPrefix, time.Now().Format(time.RFC3339), branchName)
	pushBranchCmd := exec.Command("git", "push", "origin", branchName)
	pushBranchCmd.Dir = tempDir
	if err := pushBranchCmd.Run(); err != nil {
		log.Printf("%s [%s] ERROR: Failed to push branch. Branch: %s, Error: %v", logPrefix, time.Now().Format(time.RFC3339), branchName, err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push branch"})
		return
	}

	// Push the new tag to remote
	log.Printf("%s [%s] Pushing tag to remote: %s", logPrefix, time.Now().Format(time.RFC3339), release.Tag)
	pushTagCmd := exec.Command("git", "push", "origin", release.Tag)
	pushTagCmd.Dir = tempDir
	if err := pushTagCmd.Run(); err != nil {
		log.Printf("%s [%s] ERROR: Failed to push tag. Tag: %s, Error: %v", logPrefix, time.Now().Format(time.RFC3339), release.Tag, err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push tag"})
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

// GetAllReleases handles listing all releases
// @Summary Get all releases
// @Description Get a list of all releases
// @Tags releases
// @Produce json
// @Success 200 {array} models.Release "List of releases"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases [get]
func (h *ReleaseHandler) GetAllReleases(c *gin.Context) {
	releases, err := h.releaseRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch releases: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, releases)
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

	// Create the release
	release := &models.Release{
		ProjectID: projectID,
		Tag:       req.Tag,
		Status:    models.ReleaseStatusDraft,
		Notes:     req.Notes,
	}

	// Validate the PRs before creating the release
	if err := h.releaseValidator.ValidatePRsForRelease(req.PRs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the release
	if err := h.releaseRepo.Create(release, req.PRs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create release: " + err.Error()})
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
