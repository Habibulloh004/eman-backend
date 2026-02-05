package models

import "time"

// AdminUser represents the single admin account stored in the database.
type AdminUser struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Username          string     `gorm:"uniqueIndex;size:80" json:"username"`
	PasswordHash      string     `gorm:"type:text" json:"-"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}
