package models

import (
	"time"

	"gorm.io/gorm"
)

type GalleryItem struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `json:"title"`
	TitleUz     string         `json:"title_uz"`
	Description string         `json:"description"`
	DescriptionUz string       `json:"description_uz"`
	Type        string         `json:"type"` // image, video
	URL         string         `json:"url"`
	Thumbnail   string         `json:"thumbnail"` // for videos
	Category    string         `json:"category"`  // construction, interior, exterior
	SortOrder   int            `json:"sort_order"`
	IsPublished bool           `json:"is_published" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
