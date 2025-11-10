# FeaturePlus CLI Commands

## Project Connection Commands

### `init` - Initialize FeaturePlus

Initializes FeaturePlus in the current directory by creating a `.featureplus` folder with `config.yaml`.

```bash
featureplus-pr init
```

**Output:**
```
✅ Initialized FeaturePlus in this directory.
```

**Created Files:**
- `.featureplus/config.yaml` with default configuration:
  ```yaml
  api_url: http://localhost:8080
  project_id: ""
  linked_at: ""
  ```

---

### `connect` - Connect to a Project

Links the current directory to a FeaturePlus project.

```bash
featureplus-pr connect <project-id>
```

**Example:**
```bash
featureplus-pr connect 1
```

**What it does:**
1. Reads `.featureplus/config.yaml`
2. Calls `POST http://localhost:8080/api/projects/<id>/connect` with current directory path
3. Updates `config.yaml` with `project_id` and `linked_at`
4. Prints success message

**Success Output:**
```
✅ Linked this folder to FeaturePlus project Todo App (1) at 2025-11-09 23:30:00
```

**Error Handling:**
- ❌ Config file not found → Suggests running `init`
- ❌ Invalid config format → Shows error details
- ❌ Server not reachable → Shows connection error
- ❌ Project not found → Shows 404 error

---

### `status` - Show Connection Status

Displays the current connection status and project information.

```bash
featureplus-pr status
```

**What it does:**
1. Reads `.featureplus/config.yaml`
2. If `project_id` is empty → Shows warning
3. Else → Calls `GET /api/projects/:id/status`
4. Pretty prints the status with colors

**Output (Not Connected):**
```
⚠️  Not connected to any FeaturePlus project.
   Run 'featureplus-pr connect <project-id>' to link this directory.
```

**Output (Connected):**
```
╔════════════════════════════════════════════════╗
║         FeaturePlus Connection Status         ║
╚════════════════════════════════════════════════╝

📦 Project:       Todo App (1)
🌐 Server:        http://localhost:8080
📂 Path:          d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
🔗 Connected:     yes
🕓 Linked At:     2025-11-09 23:30:00
```

**Features:**
- ✅ Color-coded output (green for connected, red for not connected, yellow for warnings)
- ✅ Emoji icons for visual clarity
- ✅ Shows project name and ID
- ✅ Displays server URL
- ✅ Shows local path
- ✅ Connection status (yes/no)
- ✅ Timestamp of when linked

---

## API Endpoints

### POST `/api/projects/:id/connect`

**Request:**
```json
{
  "path": "/path/to/directory"
}
```

**Response:**
```json
{
  "status": "linked",
  "project_id": "1",
  "project_name": "Todo App",
  "path": "/path/to/directory",
  "connected_at": "2025-11-09T23:30:00Z"
}
```

### GET `/api/projects/:id/status`

**Response (Linked):**
```json
{
  "status": "linked",
  "project_id": "1",
  "project_name": "Todo App",
  "path": "/path/to/directory",
  "connected_at": "2025-11-09T23:30:00Z"
}
```

**Response (Unlinked):**
```json
{
  "status": "unlinked"
}
```

---

## Feature Management Commands

### `pull` - Pull Feature from Backend

Fetches a feature's metadata from the FeaturePlus backend and saves it locally as a YAML manifest.

```bash
featureplus pull <feature-id>
```

**Examples:**
```bash
featureplus pull FTR-001
featureplus pull 1
```

**What it does:**
1. Checks if project is initialized (`.featureplus/config.yaml` exists)
2. Parses the feature ID (accepts both `FTR-001` and `1` formats)
3. Checks if feature already exists locally
   - If yes → Prompts user to overwrite (y/N)
   - If no → Proceeds to fetch
4. Calls `GET /api/features/{id}` to fetch feature metadata
5. Creates `.featureplus/features/` directory if it doesn't exist
6. Saves feature as `FTR-{id}.yaml` with the following structure:

**YAML Manifest Structure:**
```yaml
id: FTR-001
name: Add User Login
description: Implement JWT login for users
status: in-progress
owner: aniket
files: []
prs: []
commits: []
created_at: 2025-11-10T18:00:00Z
updated_at: 2025-11-10T19:30:00Z  # Set by 'map' command
synced_from_backend: true
```

**After running `featureplus map`:**
```yaml
id: FTR-001
name: Add User Login
description: Implement JWT login for users
status: in-progress
owner: aniket
files:
  - backend/auth/login.go
  - backend/routes.go
  - ui/pages/login.html
prs: []
commits:
  - a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  - e4f5g6h7i8j9k0l1m2n3o4p5q6r7s8t9u0v1w2x3
created_at: 2025-11-10T18:00:00Z
updated_at: 2025-11-10T19:30:00Z
synced_from_backend: true
```

