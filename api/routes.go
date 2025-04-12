// api/routes.go
package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/hopekali04/valuations/services" 
)

func SetupRoutes(app *fiber.App, propertyService *services.PropertyService, syncService *services.SyncService) { // Add syncService
	// Middleware
	app.Use(logger.New()) // Basic request logger

	// Create handlers
	propertyHandler := NewPropertyHandler(propertyService)
	syncHandler := NewSyncHandler(syncService) // Create sync handler

	// Group API routes
	api := app.Group("/api/v1") // versioning of the API

	// --- Property Routes ---
	propGroup := api.Group("/properties")

	propGroup.Post("/", propertyHandler.CreateProperty)
	propGroup.Post("/bulk", propertyHandler.CreateMultipleProperties)

	propGroup.Get("/", propertyHandler.GetAllProperties)
	propGroup.Get("/search", propertyHandler.SearchProperties)
	
	propGroup.Get("/:id", propertyHandler.GetPropertyByID)


	// --- Sync Route ---
	// This will automatically fetch from an API endpoint and sync the data with our Database
	api.Post("/sync", syncHandler.TriggerSync) // Add the sync endpoint
}