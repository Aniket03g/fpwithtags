package services

import (
	"errors"
	"fmt"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
)

// DependencyService handles business logic for dependencies
type DependencyService struct {
	repo *repositories.DependencyRepository
}

// NewDependencyService creates a new dependency service
func NewDependencyService(repo *repositories.DependencyRepository) *DependencyService {
	return &DependencyService{repo: repo}
}

// CreateDependency creates a new dependency after validating it won't create a cycle
func (s *DependencyService) CreateDependency(dependency *models.Dependency) error {
	// Validate entity types
	if !isValidEntityType(dependency.ParentType) || !isValidEntityType(dependency.ChildType) {
		return errors.New("invalid entity type")
	}

	// Prevent self-dependencies
	if dependency.ParentID == dependency.ChildID && dependency.ParentType == dependency.ChildType {
		return errors.New("an entity cannot depend on itself")
	}

	// Check for cycles
	hasCycle, err := s.repo.CheckForCycle(dependency)
	if err != nil {
		return fmt.Errorf("error checking for dependency cycles: %w", err)
	}
	if hasCycle {
		return errors.New("cannot create dependency: would create a cycle")
	}

	// Create the dependency
	return s.repo.Create(dependency)
}

// GetDependency retrieves a dependency by ID
func (s *DependencyService) GetDependency(id uint) (*models.Dependency, error) {
	return s.repo.GetByID(id)
}

// DeleteDependency removes a dependency
func (s *DependencyService) DeleteDependency(id uint) error {
	return s.repo.Delete(id)
}

// ListDependenciesByParent gets all dependencies where the entity is blocking others
func (s *DependencyService) ListDependenciesByParent(parentType models.EntityType, parentID uint) ([]models.Dependency, error) {
	if !isValidEntityType(parentType) {
		return nil, errors.New("invalid entity type")
	}
	return s.repo.ListByParent(parentType, parentID)
}

// ListDependenciesByChild gets all dependencies where the entity is blocked by others
func (s *DependencyService) ListDependenciesByChild(childType models.EntityType, childID uint) ([]models.Dependency, error) {
	if !isValidEntityType(childType) {
		return nil, errors.New("invalid entity type")
	}
	return s.repo.ListByChild(childType, childID)
}

// CheckBlocked determines if an entity is blocked by any dependencies
// Returns true if blocked, false if not blocked, and a list of blocking dependencies
func (s *DependencyService) CheckBlocked(entityType models.EntityType, entityID uint) (bool, []models.Dependency, error) {
	if !isValidEntityType(entityType) {
		return false, nil, errors.New("invalid entity type")
	}

	// Get all dependencies where this entity is the child (i.e., it's blocked by others)
	dependencies, err := s.repo.GetBlockingDependencies(entityType, entityID)
	if err != nil {
		return false, nil, err
	}

	// If there are no dependencies, the entity is not blocked
	if len(dependencies) == 0 {
		return false, nil, nil
	}

	return true, dependencies, nil
}

// GetBlockingEntities returns a map of entity types to slices of entity IDs that are blocking the given entity
func (s *DependencyService) GetBlockingEntities(entityType models.EntityType, entityID uint) (map[models.EntityType][]uint, error) {
	if !isValidEntityType(entityType) {
		return nil, errors.New("invalid entity type")
	}

	// Get all dependencies where this entity is the child (i.e., it's blocked by others)
	dependencies, err := s.repo.GetBlockingDependencies(entityType, entityID)
	if err != nil {
		return nil, err
	}

	// Group by entity type
	result := make(map[models.EntityType][]uint)
	for _, dep := range dependencies {
		result[dep.ParentType] = append(result[dep.ParentType], dep.ParentID)
	}

	return result, nil
}

// isValidEntityType checks if the entity type is valid
func isValidEntityType(entityType models.EntityType) bool {
	switch entityType {
	case models.EntityTypeFeature, models.EntityTypeTask, models.EntityTypePR:
		return true
	default:
		return false
	}
}

