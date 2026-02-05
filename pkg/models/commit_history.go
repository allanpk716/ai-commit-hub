package models

import "time"

// CommitHistory stores generated commit messages
type CommitHistory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `gorm:"index" json:"project_id"`
	Message   string    `gorm:"type:text" json:"message"`
	Provider  string    `json:"provider"`
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`

	Project GitProject `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

// TableName specifies the table name for CommitHistory
func (CommitHistory) TableName() string {
	return "commit_histories"
}
