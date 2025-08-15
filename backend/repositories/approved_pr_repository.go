package repositories

import (
	"github.com/FeaturePlus/backend/models"
	"gorm.io/gorm"
)

type ApprovedPRRepository interface {
	Create(approvedPR *models.ApprovedPR) error
	FindByID(id uint) (*models.ApprovedPR, error)
	FindByPullRequestID(prID uint) ([]models.ApprovedPR, error)
	FindByUserID(userID int) ([]models.ApprovedPR, error)
	Update(approvedPR *models.ApprovedPR) error
	Delete(id uint) error
}

type approvedPRRepository struct {
	db *gorm.DB
}

func NewApprovedPRRepository(db *gorm.DB) ApprovedPRRepository {
	return &approvedPRRepository{db: db}
}

func (r *approvedPRRepository) Create(approvedPR *models.ApprovedPR) error {
	return r.db.Create(approvedPR).Error
}

func (r *approvedPRRepository) FindByID(id uint) (*models.ApprovedPR, error) {
	var approvedPR models.ApprovedPR
	err := r.db.Preload("PullRequest").Preload("User").First(&approvedPR, id).Error
	if err != nil {
		return nil, err
	}
	return &approvedPR, nil
}

func (r *approvedPRRepository) FindByPullRequestID(prID uint) ([]models.ApprovedPR, error) {
	var approvedPRs []models.ApprovedPR
	err := r.db.Preload("User").Where("pull_request_id = ?", prID).Find(&approvedPRs).Error
	if err != nil {
		return nil, err
	}
	return approvedPRs, nil
}

func (r *approvedPRRepository) FindByUserID(userID int) ([]models.ApprovedPR, error) {
	var approvedPRs []models.ApprovedPR
	err := r.db.Preload("PullRequest").Where("user_id = ?", userID).Find(&approvedPRs).Error
	if err != nil {
		return nil, err
	}
	return approvedPRs, nil
}

func (r *approvedPRRepository) Update(approvedPR *models.ApprovedPR) error {
	return r.db.Save(approvedPR).Error
}

func (r *approvedPRRepository) Delete(id uint) error {
	return r.db.Delete(&models.ApprovedPR{}, id).Error
}
