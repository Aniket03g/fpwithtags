package models

import "gorm.io/gorm"

type PullRequest struct {
	ID          uint           `gorm:"primaryKey"`
	TaskID      uint           `gorm:"index"`
	URL         string         `gorm:"not null" form:"url"`
	Title       string         `gorm:"not null" form:"title"`
	Branch      string         `gorm:"not null" form:"branch"`
	Description string         `gorm:"type:text" form:"description"`
	Status      string         `gorm:"default:'Open'"`
	Tested      bool           `gorm:"default:false"`
	CreatedAt   int64          `gorm:"autoCreateTime:milli"`
	UpdatedAt   int64          `gorm:"autoUpdateTime:milli"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
