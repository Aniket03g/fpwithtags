package handlers

import (
	"github.com/FeaturePlus/backend/repositories"
)

// createFallbackTemplateRepo creates a fallback template repository with default data
func createFallbackTemplateRepo() *repositories.TemplateRepository {
	// Create an empty repository with minimal default data
	repo := &repositories.TemplateRepository{}
	
	// Initialize with a default template to avoid nil pointer dereference
	defaultTemplate := repositories.Template{
		ID:          "custom",
		Name:        "Custom Stack",
		Stack:       "Custom",
		Description: "Build your own custom stack",
		TechStack:   "Other",
		Features:    []repositories.TemplateFeature{},
		Tasks:       []repositories.TemplateTask{},
	}
	
	// Set up the data structure
	repo.SetDefaultTemplate(defaultTemplate)
	
	return repo
}

// createTemplateGuidanceRepo creates a fallback guidance repository for templates
func createTemplateGuidanceRepo() *repositories.GuidanceRepository {
	// Create an empty repository with minimal default data
	repo := &repositories.GuidanceRepository{}
	
	return repo
}
