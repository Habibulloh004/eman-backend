package models

import "time"

type MapIconType struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	NameRu    string    `json:"name_ru"`
	NameUz    string    `json:"name_uz"`
	NameEn    string    `json:"name_en"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
