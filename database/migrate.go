package database

import (
	"log"
	"movie-ticket-booking/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Order{},
		&models.Account{},
		&models.AccountType{},
		&models.Seat{},
		&models.Row{},
		&models.Showtime{},
		&models.Theater{},
		&models.Branch{},
		&models.Movie{},
	)
	if err != nil {
		log.Fatalf("Error during auto-migration: %v", err)
	}
}
