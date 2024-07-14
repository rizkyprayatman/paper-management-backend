package main

import (
	"log"
	"paper-management-backend/database"
	"paper-management-backend/migrations"
	"paper-management-backend/routes"
	"paper-management-backend/seeders"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://paper-management.livetest.my.id, https://www.paper-management.livetest.my.id",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))

	// Static files
	app.Static("/", "./static")

	// Database connection
	database.ConnectDB()

	if !isTableExists(database.DB, "users") {
		migrations.Migrate()
		seeders.SeedRoles()
		seeders.SeedJenisDocument()
	}

	// Setup routes
	routes.UserRoutes(app)
	routes.DocumentRoutes(app)

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.SendFile("./static/404.html")
	})

	app.Listen(":3000")
}

func isTableExists(db *gorm.DB, tableName string) bool {
	var exists bool
	db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = ?)", tableName).Scan(&exists)
	return exists
}
