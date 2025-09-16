package repositories

import (
	"github.com/FeaturePlus/backend/models"
)

// GetTasksByAssignee returns tasks assigned to a specific user
func (r *taskRepository) GetTasksByAssignee(assigneeID uint) ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Where("assignee_id = ?", assigneeID).Order("created_at desc").Find(&tasks).Error
	return tasks, err
}
