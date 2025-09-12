package models

import (
	"time"

	"gorm.io/gorm"
)

// EntityType represents the type of entity in a dependency relationship
type EntityType string

const (
	EntityTypeFeature EntityType = "feature"
	EntityTypeTask    EntityType = "task"
	EntityTypePR      EntityType = "pr"
)

// Dependency represents a dependency relationship between two entities
// where Parent blocks Child (Child depends on Parent)
type Dependency struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ParentID  uint           `gorm:"not null;index:idx_parent" json:"parent_id"`
	ParentType EntityType    `gorm:"type:varchar(20);not null;index:idx_parent" json:"parent_type"`
	ChildID   uint           `gorm:"not null;index:idx_child" json:"child_id"`
	ChildType EntityType     `gorm:"type:varchar(20);not null;index:idx_child" json:"child_type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Optional fields for additional context
	Description string `gorm:"type:text" json:"description"`
	CreatedByID uint   `json:"created_by_id"`
}

// TableName specifies the table name for the Dependency model
func (Dependency) TableName() string {
	return "dependencies"
}

// BeforeCreate hook to ensure unique dependencies
func (d *Dependency) BeforeCreate(tx *gorm.DB) error {
	// Check if a similar dependency already exists
	var count int64
	tx.Model(&Dependency{}).
		Where("parent_id = ? AND parent_type = ? AND child_id = ? AND child_type = ?",
			d.ParentID, d.ParentType, d.ChildID, d.ChildType).
		Count(&count)
	
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	
	// Prevent self-dependencies
	if d.ParentID == d.ChildID && d.ParentType == d.ChildType {
		return gorm.ErrInvalidData
	}
	
	return nil
}
