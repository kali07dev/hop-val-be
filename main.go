// main.go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/valuations/api"      
	"github.com/hopekali04/valuations/config"  
	"github.com/hopekali04/valuations/database"
	"github.com/hopekali04/valuations/services"
)

func main() {
	// 1. Load Configuration
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Connect to Database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Run Migrations
	if err := database.MigrateDB(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 4. Initialize Services
	propertyService := services.NewPropertyService(db)
	syncService := services.NewSyncService(db) // Initialize SyncService

	// 5. Create Fiber App
	app := fiber.New()

	// 6. Setup Routes
	api.SetupRoutes(app, propertyService, syncService) // Pass both services

	// 7. Start Server
	serverAddr := ":3000" // Make port configurable later
	log.Printf("Starting server on %s\n", serverAddr)
	err = app.Listen(serverAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}