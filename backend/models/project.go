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
	// Initialize Config if nil or empty
	if p.Config == nil || len(p.Config) == 0 {
		p.Config = JSONB{
			"task_types":       []string{"UI", "Dev", "DB", "Backend"},
			"feature_category": []string{"Auth", "Payment", "Tags", "Tasks", "Features"},
			"tech_stack":       "Other",
			"tags_enabled":     true,
			"context":          "Development",
		}
	} else {
		// Ensure required fields exist with defaults if Config is partially populated
		if _, exists := p.Config["task_types"]; !exists {
			p.Config["task_types"] = []string{"UI", "Dev", "DB", "Backend"}
		}
		if _, exists := p.Config["feature_category"]; !exists {
			p.Config["feature_category"] = []string{"Auth", "Payment", "Tags", "Tasks", "Features"}
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
