package repositories

import (
	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

type PullRequestRepository interface {
	Create(pr *models.PullRequest) error
	GetByTaskID(taskID uint) ([]models.PullRequest, error)
	MarkTested(prID uint) error
}

type prRepository struct {
	db *gorm.DB
}

func NewPullRequestRepository(db *gorm.DB) PullRequestRepository {
	return &prRepository{db}
}

func (r *prRepository) Create(pr *models.PullRequest) error {
	return r.db.Create(pr).Error
}

func (r *prRepository) GetByTaskID(taskID uint) ([]models.PullRequest, error) {
	var prs []models.PullRequest
	err := r.db.Where("task_id = ?", taskID).Order("created_at desc").Find(&prs).Error
	return prs, err
}

func (r *prRepository) MarkTested(prID uint) error {
	return r.db.Model(&models.PullRequest{}).Where("id = ?", prID).Update("tested", true).Error
}
