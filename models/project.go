package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TypeRu        string         `json:"type_ru"`
	TypeUz        string         `json:"type_uz"`
	AreaRu        string         `json:"area_ru"`
	AreaUz        string         `json:"area_uz"`
	DescriptionRu string         `json:"description_ru"`
	DescriptionUz string         `json:"description_uz"`
	Image         string         `json:"image"`
	SortOrder     int            `json:"sort_order"`
	IsPublished   bool           `json:"is_published" gorm:"default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
