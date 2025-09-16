# Dependency Management System Documentation

## 1. Overview

The Dependency Management system in FeaturePlus enables tracking and enforcing relationships between different entities in the application. It allows users to define dependencies between features, tasks, and pull requests, ensuring that work is completed in the correct order and preventing premature merging or release of dependent items.

The primary purpose of this system is to:
- Establish clear relationships between interdependent work items
- Provide visibility into what is blocking a particular feature or task
- Prevent work from being marked as complete when its dependencies are not yet resolved
- Improve project planning by making dependencies explicit and trackable

In Phase 1, the system provides UI components for creating, viewing, and managing dependencies, with validation to prevent circular dependencies and self-dependencies.

## 2. Architecture

The Dependency Management module follows the standard FeaturePlus architecture with clear separation of concerns:

### Components
- **Models**: Define the data structure and validation rules for dependencies
- **Repositories**: Handle database operations for dependencies
- **Services**: Implement business logic for dependency management
- **Handlers**: Process HTTP requests and responses for the dependency API
- **Templates**: Render UI components for dependency management
- **Routes**: Define API endpoints for dependency operations

### Interaction Flow
```
Client (HTMX) → Routes → Handlers → Services → Repositories → Database
                  ↑                     ↓
                  └─────── Templates ───┘
```

## 3. Data Model

The core of the system is the `Dependency` model, which represents a relationship where one entity (Parent) blocks another entity (Child).

### Dependency Model Fields

| Field       | Type       | Description |
|-------------|------------|-------------|
| ID          | uint       | Primary key |
| ParentID    | uint       | ID of the blocking entity |
| ParentType  | EntityType | Type of the blocking entity (feature, task, pr) |
| ChildID     | uint       | ID of the blocked entity |
| ChildType   | EntityType | Type of the blocked entity (feature, task, pr) |
| Description | string     | Optional context about the dependency |
| CreatedByID | uint       | User ID who created the dependency |
| CreatedAt   | time.Time  | Timestamp when the dependency was created |
| UpdatedAt   | time.Time  | Timestamp when the dependency was last updated |
| DeletedAt   | time.Time  | Soft delete timestamp |

### Entity Types
```go
type EntityType string

const (
    EntityTypeFeature EntityType = "feature"
    EntityTypeTask    EntityType = "task"
    EntityTypePR      EntityType = "pr"
)
```

## 4. Backend Components

### Repository (`dependency_repo.go`)

The repository layer handles direct database operations:

| Method | Description |
|--------|-------------|
| `Create(dependency *models.Dependency)` | Adds a new dependency |
| `GetByID(id uint)` | Retrieves a dependency by ID |
| `Delete(id uint)` | Removes a dependency by ID |
| `ListByParent(parentType, parentID)` | Gets dependencies where the entity is blocking others |
| `ListByChild(childType, childID)` | Gets dependencies where the entity is blocked by others |
| `CheckForCycle(dependency)` | Checks if adding a dependency would create a cycle |
| `GetBlockingDependencies(entityType, entityID)` | Returns dependencies blocking an entity |
| `ListAll()` | Returns all dependencies in the system |
| `ListByType(entityType)` | Returns dependencies filtered by entity type |

### Service (`dependency_service.go`)

The service layer implements business logic and validation:

| Method | Description |
|--------|-------------|
| `CreateDependency(dependency)` | Creates a dependency after validations |
| `GetDependency(id)` | Retrieves a dependency by ID |
| `DeleteDependency(id)` | Removes a dependency |
| `ListDependenciesByParent(parentType, parentID)` | Gets dependencies where entity is blocking others |
| `ListDependenciesByChild(childType, childID)` | Gets dependencies where entity is blocked by others |
| `CheckBlocked(entityType, entityID)` | Determines if an entity is blocked by dependencies |
| `GetBlockingEntities(entityType, entityID)` | Returns entities blocking a given entity |
| `ListAllDependencies()` | Returns all dependencies |
| `ListDependenciesByType(entityType)` | Returns dependencies filtered by entity type |
| `isValidEntityType(entityType)` | Validates entity type |

### Handler (`dependency_handler.go`)

The handler layer processes HTTP requests and renders responses:

| Method | Description |
|--------|-------------|
| `CreateDependency(c *gin.Context)` | Handles POST /api/dependencies |
| `ListDependencies(c *gin.Context)` | Handles GET /api/dependencies with query parameters |
| `DeleteDependency(c *gin.Context)` | Handles DELETE /api/dependencies/:id |
| `GetDependencyPanels(c *gin.Context)` | Returns HTML fragments for dependency panels |
| `GetDependencyPanelsFragment(c *gin.Context)` | Returns HTML fragment for dependency panels (no shell layout) |
| `ShowDependencyModalFragment(c *gin.Context)` | Returns HTML fragment for the add dependency modal |
| `ShowDependencyTypeSelector(c *gin.Context)` | Returns HTML fragment for the dependency type selector |
| `enhanceDependenciesWithNames(dependencies)` | Adds entity names to dependencies for display |

## 5. Frontend/HTMX Integration

The frontend uses HTMX for dynamic updates without full page reloads:

