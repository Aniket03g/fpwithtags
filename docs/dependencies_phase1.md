# Dependency Management Module Documentation (Phase 1)

## Overview

The Dependency Management module in FeaturePlus enables tracking and enforcing dependencies between different entities in the system. This documentation covers the Phase 1 implementation of this module.

## Module-Level Documentation

### Purpose of Dependency Management

Dependency Management in FeaturePlus allows teams to:

1. **Track Relationships**: Establish and visualize dependencies between different work items.
2. **Enforce Workflow**: Prevent work from proceeding until dependencies are resolved.
3. **Improve Planning**: Identify potential bottlenecks and critical paths in the development process.
4. **Enhance Visibility**: Make dependencies explicit and visible to all team members.

### Entity Types and Relationships

The following entities can participate in dependency relationships:

- **Features**: High-level functionality that may depend on other features or block tasks.
- **Tasks**: Specific work items that may depend on features or other tasks.
- **Pull Requests (PRs)**: Code changes that may depend on other PRs or tasks.

Dependencies are directional relationships where:

- **Parent Entity (Blocker)**: The entity that must be completed first.
- **Child Entity (Blocked)**: The entity that cannot proceed until the parent is completed.

### Blocking vs. Blocked Representation

Dependencies are represented as a relationship between two entities:

- **Blocking (Parent)**: An entity that blocks progress on other entities. These are shown in the "Blocking" section of the UI.
- **Blocked By (Child)**: An entity that is blocked by other entities. These are shown in the "Depends On" section of the UI.