**After running `featureplus sync`:**
```yaml
id: FTR-001
name: Add User Login
description: Implement JWT login for users
status: in-progress
owner: aniket
files:
  - backend/auth/login.go
  - backend/routes.go
  - ui/pages/login.html
prs: []
commits:
  - a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
  - e4f5g6h7i8j9k0l1m2n3o4p5q6r7s8t9u0v1w2x3
created_at: 2025-11-10T18:00:00Z
updated_at: 2025-11-10T19:30:00Z
last_synced: 2025-11-10T19:35:00Z  # ← Added by sync
synced_from_backend: true
```

**Success Output:**
```
⬇️  Fetching feature FTR-001 from backend...
✅ Pulled feature FTR-001 (Add User Login)
   Saved to .featureplus/features/FTR-001.yaml
```

**Overwrite Prompt:**
```
⚠️  Feature FTR-001 already exists locally. Overwrite? (y/N): 
```

**Error Handling:**
- ❌ Not initialized → `FeaturePlus not initialized. Run 'featureplus init' first.`
- ❌ Invalid feature ID → `Invalid feature ID: <input>`
- ❌ Feature not found → `Failed to fetch feature: API returned non-OK status: 404`
- ❌ Connection error → `Failed to fetch feature: error making request: ...`
- ❌ Authentication required → `authentication required: please login first using 'featureplus-pr login'`

**API Endpoint:**

`GET /api/features/{id}`

**Expected Response:**
```json
{
  "id": 1,
  "project_id": 1,
  "name": "Add User Login",
  "description": "Implement JWT login for users",
  "status": "in-progress"
}
```

---

### `map` - Map Git Commits to Features

Scans git history and maps commits to local feature manifests by extracting feature IDs from commit messages.

```bash
featureplus map [--commits N]
```

**Examples:**
```bash
featureplus map
featureplus map --commits 100
```

**What it does:**
1. Checks if project is initialized (`.featureplus/` exists)
2. Verifies it's a git repository (`.git/` exists)
3. Reads all local feature manifests from `.featureplus/features/`
4. Runs `git log --name-only --pretty=format:"%H|%an|%ae|%ad|%s" --date=iso -n 50`
5. For each commit:
   - Extracts feature ID from commit message (e.g., `FTR-001`)
   - Collects commit hash, author, date, and changed files
6. Aggregates data per feature:
   - Unique files touched
   - Commit count
   - Commit hashes
7. Displays summary with file and commit counts
8. **Updates local manifests automatically:**
   - Merges new files and commits (avoids duplicates)
   - Updates `updated_at` timestamp
   - Writes changes back to `.featureplus/features/{ID}.yaml`

**Flags:**
- `--commits N` - Number of commits to scan (default: 50)

**Example Output:**
```
🔍 Scanning git history (last 50 commits)...

📊 Found 2 feature(s) in git history:

📦 FTR-001 → 3 file(s), 2 commit(s)
   Name: Add User Login
   - backend/auth/login.go
   - backend/routes.go
   - ui/pages/login.html
   Commits: a1b2c3d e4f5g6h

📦 FTR-002 → 5 file(s), 4 commit(s)
   Name: Dashboard UI
   - ui/components/dashboard.tsx
   - ui/styles/dashboard.css
   - backend/api/dashboard.go
   - backend/models/widget.go
   - ui/pages/dashboard.html
   Commits: h7i8j9k l0m1n2o p3q4r5s (+1 more)

📝 Updating feature manifests...
   ✅ Updated .featureplus/features/FTR-001.yaml (3 files, 2 commits)
   ✅ Updated .featureplus/features/FTR-002.yaml (5 files, 4 commits)

✅ Mapping complete! Found 2 feature(s) with 6 total commits.
   Updated 2 manifest(s).
```

**Feature ID Pattern:**
- Matches `FTR-001`, `FTR-123`, etc. in commit messages
- Case-sensitive (must be uppercase `FTR`)
- Extracts the full ID including prefix

**Error Handling:**
- ❌ Not initialized → `FeaturePlus not initialized. Run 'featureplus init' first.`
- ❌ Not a git repo → `Not a git repository. Initialize git first.`
- ❌ No features found → `No feature manifests found in .featureplus/features/`
- ❌ Git command fails → `Failed to run git log: ...`

**Notes:**
- Only commits with feature IDs in the message are mapped
- Files are deduplicated per feature
- Results are sorted by feature ID
- File display is limited to 10 files per feature (shows "... and N more")
- Commit display is limited to 3 hashes (shows "+N more")
- **Manifests are automatically updated** with new files and commits
- Duplicate files and commits are merged intelligently
- `updated_at` timestamp is set on each update

