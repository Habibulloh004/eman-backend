package database

import (
	"eman-backend/models"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// EnsureAdminUser seeds the admin account if it does not exist.
func EnsureAdminUser(defaultUsername, defaultPassword string) error {
	var count int64
	if err := DB.Model(&models.AdminUser{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	username := strings.TrimSpace(defaultUsername)
	if username == "" {
		username = "admin"
	}

	password := defaultPassword
	if strings.TrimSpace(password) == "" {
		password = "admin123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()
	admin := models.AdminUser{
		Username:          username,
		PasswordHash:      string(hash),
		PasswordChangedAt: &now,
	}

	if err := DB.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("Seeded admin user: %s", username)
	return nil
}
