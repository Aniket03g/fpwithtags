package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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

	// Extract PR IDs and GitHub PR numbers from the release
	// Support release-first workflow: collect PRs from both direct association and via features
	var prIDs []int
	var githubPRNumbers []int
	var allPRs []models.PullRequest
	
	// Get database connection to fetch PRs from features
	db := getDB(h.releaseRepo)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal repository error"})
		return
	}
	
	// First, add PRs directly associated with the release
	for _, pr := range release.PRs {
		allPRs = append(allPRs, pr)
	}
	log.Printf("%s [%s] Found %d PRs directly associated with release", logPrefix, time.Now().Format(time.RFC3339), len(release.PRs))
	
	// Then, fetch PRs from features that are assigned to this release
	var featurePRs []models.PullRequest
	if err := db.Joins("JOIN features ON features.id = pull_requests.feature_id").
		Where("features.release_id = ?", releaseID).
		Find(&featurePRs).Error; err != nil {
		log.Printf("%s [%s] Warning: Failed to fetch PRs from features: %v", logPrefix, time.Now().Format(time.RFC3339), err)
	} else {
		log.Printf("%s [%s] Found %d PRs from features assigned to release", logPrefix, time.Now().Format(time.RFC3339), len(featurePRs))
		// Merge PRs from features, avoiding duplicates
		prIDSet := make(map[uint]bool)
		for _, pr := range allPRs {
			prIDSet[pr.ID] = true
		}
		for _, pr := range featurePRs {
			if !prIDSet[pr.ID] {
				allPRs = append(allPRs, pr)
				prIDSet[pr.ID] = true
			}
		}
	}
	
	// Extract IDs and GitHub PR numbers from all collected PRs
	for _, pr := range allPRs {
		prIDs = append(prIDs, int(pr.ID))
		
		// Extract GitHub PR number from URL
		githubPRNumber, err := extractPRNumberFromURL(pr.URL)
		if err != nil {
			log.Printf("%s [%s] Warning: Failed to extract GitHub PR number from URL %s: %v", logPrefix, time.Now().Format(time.RFC3339), pr.URL, err)
			continue
		}
		githubPRNumbers = append(githubPRNumbers, githubPRNumber)
	}
	log.Printf("%s [%s] Total PRs to finalize: %d", logPrefix, time.Now().Format(time.RFC3339), len(allPRs))
	log.Printf("%s [%s] Extracted PR DB IDs: %v", logPrefix, time.Now().Format(time.RFC3339), prIDs)
	log.Printf("%s [%s] Extracted GitHub PR numbers: %v", logPrefix, time.Now().Format(time.RFC3339), githubPRNumbers)

	// Check if there are any PRs to finalize
	if len(allPRs) == 0 {
		log.Printf("%s [%s] No PRs found for release (neither directly nor via features)", logPrefix, time.Now().Format(time.RFC3339))
		c.JSON(http.StatusBadRequest, gin.H{"error": "No PRs found for release. Please add features with PRs or directly associate PRs before finalizing."})
		return
	}
	
	// Validate that all PRs belong to the same project (only if we have PRs)
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
	
	// Validate that no PR has unresolved dependencies
	log.Printf("%s [%s] Checking for PRs with unresolved dependencies", logPrefix, time.Now().Format(time.RFC3339))
	validator := services.NewReleaseValidator(h.releaseRepo.DB())
	if err := validator.ValidatePRsNotBlocked(allPRs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: " + err.Error()})
		return
	}
	
	// Validate that all dependencies are satisfied within the release
	log.Printf("%s [%s] Validating dependencies within release", logPrefix, time.Now().Format(time.RFC3339))
	if err := validator.ValidateDependenciesInRelease(releaseID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: " + err.Error()})
		return
	}

	// IMPLEMENT GIT OPERATIONS WORKFLOW
	// Define repository settings
	baseBranch := "main" // Default base branch, could be configurable
	
	// Extract repository URL from the first PR with a valid URL
	var repoURL string
	var validPRFound bool
	
	for _, pr := range allPRs {
		if pr.URL == "" {
			continue // Skip PRs with empty URLs
		}
		
		// Extract the repository URL (everything before "/pull/")
		pullIndex := strings.Index(pr.URL, "/pull/")
		if pullIndex == -1 {
			continue // Skip PRs with invalid URL format
		}
		
		repoURL = pr.URL[:pullIndex] + ".git"
		validPRFound = true
		break
	}
	
	if !validPRFound {
		log.Printf("%s [%s] No valid PR URL found in release", logPrefix, time.Now().Format(time.RFC3339))
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid PR URL found in release"})
		return
	}
	
	log.Printf("%s [%s] Using repo URL: %s", logPrefix, time.Now().Format(time.RFC3339), repoURL)
	
	// Verify all PRs belong to the same repository
	for _, pr := range allPRs {
		if pr.URL == "" {
			continue // Skip PRs with empty URLs
		}
		
		pullIdx := strings.Index(pr.URL, "/pull/")
		if pullIdx == -1 {
			continue // Skip PRs with invalid URL format
		}
		
		currentRepoURL := pr.URL[:pullIdx] + ".git"
		if currentRepoURL != repoURL {
			// Extract GitHub PR number for better logging
			githubPRNumber, err := extractPRNumberFromURL(pr.URL)
			prIdentifier := fmt.Sprintf("db_id=%d", pr.ID)
			if err == nil {
				prIdentifier = fmt.Sprintf("db_id=%d, github_pr=%d", pr.ID, githubPRNumber)
			}
			
			log.Printf("%s [%s] PRs belong to different repositories. Expected: %s, Found: %s for PR %s", 
				logPrefix, time.Now().Format(time.RFC3339), repoURL, currentRepoURL, prIdentifier)
			c.JSON(http.StatusBadRequest, gin.H{"error": "PRs belong to different repositories"})
			return
		}
	}
	
	// Set release status to in-progress during git operations
	log.Printf("%s [%s] Setting release status to in-progress during git operations", logPrefix, time.Now().Format(time.RFC3339))
	if err := h.releaseRepo.UpdateStatus(releaseID, "in_progress"); err != nil {
		log.Printf("%s [%s] Failed to update release status to in-progress: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update release status"})
		return
	}
	
	// 1. Create a temporary directory and clone the repository
	log.Printf("%s [%s] Creating temporary directory and cloning repository", logPrefix, time.Now().Format(time.RFC3339))
	repoDir, err := ensureRepoExists(repoURL)
	if err != nil {
		log.Printf("%s [%s] Failed to create temporary repository: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
		return
	}
	
	// Ensure we clean up the temporary directory when we're done
	defer func() {
		log.Printf("%s [%s] Cleaning up temporary directory: %s", logPrefix, time.Now().Format(time.RFC3339), repoDir)
		if err := os.RemoveAll(repoDir); err != nil {
			log.Printf("%s [%s] Warning: Failed to clean up temporary directory: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		}
	}()
	
	// 2. Checkout base branch and ensure it's up to date
	log.Printf("%s [%s] Checking out base branch: %s", logPrefix, time.Now().Format(time.RFC3339), baseBranch)
	if err := runGitCmd(repoDir, "checkout", baseBranch); err != nil {
		log.Printf("%s [%s] Failed to checkout base branch: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
		return
	}
	
	// Pull latest changes
	log.Printf("%s [%s] Pulling latest changes for branch: %s", logPrefix, time.Now().Format(time.RFC3339), baseBranch)
	if err := runGitCmd(repoDir, "pull", "origin", baseBranch); err != nil {
		log.Printf("%s [%s] Failed to pull latest changes: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
		return
	}
	
	// 3. Create a release branch if it doesn't exist
	releaseBranch := fmt.Sprintf("release/%s", release.Tag)
	log.Printf("%s [%s] Creating release branch: %s", logPrefix, time.Now().Format(time.RFC3339), releaseBranch)
	
	// Check if branch exists
	branchExists := runGitCmd(repoDir, "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", releaseBranch)) == nil
	
	if branchExists {
		// Checkout existing branch
		log.Printf("%s [%s] Branch %s already exists, checking it out", logPrefix, time.Now().Format(time.RFC3339), releaseBranch)
		if err := runGitCmd(repoDir, "checkout", releaseBranch); err != nil {
			log.Printf("%s [%s] Failed to checkout existing branch: %v", logPrefix, time.Now().Format(time.RFC3339), err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
			return
		}
	} else {
		// Create and checkout new branch
		log.Printf("%s [%s] Creating new branch %s from %s", logPrefix, time.Now().Format(time.RFC3339), releaseBranch, baseBranch)
		if err := runGitCmd(repoDir, "checkout", "-b", releaseBranch, baseBranch); err != nil {
			log.Printf("%s [%s] Failed to create new branch: %v", logPrefix, time.Now().Format(time.RFC3339), err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
			return
		}
		
		// Push the release branch to GitHub
		log.Printf("%s [%s] Pushing branch %s to origin", logPrefix, time.Now().Format(time.RFC3339), releaseBranch)
		if err := runGitCmd(repoDir, "push", "origin", releaseBranch); err != nil {
			log.Printf("%s [%s] Failed to push branch to origin: %v", logPrefix, time.Now().Format(time.RFC3339), err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
			return
		}
	}
	
	// 4. For each PR, fetch and cherry-pick its merge commit
	log.Printf("%s [%s] Processing %d PRs for cherry-picking", logPrefix, time.Now().Format(time.RFC3339), len(allPRs))
	
	for i, pr := range allPRs {
		// Extract GitHub PR number from URL
		githubPRNumber, err := extractPRNumberFromURL(pr.URL)
		if err != nil {
			log.Printf("%s [%s] Failed to extract PR number from URL %s: %v", logPrefix, time.Now().Format(time.RFC3339), pr.URL, err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to extract PR number from URL: %v", err)})
			return
		}
		
		// Log both DB ID and GitHub PR number
		log.Printf("%s [%s] Processing PR db_id=%d, github_pr=%d (%d/%d): %s", logPrefix, time.Now().Format(time.RFC3339), 
			pr.ID, githubPRNumber, i+1, len(allPRs), pr.Title)
		
		// Get PR details to find the merge commit SHA
		// For this example, we'll assume the PR URL contains the necessary info
		// In a real implementation, you might need to use GitHub API to get the merge commit SHA
		
		// Extract repo owner and name from PR URL
		parts := strings.Split(pr.URL, "/")
		if len(parts) < 5 {
			log.Printf("%s [%s] Invalid PR URL format: %s", logPrefix, time.Now().Format(time.RFC3339), pr.URL)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid PR URL format"})
			return
		}
		
		// Use the GitHub PR number for git operations
		prNumberStr := strconv.Itoa(githubPRNumber)
		log.Printf("%s [%s] Fetching GitHub PR #%s", logPrefix, time.Now().Format(time.RFC3339), prNumberStr)
		
		// Clean up any existing PR branch to ensure we get fresh content
		prBranchName := fmt.Sprintf("pr-%s", prNumberStr)
		_ = runGitCmd(repoDir, "branch", "-D", prBranchName) // Ignore errors if branch doesn't exist
		
		// Fetch the PR using GitHub PR number
		// First, ensure we're fetching the correct PR by using both the PR number and the PR title for validation
		log.Printf("%s [%s] Fetching GitHub PR #%s with title: %s", logPrefix, time.Now().Format(time.RFC3339), prNumberStr, pr.Title)
		
		// Fetch the PR branch
		if err := runGitCmd(repoDir, "fetch", "origin", fmt.Sprintf("pull/%s/head:%s", prNumberStr, prBranchName)); err != nil {
			log.Printf("%s [%s] Failed to fetch GitHub PR #%s: %v", logPrefix, time.Now().Format(time.RFC3339), prNumberStr, err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch GitHub PR #%s: %v", prNumberStr, err)})
			return
		}
		
		// Get the merge commit SHA
		// In a real implementation, you would get this from GitHub API
		// For now, we'll use the HEAD of the PR branch
		mergeCommitCmd := exec.Command("git", "rev-parse", fmt.Sprintf("pr-%s", prNumberStr))
		mergeCommitCmd.Dir = repoDir
		mergeCommitOutput, err := mergeCommitCmd.Output()
		if err != nil {
			log.Printf("%s [%s] Failed to get merge commit for GitHub PR #%s: %v", logPrefix, time.Now().Format(time.RFC3339), prNumberStr, err)
			_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get merge commit for GitHub PR #%s", prNumberStr)})
			return
		}
		
		mergeCommit := strings.TrimSpace(string(mergeCommitOutput))
		log.Printf("%s [%s] Cherry-picking commit %s for GitHub PR #%s (db_id=%d)", logPrefix, time.Now().Format(time.RFC3339), mergeCommit, prNumberStr, pr.ID)
		
		// Get commit details to verify what we're cherry-picking
		commitDetailsCmd := exec.Command("git", "log", "-1", "--name-status", "--format=%s%n%b", mergeCommit)
		commitDetailsCmd.Dir = repoDir
		commitDetails, err := commitDetailsCmd.Output()
		if err != nil {
			log.Printf("%s [%s] Warning: Failed to get commit details for %s: %v", logPrefix, time.Now().Format(time.RFC3339), mergeCommit, err)
		} else {
			log.Printf("%s [%s] Commit details for %s: %s", logPrefix, time.Now().Format(time.RFC3339), mergeCommit, strings.TrimSpace(string(commitDetails)))
		}
		
		// Verify the commit matches the PR title
		commitSubjectCmd := exec.Command("git", "log", "-1", "--format=%s", mergeCommit)
		commitSubjectCmd.Dir = repoDir
		commitSubject, err := commitSubjectCmd.Output()
		if err == nil {
			subject := strings.TrimSpace(string(commitSubject))
			log.Printf("%s [%s] Commit subject: '%s', PR title: '%s'", logPrefix, time.Now().Format(time.RFC3339), subject, pr.Title)
			
			// Check if the commit subject contains the PR title or vice versa
			if !strings.Contains(strings.ToLower(subject), strings.ToLower(pr.Title)) && 
			   !strings.Contains(strings.ToLower(pr.Title), strings.ToLower(subject)) {
				log.Printf("%s [%s] WARNING: Commit subject does not match PR title. This may indicate a mismatch.", 
					logPrefix, time.Now().Format(time.RFC3339))
				
				// If we detect a mismatch between PR title and commit subject, abort the process
				// This prevents cherry-picking the wrong content
				log.Printf("%s [%s] Aborting due to mismatch between PR title and commit subject", 
					logPrefix, time.Now().Format(time.RFC3339))
				_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("PR title '%s' does not match commit subject '%s'. This indicates the wrong PR content would be cherry-picked.", pr.Title, subject)})
				return
			}
		}
		
		// Check if commit is already in the branch before cherry-picking
		commitExistsCmd := exec.Command("git", "cherry", "HEAD", mergeCommit, "-v")
		commitExistsCmd.Dir = repoDir
		commitExistsOutput, _ := commitExistsCmd.Output()
		commitExistsStr := string(commitExistsOutput)
		
		// If the commit is already in the branch, skip cherry-picking
		if strings.Contains(commitExistsStr, mergeCommit) || strings.Contains(commitExistsStr, "duplicate") {
			log.Printf("%s [%s] Commit %s for GitHub PR #%s (db_id=%d) is already in the branch, skipping cherry-pick", 
				logPrefix, time.Now().Format(time.RFC3339), mergeCommit, prNumberStr, pr.ID)
		} else {
			// Cherry-pick the commit
			cherryPickErr := runGitCmd(repoDir, "cherry-pick", mergeCommit)
			if cherryPickErr != nil {
				// Check if this is an empty cherry-pick error
				emptyErrCheck := exec.Command("git", "diff", "--cached", "--quiet")
				emptyErrCheck.Dir = repoDir
				emptyErr := emptyErrCheck.Run()
				
				if emptyErr == nil {
					// Empty cherry-pick, commit it with --allow-empty
					log.Printf("%s [%s] Empty cherry-pick detected for GitHub PR #%s (db_id=%d), committing with --allow-empty", 
						logPrefix, time.Now().Format(time.RFC3339), prNumberStr, pr.ID)
					
					if err := runGitCmd(repoDir, "commit", "--allow-empty", "-m", fmt.Sprintf("Cherry-pick PR #%s: %s (empty)", prNumberStr, pr.Title)); err != nil {
						log.Printf("%s [%s] Failed to commit empty cherry-pick for GitHub PR #%s (db_id=%d): %v", 
							logPrefix, time.Now().Format(time.RFC3339), prNumberStr, pr.ID, err)
						_ = runGitCmd(repoDir, "cherry-pick", "--abort")
						_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to commit empty cherry-pick for GitHub PR #%s: %v", prNumberStr, err)})
						return
					}
				} else {
					// Real error, abort cherry-pick
					log.Printf("%s [%s] Cherry-pick failed for GitHub PR #%s (db_id=%d): %v", 
						logPrefix, time.Now().Format(time.RFC3339), prNumberStr, pr.ID, cherryPickErr)
					_ = runGitCmd(repoDir, "cherry-pick", "--abort")
					_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Cherry-pick failed for GitHub PR #%s: %v", prNumberStr, cherryPickErr)})
					return
				}
			}
		}
	}
	
	// 5. Create a tag for the release
	tagName := release.Tag
	log.Printf("%s [%s] Creating tag %s", logPrefix, time.Now().Format(time.RFC3339), tagName)
	
	// Check if tag already exists and delete it if it does
	checkTagCmd := exec.Command("git", "tag", "-l", tagName)
	checkTagCmd.Dir = repoDir
	checkTagOutput, _ := checkTagCmd.Output()
	if strings.TrimSpace(string(checkTagOutput)) == tagName {
		log.Printf("%s [%s] Tag %s already exists, deleting it", logPrefix, time.Now().Format(time.RFC3339), tagName)
		if err := runGitCmd(repoDir, "tag", "-d", tagName); err != nil {
			log.Printf("%s [%s] Failed to delete existing tag: %v", logPrefix, time.Now().Format(time.RFC3339), err)
			// Continue anyway, the tag creation might still succeed
		}
	}
	
	// Create the tag
	if err := runGitCmd(repoDir, "tag", "-a", tagName, "-m", fmt.Sprintf("Release %s", tagName)); err != nil {
		log.Printf("%s [%s] Failed to create tag: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
		return
	}
	
	// Push the tag to GitHub
	log.Printf("%s [%s] Pushing tag %s to origin", logPrefix, time.Now().Format(time.RFC3339), tagName)
	if err := runGitCmd(repoDir, "push", "origin", tagName); err != nil {
		log.Printf("%s [%s] Failed to push tag to origin: %v", logPrefix, time.Now().Format(time.RFC3339), err)
		_ = h.releaseRepo.UpdateStatus(releaseID, models.ReleaseStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Git operation failed: " + err.Error()})
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

		// Check if user exists in context before using MustGet
		var userData interface{}
		var userExists bool
		if userData, userExists = c.Get("user"); !userExists {
			// User not in context, proceed without it
			c.HTML(http.StatusOK, "_release_row.html", gin.H{
				"Release": updatedRelease,
			})
		} else {
			// User in context, include it
			c.HTML(http.StatusOK, "_release_row.html", gin.H{
				"Release":     updatedRelease,
				"CurrentUser": userData,
			})
		}
		return
	}

	// For non-HTMX requests, return JSON
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"tag":    release.Tag,
	})
}

// runGitCmd executes a git command in the specified directory and captures logs/errors
func runGitCmd(dir string, args ...string) error {
	logPrefix := "[GitOp]"
	
	// Log the command being executed
	cmdStr := fmt.Sprintf("git %s", strings.Join(args, " "))
	log.Printf("%s Executing: %s in directory: %s", logPrefix, cmdStr, dir)
	
	// Create the command
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	
	// Capture stdout and stderr
	stdout, err := cmd.Output()
	if err != nil {
		// Try to get stderr if available
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		
		log.Printf("%s Command failed: %s\nError: %v\nStderr: %s", logPrefix, cmdStr, err, stderr)
		return fmt.Errorf("git command failed: %s: %v: %s", cmdStr, err, stderr)
	}
	
	// Log success and output
	log.Printf("%s Command succeeded: %s\nOutput: %s", logPrefix, cmdStr, string(stdout))
	return nil
}

// ensureRepoExists creates a new temporary directory and clones the repository into it
// Returns the path to the temporary directory which must be cleaned up by the caller
func ensureRepoExists(repoURL string) (string, error) {
	logPrefix := "[EnsureRepo]"
	
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "featureplus-release-*")
	if err != nil {
		log.Printf("%s Failed to create temporary directory: %v", logPrefix, err)
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}
	log.Printf("%s Created temporary directory at: %s", logPrefix, tempDir)
	
	// Clone the repository into the temporary directory
	log.Printf("%s Cloning repository from %s to %s", logPrefix, repoURL, tempDir)
	if err := runGitCmd("", "clone", repoURL, tempDir); err != nil {
		// Clean up the temporary directory if cloning fails
		cleanupErr := os.RemoveAll(tempDir)
		if cleanupErr != nil {
			log.Printf("%s Warning: Failed to clean up temporary directory after clone failure: %v", logPrefix, cleanupErr)
		}
		return "", fmt.Errorf("failed to clone repository: %v", err)
	}
	
	return tempDir, nil
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

// extractPRNumberFromURL extracts the PR number from a GitHub PR URL
// URL format: https://github.com/<owner>/<repo>/pull/<number>
func extractPRNumberFromURL(url string) (int, error) {
	if url == "" {
		return 0, fmt.Errorf("empty PR URL")
	}
	
	// Find the position of "/pull/" in the URL
	pullIndex := strings.LastIndex(url, "/pull/")
	if pullIndex == -1 {
		return 0, fmt.Errorf("invalid PR URL format (missing /pull/): %s", url)
	}
	
	// Extract the number part after "/pull/"
	numberStr := url[pullIndex+6:] // +6 to skip "/pull/"
	
	// If there are any additional path segments or query parameters, remove them
	if slashIndex := strings.Index(numberStr, "/"); slashIndex != -1 {
		numberStr = numberStr[:slashIndex]
	}
	if queryIndex := strings.Index(numberStr, "?"); queryIndex != -1 {
		numberStr = numberStr[:queryIndex]
	}
	
	// Parse the number string to an integer
	prNumber, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PR number from URL: %s, error: %v", url, err)
	}
	
	if prNumber <= 0 {
		return 0, fmt.Errorf("invalid PR number (must be positive): %d", prNumber)
	}
	
	return prNumber, nil
}

type CreateReleaseRequest struct {
	Tag       string `form:"tag" json:"tag" binding:"required"`
	PRs       []int  `form:"prs" json:"prs,omitempty"`
	PRIDs     []uint `form:"pr_ids" json:"pr_ids,omitempty" binding:"omitempty"`
	ProjectID int    `form:"project_id" json:"project_id,omitempty"` // Required when creating release without PRs
	Notes     string `form:"notes" json:"notes"`
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
		
		// Debug: Log all form values
		if err := c.Request.ParseForm(); err == nil {
			log.Printf("%s [DEBUG] Form values: %+v", logPrefix, c.Request.PostForm)
		}
		
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
			
			// Extract project_id
			projectIDStr := c.PostForm("project_id")
			if projectIDStr != "" {
				projectID, convErr := strconv.Atoi(projectIDStr)
				if convErr == nil {
					req.ProjectID = projectID
					log.Printf("%s Manually extracted project_id: %d", logPrefix, projectID)
				} else {
					log.Printf("%s Failed to parse project_id '%s': %v", logPrefix, projectIDStr, convErr)
				}
			}
			
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
			if req.Tag == "" {
				log.Printf("%s Manual extraction failed to get required field: tag", logPrefix)
				err = errors.New("missing required field: tag")
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
	log.Printf("%s Parsed request: Tag=%s, ProjectID=%d, Notes=%s, PRs=%v", logPrefix, req.Tag, req.ProjectID, req.Notes, req.PRs)
	
	// Handle both PR field formats (prs from UI, pr_ids from CLI)
	// PRs are now optional for release-first workflow
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
		log.Printf("%s No PR IDs provided - creating release without PRs (release-first workflow)", logPrefix)
	}
	
	// Perform validations using the repository directly
	// Derive ProjectID from PRs' features to scope the release per project
	// Skip validation if no PRs provided (release-first workflow)
	var projectID int
	
	if len(prIDs) > 0 {
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
		projectID = feature.ProjectID
	} else {
		// Release-first workflow: require project_id in request
		if req.ProjectID == 0 {
			log.Printf("%s No PRs provided and no project_id specified", logPrefix)
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required when creating a release without PRs"})
			return
		}
		projectID = req.ProjectID
		log.Printf("%s Using project_id from request: %d", logPrefix, projectID)
	}

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

	// Verify all PRs exist and validate them (only if PRs provided)
	if len(prIDs) > 0 {
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

// AddFeaturesToReleaseRequest represents the request to add features to a release
type AddFeaturesToReleaseRequest struct {
	FeatureIDs []uint `form:"feature_ids" json:"feature_ids" binding:"required"`
}

// AddFeaturesToRelease handles adding existing features to a release
// @Summary Add features to a release
// @Description Assign existing features to a release for release-first planning
// @Tags releases
// @Accept json
// @Produce json
// @Param id path int true "Release ID"
// @Param request body AddFeaturesToReleaseRequest true "Feature IDs to add"
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Release not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases/{id}/features [post]
func (h *ReleaseHandler) AddFeaturesToRelease(c *gin.Context) {
	logPrefix := "[AddFeaturesToRelease]"
	log.Printf("%s Start adding features to release", logPrefix)
	
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release ID"})
		return
	}
	
	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		return
	}
	
	// Verify release is in draft state
	if release.Status != models.ReleaseStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft releases can have features added"})
		return
	}
	
	// Parse request body (support both JSON and form data)
	var req AddFeaturesToReleaseRequest
	
	// Log content type for debugging
	log.Printf("%s Content-Type: %s", logPrefix, c.ContentType())
	
	if strings.Contains(c.ContentType(), "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("%s JSON binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	} else {
		// Try form binding
		if err := c.ShouldBind(&req); err != nil {
			log.Printf("%s Form binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	}
	
	log.Printf("%s Received feature IDs: %v", logPrefix, req.FeatureIDs)
	
	if len(req.FeatureIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one feature ID is required"})
		return
	}
	
	log.Printf("%s Adding %d features to release %d", logPrefix, len(req.FeatureIDs), releaseID)
	
	// Get database connection
	db := getDB(h.releaseRepo)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal repository error"})
		return
	}
	
	// Fetch all features and validate they belong to the same project as the release
	var features []models.Feature
	if err := db.Where("id IN ?", req.FeatureIDs).Find(&features).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch features: " + err.Error()})
		return
	}
	
	if len(features) != len(req.FeatureIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "One or more features not found"})
		return
	}
	
	// Validate all features belong to the same project as the release
	for _, feature := range features {
		if feature.ProjectID != release.ProjectID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Feature %d belongs to project %d, but release belongs to project %d", 
					feature.ID, feature.ProjectID, release.ProjectID),
			})
			return
		}
	}
	
	// Update all features to set their ReleaseID
	if err := db.Model(&models.Feature{}).Where("id IN ?", req.FeatureIDs).Update("release_id", releaseID).Error; err != nil {
		log.Printf("%s Failed to update features: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add features to release: " + err.Error()})
		return
	}
	
	log.Printf("%s Successfully added %d features to release %d", logPrefix, len(req.FeatureIDs), releaseID)
	
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Fetch updated release
		updatedRelease, err := h.releaseRepo.GetByID(releaseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated release"})
			return
		}
		
		// Fetch all features for this release
		var plannedFeatures []models.Feature
		if err := db.Where("release_id = ?", releaseID).Find(&plannedFeatures).Error; err != nil {
			log.Printf("%s Failed to fetch planned features: %v", logPrefix, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch planned features"})
			return
		}
		
		// Get user context
		userRole, _ := c.Get("user_role")
		currentUser := map[string]interface{}{
			"Role":      userRole,
			"IsManager": userRole == "manager",
		}
		
		// Render the planned features list and close modal
		c.HTML(http.StatusOK, "release-planned-features-oob.html", gin.H{
			"PlannedFeatures": plannedFeatures,
			"Release":         updatedRelease,
			"CurrentUser":     currentUser,
		})
		return
	}
	
	// For non-HTMX requests, return JSON
	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"release_id":  releaseID,
		"feature_ids": req.FeatureIDs,
		"message":     fmt.Sprintf("Added %d features to release", len(req.FeatureIDs)),
	})
}

