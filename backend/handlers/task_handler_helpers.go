package handlers

import (
	"github.com/FeaturePlus/backend/models"
)

// ensureTaskProjectConfig ensures that project.Config exists and has required fields
func ensureTaskProjectConfig(project *models.Project) {
	// Initialize Config if nil
	if project.Config == nil {
		project.Config = models.JSONB{}
	}

	// Ensure only truly required fields exist with defaults
	if _, exists := project.Config["task_types"]; !exists {
		project.Config["task_types"] = []string{"UI", "Dev", "DB", "Backend"}
	}
	// Don't force default feature_category - allow it to be empty
	if _, exists := project.Config["tech_stack"]; !exists {
		project.Config["tech_stack"] = "Other"
	}
	if _, exists := project.Config["tags_enabled"]; !exists {
		project.Config["tags_enabled"] = true
	}
}

// getTaskTypes safely extracts task types from project config
func getTaskTypes(project *models.Project) []string {
	ensureTaskProjectConfig(project)

	taskTypes := []string{}
	
	// Try to extract task types as interface array
	if types, ok := project.Config["task_types"].([]interface{}); ok {
		for _, t := range types {
			if tStr, ok := t.(string); ok {
				// Normalize 'Db' to 'DB'
				if tStr == "Db" {
					tStr = "DB"
				}
				taskTypes = append(taskTypes, tStr)
			}
		}
	} else if strArr, ok := project.Config["task_types"].([]string); ok {
		// Try to extract as string array
		for _, tStr := range strArr {
			// Normalize 'Db' to 'DB'
			if tStr == "Db" {
				tStr = "DB"
			}
			taskTypes = append(taskTypes, tStr)
		}
	}
	
	// If still empty, provide defaults
	if len(taskTypes) == 0 {
		taskTypes = []string{"UI", "Dev", "DB", "Backend"}
		// Update the config with defaults
		project.Config["task_types"] = taskTypes
	}
	
	return taskTypes
}
