package models

import "time"

type MapIcon struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	Name      string      `json:"name"`
	NameRu    string      `json:"name_ru"`
	NameUz    string      `json:"name_uz"`
	Latitude  float64     `json:"lat"`
	Longitude float64     `json:"lng"`
	TypeID    uint        `gorm:"index" json:"type_id"`
	Type      MapIconType `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"type"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
