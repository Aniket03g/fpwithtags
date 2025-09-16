# Dependency Management Phase 2 Integration Guide

This guide explains how to integrate the Phase 2 Dependency Management features into the FeaturePlus PR and Release lifecycle.

## Overview

Phase 2 of the Dependency Management system extends the existing dependency tracking functionality to enforce dependency rules during PR approval/merge and release finalization. This ensures that:

1. PRs cannot be approved or merged if their associated features have unresolved dependencies
2. Releases cannot be finalized unless all included PRs have their dependencies resolved
3. All dependencies are satisfied within the same release

## Integration Components

### 1. PR Dependency Status

The system now includes a computed `IsBlocked` status for PRs based on their associated feature's dependencies:

```go
// PRResponse extends the PullRequest model with additional computed fields
type PRResponse struct {
    PullRequest
    IsBlocked           bool                   `json:"is_blocked"`
    BlockingDependencies []map[string]interface{} `json:"blocking_dependencies,omitempty"`
    MergeableState      string                 `json:"mergeable_state,omitempty"`
}
```

### 2. PR Approval Validation

The `ApprovePR` handler now checks for unresolved dependencies before allowing PR approval:

```go
// In pr_handler.go
// Check if the PR is blocked by dependencies
dependencyService := services.NewDependencyService(repositories.NewDependencyRepository(db.DB))
isBlocked, blockingDeps, err := dependencyService.CheckPRBlocked(uint(id))
if err != nil {
    log.Printf("[ApprovePR] Error checking PR dependencies: %v", err)
    // Continue with approval even if dependency check fails
} else if isBlocked {
    // PR is blocked by dependencies, cannot approve
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Cannot approve PR: Feature has unresolved dependencies",
        "blocking_dependencies": blockingDeps,
    })
    return
}
```

### 3. Release Validation

The `ReleaseValidator` now includes two new validation methods:

```go
// ValidatePRsNotBlocked checks if any PRs in the release have unresolved dependencies
func (v *ReleaseValidator) ValidatePRsNotBlocked(prs []models.PullRequest) error {
    // Implementation details...
}

// ValidateDependenciesInRelease checks if all features in a release have their dependencies satisfied
func (v *ReleaseValidator) ValidateDependenciesInRelease(releaseID uint) error {
    // Implementation details...
}
```

### 4. Release Finalization Checks

The `FinalizeRelease` handler now validates dependencies before allowing finalization:

```go
// In release_handler.go
// Validate that no PR has unresolved dependencies
validator := services.NewReleaseValidator(h.releaseRepo.DB())
if err := validator.ValidatePRsNotBlocked(release.PRs); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: " + err.Error()})
    return
}

// Validate that all dependencies are satisfied within the release
if err := validator.ValidateDependenciesInRelease(releaseID); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot finalize release: " + err.Error()})
    return
}
```

## UI Components

### 1. PR Dependency Panel

A new `pr_dependencies.html` template displays dependency information on PR detail pages:

```html
<div 
  id="dependency-section-{{ .PR.ID }}" 
  hx-get="/web/fragments/pr/{{ .PR.ID }}/dependencies" 
  hx-trigger="load"
  hx-swap="innerHTML">
    <!-- Loading spinner and dependency information -->
</div>
```

### 2. Release Creation UI

The release creation form now includes dependency validation:

```html
<!-- PR Selection with dependency checks -->
<input 
  type="checkbox" 
  name="pr_ids" 
  value="{{ .ID }}" 
  hx-post="/releases/check-dependencies"
  hx-trigger="click"
  hx-target="#dependency-check-container"
  hx-include="[name='pr_ids']:checked"
  {{ if .Blocked }}disabled{{ end }}
>

<!-- Dependency check results container -->
<div id="dependency-check-container">
  <!-- Results loaded via HTMX -->
</div>
```

## Integration Steps

1. **Add PR Repository Method**:
   - Implement `GetByID` method in `PullRequestRepository`

2. **Update PR Handler**:
   - Modify `ApprovePR` to check dependencies before approval

3. **Extend Release Validator**:
   - Add `ValidatePRsNotBlocked` and `ValidateDependenciesInRelease` methods

4. **Update Release Handler**:
   - Add dependency validation to `FinalizeRelease`

5. **Register New Routes**:
   - Add routes for PR dependency panel and release dependency checks

6. **Update Templates**:
   - Add dependency panel to PR detail page
   - Update release creation UI to handle dependency checks

## Testing

1. **PR Approval Flow**:
   - Create a feature with dependencies
   - Create a PR for that feature
   - Verify that the PR cannot be approved until dependencies are resolved

2. **Release Finalization Flow**:
   - Create a release with PRs that have dependencies
   - Verify that the release cannot be finalized until all dependencies are included
   - Add the missing dependencies to the release
   - Verify that the release can now be finalized

## Error Messages

The system provides clear error messages when dependency validation fails:

- PR Approval: "Cannot approve PR: Feature has unresolved dependencies"
- Release Finalization: "Cannot finalize release: PR #X (Title) has unresolved dependencies"
- Release Dependencies: "Feature #X (Title) is missing Y dependencies"
