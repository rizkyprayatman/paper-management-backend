package main

import (
	"log"
	"paper-management-backend/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := fiber.New()

	app.Use(cors.New())

	// Static files
	app.Static("/", "./static")

	// Database connection
	database.ConnectDB()

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.SendFile("./static/404.html")
	})

	app.Listen(":3000")
}