**Data Structure (In Memory):**
```go
type MapResult struct {
    FeatureID    string   // e.g., "FTR-001"
    FeatureName  string   // From local manifest
    Files        []string // Unique file paths
    CommitCount  int      // Number of commits
    CommitHashes []string // Git commit hashes
}
```

---

### `sync` - Sync Local Mappings to Backend

Syncs local feature manifests with the FeaturePlus backend, uploading file and commit mappings.

```bash
featureplus sync
```

**What it does:**
1. Checks if project is initialized (`.featureplus/` exists)
2. Verifies authentication (requires login)
3. Reads all feature manifests from `.featureplus/features/`
4. Identifies features that need syncing:
   - Has files or commits data
   - Never synced before, OR
   - Updated after last sync
5. For each feature, sends `POST /api/features/sync`:
   ```json
   {
     "feature_id": "FTR-001",
     "files": ["backend/auth/login.go", "backend/routes.go"],
     "commits": ["a1b2c3d...", "e4f5g6h..."],
     "status": "in-progress"
   }
   ```
6. On success, updates `last_synced` timestamp in manifest
7. Displays sync results

**Example Output:**
```
☁️  Syncing features to FeaturePlus backend...

   ✅ Synced FTR-001 (3 files, 2 commits)
   ✅ Synced FTR-002 (5 files, 4 commits)

☁️  Synced 2 feature(s) with FeaturePlus backend.
```

**When No Sync Needed:**
```
☁️  Syncing features to FeaturePlus backend...
ℹ️  No features need syncing. All up to date!
```

**API Endpoint:**

`POST /api/features/sync`

**Request Payload:**
```json
{
  "feature_id": "FTR-001",
  "files": [
    "backend/auth/login.go",
    "backend/routes.go",
    "ui/pages/login.html"
  ],
  "commits": [
    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
    "e4f5g6h7i8j9k0l1m2n3o4p5q6r7s8t9u0v1w2x3"
  ],
  "status": "in-progress"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Feature synced successfully"
}
```

**Error Handling:**
- ❌ Not initialized → `FeaturePlus not initialized. Run 'featureplus init' first.`
- ❌ Not authenticated → `Not authenticated. Run 'featureplus-pr login' first.`
- ❌ No manifests found → `Failed to sync: no feature manifests found`
- ❌ API error → `Failed to sync FTR-001: API returned status 500`
- ❌ Network error → `Failed to sync FTR-001: failed to send request: ...`

**Sync Logic:**

A feature needs syncing if:
1. It has files OR commits (not empty)
2. AND one of:
   - Never synced before (`last_synced` is zero)
   - Updated after last sync (`updated_at` > `last_synced`)

**Updated Manifest After Sync:**
```yaml
id: FTR-001
name: Add User Login
files:
  - backend/auth/login.go
  - backend/routes.go
commits:
  - a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
updated_at: 2025-11-10T19:30:00Z
last_synced: 2025-11-10T19:35:00Z  # ← Set by sync command
```

**Notes:**
- Only syncs features with actual data (files/commits not empty)
- Skips features already synced and not updated since
- Updates `last_synced` timestamp only on successful sync
- Continues syncing other features if one fails
- Shows summary of successes and failures

---

## Workflow Example

```bash
# 1. Navigate to your project directory
cd /path/to/your/project

# 2. Initialize FeaturePlus
featureplus-pr init

# 3. Connect to a project
featureplus-pr connect 1

# 4. Check status
featureplus-pr status

# 5. Pull a feature from backend
featureplus pull FTR-001

# 6. View the pulled feature
cat .featureplus/features/FTR-001.yaml

# 7. Map git commits to features
featureplus map

# 8. Sync mappings to backend
featureplus sync

# 9. Map with more commits (if needed)
featureplus map --commits 100
```

---

## Adding CLI to PATH

### Windows (PowerShell)

**Build the executable:**
```powershell
cd d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr
go build -o featureplus-pr.exe
```

**Add to PATH (Permanent):**
```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr", "User")
```

**Verify:**
```powershell
featureplus-pr --version
```

### Alternative: Batch Script

Create `featureplus-pr.bat` in a directory already in PATH:

```batch
@echo off
"d:\FeaturePlus_Code_MAPPING\FeaturePlus\featureplus-pr\featureplus-pr.exe" %*
```

---

## Configuration File

**Location:** `.featureplus/config.yaml`

**Structure:**
```yaml
api_url: http://localhost:8080  # FeaturePlus server URL
project_id: "1"                 # Connected project ID
linked_at: "2025-11-09T23:30:00Z"  # Timestamp of connection
```

**Notes:**
- The config file is automatically created by `init` command
- Updated by `connect` command when linking to a project
- Read by `status` command to display information
