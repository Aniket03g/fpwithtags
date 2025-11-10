# FeaturePlus Code Mapping Workflow

Complete guide to mapping your codebase to features using the FeaturePlus CLI.

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Complete Workflow](#complete-workflow)
4. [CLI Commands Reference](#cli-commands-reference)
5. [Local Manifest Structure](#local-manifest-structure)
6. [Database Schema](#database-schema)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

---

## Overview

FeaturePlus Code Mapping allows you to:
- **Track which code files** belong to which features
- **Link git commits** to specific features
- **Sync mappings** to the backend for team visibility
- **Visualize feature progress** based on actual code changes

### How It Works

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐      ┌──────────────┐
│   Backend   │─────▶│  Local YAML  │─────▶│ Git History │─────▶│   Backend    │
│  (Features) │      │  (Manifests) │      │  (Commits)  │      │ (Files/Commits)│
└─────────────┘      └──────────────┘      └─────────────┘      └──────────────┘
   featureplus          featureplus            featureplus          featureplus
      pull                 map                    map                  sync
```

---

## Prerequisites

1. **Git repository** - Your project must be a git repository
2. **FeaturePlus backend** - Running at `http://localhost:8080` (or configured URL)
3. **FeaturePlus CLI** - Installed and in your PATH
4. **Features created** - At least one feature exists in the backend

---

## Complete Workflow

### Step 1: Initialize FeaturePlus in Your Project

Navigate to your project's root directory and initialize FeaturePlus:

```bash
cd /path/to/your/project
featureplus init
```

**What it does:**
- Creates `.featureplus/` directory
- Creates `.featureplus/config.yaml` with project configuration
- Prompts for project ID (optional)

**Example output:**
```
✅ FeaturePlus initialized in current directory
   Created .featureplus/config.yaml
```

**Generated structure:**
```
your-project/
├── .featureplus/
│   ├── config.yaml
│   └── features/        # Will be created when you pull features
├── .git/
└── ... (your code)
```

---

### Step 2: Pull Feature Metadata from Backend

Fetch a feature's metadata from the FeaturePlus backend:

```bash
featureplus pull <feature-id>
```

**Examples:**
```bash
# Using numeric ID
featureplus pull 35

# Using FTR-XXX format
featureplus pull FTR-035
```

**What it does:**
- Calls `GET /api/features/{id}` on the backend
- Fetches feature metadata (name, description, status)
- Creates `.featureplus/features/FTR-035.yaml` locally
- Initializes empty arrays for files, commits, and PRs

**Example output:**
```
⬇️  Fetching feature FTR-035 from backend...
✅ Pulled feature FTR-035 (Filtering by tags)
   Saved to .featureplus\features\FTR-035.yaml
```

**Generated YAML:**
```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag
status: todo
owner: ""
files: []
prs: []
commits: []
created_at: 2025-11-10T12:16:00Z
synced_from_backend: true
```

---

### Step 3: Work on the Feature (Code & Commit)

Make your code changes and commit them with the **feature ID in the commit message**:

```bash
# Make changes to files
git add .
git commit -m "FTR-035: Add autocomplete dropdown for tags"

# More changes
git add .
git commit -m "FTR-035: Style tag autocomplete component"

# Another commit
git add .
git commit -m "FTR-035: Add keyboard navigation to autocomplete"
```

**Important:** 
- Include `FTR-035` (or your feature ID) in the commit message
- The CLI uses regex pattern `FTR-(\d+)` to match feature IDs
- Commits without feature IDs won't be mapped

**Example commits:**
```
f3a5ffb - FTR-035: Add autocomplete dropdown for tags
a55bb85 - FTR-035: Style tag autocomplete component
c7d8e9f - FTR-035: Add keyboard navigation to autocomplete
```

---

### Step 4: Map Git History to Features

Scan your git history and map commits to local feature manifests:

```bash
featureplus map
```

**Optional: Scan more commits**
```bash
# Scan last 100 commits instead of default 50
featureplus map --commits 100
```

**What it does:**
1. Reads all local feature manifests from `.featureplus/features/`
2. Runs `git log --name-only` to get commit history
3. Matches feature IDs in commit messages using regex `FTR-(\d+)`
4. Extracts changed files from each commit
5. Aggregates files and commits per feature
6. Updates local YAML manifests with:
   - List of unique files touched
   - List of commit hashes
   - `updated_at` timestamp
7. Deduplicates files and commits automatically

**Example output:**
```
🔍 Scanning git history (last 50 commits)...

📊 Found 1 feature(s) in git history:

📦 FTR-035 → 2 file(s), 2 commit(s)
   Name: Filtering by tags
   - autocomplete.py
   - navigate.py
   Commits: f3a5ffb a55bb85

📝 Updating feature manifests...
   ✅ Updated .featureplus/features/FTR-035.yaml (2 files, 2 commits)

✅ Mapping complete! Found 1 feature(s) with 2 total commits.
   Updated 1 manifest(s).
```

**Updated YAML:**
```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag
status: todo
owner: ""
files:
  - autocomplete.py
  - navigate.py
prs: []
commits:
  - f3a5ffb
  - a55bb85
created_at: 2025-11-10T12:16:00Z
updated_at: 2025-11-10T12:29:00Z  # ← Updated by map command
synced_from_backend: true
```

---

### Step 5: Sync Mappings to Backend

Upload the file and commit mappings to the FeaturePlus backend:

```bash
featureplus sync
```

**What it does:**
1. Reads all feature manifests from `.featureplus/features/`
2. Identifies features that need syncing:
   - Has files OR commits (not empty)
   - Never synced before (`last_synced` is empty), OR
   - Updated after last sync (`updated_at` > `last_synced`)
3. For each feature, sends `POST /api/features/sync` with:
   ```json
   {
     "feature_id": "FTR-035",
     "files": ["autocomplete.py", "navigate.py"],
     "commits": ["f3a5ffb", "a55bb85"],
     "status": "todo"
   }
   ```
4. Backend stores data in `feature_files` and `feature_commits` tables
5. Updates `last_synced` timestamp in local manifest

**Example output:**
```
☁️  Syncing features to FeaturePlus backend...

   ✅ Synced FTR-035 (2 files, 2 commits)

☁️  Synced 1 feature(s) with FeaturePlus backend.
```

**Backend logs:**
```
Synced feature 35: 2 files, 2 commits stored in database
```

**Updated YAML:**
```yaml
id: FTR-035
name: Filtering by tags
files:
  - autocomplete.py
  - navigate.py
commits:
  - f3a5ffb
  - a55bb85
created_at: 2025-11-10T12:16:00Z
updated_at: 2025-11-10T12:29:00Z
last_synced: 2025-11-10T12:40:00Z  # ← Added by sync command
synced_from_backend: true
```

---

## CLI Commands Reference

### `featureplus init`

Initialize FeaturePlus in the current directory.

```bash
featureplus init
```

**Creates:**
- `.featureplus/` directory
- `.featureplus/config.yaml`

**No arguments required.**

---

### `featureplus pull <feature-id>`

Pull feature metadata from backend and save locally.

```bash
featureplus pull <feature-id>
```

**Arguments:**
- `<feature-id>` - Feature ID (numeric or FTR-XXX format)

**Examples:**
```bash
featureplus pull 35
featureplus pull FTR-035
```

**API Call:**
```
GET /api/features/{id}
```

**Creates:**
- `.featureplus/features/FTR-{id}.yaml`

---

### `featureplus map`

Scan git history and map commits to features.

```bash
featureplus map [--commits N]
```

**Flags:**
- `--commits N` - Number of commits to scan (default: 50)

**Examples:**
```bash
featureplus map
featureplus map --commits 100
```

**What it scans:**
- Git commit messages for feature IDs
- Changed files in each commit
- Commit hashes and metadata

**Updates:**
- `.featureplus/features/*.yaml` files
- Adds files and commits arrays
- Sets `updated_at` timestamp

---

### `featureplus sync`

Sync local mappings to backend.

```bash
featureplus sync
```

**No arguments required.**

**API Call:**
```
POST /api/features/sync
```

**Payload:**
```json
{
  "feature_id": "FTR-035",
  "files": ["file1.py", "file2.py"],
  "commits": ["abc123", "def456"],
  "status": "todo"
}
```

**Updates:**
- Backend database (`feature_files` and `feature_commits` tables)
- Local manifest (`last_synced` timestamp)

---

## Local Manifest Structure

### Initial State (After `pull`)

```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag
status: todo
owner: ""
files: []
prs: []
commits: []
created_at: 2025-11-10T12:16:00Z
synced_from_backend: true
```

### After `map`

```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag
status: todo
owner: ""
files:
  - autocomplete.py
  - navigate.py
prs: []
commits:
  - f3a5ffb
  - a55bb85
created_at: 2025-11-10T12:16:00Z
updated_at: 2025-11-10T12:29:00Z  # ← Added
synced_from_backend: true
```

### After `sync`

```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag
status: todo
owner: ""
files:
  - autocomplete.py
  - navigate.py
prs: []
commits:
  - f3a5ffb
  - a55bb85
created_at: 2025-11-10T12:16:00Z
updated_at: 2025-11-10T12:29:00Z
last_synced: 2025-11-10T12:40:00Z  # ← Added
synced_from_backend: true
```

---

## Database Schema

### `feature_files` Table

Stores file-to-feature mappings.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `feature_id` | INTEGER | Foreign key to `features.id` |
| `file_path` | VARCHAR(500) | Relative file path from repo root |
| `created_at` | TIMESTAMP | When mapping was created |
| `updated_at` | TIMESTAMP | When mapping was last updated |
| `deleted_at` | TIMESTAMP | Soft delete timestamp (nullable) |

**Example data:**
```
id | feature_id | file_path        | created_at              | updated_at
---+------------+------------------+-------------------------+-------------------------
1  | 35         | autocomplete.py  | 2025-11-10 12:40:15.123 | 2025-11-10 12:40:15.123
2  | 35         | navigate.py      | 2025-11-10 12:40:15.124 | 2025-11-10 12:40:15.124
```

**Indexes:**
- `feature_id` (for fast lookups)
- `feature_id + file_path` (unique constraint to prevent duplicates)

---

### `feature_commits` Table

Stores commit-to-feature mappings.

| Column | Type | Description |
|--------|------|-------------|
| `id` | INTEGER | Primary key |
| `feature_id` | INTEGER | Foreign key to `features.id` |
| `commit_hash` | VARCHAR(40) | Git SHA-1 hash (short or full) |
| `created_at` | TIMESTAMP | When mapping was created |
| `updated_at` | TIMESTAMP | When mapping was last updated |
| `deleted_at` | TIMESTAMP | Soft delete timestamp (nullable) |

**Example data:**
```
id | feature_id | commit_hash | created_at              | updated_at
---+------------+-------------+-------------------------+-------------------------
1  | 35         | f3a5ffb     | 2025-11-10 12:40:15.125 | 2025-11-10 12:40:15.125
2  | 35         | a55bb85     | 2025-11-10 12:40:15.126 | 2025-11-10 12:40:15.126
```

**Indexes:**
- `feature_id` (for fast lookups)
- `feature_id + commit_hash` (unique constraint to prevent duplicates)

---

## Best Practices

### 1. Commit Message Format

**Always include the feature ID in your commit messages:**

✅ **Good:**
```
FTR-035: Add autocomplete dropdown for tags
FTR-035: Fix styling issues in autocomplete
FTR-035: Add keyboard navigation support
```

❌ **Bad:**
```
Add autocomplete dropdown
Fix styling
Add keyboard navigation
```

**Why:** The CLI uses regex `FTR-(\d+)` to match feature IDs. Without it, commits won't be mapped.

---

### 2. Run `map` Frequently

Run `featureplus map` after making several commits:

```bash
# After 3-5 commits
git commit -m "FTR-035: Feature work"
git commit -m "FTR-035: More changes"
git commit -m "FTR-035: Final touches"
featureplus map
```

**Why:** Keeps your local manifests up-to-date and helps you track progress.

---

### 3. Sync Strategically

Run `featureplus sync` when you want to share progress with the team:

```bash
# End of day
featureplus map
featureplus sync

# Before standup
featureplus map
featureplus sync

# Feature complete
featureplus map
featureplus sync
```

**Why:** Syncing uploads data to the backend where the team can see it in the UI.

---

### 4. Multiple Features

You can work on multiple features simultaneously:

```bash
# Pull multiple features
featureplus pull 35
featureplus pull 36
featureplus pull 37

# Commit with different feature IDs
git commit -m "FTR-035: Add autocomplete"
git commit -m "FTR-036: Fix login bug"
git commit -m "FTR-037: Update documentation"

# Map all at once
featureplus map

# Sync all at once
featureplus sync
```

**Output:**
```
📊 Found 3 feature(s) in git history:

📦 FTR-035 → 2 file(s), 1 commit(s)
📦 FTR-036 → 3 file(s), 1 commit(s)
📦 FTR-037 → 1 file(s), 1 commit(s)

☁️  Synced 3 feature(s) with FeaturePlus backend.
```

---

### 5. Scan More History

If you have old commits you want to map:

```bash
# Scan last 200 commits
featureplus map --commits 200

# Scan last 500 commits
featureplus map --commits 500
```

**Note:** Larger scans take longer but ensure all historical commits are mapped.

---

## Troubleshooting

### Issue: "FeaturePlus not initialized"

**Error:**
```
❌ FeaturePlus not initialized. Run 'featureplus init' first.
```

**Solution:**
```bash
cd /path/to/your/project
featureplus init
```

---

### Issue: "Feature not found" (404)

**Error:**
```
❌ Failed to fetch feature: API returned non-OK status: 404
```

**Causes:**
1. Feature doesn't exist in backend
2. Wrong feature ID
3. Backend not running

**Solutions:**
```bash
# Check backend is running
curl http://localhost:8080/api/features/35

# List all features
featureplus list

# Create feature in backend first
```

---

### Issue: "No features need syncing"

**Message:**
```
ℹ️  No features need syncing. All up to date!
```

**Causes:**
1. No commits with feature IDs
2. Already synced
3. No changes since last sync

**Solutions:**
```bash
# Make commits with feature IDs
git commit -m "FTR-035: Your changes"

# Run map to update manifests
featureplus map

# Then sync
featureplus sync
```

---

### Issue: Commits not being mapped

**Problem:** Ran `featureplus map` but no commits found.

**Causes:**
1. Feature ID not in commit message
2. Wrong format (not `FTR-XXX`)
3. Commits older than scan limit

**Solutions:**
```bash
# Check commit messages
git log --oneline -20

# Ensure format is correct: FTR-035
git commit --amend -m "FTR-035: Your message"

# Scan more commits
featureplus map --commits 100
```

---

### Issue: Duplicate files/commits

**Problem:** Same file appears multiple times in manifest.

**Solution:** This shouldn't happen! The CLI and backend use deduplication:
- CLI: `mergeUnique()` function
- Backend: `FirstOrCreate()` in GORM

If you see duplicates, it's a bug. Please report it.

---

## Summary

### Quick Reference

```bash
# 1. Initialize
featureplus init

# 2. Pull feature
featureplus pull 35

# 3. Code and commit (include FTR-035 in message!)
git commit -m "FTR-035: Your changes"

# 4. Map commits to features
featureplus map

# 5. Sync to backend
featureplus sync

# Repeat steps 3-5 as you work
```

### Workflow Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                     FeaturePlus Workflow                      │
└──────────────────────────────────────────────────────────────┘

1. featureplus init
   └─▶ Creates .featureplus/ directory

2. featureplus pull 35
   └─▶ Fetches feature from backend
       └─▶ Creates FTR-035.yaml locally

3. git commit -m "FTR-035: Your work"
   └─▶ Make commits with feature ID

4. featureplus map
   └─▶ Scans git history
       └─▶ Updates FTR-035.yaml with files & commits

5. featureplus sync
   └─▶ Uploads to backend
       └─▶ Stores in database
           └─▶ Visible in UI

Repeat steps 3-5 as you work on the feature
```

---

## Next Steps

- **View mapped files in UI** - Display files and commits on feature detail page
- **Generate reports** - Show code coverage per feature
- **Track progress** - Visualize feature completion based on commits
- **Team collaboration** - See who's working on what files

For more information, see:
- [CLI Commands Documentation](./CLI_COMMANDS.md)
- [API Documentation](../backend/API.md)
- [Database Schema](../backend/SCHEMA.md)
