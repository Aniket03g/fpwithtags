package models

import "gorm.io/gorm"

type PullRequest struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TaskID      uint           `gorm:"index" json:"task_id"`
	URL         string         `gorm:"not null" form:"url" json:"pr_url"`
	Title       string         `gorm:"not null" form:"title" json:"title"`
	Branch      string         `gorm:"not null" form:"branch" json:"branch"`
	Description string         `gorm:"type:text" form:"description" json:"description"`
	Status      string         `gorm:"default:'Open'" json:"status"`
	Tested      bool           `gorm:"default:false" json:"is_tested"`
	Version     string         `gorm:"column:version" json:"version"`
	CreatedAt   int64          `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64          `gorm:"autoUpdateTime:milli" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
