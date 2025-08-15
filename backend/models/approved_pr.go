package models

type ApprovedPR struct {
	ID            uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	PullRequestID uint   `gorm:"not null;index" json:"pull_request_id"`
	PullRequest   PullRequest `gorm:"foreignKey:PullRequestID" json:"pull_request,omitempty"`
	UserID        int    `gorm:"not null;index" json:"user_id"`
	User          User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ApprovedAt    int64  `gorm:"not null;index" json:"approved_at"`
	Version       string `gorm:"type:text" json:"version"`
}
