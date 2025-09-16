package handlers

import (
	"strings"

	"github.com/FeaturePlus/backend/models"
)

// DependencyPair represents a parent-child dependency relationship
type DependencyPair struct {
	Parent string
	Child  string
}

// parseDependencies converts string dependencies into structured dependency pairs
func parseDependencies(dependencies []string) []DependencyPair {
	pairs := make([]DependencyPair, 0, len(dependencies))

	for _, dep := range dependencies {
		// Check if the dependency string contains a relationship indicator
		if strings.Contains(dep, "->") {
			// Format: "child -> parent" (child depends on parent)
			parts := strings.Split(dep, "->")
			if len(parts) == 2 {
				pairs = append(pairs, DependencyPair{
					Parent: strings.TrimSpace(parts[1]),
					Child:  strings.TrimSpace(parts[0]),
				})
			}
		} else if strings.Contains(dep, "<-") {
			// Format: "parent <- child" (child depends on parent)
			parts := strings.Split(dep, "<-")
			if len(parts) == 2 {
				pairs = append(pairs, DependencyPair{
					Parent: strings.TrimSpace(parts[0]),
					Child:  strings.TrimSpace(parts[1]),
				})
			}
		} else if strings.Contains(dep, ":") {
			// Format: "parent: child1, child2" (children depend on parent)
			parts := strings.Split(dep, ":")
			if len(parts) == 2 {
				parent := strings.TrimSpace(parts[0])
				children := strings.Split(parts[1], ",")
				for _, child := range children {
					pairs = append(pairs, DependencyPair{
						Parent: parent,
						Child:  strings.TrimSpace(child),
					})
				}
			}
		} else {
			// Simple dependency - treat as a standalone item that other items might depend on
			pairs = append(pairs, DependencyPair{
				Parent: strings.TrimSpace(dep),
				Child:  "Implementation", // Generic child
			})
		}
	}

	return pairs
}

// findEntityMatch tries to find an exact match for an entity name in the feature and task maps
func findEntityMatch(name string, featureMap map[string]uint, taskMap map[string]uint) (uint, models.EntityType) {
	// Normalize the name for comparison
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	
	// Check for exact match in features
	if id, exists := featureMap[normalizedName]; exists {
		return id, models.EntityTypeFeature
	}
	
	// Check for exact match in tasks
	if id, exists := taskMap[normalizedName]; exists {
		return id, models.EntityTypeTask
	}
	
	// No match found
	return 0, ""
}

// findPartialMatch tries to find a partial match for an entity name in features and tasks
func findPartialMatch(name string, features []models.Feature, tasks []models.Task) (uint, models.EntityType) {
	// Normalize the name for comparison
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	
	// Try to find a partial match in features
	for _, feature := range features {
		if strings.Contains(strings.ToLower(feature.Title), normalizedName) ||
		   strings.Contains(normalizedName, strings.ToLower(feature.Title)) ||
		   strings.Contains(strings.ToLower(feature.Category), normalizedName) ||
		   strings.Contains(normalizedName, strings.ToLower(feature.Category)) {
			return feature.ID, models.EntityTypeFeature
		}
	}
	
	// Try to find a partial match in tasks
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.TaskName), normalizedName) ||
		   strings.Contains(normalizedName, strings.ToLower(task.TaskName)) ||
		   strings.Contains(strings.ToLower(task.TaskType), normalizedName) ||
		   strings.Contains(normalizedName, strings.ToLower(task.TaskType)) {
			return task.ID, models.EntityTypeTask
		}
	}
	
	// No match found
	return 0, ""
}
