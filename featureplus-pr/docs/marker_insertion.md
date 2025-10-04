# Marker Insertion System

The Marker Insertion System is Phase 2 of the FeaturePlus CLI's marker management functionality. It allows for automated insertion of code markers around identified code blocks, making it easier to locate and modify specific parts of the codebase.

## Overview

The system works in two phases:

1. **Phase 1: Marker Awareness** - Identifies and manages existing markers in the codebase.
2. **Phase 2: Marker Insertion** - Inserts new markers around identified code blocks.

## Usage

### Basic Usage

```bash
featureplus markers insert
```

This command will read the default `markers_index.json` file and insert markers at the specified locations.

### Using a Custom Input File

```bash
featureplus markers insert --input examples/marker_candidates.json
```

This command reads marker candidates from a custom JSON file.

### Dry Run Mode

```bash
featureplus markers insert --dry-run
```

This shows what changes would be made without actually modifying any files.

### Filtering Markers

```bash
# Filter by file name
featureplus markers insert --file main.go

# Filter by directory
featureplus markers insert --dir handlers

# Filter by marker type
featureplus markers insert --type routes
```

### Customizing Marker Format

```bash
featureplus markers insert --start-format "// START: %s" --end-format "// END: %s"
```

### Adding a Prefix to Marker Names

```bash
featureplus markers insert --prefix v2
```

This will add a prefix to all marker names, e.g., `v2-API_ROUTES`.

### Repairing Malformed Markers

```bash
featureplus markers insert --repair
```

This will attempt to repair any malformed markers found in the codebase.

## Input Format

The marker insertion system accepts input in JSON format. Here's an example:

```json
{
  "candidates": [
    {
      "name": "API_ROUTES",
      "file": "backend/main.go",
      "line_number": 488,
      "type": "routes",
      "start_line": 488,
      "end_line": 526
    },
    {
      "name": "FEATURE_MODEL",
      "file": "backend/models/features.go",
      "line_number": 23,
      "type": "model",
      "start_line": 23,
      "end_line": 41
    }
  ]
}
```

Each candidate has the following properties:

- `name`: The name of the marker.
- `file`: The path to the file where the marker should be inserted.
- `line_number`: The line number where the marker should be inserted.
- `type`: The type of code block (e.g., routes, model, handler).
- `start_line`: The line number where the code block starts (optional).
- `end_line`: The line number where the code block ends (optional).

## Marker Format

By default, markers are inserted in the following format:

```
// @fp:marker-start:<name>
... code ...
// @fp:marker-end:<name>
```

You can customize this format using the `--start-format` and `--end-format` flags.

## Idempotency

The marker insertion system is designed to be idempotent. If markers already exist for a given code block, they will not be inserted again. This ensures that running the command multiple times will not result in duplicate markers.

## Error Handling

If any errors occur during the insertion process, they will be reported in the summary output. The system will continue processing other files even if errors are encountered.

## Examples

### Insert Markers for Routes Only

```bash
featureplus markers insert --type routes
```

### Insert Markers in a Specific Directory

```bash
featureplus markers insert --dir backend/handlers
```

### Dry Run with Custom Format

```bash
featureplus markers insert --dry-run --start-format "// BEGIN %s" --end-format "// END %s"
```

## Integration with Phase 1

The marker insertion system builds on the marker awareness functionality from Phase 1. You can use the `markers list` and `markers find` commands to verify that markers have been inserted correctly.
