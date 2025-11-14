# Getting Started with FeaturePlus Code Mapping

## 📖 Overview

FeaturePlus Code Mapping allows you to **link your code commits and files to features** in your project management system. This creates a powerful connection between your codebase and feature tracking, making it easy to:

- See which files belong to which feature
- Track commit history per feature
- Understand code ownership and feature implementation status
- Navigate from features to code and vice versa

This guide will walk you through the complete setup and usage process.

---

## 🎯 What You'll Learn

By the end of this guide, you'll be able to:

1. ✅ Install the FeaturePlus CLI
2. ✅ Initialize a project for code mapping
3. ✅ Pull features from FeaturePlus
4. ✅ Write commits with feature references
5. ✅ Map commits to features
6. ✅ Sync mappings to the web UI
7. ✅ View linked code in the FeaturePlus dashboard

---

## 📋 Prerequisites

Before you begin, make sure you have:

- **Git** installed on your system
- **Go 1.19+** installed (for building the CLI)
- **A FeaturePlus account** with access to a project
- **A code repository** (local or remote)
- **Terminal/Command Prompt** access

---

## 🚀 Step 1: Install the FeaturePlus CLI

### Option A: Build from Source (Recommended)

1. **Clone the CLI repository:**

```bash
git clone https://github.com/FeaturePlus/featureplus-pr.git
cd featureplus-pr
```

2. **Build the CLI:**

```bash
go build -o featureplus
```

3. **Move to system PATH (Optional but recommended):**

**On Linux/macOS:**
```bash
sudo mv featureplus /usr/local/bin/
```

**On Windows:**
```powershell
# Move to a directory in your PATH, e.g., C:\Program Files\FeaturePlus\
mkdir "C:\Program Files\FeaturePlus"
move featureplus.exe "C:\Program Files\FeaturePlus\"
# Add to PATH in System Environment Variables
```

4. **Verify installation:**

```bash
featureplus --version
```

**Expected output:**
```
featureplus-pr version 0.1.0
```

---

### Option B: Download Pre-built Binary

*(Coming soon - check releases page)*

---

## 🔧 Step 2: Configure the CLI

### 2.1 Set API URL

The CLI needs to know where your FeaturePlus backend is running.

**For local development:**
```bash
export FEATUREPLUS_API_URL=http://localhost:8080
```

**For production:**
```bash
export FEATUREPLUS_API_URL=https://featureplus.yourcompany.com
```

**Make it permanent (add to ~/.bashrc or ~/.zshrc):**
```bash
echo 'export FEATUREPLUS_API_URL=http://localhost:8080' >> ~/.bashrc
source ~/.bashrc
```

**On Windows (PowerShell):**
```powershell
$env:FEATUREPLUS_API_URL = "http://localhost:8080"
# Or set permanently via System Environment Variables
```

---

### 2.2 Authenticate (Optional for some commands)

Some commands require authentication. Login to get a token:

```bash
featureplus login
```

**You'll be prompted for:**
- Email: `your-email@example.com`
- Password: `your-password`

**Expected output:**
```
✓ Login successful!
Token saved to: ~/.featureplus/config.json
```

