package models

import (
	"time"
)

// ProjectConnection represents a connection between a project and a local directory
type ProjectConnection struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ProjectID   string    `gorm:"not null;index" json:"project_id"`
	Path        string    `gorm:"not null" json:"path"`
	ConnectedAt time.Time `gorm:"autoCreateTime" json:"connected_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
