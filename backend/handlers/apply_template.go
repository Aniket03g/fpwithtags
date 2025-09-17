package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/FeaturePlus/backend/models"
	"github.com/FeaturePlus/backend/repositories"
	"github.com/FeaturePlus/backend/services"
)

// ApplyTemplate applies a template to a project
func (h *ProjectHandler) ApplyTemplate(project *models.Project, templateID string) error {
	if templateID == "" || h.templateRepo == nil {
		return fmt.Errorf("invalid template ID or repository not initialized")
	}

	log.Printf("INFO: Applying template %s to project %d", templateID, project.ID)

	// Get the template
	template, err := h.templateRepo.GetTemplateByID(templateID)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}

	if template == nil {
		return fmt.Errorf("template not found")
	}

	// Update project config with template's feature categories and task types if available
	if project.Config == nil {
		project.Config = models.JSONB{}
	}

	// Update feature categories if provided in the template
	if len(template.FeatureCategories) > 0 {
		log.Printf("INFO: Setting feature categories from template: %v", template.FeatureCategories)
		project.Config["feature_category"] = template.FeatureCategories
		
		// Update the project in the database
		if err := h.repo.UpdateProject(project); err != nil {
			log.Printf("WARNING: Failed to update project with template feature categories: %v", err)
		}
	}

	// Update task types if provided in the template
	if len(template.TaskTypes) > 0 {
		log.Printf("INFO: Setting task types from template: %v", template.TaskTypes)
		project.Config["task_types"] = template.TaskTypes
		
		// Update the project in the database
		if err := h.repo.UpdateProject(project); err != nil {
			log.Printf("WARNING: Failed to update project with template task types: %v", err)
		}
	}

	// Log template details for debugging
	log.Printf("INFO: Applying template '%s' -> inserting %d features, %d tasks, %d dependencies", 
		template.Name, len(template.Features), len(template.Tasks), len(template.Dependencies))

	// Create features from template
	createdFeatures := []models.Feature{}
	featureMap := make(map[string]uint) // Map feature names to IDs for dependency creation

	for i, templateFeature := range template.Features {
		log.Printf("DEBUG: Creating feature %d/%d: %s", i+1, len(template.Features), templateFeature.Name)
		feature := models.Feature{
			Title:       templateFeature.Name,
			Description: templateFeature.Description,
			Category:    templateFeature.Category,
			ProjectID:   int(project.ID),
			Status:      models.StatusTodo,
			Priority:    models.PriorityMedium,
		}

		if err := h.featureRepo.CreateFeature(&feature); err != nil {
			log.Printf("ERROR: Failed to create feature %s: %v", templateFeature.Name, err)
			continue
		}
		
		log.Printf("DEBUG: Successfully created feature with ID: %d", feature.ID)
		createdFeatures = append(createdFeatures, feature)
		
		// Store feature ID by name for dependency creation
		featureMap[templateFeature.Name] = feature.ID
	}
	
	log.Printf("INFO: Created %d/%d features successfully", len(createdFeatures), len(template.Features))

	// Create tasks from template with feature assignment
	createdTasks := []models.Task{}
	taskMap := make(map[string]uint) // Map task names to IDs for dependency creation

	// Create a map of feature categories for quick lookup
	featureCategories := make(map[string][]models.Feature)
	for _, feature := range createdFeatures {
		featureCategories[feature.Category] = append(featureCategories[feature.Category], feature)
	}

	for i, templateTask := range template.Tasks {
		log.Printf("DEBUG: Creating task %d/%d: %s (Type: %s)", i+1, len(template.Tasks), templateTask.Name, templateTask.Type)
		
		// Find the appropriate feature for this task based on type matching
		var featureID uint
		if len(createdFeatures) > 0 {
			// Try exact match first
			if features, exists := featureCategories[templateTask.Type]; exists && len(features) > 0 {
				featureID = features[0].ID
				log.Printf("DEBUG: Exact match - Task type '%s' with feature category '%s' (ID: %d)", 
					templateTask.Type, features[0].Category, features[0].ID)
			} else {
				// Try case-insensitive match
				for _, feature := range createdFeatures {
					if strings.EqualFold(feature.Category, templateTask.Type) {
						featureID = feature.ID
						log.Printf("DEBUG: Case-insensitive match - Task type '%s' with feature category '%s' (ID: %d)", 
							templateTask.Type, feature.Category, feature.ID)
						break
					}
				}
				
				// If still no match, use the first feature
				if featureID == 0 {
					featureID = createdFeatures[0].ID
					log.Printf("DEBUG: No match found, assigning task to first feature (ID: %d)", featureID)
				}
			}
		} else {
			log.Printf("WARNING: No features available to assign task to")
		}

		task := models.Task{
			TaskName:    templateTask.Name,
			Description: templateTask.Description,
			TaskType:    templateTask.Type,
			FeatureID:   featureID,
		}

		if err := h.taskRepo.Create(&task); err != nil {
			log.Printf("ERROR: Failed to create task %s: %v", templateTask.Name, err)
			continue
		}
		
		log.Printf("DEBUG: Successfully created task with ID: %d", task.ID)
		createdTasks = append(createdTasks, task)
		
		// Store task ID by name for dependency creation
		taskMap[templateTask.Name] = task.ID
	}
	
	log.Printf("INFO: Created %d/%d tasks successfully", len(createdTasks), len(template.Tasks))

	// Create dependencies if we have a dependency repository
	if len(template.Dependencies) > 0 && h.db != nil {
		dependencyService := services.NewDependencyService(repositories.NewDependencyRepository(h.db))
		createdDependencies := 0
		
		for _, dep := range template.Dependencies {
			// Create a simple dependency record
			dependency := models.Dependency{
				Description: fmt.Sprintf("Template dependency: %s", dep),
				ParentType:  models.EntityTypeFeature,
				ChildType:   models.EntityTypeFeature,
			}
			
			// Try to find matching features
			if len(createdFeatures) > 0 {
				dependency.ParentID = createdFeatures[0].ID
				
				if len(createdFeatures) > 1 {
					dependency.ChildID = createdFeatures[1].ID
				} else {
					dependency.ChildID = createdFeatures[0].ID
				}
				
				if err := dependencyService.CreateDependency(&dependency); err != nil {
					log.Printf("ERROR: Failed to create dependency: %v", err)
					continue
				}
				
				createdDependencies++
			}
		}
		
		log.Printf("INFO: Created %d dependencies", createdDependencies)
	}

	log.Printf("INFO: Successfully applied template %s to project %d: %d features, %d tasks", 
		template.Name, project.ID, len(createdFeatures), len(createdTasks))
	
	return nil
}
