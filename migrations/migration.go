package migrations

import (
	"log"
	"paper-management-backend/database"
	"paper-management-backend/models"
)

func Migrate() {
	err := database.DB.AutoMigrate(&models.Role{},
		&models.User{}, &models.Document{}, &models.JenisDocument{})
	if err != nil {
		log.Fatal("Migration failed", err)
	}
}
