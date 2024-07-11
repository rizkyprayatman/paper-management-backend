package routes

import (
	"paper-management-backend/controllers"
	"paper-management-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func DocumentRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Use(middleware.Authorization)

	api.Get("/jenis-documents", controllers.GetAllJenisDocuments)
	api.Get("/get-all-documents", controllers.GetAllDocuments)
	api.Post("/upload-document", controllers.UploadDocument)
	api.Post("/share-document/:id", controllers.ShareDocument)
}
