package seeders

import (
	"paper-management-backend/database"
	"paper-management-backend/models"
)

func SeedRoles() {
	roles := []models.Role{
		{Name: "user"},
		{Name: "admin"},
		{Name: "superuser"},
	}

	for _, role := range roles {
		database.DB.FirstOrCreate(&role, models.Role{Name: role.Name})
	}
}