// CreateFeatureUnderReleaseRequest represents the request to create a feature under a release
type CreateFeatureUnderReleaseRequest struct {
	Title       string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description"`
	Category    string `form:"category" json:"category"`
	Priority    string `form:"priority" json:"priority"`
	AssigneeID  uint   `form:"assignee_id" json:"assignee_id"`
}

// CreateFeatureUnderRelease handles creating a new feature directly under a release
// @Summary Create a feature under a release
// @Description Create a new feature and automatically assign it to a release
// @Tags releases
// @Accept json
// @Produce json
// @Param id path int true "Release ID"
// @Param request body CreateFeatureUnderReleaseRequest true "Feature data"
// @Success 200 {object} models.Feature "Created feature"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Release not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases/{id}/features/create [post]
func (h *ReleaseHandler) CreateFeatureUnderRelease(c *gin.Context) {
	logPrefix := "[CreateFeatureUnderRelease]"
	log.Printf("%s Start creating feature under release", logPrefix)
	
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release ID"})
		return
	}
	
	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		return
	}
	
	// Verify release is in draft state
	if release.Status != models.ReleaseStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft releases can have features created"})
		return
	}
	
	// Parse request body (support both JSON and form data)
	var req CreateFeatureUnderReleaseRequest
	
	// Log content type for debugging
	log.Printf("%s Content-Type: %s", logPrefix, c.ContentType())
	
	if strings.Contains(c.ContentType(), "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("%s JSON binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	} else {
		// Try form binding
		if err := c.ShouldBind(&req); err != nil {
			log.Printf("%s Form binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	}
	
	log.Printf("%s Creating feature '%s' under release %d", logPrefix, req.Title, releaseID)
	
	// Get database connection
	db := getDB(h.releaseRepo)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal repository error"})
		return
	}
	
	// Set default values
	category := req.Category
	if category == "" {
		category = "general"
	}
	
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	
	// Create the feature
	feature := models.Feature{
		ProjectID:   release.ProjectID,
		ReleaseID:   &releaseID,
		Title:       req.Title,
		Description: req.Description,
		Category:    category,
		Status:      models.StatusTodo,
		Priority:    models.FeaturePriority(priority),
		AssigneeID:  req.AssigneeID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := db.Create(&feature).Error; err != nil {
		log.Printf("%s Failed to create feature: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create feature: " + err.Error()})
		return
	}
	
	log.Printf("%s Successfully created feature %d under release %d", logPrefix, feature.ID, releaseID)
	
	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Fetch updated release
		updatedRelease, err := h.releaseRepo.GetByID(releaseID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated release"})
			return
		}
		
		// Fetch all features for this release
		var plannedFeatures []models.Feature
		if err := db.Where("release_id = ?", releaseID).Find(&plannedFeatures).Error; err != nil {
			log.Printf("%s Failed to fetch planned features: %v", logPrefix, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch planned features"})
			return
		}
		
		// Get user context
		userRole, _ := c.Get("user_role")
		currentUser := map[string]interface{}{
			"Role":      userRole,
			"IsManager": userRole == "manager",
		}
		
		// Render the planned features list and clear the form
		c.HTML(http.StatusOK, "release-planned-features-oob.html", gin.H{
			"PlannedFeatures": plannedFeatures,
			"Release":         updatedRelease,
			"CurrentUser":     currentUser,
		})
		return
	}
	
	// For non-HTMX requests, return JSON
	c.JSON(http.StatusOK, feature)
}

// UpdateNotesRequest represents the request to update release notes
type UpdateNotesRequest struct {
	Notes string `form:"notes" json:"notes"`
}

// UpdateReleaseNotes handles updating the notes for a release
// @Summary Update release notes
// @Description Update the notes/description for a draft release
// @Tags releases
// @Accept json
// @Produce json
// @Param id path int true "Release ID"
// @Param request body UpdateNotesRequest true "Notes data"
// @Success 200 {object} models.Release "Updated release"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 404 {object} map[string]string "Release not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/releases/{id}/notes [put]
func (h *ReleaseHandler) UpdateReleaseNotes(c *gin.Context) {
	logPrefix := "[UpdateReleaseNotes]"
	log.Printf("%s Start updating release notes", logPrefix)
	
	// Parse release ID from URL
	releaseID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid release ID"})
		return
	}
	
	// Get release by ID
	release, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Release not found"})
		return
	}
	
	// Verify release is in draft state
	if release.Status != models.ReleaseStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only draft releases can have notes updated"})
		return
	}
	
	// Parse request body (support both JSON and form data)
	var req UpdateNotesRequest
	
	// Log content type for debugging
	log.Printf("%s Content-Type: %s", logPrefix, c.ContentType())
	
	if strings.Contains(c.ContentType(), "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("%s JSON binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	} else {
		// Try form binding
		if err := c.ShouldBind(&req); err != nil {
			log.Printf("%s Form binding error: %v", logPrefix, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}
	}
	
	log.Printf("%s Updating notes for release %d", logPrefix, releaseID)
	
	// Get database connection
	db := getDB(h.releaseRepo)
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal repository error"})
		return
	}
	
	// Update the notes
	if err := db.Model(&models.Release{}).Where("id = ?", releaseID).Update("notes", req.Notes).Error; err != nil {
		log.Printf("%s Failed to update notes: %v", logPrefix, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notes: " + err.Error()})
		return
	}
	
	// Fetch updated release
	updatedRelease, err := h.releaseRepo.GetByID(releaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated release"})
		return
	}
	
	log.Printf("%s Successfully updated notes for release %d", logPrefix, releaseID)
	
	c.JSON(http.StatusOK, updatedRelease)
}
