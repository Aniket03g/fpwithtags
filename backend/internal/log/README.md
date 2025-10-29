# Structured Logging with Logrus

This package provides a centralized, structured logging solution for FeaturePlus using [logrus](https://github.com/sirupsen/logrus).

## Features

- ✅ **JSON formatted logs** for easy parsing and analysis
- ✅ **Dual output**: Writes to both stdout and file (`/var/log/featureplus/app.log`)
- ✅ **Automatic directory creation**: Creates log directory if missing
- ✅ **Structured fields**: Add context to logs with key-value pairs
- ✅ **Environment-based log levels**: Configure via `LOG_LEVEL` env var
- ✅ **Docker-friendly**: Logs appear in host's `./logs/` directory

## Quick Start

### Basic Usage

```go
import "backend/internal/log"

func main() {
    log.Info("Application started")
    log.Debugf("Processing request for user: %s", userID)
    log.Error("Failed to connect to database")
}
```

### Structured Logging

```go
import (
    "backend/internal/log"
    "github.com/sirupsen/logrus"
)

// Add context with fields
log.WithFields(logrus.Fields{
    "user_id": 123,
    "action": "login",
    "ip": "192.168.1.1",
}).Info("User logged in")

// Single field
log.WithField("module", "auth").Info("Authentication successful")

// Error context
if err != nil {
    log.WithError(err).Error("Failed to process request")
}
```

### Using the Logger Instance

```go
logger := log.Logger()
logger.WithField("module", "database").Info("Connection established")
```

## Configuration

### Log Levels

Set the `LOG_LEVEL` environment variable to control verbosity:

- `debug` - Most verbose, includes all logs
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages only

**Example:**
```bash
export LOG_LEVEL=debug
go run main.go
```

Or in Docker:
```yaml
environment:
  - LOG_LEVEL=debug
```

### Log Output

Logs are written to two destinations:

1. **stdout** - Console output for real-time monitoring
2. **File** - `/var/log/featureplus/app.log` for persistence

In Docker, the file logs are mapped to the host's `./logs/` directory via volume mounting.

## JSON Log Format

All logs are formatted as JSON for easy parsing:

```json
{
  "level": "info",
  "msg": "User logged in",
  "time": "2025-01-15 14:30:45",
  "user_id": 123,
  "action": "login",
  "ip": "192.168.1.1"
}
```

## API Reference

### Logging Functions

- `Info(args...)` / `Infof(format, args...)` - Info level
- `Debug(args...)` / `Debugf(format, args...)` - Debug level
- `Warn(args...)` / `Warnf(format, args...)` - Warning level
- `Error(args...)` / `Errorf(format, args...)` - Error level
- `Fatal(args...)` / `Fatalf(format, args...)` - Fatal level (exits)
- `Panic(args...)` / `Panicf(format, args...)` - Panic level (panics)

### Context Functions

- `WithFields(fields)` - Add multiple fields
- `WithField(key, value)` - Add single field
- `WithError(err)` - Add error context

### Utility Functions

- `Logger()` - Get logger instance
- `SetLevel(level)` - Change log level
- `GetLevel()` - Get current log level

## Docker Integration

The `docker-compose.yml` includes a volume mapping:

```yaml
volumes:
  - ./logs:/var/log/featureplus
```

This ensures logs written inside the container appear in the host's `./logs/` directory.

## Best Practices

1. **Use structured fields** instead of string formatting:
   ```go
   // Good
   log.WithFields(logrus.Fields{
       "user_id": userID,
       "action": "login",
   }).Info("User action")
   
   // Avoid
   log.Infof("User %d performed %s", userID, "login")
   ```

2. **Add context to errors**:
   ```go
   if err != nil {
       log.WithError(err).WithField("user_id", userID).Error("Failed to save user")
   }
   ```

3. **Use appropriate log levels**:
   - `Debug` - Detailed debugging information
   - `Info` - General informational messages
   - `Warn` - Warning messages (non-critical issues)
   - `Error` - Error messages (failures)
   - `Fatal` - Critical errors (application exits)

4. **Include relevant context**:
   ```go
   log.WithFields(logrus.Fields{
       "module": "auth",
       "function": "Login",
       "user_id": userID,
   }).Info("Authentication successful")
   ```

## Examples

See `logger_example.go` for more usage examples.

## Migration from Standard Log

Replace standard library log calls:

```go
// Before
import "log"
log.Printf("User %d logged in", userID)

// After
import "backend/internal/log"
log.WithField("user_id", userID).Info("User logged in")
```

## Troubleshooting

### Logs not appearing in file

1. Check directory permissions: `/var/log/featureplus` must be writable
2. In Docker, ensure volume mapping is correct in `docker-compose.yml`
3. Check logs in stdout for initialization errors

### Log level not changing

Ensure `LOG_LEVEL` environment variable is set before the application starts:

```bash
export LOG_LEVEL=debug
go run main.go
```

Or in Docker:
```yaml
environment:
  - LOG_LEVEL=debug
```
