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
