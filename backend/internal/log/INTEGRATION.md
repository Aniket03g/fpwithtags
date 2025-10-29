# Integration Guide

This guide shows how to integrate the structured logger into your FeaturePlus application.

## Step 1: Import the Logger

Replace standard log imports with the structured logger:

```go
// Before
import "log"

// After
import "backend/internal/log"
```

## Step 2: Update main.go

Add logger initialization message in your main function:

```go
package main

import (
    "backend/internal/log"
    "github.com/sirupsen/logrus"
)

func main() {
    // Set log level from environment or default to info
    log.Info("FeaturePlus application starting...")
    
    // Your existing initialization code
    // ...
    
    log.WithFields(logrus.Fields{
        "port": 8080,
        "environment": os.Getenv("ENV"),
    }).Info("Server started successfully")
}
```

## Step 3: Update Handlers

Add structured logging to your HTTP handlers:

```go
func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        log.WithError(err).Error("Failed to parse login request")
        c.JSON(400, gin.H{"error": "Invalid request"})
        return
    }
    
    log.WithFields(logrus.Fields{
        "username": req.Username,
        "ip": c.ClientIP(),
    }).Info("Login attempt")
    
    // Your login logic
    // ...
    
    log.WithField("user_id", user.ID).Info("User logged in successfully")
}
```

## Step 4: Update Database Operations

Add logging to database operations:

```go
func CreateRelease(release *models.Release) error {
    log.WithFields(logrus.Fields{
        "release_name": release.Name,
        "tag": release.Tag,
    }).Debug("Creating new release")
    
    if err := db.Create(release).Error; err != nil {
        log.WithError(err).WithField("release_name", release.Name).Error("Failed to create release")
        return err
    }
    
    log.WithField("release_id", release.ID).Info("Release created successfully")
    return nil
}
```

## Step 5: Update Middleware

Add request logging middleware:

```go
func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        // Process request
        c.Next()
        
        // Log request details
        log.WithFields(logrus.Fields{
            "method": c.Request.Method,
            "path": path,
            "status": c.Writer.Status(),
            "duration_ms": time.Since(start).Milliseconds(),
            "ip": c.ClientIP(),
            "user_agent": c.Request.UserAgent(),
        }).Info("HTTP request")
    }
}

// Register middleware
router.Use(LoggingMiddleware())
```

## Step 6: Update Error Handling

Improve error logging with context:

```go
func ProcessPullRequest(prID int) error {
    logger := log.WithField("pr_id", prID)
    
    logger.Debug("Fetching pull request")
    pr, err := fetchPR(prID)
    if err != nil {
        logger.WithError(err).Error("Failed to fetch pull request")
        return err
    }
    
    logger.WithField("pr_title", pr.Title).Info("Processing pull request")
    
    // Processing logic
    // ...
    
    logger.Info("Pull request processed successfully")
    return nil
}
```

## Step 7: Environment Configuration

Set log level in your environment:

### Development (.env)
```bash
LOG_LEVEL=debug
```

### Production (.env.docker)
```bash
LOG_LEVEL=info
```

### Docker Compose
```yaml
services:
  featureplus:
    environment:
      - LOG_LEVEL=info
```

## Step 8: Viewing Logs

### Local Development
Logs appear in:
- Console (stdout)
- File: `/var/log/featureplus/app.log` (or `\var\log\featureplus\app.log` on Windows)

### Docker
Logs appear in:
- Docker logs: `docker-compose logs -f`
- Host file: `./logs/app.log`

### Parsing JSON Logs

Use `jq` to parse and filter JSON logs:

```bash
# View all logs
cat logs/app.log | jq

# Filter by level
cat logs/app.log | jq 'select(.level=="error")'

# Filter by field
cat logs/app.log | jq 'select(.user_id==123)'

# View recent errors
tail -f logs/app.log | jq 'select(.level=="error")'
```

## Migration Checklist

- [ ] Replace all `log.Printf` with `log.Infof`
- [ ] Replace all `log.Println` with `log.Info`
- [ ] Add structured fields to important logs
- [ ] Add error context with `WithError(err)`
- [ ] Update middleware to use structured logging
- [ ] Set appropriate log levels (Debug for dev, Info for prod)
- [ ] Test log output in both console and file
- [ ] Verify Docker volume mapping works
- [ ] Document important log fields for your team

## Best Practices Summary

1. **Always add context**: Use `WithFields` to add relevant information
2. **Use appropriate levels**: Debug < Info < Warn < Error < Fatal
3. **Include error details**: Use `WithError(err)` for error logging
4. **Be consistent**: Use the same field names across the application
5. **Don't log sensitive data**: Avoid passwords, tokens, etc.
6. **Use structured fields**: Prefer fields over string formatting
7. **Log at boundaries**: Log at API entry/exit, database operations, external calls

## Common Patterns

### API Request Logging
```go
log.WithFields(logrus.Fields{
    "method": req.Method,
    "path": req.URL.Path,
    "user_id": userID,
}).Info("API request")
```

### Database Operation Logging
```go
log.WithFields(logrus.Fields{
    "operation": "create",
    "table": "releases",
    "id": release.ID,
}).Debug("Database operation")
```

### Error Logging
```go
log.WithError(err).WithFields(logrus.Fields{
    "operation": "finalize_release",
    "release_id": releaseID,
}).Error("Operation failed")
```

### Performance Logging
```go
start := time.Now()
// ... operation ...
log.WithFields(logrus.Fields{
    "operation": "cherry_pick",
    "duration_ms": time.Since(start).Milliseconds(),
}).Debug("Operation completed")
```