**Note:** The `sync` command does NOT require authentication (it's a public endpoint).

---

## 📂 Step 3: Initialize Your Project

Navigate to your code repository and initialize FeaturePlus tracking.

```bash
cd /path/to/your/project
featureplus init
```

**You'll be prompted for:**

1. **Project ID:** The ID of your FeaturePlus project (e.g., `8`)
   - Find this in the FeaturePlus web UI URL: `/projects/8/features`

2. **Repository Path:** Path to your git repository (default: current directory)
   - Press Enter to use current directory
   - Or specify a different path

**Expected output:**
```
✓ Project initialized successfully!
Project ID: 8
Repository: /path/to/your/project
Config saved to: .featureplus/config.json
```

**What this does:**
- Creates a `.featureplus/` directory in your project
- Stores project configuration
- Prepares the project for feature tracking

---

## 📥 Step 4: Pull Features

Pull features from FeaturePlus to start tracking them locally.

### 4.1 Pull All Features

```bash
featureplus pull
```

**Expected output:**
```
Fetching features for project 8...
✓ Pulled feature FTR-035: "Filtering by tags"
✓ Pulled feature FTR-036: "Autocomplete for tags"
✓ Pulled feature FTR-037: "Add tags support"

Total: 3 features pulled
Manifests saved to: .featureplus/features/
```

---

### 4.2 Pull a Specific Feature

```bash
featureplus pull 35
```

**Expected output:**
```
✓ Pulled feature FTR-035: "Filtering by tags"
Manifest saved to: .featureplus/features/FTR-035.yaml
```

---

### 4.3 View Pulled Features

Check what features are tracked locally:

```bash
ls .featureplus/features/
```

**Expected output:**
```
FTR-035.yaml
FTR-036.yaml
FTR-037.yaml
```

---

### 4.4 Inspect a Feature Manifest

```bash
cat .featureplus/features/FTR-035.yaml
```

**Example manifest:**
```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag. TEST
status: todo
priority: medium
project_id: 8
created_at: "2025-06-03T00:00:00Z"
updated_at: "2025-11-10T12:40:00Z"
files: []
commits: []
last_synced: ""
```

---

## 💻 Step 5: Write Code with Feature References

Now that you have features pulled, start working on them!

### 5.1 Commit Message Convention

**IMPORTANT:** Include the feature ID in your commit messages using the format `FTR-XXX`

**Examples:**

✅ **Good commit messages:**
```bash
git commit -m "FTR-035: Add tag filtering UI component"
git commit -m "FTR-035: Implement tag click handler"
git commit -m "FTR-035 FTR-036: Add autocomplete dropdown for tags"
```

✅ **Also valid:**
```bash
git commit -m "Add tag filtering (FTR-035)"
git commit -m "[FTR-035] Style tag filter component"
```

❌ **Bad commit messages (won't be tracked):**
```bash
git commit -m "Add tag filtering"  # Missing feature ID
git commit -m "FTR035: Add tags"   # Missing hyphen
git commit -m "ftr-035: Add tags"  # Lowercase (won't match)
```

---

### 5.2 Example Workflow

Let's work on feature FTR-035 (Filtering by tags):

```bash
# 1. Create/edit files for the feature
echo "def filter_by_tag(tag):" > autocomplete.py
echo "    return Feature.objects.filter(tags__contains=tag)" >> autocomplete.py

# 2. Stage changes
git add autocomplete.py

# 3. Commit with feature reference
git commit -m "FTR-035: Add tag filtering function"

# 4. Continue working
echo "def navigate_to_tag(tag_id):" > navigate.py
echo "    return redirect(f'/features/tags/{tag_id}')" >> navigate.py

git add navigate.py
git commit -m "FTR-035: Add tag navigation helper"

# 5. Make more commits as needed
git commit -m "FTR-035: Add unit tests for tag filtering"
git commit -m "FTR-035: Update documentation"
```

---

## 🗺️ Step 6: Map Commits to Features

After making commits, map them to features using the CLI.

### 6.1 Run the Map Command

```bash
featureplus map
```

**What this does:**
1. Scans your git commit history
2. Finds commits with feature IDs (e.g., `FTR-035`)
3. Extracts file changes from those commits
4. Updates local feature manifests with files and commits

**Expected output:**
```
Scanning git commits for feature references...
Found 4 commits for FTR-035
Found 2 commits for FTR-036

Updating manifests...
✓ Updated FTR-035: 2 files, 4 commits
✓ Updated FTR-036: 1 file, 2 commits

Total: 2 features updated
```

---

### 6.2 Verify Mapping

Check the updated manifest:

```bash
cat .featureplus/features/FTR-035.yaml
```

**Updated manifest:**
```yaml
id: FTR-035
name: Filtering by tags
description: Clicking on tag in a feature should display all features with that tag. TEST
status: todo
priority: medium
project_id: 8
created_at: "2025-06-03T00:00:00Z"
updated_at: "2025-11-10T12:40:00Z"
files:
  - path: autocomplete.py
    last_seen_commit: f3a5ffb
  - path: navigate.py
    last_seen_commit: f3a5ffb
commits:
  - hash: f3a5ffb1234567890abcdef
    message: "FTR-035: Add tag filtering function"
    author: aniket
    timestamp: "2025-11-10T12:30:00Z"
  - hash: a55bb851234567890abcdef
    message: "FTR-035: Add tag navigation helper"
    author: aniket
    timestamp: "2025-11-10T12:35:00Z"
last_synced: ""
```

---

### 6.3 Map Options

**Map specific feature:**
```bash
featureplus map --feature FTR-035
```

**Map with custom commit range:**
```bash
featureplus map --since "2 weeks ago"
featureplus map --since "2025-11-01"
```

**Dry run (preview without updating):**
```bash
featureplus map --dry-run
```

---

## ☁️ Step 7: Sync to FeaturePlus

Upload your local mappings to the FeaturePlus web UI.

### 7.1 Run Sync Command

```bash
featureplus sync
```

**What this does:**
1. Reads local feature manifests
2. Sends files and commits to FeaturePlus backend
3. Updates the web UI with linked code data
4. Updates `last_synced` timestamp in manifests

**Expected output:**
```
Syncing features to FeaturePlus...

✓ Synced FTR-035: 2 files, 4 commits
✓ Synced FTR-036: 1 file, 2 commits
✗ Skipped FTR-037: No changes since last sync

Successfully synced: 2 features
Failed: 0 features
Skipped: 1 feature (no changes)
```

---

### 7.2 Verify Sync

Check the updated manifest:

```bash
cat .featureplus/features/FTR-035.yaml
```

**After sync:**
```yaml
last_synced: "2025-11-10T12:40:00Z"  # ← Updated!
```

---

### 7.3 Sync Options

**Force sync (even if no changes):**
```bash
featureplus sync --force
```

**Sync specific feature:**
```bash
featureplus sync --feature FTR-035
```

**Dry run (preview without syncing):**
```bash
featureplus sync --dry-run
```

---

## 🌐 Step 8: View in Web UI

Now check the FeaturePlus web dashboard to see your linked code!

### 8.1 Navigate to Feature

1. Open FeaturePlus in your browser: `http://localhost:8080`
2. Login to your account
3. Go to **Projects** → Select your project
4. Click on **Features** tab
5. Click on a feature (e.g., "Filtering by tags")

---

### 8.2 View Linked Code Section

Scroll down to the **"Linked Code"** section (below "Associated Tasks").

**You should see:**

#### 📁 Files Section
```
📁 Files (2)
┌─────────────────────────────────────┐
│ 📄 autocomplete.py                  │
│    Last seen in f3a5ffb             │
│                                      │
│ 📄 navigate.py                       │
│    Last seen in f3a5ffb             │
└─────────────────────────────────────┘
```

#### ✓ Commits Section
```
✓ Commits (4)
┌─────────────────────────────────────┐
│ ✓ f3a5ffb                           │
│   FTR-035: Add tag filtering...     │
│   by aniket                          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ ✓ a55bb85                           │
│   FTR-035: Add tag navigation...    │
│   by aniket                          │
└─────────────────────────────────────┘
```

#### 🕐 Last Synced
```
🕐 Last synced: Nov 10, 2025 at 12:40 PM
```

---

### 8.3 Refresh Mapping

If you make new commits and sync again, click the **"Refresh Mapping"** button in the UI to reload the data without refreshing the page.

---

## 🔄 Complete Workflow Example

Here's a complete end-to-end example:

### Scenario: Implement "Tag Autocomplete" Feature

```bash
# 1. Pull the feature
featureplus pull 36

# 2. Create a new branch (optional but recommended)
git checkout -b feature/tag-autocomplete

# 3. Write code
cat > tag_autocomplete.js << 'EOF'
function autocompleteTag(input) {
  const tags = fetchTags();
  return tags.filter(t => t.startsWith(input));
}
EOF

# 4. Commit with feature reference
git add tag_autocomplete.js
git commit -m "FTR-036: Add tag autocomplete function"

# 5. Add more code
cat > tag_autocomplete.css << 'EOF'
.tag-autocomplete {
  position: absolute;
  background: white;
  border: 1px solid #ccc;
}
EOF

git add tag_autocomplete.css
git commit -m "FTR-036: Style autocomplete dropdown"

# 6. Map commits to feature
featureplus map

# Expected output:
# ✓ Updated FTR-036: 2 files, 2 commits

# 7. Sync to FeaturePlus
featureplus sync

# Expected output:
# ✓ Synced FTR-036: 2 files, 2 commits

# 8. View in web UI
# Navigate to http://localhost:8080/projects/8/features/36
# Scroll to "Linked Code" section
# See your files and commits!
```

---

## 🎓 Advanced Usage

### Multiple Features in One Commit

You can reference multiple features in a single commit:

```bash
git commit -m "FTR-035 FTR-036: Shared tag utility functions"
```

This commit will be mapped to **both** FTR-035 and FTR-036.

---

### Incremental Mapping

The `map` command is incremental - it only processes new commits:

```bash
# First time: scans all commits
featureplus map

# Make new commits
git commit -m "FTR-035: Add more features"

# Second time: only scans new commits
featureplus map
```

---

### Checking Sync Status

View which features need syncing:

```bash
featureplus status
```

**Expected output:**
```
Feature Status:
  FTR-035: ✓ Synced (2 files, 4 commits) - Last synced: 5 minutes ago
  FTR-036: ⚠ Modified (1 file, 2 commits) - Last synced: 2 hours ago
  FTR-037: ✗ Never synced (0 files, 0 commits)
```

---

### Viewing Manifest Stats

Get a summary of mapped code:

```bash
featureplus stats
```

**Expected output:**
```
Code Mapping Statistics:
  Total features: 3
  Total files: 5
  Total commits: 8
  Last sync: 5 minutes ago

Top features by commits:
  1. FTR-035: 4 commits
  2. FTR-036: 2 commits
  3. FTR-037: 2 commits
```

---

## 🐛 Troubleshooting

### Issue 1: "Feature not found"

**Error:**
```
Error: Feature FTR-035 not found
```

**Solution:**
```bash
# Make sure you pulled the feature first
featureplus pull 35

# Or pull all features
featureplus pull
```

---

### Issue 2: "No commits found"

**Error:**
```
Warning: No commits found for any features
```

**Solution:**
- Check your commit messages include feature IDs (e.g., `FTR-035`)
- Ensure you're using uppercase `FTR-XXX` format
- Verify you're in a git repository: `git log`

---

### Issue 3: "Sync failed: 401 Unauthorized"

**Error:**
```
Error: Sync failed: 401 Unauthorized
```

**Solution:**
The sync endpoint is public and doesn't require auth. This error shouldn't occur. If it does:

```bash
# Check API URL is correct
echo $FEATUREPLUS_API_URL

# Verify backend is running
curl http://localhost:8080/health
```

---

### Issue 4: "Template not found" in Web UI

**Error:**
```
Error #01: html/template: "linked-code.html" is undefined
```

**Solution:**
This is a backend issue. The template file needs to be loaded:

1. Check `backend/templates/linked-code.html` exists
2. Verify it's in the `LoadHTMLFiles` list in `backend/main.go`
3. Restart the backend server

---

### Issue 5: Empty Linked Code Section

**Symptom:** Linked Code section shows "No code mappings found yet"

**Solution:**
```bash
# 1. Verify feature has commits
cat .featureplus/features/FTR-035.yaml

# 2. Check if commits are in manifest
# Should see files: and commits: arrays

# 3. Re-sync
featureplus sync --force

# 4. Refresh browser (Ctrl+Shift+R)
```

---

## 📚 Command Reference

### Quick Command List

| Command | Description |
|---------|-------------|
| `featureplus init` | Initialize project |
| `featureplus pull` | Pull all features |
| `featureplus pull <id>` | Pull specific feature |
| `featureplus map` | Map commits to features |
| `featureplus sync` | Sync to FeaturePlus |
| `featureplus status` | Check sync status |
| `featureplus stats` | View mapping statistics |
| `featureplus login` | Authenticate CLI |
| `featureplus --help` | Show help |

---

### Detailed Command Help

Get help for any command:

```bash
featureplus --help
featureplus pull --help
featureplus map --help
featureplus sync --help
```

---

## 🎯 Best Practices

### 1. Commit Message Convention

✅ **Always include feature ID in commits:**
```bash
git commit -m "FTR-035: Descriptive message"
```

✅ **Use present tense:**
```bash
git commit -m "FTR-035: Add feature" # Good
git commit -m "FTR-035: Added feature" # Avoid
```

✅ **Be descriptive:**
```bash
git commit -m "FTR-035: Add tag filtering with autocomplete support"
```

---

### 2. Regular Syncing

Sync frequently to keep the web UI up-to-date:

```bash
# After each mapping session
featureplus map && featureplus sync

# Or create an alias
alias fp-sync="featureplus map && featureplus sync"
```

---

### 3. Feature Branch Workflow

Use feature branches for better organization:

```bash
# Create branch for feature
git checkout -b feature/FTR-035-tag-filtering

# Work on feature
git commit -m "FTR-035: Add filtering logic"
git commit -m "FTR-035: Add tests"

# Map and sync
featureplus map
featureplus sync

# Merge when done
git checkout main
git merge feature/FTR-035-tag-filtering
```

---

### 4. Team Collaboration

When working in a team:

1. **Each developer runs their own mapping:**
   ```bash
   featureplus pull
   featureplus map
   featureplus sync
   ```

2. **Mappings are merged on the backend** - multiple developers can sync the same feature

3. **Last sync wins** - the most recent sync updates the web UI

---

## 📊 Example Output Gallery

### Successful Init
```
$ featureplus init
Enter Project ID: 8
Enter Repository Path [.]: 
✓ Project initialized successfully!
Project ID: 8
Repository: /Users/aniket/projects/myapp
Config saved to: .featureplus/config.json
```

---

### Successful Pull
```
$ featureplus pull
Fetching features for project 8...
✓ Pulled feature FTR-035: "Filtering by tags"
✓ Pulled feature FTR-036: "Autocomplete for tags"
✓ Pulled feature FTR-037: "Add tags support"

Total: 3 features pulled
Manifests saved to: .featureplus/features/
```

---

### Successful Map
```
$ featureplus map
Scanning git commits for feature references...
Found commits:
  FTR-035: 4 commits (autocomplete.py, navigate.py)
  FTR-036: 2 commits (tag_autocomplete.js, tag_autocomplete.css)

Updating manifests...
✓ Updated FTR-035: 2 files, 4 commits
✓ Updated FTR-036: 2 files, 2 commits

Total: 2 features updated
```

---

### Successful Sync
```
$ featureplus sync
Syncing features to FeaturePlus...

Syncing FTR-035...
  Files: autocomplete.py, navigate.py
  Commits: 4
  ✓ Synced successfully

Syncing FTR-036...
  Files: tag_autocomplete.js, tag_autocomplete.css
  Commits: 2
  ✓ Synced successfully

Summary:
  ✓ Successfully synced: 2 features
  ✗ Failed: 0 features
  ⊘ Skipped: 0 features
```

---

## 🎉 Success Checklist

You've successfully set up code mapping when you can:

- [x] Install and run the CLI
- [x] Initialize a project
- [x] Pull features from FeaturePlus
- [x] Make commits with feature references
- [x] Map commits to features locally
- [x] Sync mappings to the backend
- [x] View linked code in the web UI
- [x] See files and commits for each feature
- [x] Refresh mappings without page reload

---

## 🆘 Getting Help

### Documentation
- **CLI Commands:** See `CLI_COMMANDS.md`
- **Workflow Details:** See `FEATURE_MAPPING_WORKFLOW.md`
- **UI Features:** See `LINKED_CODE_FEATURE.md`

### Support
- **GitHub Issues:** Report bugs or request features
- **Email:** support@featureplus.com
- **Slack:** #featureplus-support

---

## 🚀 Next Steps

Now that you've set up code mapping, explore these advanced features:

1. **PR Linking:** Link pull requests to features
2. **Task Tracking:** Break features into tasks
3. **Dependencies:** Track feature dependencies
4. **Releases:** Plan and track releases
5. **Analytics:** View code metrics and insights

---

## 📝 Feedback

We'd love to hear your feedback! Please share:

- What worked well?
- What was confusing?
- What features would you like to see?

Submit feedback via GitHub issues or email us at feedback@featureplus.com

---

**Happy Mapping! 🎉**
