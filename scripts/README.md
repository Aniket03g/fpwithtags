# Database Backup & Restore Scripts

This folder contains shell scripts for managing FeaturePlus database backups.

## Scripts

### 1. backup.sh
Creates a timestamped backup of the `test.db` database.

**Usage:**
```bash
./scripts/backup.sh
```

**What it does:**
- Copies `backend/test.db` to `backups/featureplus_YYYY-MM-DD_HH-MM.db`
- Creates the `backups/` directory if it doesn't exist
- Displays backup size and lists recent backups
- Exits gracefully if database doesn't exist

### 2. restore.sh
Restores the database from a backup file.

**Usage:**
```bash
./scripts/restore.sh <backup_file>
```

**Example:**
```bash
./scripts/restore.sh backups/featureplus_2025-01-15_14-30.db
```

**What it does:**
- Creates a safety backup of the current database before restoring
- Replaces `backend/test.db` with the specified backup file
- Lists available backups if no file is specified or file doesn't exist
- Exits gracefully if backup file doesn't exist

## Setup

### On Linux/Mac:
Make the scripts executable:
```bash
chmod +x scripts/backup.sh scripts/restore.sh
```

### On Windows (Git Bash/WSL):
The scripts will work in Git Bash or WSL. If using Git Bash:
```bash
bash scripts/backup.sh
bash scripts/restore.sh backups/featureplus_2025-01-15_14-30.db
```

## Features

- ✅ Automatic timestamp format: `YYYY-MM-DD_HH-MM`
- ✅ Creates `backups/` folder automatically
- ✅ Safety backup before restore
- ✅ Clear success/failure messages with colors
- ✅ Validates file existence before operations
- ✅ Lists available backups when needed
- ✅ Displays file sizes

## Directory Structure

```
FeaturePlus/
├── backend/
│   └── test.db          # Main database
├── backups/             # Created automatically
│   ├── featureplus_2025-01-15_14-30.db
│   └── featureplus_pre-restore_2025-01-15_15-00.db
└── scripts/
    ├── backup.sh
    ├── restore.sh
    └── README.md
```

## Notes

- Backups are stored in the `backups/` directory at the project root
- The restore script creates a safety backup before overwriting the current database
- All scripts must be run from the project root directory
- Timestamps use 24-hour format
