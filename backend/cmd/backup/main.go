package main

import (
	"fmt"
	"os"

	"github.com/FeaturePlus/backend/internal/log"
	"github.com/FeaturePlus/backend/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "backup":
		performBackup()
	case "restore":
		if len(os.Args) < 3 {
			fmt.Println("Error: restore requires a backup filename")
			fmt.Println("Usage: backup restore <filename>")
			os.Exit(1)
		}
		performRestore(os.Args[2])
	case "list":
		listBackups()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FeaturePlus Database Backup Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  backup backup           - Create a new backup")
	fmt.Println("  backup restore <file>   - Restore from a backup file")
	fmt.Println("  backup list             - List available backups")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  backup backup")
	fmt.Println("  backup list")
	fmt.Println("  backup restore featureplus_2025-10-29_13-45.db")
}

func performBackup() {
	log.Info("Starting database backup...")
	
	filename, err := utils.BackupDatabase()
	if err != nil {
		log.WithError(err).Error("Backup failed")
		os.Exit(1)
	}
	
	log.WithField("filename", filename).Info("Backup completed successfully")
	fmt.Printf("\n✓ Backup created: %s\n", filename)
}

func performRestore(filename string) {
	log.WithField("filename", filename).Info("Starting database restore...")
	
	err := utils.RestoreDatabase(filename)
	if err != nil {
		log.WithError(err).Error("Restore failed")
		os.Exit(1)
	}
	
	log.Info("Restore completed successfully")
	fmt.Printf("\n✓ Database restored from: %s\n", filename)
}

func listBackups() {
	backups, err := utils.ListBackups()
	if err != nil {
		log.WithError(err).Error("Failed to list backups")
		os.Exit(1)
	}
	
	if len(backups) == 0 {
		fmt.Println("No backups found")
		return
	}
	
	fmt.Println("Available backups:")
	fmt.Println()
	
	for _, backup := range backups {
		info, err := utils.GetBackupInfo(backup)
		if err != nil {
			fmt.Printf("  - %s (error getting info)\n", backup)
			continue
		}
		
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("  - %s (%.2f MB, modified: %s)\n", 
			backup, 
			sizeMB, 
			info.ModTime().Format("2006-01-02 15:04:05"))
	}
}
