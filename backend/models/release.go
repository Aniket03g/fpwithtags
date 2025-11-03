package models

import (
	"time"
)

type ReleaseStatus string

const (
	ReleaseStatusDraft     ReleaseStatus = "draft"
	ReleaseStatusPublished ReleaseStatus = "published"
	ReleaseStatusFailed    ReleaseStatus = "failed"
)

type Release struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// Make tag unique per project using a composite unique index
	ProjectID int           `gorm:"index;uniqueIndex:idx_release_project_tag" json:"project_id"`
	Tag       string        `gorm:"size:50;not null;uniqueIndex:idx_release_project_tag" json:"tag"`
	Project   Project       `gorm:"foreignKey:ProjectID" json:"project"`
	Status    ReleaseStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	PRs       []PullRequest `gorm:"many2many:release_prs;" json:"prs"`
	Notes     string        `gorm:"type:text" json:"notes"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