For example, if Feature A depends on Feature B:
- Feature B is the parent/blocker (appears in Feature B's "Blocking" list)
- Feature A is the child/blocked (appears in Feature A's "Depends On" list)

## Developer Documentation

### Architecture Overview

The Dependency Management module follows the layered architecture of FeaturePlus:

```
Models → Repositories → Services → Handlers → Templates
```

Each layer has specific responsibilities:

1. **Models**: Define the data structure and relationships.
2. **Repositories**: Handle database operations and queries.
3. **Services**: Implement business logic and validation.
4. **Handlers**: Process HTTP requests and responses.
5. **Templates**: Render the UI components.

### Model Layer

The Dependency model is defined in `models/dependency.go`:

```go
type Dependency struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    ParentType  EntityType `json:"parent_type"`
    ParentID    uint       `json:"parent_id"`
    ChildType   EntityType `json:"child_type"`
    ChildID     uint       `json:"child_id"`
    Description string     `json:"description"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

The `EntityType` is an enumeration defined as:

```go
type EntityType string

const (
    EntityTypeFeature EntityType = "feature"
    EntityTypeTask    EntityType = "task"
    EntityTypePR      EntityType = "pr"
)
```

### Repository Layer

The repository layer (`repositories/dependency_repo.go`) provides the following functionality:

#### Core CRUD Operations

- `Create(dependency *models.Dependency) error`: Creates a new dependency.
- `GetByID(id uint) (*models.Dependency, error)`: Retrieves a dependency by ID.
- `Delete(id uint) error`: Deletes a dependency by ID.

#### Query Operations

- `ListByParent(parentType models.EntityType, parentID uint) ([]models.Dependency, error)`: Lists dependencies where the specified entity is blocking others.
- `ListByChild(childType models.EntityType, childID uint) ([]models.Dependency, error)`: Lists dependencies where the specified entity is blocked by others.
- `GetBlockingDependencies(entityType models.EntityType, entityID uint) ([]models.Dependency, error)`: Gets all dependencies blocking a specific entity.
- `ListAll() ([]models.Dependency, error)`: Lists all dependencies in the system.
- `ListByType(entityType models.EntityType) ([]models.Dependency, error)`: Lists dependencies filtered by entity type.

#### Cycle Detection

The `CheckForCycle(dependency *models.Dependency) (bool, error)` method is a critical component that prevents circular dependencies. It works by:

1. Starting with the child entity of the new dependency.
2. Traversing the dependency graph to find all entities that the child blocks.
3. Checking if any of those entities is the parent of the new dependency.
4. If a cycle is detected, the method returns `true`.

This prevents situations where, for example:
- Feature A depends on Feature B
- Feature B depends on Feature C
- Feature C depends on Feature A (would create a cycle)

### Service Layer

The service layer (`services/dependency_service.go`) implements business logic and validation:

#### Core Operations

- `CreateDependency(dependency *models.Dependency) error`: Creates a dependency after validation.
- `GetDependency(id uint) (*models.Dependency, error)`: Retrieves a dependency by ID.
- `DeleteDependency(id uint) error`: Deletes a dependency.

#### Query Operations

- `ListDependenciesByParent(parentType models.EntityType, parentID uint) ([]models.Dependency, error)`: Lists dependencies where an entity is blocking others.
- `ListDependenciesByChild(childType models.EntityType, childID uint) ([]models.Dependency, error)`: Lists dependencies where an entity is blocked by others.
- `ListAllDependencies() ([]models.Dependency, error)`: Lists all dependencies.
- `ListDependenciesByType(entityType string) ([]models.Dependency, error)`: Lists dependencies filtered by entity type.

#### Validation and Business Logic

- `CheckBlocked(entityType models.EntityType, entityID uint) (bool, []models.Dependency, error)`: Determines if an entity is blocked by any dependencies. Returns:
  - A boolean indicating if the entity is blocked
  - A list of blocking dependencies
  - Any error that occurred

- Validation rules enforced by the service:
  1. Entity types must be valid (feature, task, or PR)
  2. An entity cannot depend on itself
  3. Dependencies cannot create cycles

### Handler Layer

The handler layer (`handlers/dependency_handler.go`) processes HTTP requests and renders responses:

#### REST API Endpoints

- `CreateDependency(c *gin.Context)`: Handles `POST /api/dependencies`
- `ListDependencies(c *gin.Context)`: Handles `GET /api/dependencies` with query parameters
- `DeleteDependency(c *gin.Context)`: Handles `DELETE /api/dependencies/:id`

#### HTMX Fragment Endpoints

- `GetDependencyPanels(c *gin.Context)`: Handles `GET /api/dependencies/panels`
- `ShowDependencyModal(c *gin.Context)`: Handles `GET /api/dependencies/modal`
- `GetDependencyPanelsFragment(c *gin.Context)`: Handles `GET /web/fragments/dependencies/panels`
- `ShowDependencyModalFragment(c *gin.Context)`: Handles `GET /web/fragments/dependencies/modal`
- `GetDependenciesListFragment(c *gin.Context)`: Handles `GET /web/fragments/dependencies`

### Templates

The UI is implemented using HTMX, Alpine.js, and Tailwind CSS:

#### dependency_panels.html

This template displays two panels:
1. **Depends On**: Shows entities that are blocking the current entity.
2. **Blocking**: Shows entities that are blocked by the current entity.

Each panel includes:
- A list of dependencies with their types and IDs
- Delete buttons for removing dependencies
- An "Add Dependency" button that triggers the dependency modal

The template uses HTMX to:
- Load dependency data asynchronously
- Handle delete operations without page reloads
- Show the dependency modal for adding new dependencies

#### dependency_modal.html

This template provides a form for adding new dependencies:
- Entity type selection (feature, task, PR)
- Entity ID input
- Optional description field
- Submit and cancel buttons

The form uses HTMX to submit the data asynchronously and update the dependency panels without a page reload.

## Usage Documentation

### API Usage Examples

#### Creating a Dependency

```http
POST /api/dependencies
Content-Type: application/json

{
  "parent_type": "feature",
  "parent_id": 1,
  "child_type": "task",
  "child_id": 5,
  "description": "Task 5 depends on Feature 1 being completed"
}
```

Response:
```json
{
  "id": 10,
  "parent_type": "feature",
  "parent_id": 1,
  "child_type": "task",
  "child_id": 5,
  "description": "Task 5 depends on Feature 1 being completed",
  "created_at": "2025-09-12T12:00:00Z",
  "updated_at": "2025-09-12T12:00:00Z"
}
```

#### Listing Dependencies for an Entity

To list dependencies where an entity is blocked by others:

```http
GET /api/dependencies?child_type=task&child_id=5
```

To list dependencies where an entity is blocking others:

```http
GET /api/dependencies?parent_type=feature&parent_id=1
```

Response:
```json
[
  {
    "id": 10,
    "parent_type": "feature",
    "parent_id": 1,
    "child_type": "task",
    "child_id": 5,
    "description": "Task 5 depends on Feature 1 being completed",
    "created_at": "2025-09-12T12:00:00Z",
    "updated_at": "2025-09-12T12:00:00Z"
  }
]
```

#### Deleting a Dependency

```http
DELETE /api/dependencies/10
```

Response:
```json
{
  "message": "Dependency deleted successfully"
}
```

### HTMX Integration Example

To integrate dependency panels into a feature detail page:

```html
<!-- Dependency Section -->
<div class="mt-8">
  <h3 class="text-xl font-semibold text-gray-800 mb-4">Dependencies</h3>
  <div 
    id="dependency-section" 
    hx-get="http://localhost:8080/web/fragments/dependencies/panels?entity_type=feature&entity_id={{ .Feature.ID }}" 
    hx-trigger="load"
    hx-swap="innerHTML">
      <div class="text-center py-8">
        <svg class="mx-auto animate-spin h-8 w-8 text-blue-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z"></path></svg>
        <p class="text-gray-400 mt-2">Loading dependencies...</p>
      </div>
  </div>
</div>
```

This code:
1. Creates a container for the dependency section
2. Uses HTMX to load the dependency panels when the page loads
3. Shows a loading spinner while the panels are being loaded

### Validation Rules and Cycle Detection

When creating dependencies, the following validation rules are enforced:

1. **Valid Entity Types**: Both parent and child must be one of: feature, task, or PR.
2. **No Self-Dependencies**: An entity cannot depend on itself.
3. **No Cycles**: Dependencies cannot create circular relationships.

The cycle detection algorithm works by:
1. Starting with the child entity of the new dependency
2. Traversing all dependencies where this entity is the parent
3. Checking if any of those paths lead back to the parent of the new dependency

For example, these dependencies would be rejected:
- Feature 1 depends on Feature 1 (self-dependency)
- Feature 1 depends on Feature 2, Feature 2 depends on Feature 1 (direct cycle)
- Feature 1 depends on Feature 2, Feature 2 depends on Feature 3, Feature 3 depends on Feature 1 (indirect cycle)

## Conclusion

The Dependency Management module provides a robust way to track and enforce dependencies between different entities in FeaturePlus. The implementation follows the layered architecture of the application and provides both API endpoints and UI components for managing dependencies.

Future enhancements could include:
- Dependency visualization (e.g., dependency graphs)
- Dependency impact analysis
- Bulk dependency management
- Automatic dependency resolution tracking
