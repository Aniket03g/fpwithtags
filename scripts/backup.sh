#!/bin/bash

# FeaturePlus Database Backup Script
# Copies test.db from backend folder to backups/ with timestamp

set -e

# Configuration
DB_PATH="backend/test.db"
BACKUP_DIR="backups"
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M")
BACKUP_FILE="${BACKUP_DIR}/featureplus_${TIMESTAMP}.db"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "FeaturePlus Database Backup"
echo "============================"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}Error: Database file not found at ${DB_PATH}${NC}"
    echo "Please ensure the database exists before running backup."
    exit 1
fi

# Create backups directory if it doesn't exist
if [ ! -d "$BACKUP_DIR" ]; then
    echo -e "${YELLOW}Creating backups directory...${NC}"
    mkdir -p "$BACKUP_DIR"
fi

# Perform backup
echo "Backing up database..."
echo "Source: $DB_PATH"
echo "Destination: $BACKUP_FILE"

if cp "$DB_PATH" "$BACKUP_FILE"; then
    DB_SIZE=$(du -h "$DB_PATH" | cut -f1)
    echo -e "${GREEN}✓ Backup successful!${NC}"
    echo "Backup file: $BACKUP_FILE"
    echo "Size: $DB_SIZE"
    
    # List recent backups
    echo ""
    echo "Recent backups:"
    ls -lht "$BACKUP_DIR" | head -6
else
    echo -e "${RED}✗ Backup failed!${NC}"
    exit 1
fi

exit 0
