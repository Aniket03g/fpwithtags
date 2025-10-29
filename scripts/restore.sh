#!/bin/bash

# FeaturePlus Database Restore Script
# Replaces current test.db with a selected backup file

set -e

# Configuration
DB_PATH="backend/test.db"
BACKUP_DIR="backups"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "FeaturePlus Database Restore"
echo "============================"

# Check if backup file argument is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: No backup file specified${NC}"
    echo ""
    echo "Usage: $0 <backup_file>"
    echo "Example: $0 backups/featureplus_2025-01-15_14-30.db"
    echo ""
    
    # List available backups if directory exists
    if [ -d "$BACKUP_DIR" ]; then
        echo "Available backups:"
        ls -lht "$BACKUP_DIR"/*.db 2>/dev/null || echo "No backup files found in $BACKUP_DIR"
    else
        echo -e "${YELLOW}No backups directory found. Run backup.sh first.${NC}"
    fi
    
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo -e "${RED}Error: Backup file not found: ${BACKUP_FILE}${NC}"
    echo ""
    
    # List available backups
    if [ -d "$BACKUP_DIR" ]; then
        echo "Available backups:"
        ls -lht "$BACKUP_DIR"/*.db 2>/dev/null || echo "No backup files found in $BACKUP_DIR"
    fi
    
    exit 1
fi

# Check if database path exists (create directory if needed)
DB_DIR=$(dirname "$DB_PATH")
if [ ! -d "$DB_DIR" ]; then
    echo -e "${YELLOW}Creating database directory: ${DB_DIR}${NC}"
    mkdir -p "$DB_DIR"
fi

# Create backup of current database if it exists
if [ -f "$DB_PATH" ]; then
    TIMESTAMP=$(date +"%Y-%m-%d_%H-%M")
    SAFETY_BACKUP="${BACKUP_DIR}/featureplus_pre-restore_${TIMESTAMP}.db"
    
    # Create backups directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"
    
    echo -e "${YELLOW}Creating safety backup of current database...${NC}"
    if cp "$DB_PATH" "$SAFETY_BACKUP"; then
        echo -e "${GREEN}✓ Safety backup created: ${SAFETY_BACKUP}${NC}"
    else
        echo -e "${RED}Warning: Failed to create safety backup${NC}"
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "Restore cancelled."
            exit 1
        fi
    fi
fi

# Perform restore
echo ""
echo "Restoring database..."
echo "Source: $BACKUP_FILE"
echo "Destination: $DB_PATH"

if cp "$BACKUP_FILE" "$DB_PATH"; then
    DB_SIZE=$(du -h "$DB_PATH" | cut -f1)
    echo -e "${GREEN}✓ Restore successful!${NC}"
    echo "Database restored: $DB_PATH"
    echo "Size: $DB_SIZE"
else
    echo -e "${RED}✗ Restore failed!${NC}"
    exit 1
fi

exit 0
