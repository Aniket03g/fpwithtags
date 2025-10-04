# Contextual Patch System

The Contextual Patch System (Phase 3) enhances the FeaturePlus CLI with the ability to propose code changes around markers in a controlled, PR-like workflow. Instead of directly modifying files, developers create `.fppatch` files that describe the intended snippet insertion, which can be reviewed before application.

## Overview

The system works in two main steps:

1. **Create a Patch**: Locate a marker, show context, and prompt for a code snippet to insert.
2. **Apply a Patch**: Find the marker in the original file and insert the snippet with proper formatting.

## Creating Patches

To create a patch, use the `insert` command with a marker name:

```bash
featureplus markers insert MARKER_NAME
```

This will:
1. Locate the marker in the codebase
2. Show ~5 lines of context above and below the marker
3. Prompt you to enter your code snippet
4. Save the snippet to a `.fppatch` file

### Example

```bash
$ featureplus markers insert SWITCH_CASES

Context for marker 'SWITCH_CASES' in file 'backend/handlers/feature_handler.go':

--- Before ---
fmt.Printf("Existing feature: %+v\n", existingFeature)

// MARKER:UPDATE_FEATURE_SWITCH Switch for updating feature fields
// Validate and update the specific field
switch updateData.Field {

--- Insertion Point (line 454) ---

--- After ---
case "title":
    if title, ok := updateData.Value.(string); ok {
        existingFeature.Title = title
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title value"})
        return
    }

Enter your code snippet (type 'EOF' on a new line to finish):
case "delete":
    if _, ok := updateData.Value.(bool); ok {
        existingFeature.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
    } else {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delete value"})
        return
    }
EOF

Patch saved to patch-switch-cases.fppatch

To apply this patch, run:
  featureplus markers apply patch-switch-cases.fppatch
```

### Options

- `--context N`: Show N lines of context (default: 5)
- `--output FILE`: Specify the output filename (default: auto-generated)
- `--verbose`: Show verbose output

## Applying Patches

To apply a patch, use the `apply` command with the patch file:

```bash
featureplus markers apply PATCH_FILE
```

This will:
1. Find the marker in the original file
2. Insert the snippet with a `// inserted by featureplus` comment above it
3. Run `go fmt` automatically for Go files
4. Safely overwrite the original file

### Example

```bash
$ featureplus markers apply patch-switch-cases.fppatch

Applying patch to marker 'SWITCH_CASES' in file 'backend/handlers/feature_handler.go'...
Patch applied successfully!
```

## Patch File Format

Patch files use YAML format and have the following structure:

```yaml
marker: SWITCH_CASES
snippet: |
  case "delete":
    if _, ok := updateData.Value.(bool); ok {
      existingFeature.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
    } else {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid delete value"})
      return
    }
target_file: /path/to/backend/handlers/feature_handler.go
context:
  before:
    - 'fmt.Printf("Existing feature: %+v\n", existingFeature)'
    - '// MARKER:UPDATE_FEATURE_SWITCH Switch for updating feature fields'
    - '// Validate and update the specific field'
    - 'switch updateData.Field {'
  after:
    - 'case "title":'
    - '    if title, ok := updateData.Value.(string); ok {'
    - '        existingFeature.Title = title'
    - '    } else {'
    - '        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid title value"})'
    - '        return'
    - '    }'
```

## Safety Features

The Contextual Patch System includes several safety features:

1. **Context Validation**: Before applying a patch, the system validates that the context around the marker hasn't changed since the patch was created.

2. **Temporary Files**: Changes are first written to a temporary file before replacing the original, ensuring that the original file is only modified if all steps succeed.

3. **Formatting**: For Go files, `go fmt` is automatically run to ensure proper formatting.

4. **Comments**: Inserted code is marked with a comment to indicate that it was added by FeaturePlus.

## Workflow Benefits

This approach provides several benefits:

1. **Review Before Apply**: Developers can review the patch file before applying it, similar to a PR workflow.

2. **Version Control**: Patch files can be committed to version control, allowing for review and collaboration.

3. **Automation**: The system can be integrated into CI/CD pipelines to automatically apply patches.

4. **Safety**: The original file is only modified when explicitly requested, reducing the risk of accidental changes.

## Example Workflow

1. **Identify a Marker**: Find a marker where you want to add code.

2. **Create a Patch**:
   ```bash
   featureplus markers insert SWITCH_CASES
   ```

3. **Review the Patch**:
   ```bash
   cat patch-switch-cases.fppatch
   ```

4. **Apply the Patch**:
   ```bash
   featureplus markers apply patch-switch-cases.fppatch
   ```

5. **Verify the Changes**:
   ```bash
   git diff backend/handlers/feature_handler.go
   ```
