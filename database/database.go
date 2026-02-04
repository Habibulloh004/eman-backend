package database

import (
	"eman-backend/models"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect(dsn string) error {
	if strings.TrimSpace(dsn) == "" {
		return fmt.Errorf("database DSN is empty")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	log.Println("Database connected")
	return nil
}

func Migrate() error {
	err := DB.AutoMigrate(
		&models.GalleryItem{},
		&models.MapIconType{},
		&models.MapIcon{},
		&models.Project{},
		&models.ContactSubmission{},
		&models.SiteSetting{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migration completed")

	// Seed default settings if table is empty or missing keys
	if err := SeedSettings(); err != nil {
		log.Printf("Warning: Failed to seed settings: %v", err)
	}

	return nil
}

// SeedSettings populates the site_settings table with default values if empty
func SeedSettings() error {
	var count int64
	DB.Model(&models.SiteSetting{}).Count(&count)

	defaults := models.DefaultSettings()
	if count > 0 {
		var existing []models.SiteSetting
		if err := DB.Select("id", "key", "value").Find(&existing).Error; err != nil {
			return err
		}

		existingByKey := make(map[string]models.SiteSetting, len(existing))
		for _, setting := range existing {
			existingByKey[setting.Key] = setting
		}

		added := 0
		updated := 0
		for _, setting := range defaults {
			current, found := existingByKey[setting.Key]
			if !found {
				if err := DB.Create(&setting).Error; err != nil {
					log.Printf("Warning: Failed to create setting %s: %v", setting.Key, err)
					continue
				}
				added++
				continue
			}

			// Backfill empty FAQ JSON only on startup.
			if setting.Type == models.TypeJSON && (setting.Key == "faq_items" || setting.Key == "faq_items_uz") {
				trimmed := strings.TrimSpace(current.Value)
				if trimmed == "" || trimmed == "[]" || trimmed == "{}" || trimmed == "null" {
					if err := DB.Model(&models.SiteSetting{}).
						Where("id = ?", current.ID).
						Update("value", setting.Value).Error; err != nil {
						log.Printf("Warning: Failed to update setting %s: %v", setting.Key, err)
					} else {
						updated++
					}
				}
			}
		}

		if added > 0 || updated > 0 {
			log.Printf("Added %d missing default settings, updated %d empty FAQ settings", added, updated)
		} else {
			log.Printf("Settings already exist (%d items), no defaults added", count)
		}
		return nil
	}

	for _, setting := range defaults {
		if err := DB.Create(&setting).Error; err != nil {
			log.Printf("Warning: Failed to create setting %s: %v", setting.Key, err)
		}
	}

	log.Printf("Seeded %d default settings", len(defaults))
	return nil
}
