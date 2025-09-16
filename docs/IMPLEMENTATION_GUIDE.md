# FeaturePlus Implementation Guide
## Complete Documentation for Phases 1-4

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Phase 1: Core Project Management](#phase-1-core-project-management)
3. [Phase 2: Tech Stack Integration](#phase-2-tech-stack-integration)
4. [Phase 3: Dynamic Guidance System](#phase-3-dynamic-guidance-system)
5. [Phase 4: Stack-Based Templates](#phase-4-stack-based-templates)
6. [API Reference](#api-reference)
7. [HTMX Integration Guide](#htmx-integration-guide)
8. [Extension Guide](#extension-guide)

---

## Architecture Overview

### Technology Stack
- **Backend**: Go with Gin framework
- **Database**: PostgreSQL with GORM ORM
- **Frontend**: HTMX + TailwindCSS
- **Data Storage**: JSON files for templates and guidance
- **Authentication**: JWT-based

### Data Model
```
Projects (1:N) → Features (1:N) → Tasks
    ↓                                ↓
Tech Stack                      Guidance
    ↓                                ↑
Templates ←──────────────────────────┘
```

---

## Phase 1: Core Project Management

### Features Implemented
- Project CRUD operations
- Feature management with categories
- Task tracking with types and status
- Comments and attachments
- Pull request linking

### Database Schema
```go
type Project struct {
    ID          uint
    Title       string
    Description string
    Config      JSONB  // Stores tech_stack and other metadata
    Features    []Feature
}

type Feature struct {
    ID          uint
    Title       string
    Description string
    Category    string
    ProjectID   int
    Tasks       []Task
}

type Task struct {
    ID          uint
    TaskName    string
    Description string
    TaskType    string
    FeatureID   uint
    Attachments []Attachment
    Comments    []Comment
}
```

### HTMX Flow Example
```html
<!-- Project Creation Form -->
<form hx-post="/api/projects" 
      hx-target="#project-list" 
      hx-swap="innerHTML">
    <input name="title" required>
    <button type="submit">Create Project</button>
</form>
```

---

## Phase 2: Tech Stack Integration

### Implementation Details

#### 1. Tech Stack Options
- React
- Go
- Node.js
- Python
- Vue
- Django
- Other (default)

#### 2. Storage Method
Projects store tech_stack in Config JSONB field:
```json
{
  "tech_stack": "React",
  "template_id": "react-firebase"
}
```

#### 3. Visual Representation
```html
<!-- Tech Stack Badge -->
<span class="badge badge-react">React</span>
```

#### 4. Migration for Existing Data
```go
func MigrateTechStackField(db *gorm.DB) {
    var projects []models.Project
    db.Find(&projects)
    for _, project := range projects {
        if project.Config == nil {
            project.Config = models.JSONB{}
        }
        if project.Config["tech_stack"] == nil {
            project.Config["tech_stack"] = "Other"
            db.Save(&project)
        }
    }
}
```

---

## Phase 3: Dynamic Guidance System

### Guidance JSON Structure
```json
{
  "guidances": [
    {
      "stack": "React",
      "task_type": "UI",
      "title": "React Component Development",
      "description": "Create reusable React components",
      "snippet": "import React from 'react';...",
      "language": "javascript",
      "commands": [
        "npx create-react-app my-app",
        "npm install prop-types"
      ],
      "setup_steps": [
        "Install Node.js",
        "Create React app",
        "Install dependencies"
      ],
      "docs_link": "https://react.dev/",
      "starter_repo": "https://github.com/facebook/create-react-app"
    }
  ],
  "default_guidance": {
    "title": "General Development Guidelines",
    "description": "Best practices for development"
  }
}
```

### Repository Pattern
```go
type GuidanceRepository struct {
    data     *GuidanceData
    dataPath string
    mu       sync.RWMutex
}

func (r *GuidanceRepository) GetGuidance(stack, taskType string) GuidanceEntry {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    for _, guidance := range r.data.Guidances {
        if guidance.Stack == stack && guidance.TaskType == taskType {
            return guidance
        }
    }
    return r.data.DefaultGuidance
}
```

### HTMX Integration
```html
<!-- Show Guidance Button -->
<button hx-get="/web/tasks/{{.Task.ID}}/guidance?stack={{.ProjectTechStack}}"
        hx-target="#guidance-{{.Task.ID}}"
        hx-swap="innerHTML">
    Show Guidance
</button>

<!-- Guidance Container -->
<div id="guidance-{{.Task.ID}}"></div>
```

---

## Phase 4: Stack-Based Templates

### Template JSON Structure
```json
{
  "templates": [
    {
      "id": "react-firebase",
      "name": "React + Firebase",
      "stack": "React + Firebase",
      "tech_stack": "React",
      "features": [
        {
          "name": "Authentication",
          "category": "Auth",
          "description": "Firebase Authentication"
        }
      ],
      "tasks": [
        {
          "name": "Setup Firebase Project",
          "type": "Backend",
          "description": "Create Firebase project",
          "priority": "high"
        }
      ],
      "dependencies": ["react", "firebase"],
      "setup_steps": ["Install Node.js", "Create React app"],
      "environment_variables": ["REACT_APP_FIREBASE_API_KEY"],
      "starter_repo": "https://github.com/firebase/quickstart-js",
      "docs_links": ["https://firebase.google.com/docs"]
    }
  ]
}
```

### Template Storage and Management
Templates are stored in `backend/data/templates.json` and loaded at application startup. The template repository provides methods to access and manage templates:

```go
// TemplateRepository handles template data operations
type TemplateRepository struct {
    data     *TemplateData
    dataPath string
    mu       sync.RWMutex
}

// GetTemplateByID retrieves a template by its ID
func (r *TemplateRepository) GetTemplateByID(id string) (*Template, error) {
    // Returns template with matching ID or default template
}
```

To reload templates after making changes to the JSON file:
```bash
curl -X POST http://localhost:8080/api/templates/reload
```

### Template Application Process
When a user creates a new project with a template selected, the following process occurs:

```go
// Project creation with template application
func (h *ProjectHandler) CreateProjectFromForm(c *gin.Context) {
    // Create base project record
    project := models.Project{
        Name:        name,
        Description: description,
        OwnerID:     userID,
        Config:      customConfig,
    }
    h.repo.CreateProject(&project)
    
    // Apply template if selected
    if templateID != "" {
        // 1. Get template by ID
        template, err := h.templateRepo.GetTemplateByID(templateID)
        
        // 2. Create features from template
        for _, templateFeature := range template.Features {
            feature := models.Feature{
                Title:       templateFeature.Name,
                Description: templateFeature.Description,
                Category:    templateFeature.Category,
                ProjectID:   int(project.ID),
                Status:      models.StatusTodo,
                Priority:    models.PriorityMedium,
            }
            h.featureRepo.CreateFeature(&feature)
            featureMap[templateFeature.Name] = feature.ID
        }
        
        // 3. Create tasks with intelligent feature assignment
        for _, templateTask := range template.Tasks {
            // Match task type with feature category when possible
            var featureID uint
            for _, feature := range createdFeatures {
                if strings.EqualFold(feature.Category, templateTask.Type) {
                    featureID = feature.ID
                    break
                }
            }
            
            task := models.Task{
                TaskName:    templateTask.Name,
                Description: templateTask.Description,
                TaskType:    templateTask.Type,
                FeatureID:   featureID,
            }
            h.taskRepo.Create(&task)
        }
        
        // 4. Create dependencies from template
        for _, dependencyName := range template.Dependencies {
            dependency := models.Dependency{
                Description: fmt.Sprintf("Project dependency: %s", dependencyName),
                ParentType:  models.EntityTypeFeature,
                ChildType:   models.EntityTypeFeature,
                // Find appropriate features to link
            }
            dependencyService.CreateDependency(&dependency)
        }
        
        log.Printf("Project %d created with template '%s': %d features, %d tasks, %d dependencies", 
            project.ID, template.Name, len(createdFeatures), len(createdTasks), len(template.Dependencies))
    }
}
```

### HTMX Template Selection
```html
<select name="template_id" 
        hx-get="/web/templates/{value}/details"
        hx-target="#template-details"
        hx-trigger="change">
    <option value="">No Template</option>
    <option value="react-firebase">React + Firebase</option>
    <option value="go-postgresql">Go + PostgreSQL</option>
</select>

<div id="template-details"></div>
```

### Template Details Fragment
When a user selects a template, the details are shown via HTMX:

```html
<!-- template-details-fragment.html -->
<div class="template-details">
    <h3>{{.Name}}</h3>
    <p>{{.Description}}</p>
    
    <h4>Features ({{len .Features}})</h4>
    <ul>
        {{range .Features}}
        <li><strong>{{.Name}}</strong> - {{.Description}}</li>
        {{end}}
    </ul>
    
    <h4>Tasks ({{len .Tasks}})</h4>
    <ul>
        {{range .Tasks}}
        <li><strong>{{.Name}}</strong> ({{.Type}}) - {{.Description}}</li>
        {{end}}
    </ul>
    
    <h4>Dependencies</h4>
    <ul>
        {{range .Dependencies}}
        <li>{{.}}</li>
        {{end}}
    </ul>
</div>
```

---

## API Reference

### Project Endpoints
```yaml
GET /api/projects:
  description: List all projects
  response:
    - id: 1
      title: "My Project"
      tech_stack: "React"

POST /api/projects:
  body:
    title: "New Project"
    description: "Description"
    tech_stack: "Go"
    template_id: "go-postgresql"
  response:
    id: 2
    title: "New Project"
```

### Guidance Endpoints
```yaml
GET /web/tasks/{id}/guidance:
  params:
    stack: "React"
  response: HTML fragment with guidance

POST /api/guidance/:
  body:
    stack: "Ruby"
    task_type: "Backend"
    title: "Ruby on Rails"
    snippet: "class ApplicationController..."
```

### Template Endpoints
```yaml
GET /api/templates:
  response:
    templates:
      - id: "react-firebase"
        name: "React + Firebase"

POST /api/templates/apply:
  body:
    project_id: 1
    template_id: "react-firebase"
  response:
    message: "Template applied successfully"
    created_features: 4
    created_tasks: 6
```

---

## HTMX Integration Guide

### Key Attributes Used
- `hx-get`: Fetch content
- `hx-post`: Submit forms
- `hx-target`: Where to place response
- `hx-swap`: How to swap content
- `hx-trigger`: When to trigger request

### Common Patterns

#### 1. Modal Forms
```html
<button hx-get="/web/projects/new" 
        hx-target="#modal-container">
    New Project
</button>
```

#### 2. Dynamic Lists
```html
<div id="project-list" 
     hx-get="/web/projects" 
     hx-trigger="load">
</div>
```

#### 3. Inline Editing
```html
<div hx-get="/web/tasks/{{.ID}}/edit" 
     hx-target="this" 
     hx-swap="outerHTML">
</div>
```

---

## Extension Guide

### Adding a New Tech Stack

#### Step 1: Update Options
```html
<!-- create_project.html -->
<option value="rust">Rust</option>
```

#### Step 2: Add Guidance
```json
// guidance.json
{
  "stack": "Rust",
  "task_type": "Backend",
  "title": "Rust API Development",
  "snippet": "use actix_web::{web, App, HttpServer};",
  "commands": ["cargo new myapp", "cargo add actix-web"]
}
```

#### Step 3: Create Template
```json
// templates.json
{
  "id": "rust-actix",
  "name": "Rust + Actix Web",
  "tech_stack": "Rust",
  "features": [
    {
      "name": "REST API",
      "category": "Backend"
    }
  ]
}
```

### Adding Custom Guidance via API
```bash
curl -X POST http://localhost:8080/api/guidance/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d @guidance.json
```

### Hot Reloading Data
```bash
# Reload guidance
curl -X POST http://localhost:8080/api/guidance/reload

# Reload templates
curl -X POST http://localhost:8080/api/templates/reload
```

---

## Best Practices

### 1. Thread Safety
Always use mutex locks when accessing shared data:
```go
func (r *Repository) GetData() Data {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.data
}
```

### 2. Error Handling
Provide meaningful error messages:
```go
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": "Failed to create project",
        "details": err.Error(),
    })
}
```

### 3. HTMX Response Headers
Set proper headers for HTMX:
```go
c.Header("HX-Trigger", "projectCreated")
c.Header("HX-Retarget", "#project-list")
```

### 4. Data Validation
Validate before processing:
```go
if template.ID == "" {
    return errors.New("template ID required")
}
```

---

## Troubleshooting

### Common Issues

#### 1. HTMX Not Updating
- Check `hx-target` selector exists
- Verify `hx-swap` method is correct
- Ensure response is valid HTML

#### 2. Guidance Not Loading
- Verify guidance.json exists
- Check file permissions
- Confirm stack/task_type match

#### 3. Template Not Applying
- Check project exists
- Verify template ID is valid
- Ensure features can be created

### Debug Tips
```go
// Enable debug logging
gin.SetMode(gin.DebugMode)

// Log HTMX requests
router.Use(func(c *gin.Context) {
    if c.GetHeader("HX-Request") == "true" {
        log.Printf("HTMX Request: %s %s", c.Request.Method, c.Request.URL)
    }
    c.Next()
})
```

---

## Deployment Checklist

- [ ] Set production environment variables
- [ ] Configure database connection pool
- [ ] Enable CORS for production domain
- [ ] Set up SSL/TLS certificates
- [ ] Configure rate limiting
- [ ] Enable request logging
- [ ] Set up monitoring/alerting
- [ ] Backup guidance.json and templates.json
- [ ] Test all HTMX flows
- [ ] Verify admin endpoints are protected

---

## Version History

- **v1.0**: Phase 1 - Core Project Management
- **v1.1**: Phase 2 - Tech Stack Integration
- **v1.2**: Phase 3 - Dynamic Guidance System
- **v1.3**: Phase 3.1 - Guidance Refactoring
- **v1.4**: Phase 4 - Stack-Based Templates

---

*Last Updated: December 2024*
*Documentation Version: 1.0*
