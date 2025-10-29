package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultDBPath is the default path to the SQLite database
	DefaultDBPath = "test.db"
	// BackupDir is the directory where backups are stored
	BackupDir = "backups"
)

// BackupDatabase creates a timestamped copy of the SQLite database
// Returns the backup filename and any error encountered
func BackupDatabase() (string, error) {
	return BackupDatabaseWithPath(DefaultDBPath)
}

// BackupDatabaseWithPath creates a timestamped copy of the specified database file
func BackupDatabaseWithPath(dbPath string) (string, error) {
	// Generate timestamp in YYYY-MM-DD_HH-MM format
	timestamp := time.Now().Format("2006-01-02_15-04")
	
	// Create backup filename
	backupFilename := fmt.Sprintf("featureplus_%s.db", timestamp)
	backupPath := filepath.Join(BackupDir, backupFilename)
	
	// Ensure backup directory exists
	if err := os.MkdirAll(BackupDir, 0755); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Check if source database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Printf("Database file not found: %s", dbPath)
		return "", fmt.Errorf("database file not found: %s", dbPath)
	}
	
	// Open source file
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		log.Printf("Failed to open source database: %v", err)
		return "", fmt.Errorf("failed to open source database: %w", err)
	}
	defer sourceFile.Close()
	
	// Create destination file
	destFile, err := os.Create(backupPath)
	if err != nil {
		log.Printf("Failed to create backup file: %v", err)
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()
	
	// Copy the file
	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		log.Printf("Failed to copy database: %v", err)
		return "", fmt.Errorf("failed to copy database: %w", err)
	}
	
	// Sync to ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		log.Printf("Failed to sync backup file: %v", err)
		return "", fmt.Errorf("failed to sync backup file: %w", err)
	}
	
	log.Printf("✓ Database backup successful: %s (%.2f MB)", backupPath, float64(bytesWritten)/(1024*1024))
	return backupFilename, nil
}

// RestoreDatabase replaces the current database with the selected backup file
func RestoreDatabase(filename string) error {
	return RestoreDatabaseWithPath(filename, DefaultDBPath)
}

// RestoreDatabaseWithPath replaces the specified database with a backup file
func RestoreDatabaseWithPath(filename, dbPath string) error {
	// Construct full backup path
	backupPath := filepath.Join(BackupDir, filename)
	
	// Check if backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		log.Printf("Backup file not found: %s", backupPath)
		return fmt.Errorf("backup file not found: %s", backupPath)
	}
	
	// Create safety backup of current database if it exists
	if _, err := os.Stat(dbPath); err == nil {
		safetyTimestamp := time.Now().Format("2006-01-02_15-04")
		safetyBackup := fmt.Sprintf("featureplus_pre-restore_%s.db", safetyTimestamp)
		safetyPath := filepath.Join(BackupDir, safetyBackup)
		
		// Ensure backup directory exists
		if err := os.MkdirAll(BackupDir, 0755); err != nil {
			log.Printf("Failed to create backup directory for safety backup: %v", err)
			return fmt.Errorf("failed to create backup directory: %w", err)
		}
		
		// Create safety backup
		if err := copyFile(dbPath, safetyPath); err != nil {
			log.Printf("Warning: Failed to create safety backup: %v", err)
			// Continue anyway - user might want to proceed
		} else {
			log.Printf("✓ Safety backup created: %s", safetyPath)
		}
	}
	
	// Open backup file
	sourceFile, err := os.Open(backupPath)
	if err != nil {
		log.Printf("Failed to open backup file: %v", err)
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer sourceFile.Close()
	
	// Create/overwrite destination file
	destFile, err := os.Create(dbPath)
	if err != nil {
		log.Printf("Failed to create database file: %v", err)
		return fmt.Errorf("failed to create database file: %w", err)
	}
	defer destFile.Close()
	
	// Copy the file
	bytesWritten, err := io.Copy(destFile, sourceFile)
	if err != nil {
		log.Printf("Failed to restore database: %v", err)
		return fmt.Errorf("failed to restore database: %w", err)
	}
	
	// Sync to ensure data is written to disk
	if err := destFile.Sync(); err != nil {
		log.Printf("Failed to sync restored database: %v", err)
		return fmt.Errorf("failed to sync restored database: %w", err)
	}
	
	log.Printf("✓ Database restore successful: %s → %s (%.2f MB)", backupPath, dbPath, float64(bytesWritten)/(1024*1024))
	return nil
}

// copyFile is a helper function to copy a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}
	
	return destFile.Sync()
}

// ListBackups returns a list of available backup files
func ListBackups() ([]string, error) {
	// Check if backup directory exists
	if _, err := os.Stat(BackupDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	// Read directory
	entries, err := os.ReadDir(BackupDir)
	if err != nil {
		log.Printf("Failed to read backup directory: %v", err)
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	// Filter for .db files
	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".db" {
			backups = append(backups, entry.Name())
		}
	}
	
	return backups, nil
}

// GetBackupInfo returns information about a backup file
func GetBackupInfo(filename string) (os.FileInfo, error) {
	backupPath := filepath.Join(BackupDir, filename)
	return os.Stat(backupPath)
}
