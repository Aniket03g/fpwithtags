# Quick Start: Import Project API

## 🚀 Quick Test

### 1. Start the Server
```bash
cd backend
go run main.go
```

### 2. Get Authentication Token
Login to get your JWT token:
```bash
POST http://localhost:8080/api/auth/login
{
  "username": "your_username",
  "password": "your_password"
}
```

### 3. Import the Demo Project
```bash
POST http://localhost:8080/api/imports/import
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "project_id": "github_project_demo",
  "project_name": "My First Import",
  "description": "Testing the import feature"
}
```

### 4. Verify Import
Check the response for:
- ✅ `project_id` - The new project's database ID
- ✅ `features_created` - Should be 4
- ✅ `tasks_created` - Should be 6

## 📁 File Structure

```
backend/data/imports/
├── README.md                      # Full documentation
├── QUICK_START.md                 # This file
├── github_project_demo.json       # Sample import template
├── test_import.sh                 # Bash test script
└── test_import.ps1                # PowerShell test script
```

## 🔧 Adding Your Own Import

### Step 1: Create JSON File
Create `backend/data/imports/my_project.json`:
```json
{
  "id": "my_project",
  "name": "My Project",
  "tech_stack": "React",
  "features": [
    {
      "name": "Feature 1",
      "category": "Auth",
      "description": "Description here",
      "context": "Development"
    }
  ],
  "tasks": [
    {
      "name": "Task 1",
      "type": "Backend",
      "description": "Task description",
      "priority": "high",
      "context": "Development"
    }
  ]
}
```

### Step 2: Import It
```bash
POST /api/imports/import
{
  "project_id": "my_project",
  "project_name": "My Awesome Project"
}
```

## 🎯 Common Use Cases

### Use Case 1: GitHub MCP Integration
1. GitHub MCP analyzes a repository
2. Generates JSON with features/tasks
3. Saves to `backend/data/imports/<repo_name>.json`
4. Calls import API to create project

### Use Case 2: Manual Project Template
1. Create JSON file manually
2. Place in imports folder
3. Import via API
4. Customize features/tasks in UI

### Use Case 3: Batch Import
```bash
# List all available imports
GET /api/imports

# Import each one
for each import in list:
  POST /api/imports/import
```

## ⚠️ Important Notes

1. **Authentication Required**: All endpoints need valid JWT token
2. **File Naming**: Use lowercase with underscores (e.g., `my_project.json`)
3. **No Preloading**: Files are read on-demand, not at startup
4. **Context Filtering**: Features/tasks filtered by project context (Development/Production)

## 🐛 Troubleshooting

### Error: "Import template not found"
- Check file exists: `backend/data/imports/<project_id>.json`
- Verify file name matches `project_id` in request
- Check file permissions (should be readable)

### Error: "Failed to parse import template"
- Validate JSON syntax
- Ensure all required fields are present
- Check for trailing commas

### Error: "Authentication required"
- Include `Authorization: Bearer <token>` header
- Verify token is valid and not expired
- Login again if needed

## 📊 Response Format

### Success
```json
{
  "status": "success",
  "message": "Project imported successfully",
  "project_id": 12,
  "project_name": "My Project",
  "features_created": 4,
  "tasks_created": 6
}
```

### Error
```json
{
  "status": "error",
  "message": "Import template not found: my_project",
  "error": "detailed error message"
}
```

## 🔗 Related Endpoints

- `GET /api/imports` - List all available imports
- `POST /api/imports/save` - Save new import template
- `DELETE /api/imports/:id` - Delete import template
- `GET /api/projects` - List all projects (to verify import)

## 💡 Tips

1. **Test First**: Use the sample `github_project_demo.json` to test
2. **Validate JSON**: Use online JSON validators before importing
3. **Check Logs**: Server logs show detailed import progress
4. **Start Small**: Begin with 1-2 features, then expand

## 📚 Next Steps

1. Read full documentation: `README.md`
2. Review sample template: `github_project_demo.json`
3. Run test script: `test_import.ps1` or `test_import.sh`
4. Create your first import template
5. Integrate with GitHub MCP
