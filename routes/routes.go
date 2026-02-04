package routes

import (
	"eman-backend/config"
	"eman-backend/handlers"
	"eman-backend/middleware"
	"eman-backend/services"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, cfg *config.Config) {
	// Services
	macroService := services.NewMacroService(cfg)
	storageService := services.NewStorageService(cfg)

	// Handlers
	estateHandler := handlers.NewEstateHandler(macroService)
	authHandler := handlers.NewAuthHandler(cfg)
	galleryHandler := handlers.NewGalleryHandler(storageService)
	projectsHandler := handlers.NewProjectsHandler(storageService)
	submissionsHandler := handlers.NewSubmissionsHandler()
	settingsHandler := handlers.NewSettingsHandler()
	uploadHandler := handlers.NewUploadHandler(storageService)
	mapIconHandler := handlers.NewMapIconHandler()
	mapIconTypeHandler := handlers.NewMapIconTypeHandler(storageService)

	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ============ PUBLIC ROUTES ============

	// Estate routes (public)
	estate := api.Group("/estate")
	estate.Get("/complexes", estateHandler.GetComplexes)
	estate.Get("/list", estateHandler.GetEstates)

	// Gallery (public - only published)
	api.Get("/gallery", galleryHandler.ListPublic)

	// Projects (public - only published)
	api.Get("/projects", projectsHandler.ListPublic)

	// Map icons (public)
	api.Get("/map-icons", mapIconHandler.ListPublic)

	// Submissions (public - create only)
	api.Post("/submissions", submissionsHandler.Create)

	// Settings (public - read only)
	api.Get("/settings", settingsHandler.GetPublic)
	api.Get("/settings/:category", settingsHandler.GetByCategory)

	// ============ AUTH ROUTES ============

	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// ============ ADMIN ROUTES (protected) ============

	admin := api.Group("/admin", middleware.AuthRequired(cfg))

	// Auth - me
	admin.Get("/me", authHandler.Me)

	// Gallery management
	adminGallery := admin.Group("/gallery")
	adminGallery.Get("/", galleryHandler.List)
	adminGallery.Get("/:id", galleryHandler.Get)
	adminGallery.Post("/", galleryHandler.Create)
	adminGallery.Put("/:id", galleryHandler.Update)
	adminGallery.Delete("/:id", galleryHandler.Delete)
	adminGallery.Post("/upload", galleryHandler.Upload)
	adminGallery.Post("/reorder", galleryHandler.Reorder)

	// Projects management
	adminProjects := admin.Group("/projects")
	adminProjects.Get("/", projectsHandler.List)
	adminProjects.Get("/:id", projectsHandler.Get)
	adminProjects.Post("/", projectsHandler.Create)
	adminProjects.Put("/:id", projectsHandler.Update)
	adminProjects.Delete("/:id", projectsHandler.Delete)
	adminProjects.Post("/upload", projectsHandler.Upload)

	// Map icon types management
	adminMapIconTypes := admin.Group("/map-icon-types")
	adminMapIconTypes.Get("/", mapIconTypeHandler.List)
	adminMapIconTypes.Post("/", mapIconTypeHandler.Create)
	adminMapIconTypes.Put("/:id", mapIconTypeHandler.Update)
	adminMapIconTypes.Delete("/:id", mapIconTypeHandler.Delete)
	adminMapIconTypes.Post("/upload", mapIconTypeHandler.Upload)

	// Map icons management
	adminMapIcons := admin.Group("/map-icons")
	adminMapIcons.Get("/", mapIconHandler.List)
	adminMapIcons.Post("/", mapIconHandler.Create)
	adminMapIcons.Put("/:id", mapIconHandler.Update)
	adminMapIcons.Delete("/:id", mapIconHandler.Delete)

	// Submissions management
	adminSubmissions := admin.Group("/submissions")
	adminSubmissions.Get("/", submissionsHandler.List)
	adminSubmissions.Get("/stats", submissionsHandler.Stats)
	adminSubmissions.Get("/:id", submissionsHandler.Get)
	adminSubmissions.Put("/:id", submissionsHandler.Update)
	adminSubmissions.Delete("/:id", submissionsHandler.Delete)

	// Settings management
	adminSettings := admin.Group("/settings")
	adminSettings.Get("/", settingsHandler.List)
	adminSettings.Get("/categories", settingsHandler.GetCategories)
	adminSettings.Get("/:key", settingsHandler.Get)
	adminSettings.Put("/:key", settingsHandler.Update)
	adminSettings.Post("/bulk", settingsHandler.BulkUpdate)
	adminSettings.Post("/seed", settingsHandler.Seed)

	// File upload (general purpose)
	admin.Post("/upload", uploadHandler.Upload)
	admin.Post("/upload/multiple", uploadHandler.UploadMultiple)

	// Serve uploaded files with byte-range support for large media
	app.Static("/uploads", cfg.UploadDir, fiber.Static{
		ByteRange: true,
	})

	// ============ WEBSOCKET ============
	wsHandler := handlers.NewWebSocketHandler()

	app.Use("/ws", wsHandler.Upgrade)
	app.Get("/ws", websocket.New(wsHandler.Handle))
}
