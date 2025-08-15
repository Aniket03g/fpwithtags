package services

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

// ReleaseValidator provides methods to validate releases.
type ReleaseValidator struct {
	db *gorm.DB
}

// NewReleaseValidator creates a new ReleaseValidator.
func NewReleaseValidator(db *gorm.DB) *ReleaseValidator {
	return &ReleaseValidator{db: db}
}

// ValidatePRsForRelease runs all validation checks for PRs in a release.
func (v *ReleaseValidator) ValidatePRsForRelease(prIDs []int) error {
	// Fetch PRs from the database
	var prs []models.PullRequest
	if err := v.db.Where("id IN ?", prIDs).Find(&prs).Error; err != nil {
		return fmt.Errorf("failed to fetch pull requests: %w", err)
	}

	if len(prs) != len(prIDs) {
		return fmt.Errorf("one or more pull requests not found")
	}

	// 1. Validate that all PRs are from the same repository.
	if err := v.validatePRsFromSameRepo(prs); err != nil {
		return err
	}

	// 2. Validate that no PR is in another draft release.
	if err := v.validatePRsNotInOtherDrafts(prIDs); err != nil {
		return err
	}

	return nil
}

// validatePRsFromSameRepo checks if all pull requests belong to the same repository.
func (v *ReleaseValidator) validatePRsFromSameRepo(prs []models.PullRequest) error {
	if len(prs) == 0 {
		return nil
	}

	var firstRepo string
	for i, pr := range prs {
		repoPath, err := getRepoPathFromURL(pr.URL)
		if err != nil {
			return fmt.Errorf("could not parse URL for PR #%d: %w", pr.ID, err)
		}

		if i == 0 {
			firstRepo = repoPath
		} else if repoPath != firstRepo {
			return fmt.Errorf("all PRs in a release must belong to the same repository")
		}
	}
	return nil
}

// validatePRsNotInOtherDrafts checks if any of the PRs are in other draft releases.
func (v *ReleaseValidator) validatePRsNotInOtherDrafts(prIDs []int) error {
	var conflictingPRs []int

	// Find PRs that are in a release with 'draft' status.
	err := v.db.Table("release_prs").
		Select("release_prs.pull_request_id").
		Joins("JOIN releases ON releases.id = release_prs.release_id").
		Where("releases.status = ?", models.ReleaseStatusDraft).
		Where("release_prs.pull_request_id IN ?", prIDs).
		Pluck("release_prs.pull_request_id", &conflictingPRs).Error

	if err != nil {
		return fmt.Errorf("failed to check for PRs in other draft releases: %w", err)
	}

	if len(conflictingPRs) > 0 {
		return fmt.Errorf("PRs already in another draft release: %v", conflictingPRs)
	}

	return nil
}

// getRepoPathFromURL parses a GitHub pull request URL and returns the 'owner/repo' path.
func getRepoPathFromURL(prURL string) (string, error) {
	u, err := url.Parse(prURL)
	if err != nil {
		return "", err
	}

	// Expected path format: /owner/repo/pull/number
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 4 || parts[2] != "pull" {
		return "", fmt.Errorf("invalid GitHub PR URL format: %s", prURL)
	}

	return fmt.Sprintf("%s/%s", parts[0], parts[1]), nil
}
