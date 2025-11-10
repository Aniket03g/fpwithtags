# Git-Based Code Mapping Architecture

## Overview

The `featureplus map` command implements git-based code mapping to automatically track which files and commits belong to each feature by analyzing git history.

## Architecture

### Components

```
cmd/map.go                    → CLI command handler
internal/git/scanner.go       → Git log parsing and commit extraction
internal/feature/mapper.go    → Feature-commit aggregation logic
internal/feature/pull.go      → Feature manifest I/O
```

### Data Flow

```
1. User runs: featureplus map --commits 50

2. cmd/map.go
   ├─ Validates project initialization
   ├─ Validates git repository
   └─ Calls feature.MapGitHistoryToFeatures()

3. feature.MapGitHistoryToFeatures()
   ├─ Loads all local manifests from .featureplus/features/
   ├─ Calls git.ScanGitHistory(50)
   └─ Calls git.MapCommitsToFeatures()

4. git.ScanGitHistory()
   ├─ Runs: git log --name-only --pretty=format:"%H|%an|%ae|%ad|%s" --date=iso -n 50
   ├─ Parses output into CommitInfo structs
   └─ Extracts feature IDs using regex: FTR-(\d+)

5. git.MapCommitsToFeatures()
   ├─ Groups commits by feature ID
   ├─ Deduplicates files per feature
   └─ Returns map[string]*FeatureMapping

6. cmd/map.go
   ├─ Sorts results by feature ID
   ├─ Displays formatted output
   └─ Shows summary statistics
```

## Data Structures

### CommitInfo
```go
type CommitInfo struct {
    Hash      string    // Git commit hash
    Message   string    // Commit message
    Author    string    // Commit author
    Date      time.Time // Commit date
    Files     []string  // Changed files
    FeatureID string    // Extracted feature ID (e.g., "FTR-001")
}
```

### FeatureMapping
```go
type FeatureMapping struct {
    FeatureID string                // e.g., "FTR-001"
    Commits   []CommitInfo          // All commits for this feature
    Files     map[string]bool       // Unique files (map for deduplication)
}
```

### MapResult
```go
type MapResult struct {
    FeatureID    string   // e.g., "FTR-001"
    FeatureName  string   // From local manifest
    Files        []string // Unique file paths
    CommitCount  int      // Number of commits
    CommitHashes []string // Git commit hashes
}
```

## Git Log Format

### Command
```bash
git log --name-only --pretty=format:"%H|%an|%ae|%ad|%s" --date=iso -n 50
```

### Output Format
```
abc123def456|John Doe|john@example.com|2025-11-10 12:00:00 +0530|FTR-001: Add user login
backend/auth/login.go
backend/routes.go

ghi789jkl012|Jane Smith|jane@example.com|2025-11-10 13:00:00 +0530|FTR-002: Dashboard UI
ui/components/dashboard.tsx
ui/styles/dashboard.css
```

### Parsing Logic
1. Lines with `|` are commit headers
2. Split by `|` to extract: hash, author, email, date, message
3. Lines without `|` are file paths
4. Empty lines separate commits
5. Extract feature ID from message using regex

## Feature ID Extraction

### Pattern
```go
var featureIDPattern = regexp.MustCompile(`FTR-(\d+)`)
```

### Examples
- `FTR-001: Add login` → `FTR-001`
- `[FTR-123] Fix bug` → `FTR-123`
- `Implement (FTR-042)` → `FTR-042`
- `No feature ID` → `""` (empty)

### Rules
- Case-sensitive (must be uppercase `FTR`)
- Matches `FTR-` followed by one or more digits
- Returns first match if multiple IDs present
- Returns empty string if no match

## File Deduplication

Files are deduplicated per feature using a map:

```go
mapping.Files = make(map[string]bool)

for _, commit := range commits {
    for _, file := range commit.Files {
        mapping.Files[file] = true  // Automatically deduplicates
    }
}
```

## Output Formatting

### Summary Header
```
🔍 Scanning git history (last 50 commits)...
📊 Found 2 feature(s) in git history:
```

### Per-Feature Display
```
📦 FTR-001 → 3 file(s), 2 commit(s)
   Name: Add User Login
   - backend/auth/login.go
   - backend/routes.go
   - ui/pages/login.html
   Commits: a1b2c3d e4f5g6h
```

### Display Limits
- **Files**: Show first 10, then "... and N more"
- **Commits**: Show first 3 hashes (7 chars), then "(+N more)"
- **Sorting**: Features sorted by ID, files sorted alphabetically

### Footer
```
✅ Mapping complete! Found 2 feature(s) with 6 total commits.
```

## Error Handling

### Validation Errors
- Not initialized → Check `.featureplus/` exists
- Not a git repo → Check `.git/` exists
- No manifests → Check `.featureplus/features/` has YAML files

### Git Errors
- Git not installed → `exec: "git": executable file not found`
- Git command fails → Parse stderr and display
- Invalid git repo → Git returns exit code 128

### Parsing Errors
- Invalid date format → Use zero time, continue
- Malformed log line → Skip line, log warning
- Empty output → Return empty commits list

## Testing

### Unit Tests
Located in `internal/git/scanner_test.go`:

1. **TestExtractFeatureID** - Tests feature ID extraction
2. **TestParseGitLog** - Tests git log parsing
3. **TestMapCommitsToFeatures** - Tests commit aggregation

### Test Coverage
```bash
go test ./internal/git/... -cover
```

### Manual Testing
```bash
# In a git repository with feature commits
featureplus init
featureplus pull FTR-001
featureplus map
featureplus map --commits 100
```

## Future Enhancements

### Phase 4: Manifest Updates
- Store mapping results back to YAML manifests
- Update `files`, `commits`, `prs` arrays
- Track last sync timestamp

### Potential Features
- Filter by date range: `--since`, `--until`
- Filter by author: `--author`
- Export to JSON: `--format json`
- Incremental updates: only scan new commits
- Support multiple feature ID patterns
- Integration with GitHub API for PR data

## Performance Considerations

### Commit Limit
- Default: 50 commits (fast, covers recent work)
- Large repos: Use `--commits 1000` (slower but comprehensive)
- Very large repos: Consider incremental scanning

### Memory Usage
- Each commit stores: hash, message, author, date, files
- Files are deduplicated per feature (not globally)
- Typical usage: ~1MB for 1000 commits

### Optimization Opportunities
- Cache parsed git log between runs
- Use `git log --since` for incremental updates
- Stream processing for very large histories
- Parallel processing of commits

## Integration Points

### Current
- Reads from: `.featureplus/features/*.yaml`
- Writes to: stdout (display only)
- Depends on: git CLI, local manifests

### Future
- Write to: `.featureplus/features/*.yaml` (update manifests)
- Sync to: FeaturePlus backend API
- Integrate with: PR tracking, release management
