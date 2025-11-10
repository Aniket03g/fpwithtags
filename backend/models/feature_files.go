package models

import (
	"time"

	"gorm.io/gorm"
)

// FeatureFile represents a file mapped to a feature
type FeatureFile struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FeatureID  uint           `gorm:"not null;index" json:"feature_id"`
	FilePath   string         `gorm:"type:varchar(500);not null" json:"file_path"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Feature Feature `gorm:"foreignKey:FeatureID" json:"-"`
}

// FeatureCommit represents a git commit mapped to a feature
type FeatureCommit struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	FeatureID  uint           `gorm:"not null;index" json:"feature_id"`
	CommitHash string         `gorm:"type:varchar(40);not null" json:"commit_hash"` // Git SHA-1 hash
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	Feature Feature `gorm:"foreignKey:FeatureID" json:"-"`
}
