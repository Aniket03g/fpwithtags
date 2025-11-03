package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type JSONB map[string]interface{}

func (j *JSONB) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("Failed to unmarshal JSONB value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, j)
}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type Project struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	OwnerID     int       `gorm:"not null;index" json:"owner_id"` // Foreign key
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Association to User model (already in your models package)
	Owner User `gorm:"foreignKey:OwnerID" json:"owner"`

	Features []Feature `gorm:"foreignKey:ProjectID" json:"features,omitempty"`

	Config JSONB `gorm:"type:TEXT" json:"config"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	// Initialize Config if nil
	if p.Config == nil {
		p.Config = JSONB{}
	}
	
	// Only set defaults for fields that don't exist and are truly required
	// Don't force default values if Config is partially populated
	if len(p.Config) == 0 {
		// Only set minimal defaults when Config is completely empty
		p.Config["task_types"] = []string{"UI", "Dev", "DB", "Backend"}
		p.Config["tech_stack"] = "Other"
		p.Config["tags_enabled"] = true
		p.Config["context"] = "Development"
		// Don't set default feature_category - let it be empty if user doesn't provide
	} else {
		// Ensure only truly required fields exist
		if _, exists := p.Config["task_types"]; !exists {
			p.Config["task_types"] = []string{"UI", "Dev", "DB", "Backend"}
		}
		if _, exists := p.Config["tech_stack"]; !exists {
			p.Config["tech_stack"] = "Other"
		}
		if _, exists := p.Config["tags_enabled"]; !exists {
			p.Config["tags_enabled"] = true
		}
		if _, exists := p.Config["context"]; !exists {
			p.Config["context"] = "Development"
		}
	}
	return nil
}