// ListAllDependencies returns all dependencies in the system
func (s *DependencyService) ListAllDependencies() ([]models.Dependency, error) {
	return s.repo.ListAll()
}

// ListDependenciesByType returns dependencies filtered by entity type (either parent or child)
func (s *DependencyService) ListDependenciesByType(entityType string) ([]models.Dependency, error) {
	if !isValidEntityType(models.EntityType(entityType)) {
		return nil, errors.New("invalid entity type")
	}
	return s.repo.ListByType(models.EntityType(entityType))
}

// CheckPRBlocked determines if a PR is blocked based on its associated feature's dependencies
// Returns:
// - bool: true if PR is blocked, false if not
// - []models.Dependency: list of blocking dependencies if blocked
// - error: any error that occurred during the check
func (s *DependencyService) CheckPRBlocked(prID uint) (bool, []models.Dependency, error) {
	// Get the PR from the database to find its associated feature
	pr, err := s.repo.GetPRByID(prID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Check if the PR's feature has unresolved dependencies
	isBlocked, dependencies, err := s.CheckBlocked(models.EntityTypeFeature, pr.FeatureID)
	return isBlocked, dependencies, err
}

// GetPRBlockingDependencies returns detailed information about dependencies blocking a PR
// This includes the dependency objects and the names of the blocking features
func (s *DependencyService) GetPRBlockingDependencies(prID uint) ([]map[string]interface{}, error) {
	// Get the PR from the database
	pr, err := s.repo.GetPRByID(prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Check if the PR's feature has unresolved dependencies
	_, dependencies, err := s.CheckBlocked(models.EntityTypeFeature, pr.FeatureID)
	if err != nil {
		return nil, err
	}

	// Enhance dependencies with feature names
	result := make([]map[string]interface{}, 0, len(dependencies))
	for _, dep := range dependencies {
		item := map[string]interface{}{
			"ID":          dep.ID,
			"ParentType":  dep.ParentType,
			"ParentID":    dep.ParentID,
			"ChildType":   dep.ChildType,
			"ChildID":     dep.ChildID,
			"Description": dep.Description,
		}

		// Get feature name if parent is a feature
		if dep.ParentType == models.EntityTypeFeature {
			feature, err := s.repo.GetFeatureByID(dep.ParentID)
			if err == nil {
				item["ParentName"] = feature.Title
			}
		}

		result = append(result, item)
	}

	return result, nil
}

// ValidateDependenciesInRelease checks if all features in a release have their dependencies satisfied
// Returns:
// - bool: true if all dependencies are satisfied, false otherwise
// - map[uint][]uint: map of feature IDs to their missing dependency feature IDs
// - error: any error that occurred during validation
func (s *DependencyService) ValidateDependenciesInRelease(releaseID uint) (bool, map[uint][]uint, error) {
	// Get all features in the release
	features, err := s.repo.GetFeaturesInRelease(releaseID)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get features in release: %w", err)
	}

	// Create a set of feature IDs in the release for quick lookup
	featureIDsInRelease := make(map[uint]bool)
	for _, feature := range features {
		featureIDsInRelease[feature.ID] = true
	}

	// Check each feature's dependencies
	missingDependencies := make(map[uint][]uint)
	for _, feature := range features {
		// Get dependencies where this feature is the child (i.e., it depends on other features)
		dependencies, err := s.repo.GetBlockingDependencies(models.EntityTypeFeature, feature.ID)
		if err != nil {
			return false, nil, err
		}

		// Check if each dependency is satisfied (i.e., the parent feature is in the release)
		for _, dep := range dependencies {
			if dep.ParentType == models.EntityTypeFeature {
				if !featureIDsInRelease[dep.ParentID] {
					// This dependency is not satisfied
					missingDependencies[feature.ID] = append(missingDependencies[feature.ID], dep.ParentID)
				}
			}
		}
	}

	return len(missingDependencies) == 0, missingDependencies, nil
}
