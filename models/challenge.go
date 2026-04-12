package models

import (
	"time"

	"gorm.io/gorm"
)

type Challenge struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Title         string         `json:"title"`
	TitleUz       string         `json:"title_uz"`
	Description   string         `json:"description"`
	DescriptionUz string         `json:"description_uz"`
	Image         string         `json:"image"`
	Prize         string         `json:"prize"`
	PrizeUz       string         `json:"prize_uz"`
	StartDate     time.Time      `json:"start_date"`
	EndDate       time.Time      `json:"end_date"`
	MaxUsers      int            `json:"max_users"`
	SortOrder     int            `json:"sort_order"`
	IsPublished   bool           `json:"is_published" gorm:"default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ChallengeParticipant struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	ChallengeID uint          `gorm:"index" json:"challenge_id"`
	Phone       string         `gorm:"index" json:"phone"`
	Name        string         `json:"name"`
	Status      string         `json:"status" gorm:"default:active"` // active, completed, cancelled
	JoinedAt    time.Time      `json:"joined_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Challenge Challenge `gorm:"foreignKey:ChallengeID" json:"challenge,omitempty"`
}
