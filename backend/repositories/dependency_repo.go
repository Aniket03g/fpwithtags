package repositories

import (
	"errors"
	"fmt"

	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

// DependencyRepository handles database operations for dependencies
type DependencyRepository struct {
	db *gorm.DB
}

// NewDependencyRepository creates a new dependency repository
func NewDependencyRepository(db *gorm.DB) *DependencyRepository {
	return &DependencyRepository{db: db}
}

// Create adds a new dependency
func (r *DependencyRepository) Create(dependency *models.Dependency) error {
	return r.db.Create(dependency).Error
}

// GetByID retrieves a dependency by ID
func (r *DependencyRepository) GetByID(id uint) (*models.Dependency, error) {
	var dependency models.Dependency
	if err := r.db.First(&dependency, id).Error; err != nil {
		return nil, err
	}
	return &dependency, nil
}

// Delete removes a dependency by ID
func (r *DependencyRepository) Delete(id uint) error {
	return r.db.Delete(&models.Dependency{}, id).Error
}

// ListByParent gets all dependencies where the specified entity is the parent (blocking)
func (r *DependencyRepository) ListByParent(parentType models.EntityType, parentID uint) ([]models.Dependency, error) {
	var dependencies []models.Dependency
	err := r.db.Where("parent_type = ? AND parent_id = ?", parentType, parentID).Find(&dependencies).Error
	return dependencies, err
}

// ListByChild gets all dependencies where the specified entity is the child (blocked)
func (r *DependencyRepository) ListByChild(childType models.EntityType, childID uint) ([]models.Dependency, error) {
	var dependencies []models.Dependency
	err := r.db.Where("child_type = ? AND child_id = ?", childType, childID).Find(&dependencies).Error
	return dependencies, err
}

// CheckForCycle checks if adding a new dependency would create a cycle
func (r *DependencyRepository) CheckForCycle(dependency *models.Dependency) (bool, error) {
	// Start with the child as our current node
	currentType := dependency.ChildType
	currentID := dependency.ChildID
	
	// Keep track of visited nodes to detect cycles
	visited := make(map[string]bool)
	
	// Maximum depth to prevent infinite loops (adjust based on expected dependency depth)
	maxDepth := 100
	depth := 0
	
	for depth < maxDepth {
		// Mark current node as visited
		nodeKey := fmt.Sprintf("%s-%d", currentType, currentID)
		visited[nodeKey] = true
		
		// Find all dependencies where current node is the parent
		var childDeps []models.Dependency
		if err := r.db.Where("parent_type = ? AND parent_id = ?", currentType, currentID).Find(&childDeps).Error; err != nil {
			return false, err
		}
		
		// If no more children, we've reached the end of this path without finding a cycle
		if len(childDeps) == 0 {
			break
		}
		
		// Check each child
		hasCycle := false
		for _, childDep := range childDeps {
			// If this child is our original parent, we have a cycle
			if childDep.ChildType == dependency.ParentType && childDep.ChildID == dependency.ParentID {
				return true, nil
			}
			
			// If we've already visited this child, skip it to avoid loops
			childKey := fmt.Sprintf("%s-%d", childDep.ChildType, childDep.ChildID)
			if visited[childKey] {
				continue
			}
			
			// Recursively check this child
			currentType = childDep.ChildType
			currentID = childDep.ChildID
			hasCycle = true
			break
		}
		
		// If we didn't find a child to follow, we're done with this path
		if !hasCycle {
			break
		}
		
		depth++
	}
	
	// If we reached max depth, something is wrong
	if depth >= maxDepth {
		return false, errors.New("dependency chain too deep, possible cycle detected")
	}
	
	return false, nil
}

// GetBlockingDependencies returns all dependencies that are blocking the given entity
func (r *DependencyRepository) GetBlockingDependencies(entityType models.EntityType, entityID uint) ([]models.Dependency, error) {
	var dependencies []models.Dependency
	err := r.db.Where("child_type = ? AND child_id = ?", entityType, entityID).Find(&dependencies).Error
	return dependencies, err
}

// DB returns the underlying database connection
func (r *DependencyRepository) DB() *gorm.DB {
	return r.db
}

// ListAll returns all dependencies in the system
func (r *DependencyRepository) ListAll() ([]models.Dependency, error) {
	var dependencies []models.Dependency
	err := r.db.Find(&dependencies).Error
	return dependencies, err
}

// ListByType returns dependencies filtered by entity type (either parent or child)
func (r *DependencyRepository) ListByType(entityType models.EntityType) ([]models.Dependency, error) {
	var dependencies []models.Dependency
	err := r.db.Where("parent_type = ? OR child_type = ?", entityType, entityType).Find(&dependencies).Error
	return dependencies, err
}

// GetPRByID retrieves a pull request by ID
func (r *DependencyRepository) GetPRByID(id uint) (*models.PullRequest, error) {
	var pr models.PullRequest
	if err := r.db.First(&pr, id).Error; err != nil {
		return nil, err
	}
	return &pr, nil
}

// GetFeatureByID retrieves a feature by ID
func (r *DependencyRepository) GetFeatureByID(id uint) (*models.Feature, error) {
	var feature models.Feature
	if err := r.db.First(&feature, id).Error; err != nil {
		return nil, err
	}
	return &feature, nil
}

// GetFeaturesInRelease retrieves all features associated with a release through PRs
func (r *DependencyRepository) GetFeaturesInRelease(releaseID uint) ([]models.Feature, error) {
	// First get all PRs in the release
	var prs []models.PullRequest
	if err := r.db.Joins("JOIN release_prs ON release_prs.pull_request_id = pull_requests.id").Where("release_prs.release_id = ?", releaseID).Find(&prs).Error; err != nil {
		return nil, err
	}

	// Extract feature IDs from PRs
	featureIDs := make([]uint, 0, len(prs))
	featureIDMap := make(map[uint]bool) // To prevent duplicates
	for _, pr := range prs {
		if !featureIDMap[pr.FeatureID] {
			featureIDs = append(featureIDs, pr.FeatureID)
			featureIDMap[pr.FeatureID] = true
		}
	}

	// If no features found, return empty slice
	if len(featureIDs) == 0 {
		return []models.Feature{}, nil
	}

	// Get all features by their IDs
	var features []models.Feature
	if err := r.db.Where("id IN ?", featureIDs).Find(&features).Error; err != nil {
		return nil, err
	}

	return features, nil
}
