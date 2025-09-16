package repositories

import (
	"github.com/FeaturePlus/backend/models"
)

// GetByCreatorID returns PRs created by a specific user
func (r *prRepository) GetByCreatorID(creatorID uint) ([]models.PullRequest, error) {
	var prs []models.PullRequest
	err := r.db.Where("created_by_user = ?", creatorID).Order("created_at desc").Find(&prs).Error
	return prs, err
}
