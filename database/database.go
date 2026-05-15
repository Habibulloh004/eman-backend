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

// ensureDatabase connects to the postgres maintenance DB and creates the target
// user + database if they do not already exist.
func ensureDatabase(dsn string) {
	// Parse key=value pairs from DSN to extract connection fields.
	fields := map[string]string{}
	for _, part := range strings.Fields(dsn) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			fields[kv[0]] = kv[1]
		}
	}

	host := fields["host"]
	port := fields["port"]
	user := fields["user"]
	password := fields["password"]
	dbName := fields["dbname"]
	sslmode := fields["sslmode"]
	timezone := fields["TimeZone"]

	if host == "" || dbName == "" || user == "" {
		return
	}
	if port == "" {
		port = "5432"
	}
	if sslmode == "" {
		sslmode = "disable"
	}
	if timezone == "" {
		timezone = "UTC"
	}

	// Connect to the "postgres" maintenance database to run DDL.
	adminDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s TimeZone=%s",
		host, user, password, port, sslmode, timezone,
	)
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		// Try connecting as the OS superuser without a password (local trust auth).
		adminDSN = fmt.Sprintf(
			"host=%s dbname=postgres port=%s sslmode=%s TimeZone=%s",
			host, port, sslmode, timezone,
		)
		adminDB, err = gorm.Open(postgres.Open(adminDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Printf("[db-init] Cannot connect to maintenance DB to check existence: %v", err)
			return
		}
	}

	sqlDB, err := adminDB.DB()
	if err != nil {
		log.Printf("[db-init] Failed to get raw DB: %v", err)
		return
	}
	defer sqlDB.Close()

	// Create the role if it does not exist.
	var roleExists int
	adminDB.Raw("SELECT 1 FROM pg_roles WHERE rolname = ?", user).Scan(&roleExists)
	if roleExists == 0 {
		sql := fmt.Sprintf(
			"CREATE USER %s WITH PASSWORD '%s'",
			sanitizeIdentifier(user), escapeSingleQuote(password),
		)
		if err := adminDB.Exec(sql).Error; err != nil {
			log.Printf("[db-init] Failed to create user %q: %v", user, err)
		} else {
			log.Printf("[db-init] Created PostgreSQL user %q", user)
		}
	}

	// Create the database if it does not exist.
	var dbExists int
	adminDB.Raw("SELECT 1 FROM pg_database WHERE datname = ?", dbName).Scan(&dbExists)
	if dbExists == 0 {
		sql := fmt.Sprintf(
			"CREATE DATABASE %s OWNER %s",
			sanitizeIdentifier(dbName), sanitizeIdentifier(user),
		)
		if err := adminDB.Exec(sql).Error; err != nil {
			log.Printf("[db-init] Failed to create database %q: %v", dbName, err)
		} else {
			log.Printf("[db-init] Created PostgreSQL database %q", dbName)
		}
	} else {
		log.Printf("[db-init] Database %q already exists", dbName)
	}
}

func sanitizeIdentifier(s string) string {
	// Allow only alphanumeric and underscores to prevent SQL injection.
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func Connect(dsn string) error {
	if strings.TrimSpace(dsn) == "" {
		return fmt.Errorf("database DSN is empty")
	}

	// Auto-create the database and user if they don't exist.
	ensureDatabase(dsn)

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
		&models.AdminUser{},
		&models.GalleryItem{},
		&models.MapIconType{},
		&models.MapIcon{},
		&models.Project{},
		&models.ContactSubmission{},
		&models.SiteSetting{},
		&models.Challenge{},
		&models.ChallengeParticipant{},
	)
	if err != nil {
		return err
	}

	log.Println("Database migration completed")

	// Seed default settings if table is empty or missing keys
	if err := SeedSettings(); err != nil {
		log.Printf("Warning: Failed to seed settings: %v", err)
	}

	// Seed default projects if table is empty or has empty fields
	if err := SeedProjects(); err != nil {
		log.Printf("Warning: Failed to seed projects: %v", err)
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
			if setting.Type == models.TypeJSON && (setting.Key == "faq_items" || setting.Key == "faq_items_uz" || setting.Key == "faq_items_en") {
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
