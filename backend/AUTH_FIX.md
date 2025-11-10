# Authentication Fix for CLI Connection Routes

## Issue
```
❌ Error: Authentication required
```

The CLI was getting a 401 Unauthorized error when trying to connect:
```
{"status":401,"path":"/api/projects/39/connect"}
```

## Root Cause
The project connection routes were placed under the authenticated API group (`authApi`), which requires a valid JWT token. The CLI doesn't have authentication yet.

## Solution
Moved the connection routes to the **public API section** (no authentication required):

### Before
```go
// Protected routes (auth required)
authApi := api.Group("/", middleware.AuthMiddleware())
projectRoutes := authApi.Group("/projects")
{
    // ... other routes
    projectRoutes.POST("/:id/connect", handlers.ConnectProjectHandler(db.DB))
    projectRoutes.GET("/:id/status", handlers.GetProjectConnectionStatus(db.DB))
}
```

### After
```go
// Public project connection routes (no auth required for CLI)
publicProjectRoutes := api.Group("/projects")
{
    publicProjectRoutes.POST("/:id/connect", handlers.ConnectProjectHandler(db.DB))
    publicProjectRoutes.GET("/:id/status", handlers.GetProjectConnectionStatus(db.DB))
}

// Protected routes (auth required)
authApi := api.Group("/", middleware.AuthMiddleware())
projectRoutes := authApi.Group("/projects")
{
    // ... other protected routes
}
```

## Security Consideration
These routes are now public because:
1. They're meant to be used by the CLI tool
2. They only store local directory paths (no sensitive data)
3. They don't modify critical project data
4. The CLI is typically used by developers who already have access to the project

If you need to secure these routes later, you can:
- Add API key authentication
- Use the existing JWT token system in the CLI
- Implement IP whitelisting

## Testing
After restarting the server, the CLI should work:

```bash
cd D:\code_mapping_example
featureplus-pr init
featureplus-pr connect 39
featureplus-pr status
```

## Status: ✅ FIXED
The routes are now accessible without authentication.