### HTML Fragments

| Fragment | Purpose |
|----------|---------|
| `dependency_panels.html` | Shows "Depends On" and "Blocking" lists for an entity |
| `dependency_type_selector.html` | Modal for selecting dependency type (feature, task, etc.) |
| `dependency_modal.html` | Modal form for adding a new dependency |

### HTMX Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/web/fragments/dependencies/panels` | GET | Returns dependency panels HTML |
| `/web/fragments/dependencies/type_selector` | GET | Returns dependency type selector modal |
| `/web/fragments/dependencies/modal` | GET | Returns dependency creation modal |
| `/api/dependencies` | POST | Creates a new dependency |
| `/api/dependencies/:id` | DELETE | Deletes a dependency |

### HTMX Attributes

| Attribute | Example | Purpose |
|-----------|---------|---------|
| `hx-get` | `hx-get="/web/fragments/dependencies/type_selector"` | Fetches HTML fragments |
| `hx-post` | `hx-post="/api/dependencies"` | Submits form data |
| `hx-target` | `hx-target="#modal"` | Specifies where to insert the response |
| `hx-swap` | `hx-swap="innerHTML"` | Defines how to insert the response |
| `hx-trigger` | `hx-trigger="click"` | Defines when to trigger the request |
| `hx-on::after-request` | `hx-on::after-request="..."` | Executes JavaScript after request completes |

## 6. UI Flow

### Adding a Dependency

1. User navigates to a feature/task detail page
2. User clicks "Add Dependency" button in the dependency panel
   - HTMX sends GET to `/web/fragments/dependencies/type_selector`
   - Response is inserted into `#modal` div
3. User selects "Features" in the type selector modal
   - HTMX sends GET to `/web/fragments/dependencies/modal`
   - Response replaces the modal content
4. User selects a feature from the dropdown and adds optional description
5. User clicks "Add Dependency" button in the modal form
   - HTMX sends POST to `/api/dependencies`
   - Form data includes `parent_type`, `parent_id`, `child_type`, `child_id`, and `description`
6. On successful creation:
   - The modal is closed
   - The dependency panels are refreshed via HTMX
   - A success message is shown

### Viewing Dependencies

1. When a feature/task detail page loads, it includes dependency panels
2. The panels show:
   - "Depends On" section: Items that this entity depends on (blocking this entity)
   - "Blocking" section: Items that depend on this entity (blocked by this entity)
3. If the entity is blocked by dependencies, a warning message is displayed

### Removing a Dependency

1. User clicks the delete icon next to a dependency in the "Depends On" section
2. HTMX sends DELETE to `/api/dependencies/:id`
3. On successful deletion:
   - The dependency panels are refreshed via HTMX
   - The dependency is removed from the list

## 7. Business Rules & Validations

The system enforces several business rules to maintain data integrity:

### Entity Type Validation
- Only predefined entity types are allowed: `feature`, `task`, `pr`
- Validation occurs in `isValidEntityType()` in the service layer

### Self-Dependency Prevention
- An entity cannot depend on itself
- Validation occurs in both the model's `BeforeCreate` hook and the service layer

### Cycle Detection
- Dependencies cannot form cycles (e.g., A depends on B, B depends on C, C depends on A)
- Implemented in `CheckForCycle()` in the repository layer
- Uses a graph traversal algorithm to detect potential cycles

### Duplicate Prevention
- The same dependency relationship cannot be created multiple times
- Enforced by the model's `BeforeCreate` hook

### Dependency Status Indication
- Entities with unresolved dependencies are marked as "blocked"
- Visual indicators show when an entity is blocked

## 8. Current Limitations

The Phase 1 implementation has the following limitations:

1. **No PR/Release Integration**: Dependencies are not enforced during PR approval/merge or release creation
2. **Limited Entity Types**: While the system supports feature, task, and PR entity types, only feature dependencies are fully implemented in the UI
3. **No Transitive Dependency Resolution**: The system doesn't automatically resolve dependencies when a transitive dependency is resolved
4. **No Bulk Operations**: Dependencies must be created and deleted individually
5. **No Dependency Reporting**: There are no dedicated reports or visualizations for dependency relationships
6. **Manual Enforcement**: Users must manually check dependency status; the system doesn't prevent actions on blocked items

## 9. Future Enhancements (Phase 2 Preview)

Phase 2 will address the current limitations and add new capabilities:

### PR Integration
- Prevent PR approval/merge when the associated feature has unresolved dependencies
- Add dependency status to PR review interface
- Automatically update dependency status when PRs are merged

### Release Integration
- Validate that all interdependent features are included in the same release
- Warn when creating a release with incomplete dependency chains
- Prevent finalizing releases with unresolved dependencies

### Enhanced Visualization
- Dependency graph visualization to show relationships between entities
- Dashboard widgets for dependency status
- Dependency impact analysis for changes

### Automation
- Automatic notification when blocking dependencies are resolved
- Dependency-aware scheduling and prioritization
- Bulk dependency creation and management

### API Enhancements
- Expanded API for programmatic dependency management
- Webhooks for dependency status changes
- Integration with external tools and systems
