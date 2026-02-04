package models

import (
	"time"

	"gorm.io/gorm"
)

type ContactSubmission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Phone       string         `json:"phone"`
	Email       string         `json:"email"`
	Message     string         `json:"message"`
	Source      string         `json:"source"` // contact_page, catalog_request, callback
	EstateID    *int           `json:"estate_id"`
	PaymentPlan string         `json:"payment_plan"` // Selected payment plan (e.g., "Ипотека", "Рассрочка")
	Status      string         `json:"status" gorm:"default:'new'"` // new, contacted, closed
	Notes       string         `json:"notes"`
	IPAddress   string         `json:"ip_address"`
	UserAgent   string         `json:"user_agent"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
