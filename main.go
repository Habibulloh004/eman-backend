package main

import (
	"log"
	"os"

	"eman-backend/config"
	"eman-backend/database"
	"eman-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg := config.Load()

	// Connect to database
	if err := database.Connect(cfg.DBDSN); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	// Seed admin user if missing
	if err := database.EnsureAdminUser(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	// Create upload directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:           "Eman Backend API",
		BodyLimit:         cfg.MaxUploadSizeMB * 1024 * 1024,
		StreamRequestBody: true,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://127.0.0.1:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Filename",
		AllowCredentials: true,
	}))

	routes.Setup(app, cfg)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
