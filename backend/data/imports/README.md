# Import Templates Directory

This directory contains GitHub MCP-generated project templates that are loaded **dynamically on-demand**.

## Key Differences from `/backend/data/templates.json`

| Feature | templates.json | imports/*.json |
|---------|---------------|----------------|
| **Loading** | Preloaded at startup | Loaded on-demand |
| **Caching** | Cached in memory | Not cached |
| **Use Case** | Tech stack templates | GitHub-imported projects |
| **API Endpoint** | `/api/templates/*` | `/api/imports/import` |

## File Format

Each JSON file should follow the `Template` struct format:

```json
{
  "id": "unique_project_id",
  "name": "Project Name",
  "stack": "Tech Stack",
  "description": "Project description",
  "tech_stack": "Primary tech stack",
  "feature_categories": ["Category1", "Category2"],
  "task_types": ["Type1", "Type2"],
  "features": [
    {
      "name": "Feature Name",
      "category": "Category",
      "description": "Feature description",
      "context": "Development"
    }
  ],
  "tasks": [
    {
      "name": "Task Name",
      "type": "TaskType",
      "description": "Task description",
      "priority": "high|medium|low",
      "context": "Development"
    }
  ],
  "dependencies": [],
  "setup_steps": [],
  "environment_variables": [],
  "starter_repo": "",
  "docs_links": []
}
```

## API Usage

### 1. List Available Imports
```bash
GET /api/imports
Authorization: Bearer <token>
```

**Response:**
```json
{
  "status": "success",
  "imports": ["github_project_demo", "another_project"],
  "count": 2
}
```

### 2. Import a Project
```bash
POST /api/imports/import
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_id": "github_project_demo",
  "project_name": "My Imported Project",
  "description": "Optional description"
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Project imported successfully",
  "project_id": 12,
  "project_name": "My Imported Project",
  "features_created": 4,
  "tasks_created": 6
}
```

### 3. Save a New Import Template
```bash
POST /api/imports/save
Authorization: Bearer <token>
Content-Type: application/json

{
  "project_id": "new_project",
  "template": { ... }
}
```

### 4. Delete an Import Template
```bash
DELETE /api/imports/:id
Authorization: Bearer <token>
```

## Error Responses

### 404 - Template Not Found
```json
{
  "status": "error",
  "message": "Import template not found: project_id",
  "error": "..."
}
```

### 400 - Invalid JSON
```json
{
  "status": "error",
  "message": "Invalid request body",
  "error": "..."
}
```

### 500 - Server Error
```json
{
  "status": "error",
  "message": "Failed to create project",
  "error": "..."
}
```

## File Naming Convention

- Use lowercase with underscores: `github_project_demo.json`
- Avoid special characters except underscore and hyphen
- File name should match the `id` field in the JSON (optional but recommended)

## Security Notes

- All endpoints require authentication
- Project IDs are sanitized to prevent path traversal attacks
- Only `.json` files are recognized
- Files are read with restricted permissions (0644)

## Example Workflow

1. **GitHub MCP generates project JSON** → Saves to `/backend/data/imports/my_repo.json`
2. **User triggers import** → `POST /api/imports/import` with `project_id: "my_repo"`
3. **Backend dynamically reads** → `/backend/data/imports/my_repo.json`
4. **Creates project** → New project with features and tasks in database
5. **Returns success** → Project ID and creation summary

## Implementation Files

- `repositories/import_repository.go` - Dynamic file loading
- `handlers/import_handler.go` - API handlers
- `routes/import_routes.go` - Route registration
- `main.go:545` - Route registration call
