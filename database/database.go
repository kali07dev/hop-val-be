// database/database.go
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/hopekali04/valuations/config"
	"github.com/hopekali04/valuations/models"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	cfg := config.Cfg.Database

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode, cfg.TimeZone)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level (Silent, Error, Warn, Info)
			IgnoreRecordNotFoundError: false,       // Don't ignore ErrRecordNotFound error
			Colorful:                  true,        // Enable color
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	DB = db
	fmt.Println("Database connection established.")
	return DB, nil
}

func MigrateDB(db *gorm.DB) error {
	fmt.Println("Running database migrations...")

	// database models
	err := db.AutoMigrate(
		&models.User{},
		&models.Agent{},
		&models.Location{},
		&models.CoverPhoto{},
		&models.Property{},
	)
	if err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	fmt.Println("Database migration completed.")
	return nil
}
