package migrations

import (
	"log"

	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

// MigrateProjectConfig updates all existing projects to ensure they have
// the required fields in their Config JSONB object, including tech_stack
func MigrateProjectConfig(db *gorm.DB) error {
	log.Println("Starting migration: Updating Project.Config fields")

	// Get all projects
	var projects []models.Project
	if err := db.Find(&projects).Error; err != nil {
		return err
	}

	log.Printf("Found %d projects to update", len(projects))
	
	// Update each project
	for i, project := range projects {
		updated := false
		
		// Initialize Config if nil
		if project.Config == nil {
			project.Config = models.JSONB{}
			updated = true
		}
		
		// Ensure tech_stack exists
		if _, exists := project.Config["tech_stack"]; !exists {
			project.Config["tech_stack"] = "Other"
			updated = true
		}
		
		// Ensure tags_enabled exists
		if _, exists := project.Config["tags_enabled"]; !exists {
			project.Config["tags_enabled"] = true
			updated = true
		}
		
		// Ensure task_types exists
		if _, exists := project.Config["task_types"]; !exists {
			project.Config["task_types"] = []string{"UI", "Dev", "DB", "Backend"}
			updated = true
		}
		
		// Don't force default feature_category in migration
		// Allow it to be empty if user hasn't set it
		
		// Save if updated
		if updated {
			if err := db.Save(&project).Error; err != nil {
				log.Printf("Error updating project %d: %v", project.ID, err)
				continue
			}
			log.Printf("Updated project %d/%d (ID: %d, Name: %s)", i+1, len(projects), project.ID, project.Name)
		}
	}
	
	log.Println("Migration completed: Project.Config fields updated")
	return nil
}

// RegisterMigrations adds all migrations to be run at startup
func RegisterMigrations(db *gorm.DB) {
	// Run the project config migration
	if err := MigrateProjectConfig(db); err != nil {
		log.Printf("Error in project config migration: %v", err)
	}
}
