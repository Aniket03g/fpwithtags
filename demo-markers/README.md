# Demo Markers

A simple Go project for testing the FeaturePlus CLI marker/patch system.

## Project Structure

```
demo-markers/
├── cli/
│   └── commands.go      # CLI command handling with CLI_COMMANDS marker
├── routes/
│   └── routes.go        # HTTP routes with API_ROUTES marker
├── main.go              # Main application entry point with MAIN_INIT marker
├── go.mod               # Go module definition
├── markers_index.json   # Marker definitions for FeaturePlus CLI
└── README.md            # This file
```

## Markers

This project contains the following markers:

1. `CLI_COMMANDS` - Located in `cli/commands.go` inside the switch statement for handling CLI commands
2. `API_ROUTES` - Located in `routes/routes.go` inside the Gin HTTP router setup
3. `MAIN_INIT` - Located in `main.go` in the initialization logic

## Testing with FeaturePlus CLI

To test the FeaturePlus CLI with this project:

1. List all available markers:
   ```
   featureplus markers list
   ```

2. Find a specific marker:
   ```
   featureplus markers find API_ROUTES
   ```

3. Insert code at a marker:
   ```
   featureplus markers insert API_ROUTES
   ```
   This will show the context around the marker and prompt you to enter a code snippet.

4. Apply a patch:
   ```
   featureplus markers apply your-patch-file.fppatch
   ```

## Example Snippets

Here are some example snippets you can use when testing:

### For CLI_COMMANDS marker:

```go
case "status":
    fmt.Println("Checking status...")
    // Add status checking logic here
```

### For API_ROUTES marker:

```go
// User routes
r.GET("/api/users", UserHandler)
r.POST("/api/users", AuthMiddleware(), UserHandler)

// Product routes
r.GET("/api/products", ProductHandler)
```

### For MAIN_INIT marker:

```go
// Initialize logger
log := logrus.New()
log.SetFormatter(&logrus.JSONFormatter{})
log.SetLevel(logrus.InfoLevel)
```

## Running the Application

```
go run main.go
```

Or with CLI commands:

```
go run main.go help
go run main.go version
```
