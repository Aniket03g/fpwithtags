package utils

// Example usage of backup functions
//
// To create a backup:
//   filename, err := utils.BackupDatabase()
//   if err != nil {
//       log.Printf("Backup failed: %v", err)
//   } else {
//       log.Printf("Backup created: %s", filename)
//   }
//
// To restore from a backup:
//   err := utils.RestoreDatabase("featureplus_2025-01-15_14-30.db")
//   if err != nil {
//       log.Printf("Restore failed: %v", err)
//   } else {
//       log.Printf("Restore successful")
//   }
//
// To list available backups:
//   backups, err := utils.ListBackups()
//   if err != nil {
//       log.Printf("Failed to list backups: %v", err)
//   } else {
//       for _, backup := range backups {
//           info, _ := utils.GetBackupInfo(backup)
//           log.Printf("- %s (%.2f MB)", backup, float64(info.Size())/(1024*1024))
//       }
//   }
