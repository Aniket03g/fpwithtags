# FeaturePlus Code Mapping - Quick Start Guide

## 🚀 5-Minute Setup

Get started with FeaturePlus code mapping in just 5 minutes!

---

## Step 1: Install CLI (2 minutes)

```bash
# Clone and build
git clone https://github.com/FeaturePlus/featureplus-pr.git
cd featureplus-pr
go build -o featureplus

# Add to PATH (optional)
sudo mv featureplus /usr/local/bin/  # Linux/macOS
# OR move to C:\Program Files\FeaturePlus\ on Windows

# Verify
featureplus --version
```

---

## Step 2: Configure (1 minute)

```bash
# Set API URL
export FEATUREPLUS_API_URL=http://localhost:8080

# Initialize your project
cd /path/to/your/code
featureplus init
# Enter Project ID: 8 (get from FeaturePlus web UI)
```

---

## Step 3: Pull Features (30 seconds)

```bash
# Pull all features from your project
featureplus pull
```

---

## Step 4: Write Code (Your time!)

```bash
# Make commits with feature IDs
git commit -m "FTR-035: Add tag filtering feature"
git commit -m "FTR-036: Add autocomplete dropdown"
```

**Important:** Always include `FTR-XXX` in commit messages!

---

## Step 5: Map & Sync (1 minute)

```bash
# Map commits to features
featureplus map

# Sync to FeaturePlus
featureplus sync
```

---

## Step 6: View in Web UI (30 seconds)

1. Open FeaturePlus: `http://localhost:8080`
2. Go to your project → Features
3. Click on a feature
4. Scroll to **"Linked Code"** section
5. See your files and commits! 🎉

---

## 📋 Daily Workflow

```bash
# 1. Pull latest features
featureplus pull

# 2. Work on features (include FTR-XXX in commits)
git commit -m "FTR-035: Your changes"

# 3. Map and sync
featureplus map && featureplus sync
```

---

## 🎯 Commit Message Format

✅ **Correct:**
```
FTR-035: Add feature
FTR-035 FTR-036: Shared code
[FTR-035] Add feature
Add feature (FTR-035)
```

❌ **Wrong:**
```
Add feature          # No feature ID
FTR035: Add feature  # Missing hyphen
ftr-035: Add feature # Lowercase
```

---

## 🆘 Common Issues

### "Feature not found"
```bash
featureplus pull 35  # Pull the feature first
```

### "No commits found"
- Check commit messages have `FTR-XXX` format
- Use uppercase letters

### Empty Linked Code section
```bash
featureplus sync --force  # Force re-sync
# Then refresh browser (Ctrl+Shift+R)
```

---

## 📚 Full Documentation

For detailed information, see:
- **Complete Guide:** `GETTING_STARTED_CODE_MAPPING.md`
- **CLI Commands:** `CLI_COMMANDS.md`
- **Workflow Details:** `FEATURE_MAPPING_WORKFLOW.md`

---

**That's it! You're ready to start mapping code to features! 🚀**
