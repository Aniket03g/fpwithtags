# Markers Commands

The `markers` commands allow you to work with code markers defined in the `markers_index.json` file. These markers indicate specific locations in the codebase where code can be inserted or modified.

## Available Commands

### List Markers

Lists all markers defined in the project.

```
featureplus markers list
```

**Options:**
- `-v, --verbose`: Enable verbose output with additional information

**Example:**
```
> featureplus markers list

NAME                     FILE                                      LINE_HINT
API_ROUTES               FeaturePlus/backend/main.go               // MARKER:API_ROUTES
WEB_ROUTES               FeaturePlus/backend/main.go               // MARKER:WEB_ROUTES
FRAGMENT_ROUTES          FeaturePlus/backend/main.go               // MARKER:FRAGMENT_ROUTES
FRAGMENT_URL_MAPPING     FeaturePlus/backend/main.go               // MARKER:FRAGMENT_URL_MAPPING
MODEL_MIGRATIONS         FeaturePlus/backend/main.go               // MARKER:MODEL_MIGRATIONS
TEMPLATE_API_ROUTES      FeaturePlus/backend/routes/template_routes.go  // MARKER:TEMPLATE_API_ROUTES
PROJECT_CONFIG_DEFAULTS  FeaturePlus/backend/models/project.go     // MARKER:PROJECT_CONFIG_DEFAULTS
FEATURE_FIELD_UPDATES    FeaturePlus/backend/handlers/feature_handler.go  // MARKER:FEATURE_FIELD_UPDATES
AUTH_ROUTES              FeaturePlus/backend/routes/auth_routes.go  // MARKER:AUTH_ROUTES
DEPENDENCY_FRAGMENT_ROUTES  FeaturePlus/backend/routes/dependency_routes.go  // MARKER:DEPENDENCY_FRAGMENT_ROUTES

Found 10 markers in markers_index.json
```

### Find Marker

Find a specific marker by name and display its details.

```
featureplus markers find <name>
```

**Arguments:**
- `name`: The name of the marker to find (required)

**Options:**
- `-v, --verbose`: Enable verbose output with additional information
- `-f, --format`: Output format (text or json)

**Example:**
```
> featureplus markers find API_ROUTES

Name: API_ROUTES
File: FeaturePlus/backend/main.go
Line Hint: // MARKER:API_ROUTES
```

**JSON Output Example:**
```
> featureplus markers find API_ROUTES -f json

{"name": "API_ROUTES", "file": "FeaturePlus/backend/main.go", "line_hint": "// MARKER:API_ROUTES"}
```

## Notes

- The `markers_index.json` file should be located in the project root directory.
- If the file is not found, the command will search up to 4 directory levels up from the current directory.
- Run the commands from the project root directory for best results.
