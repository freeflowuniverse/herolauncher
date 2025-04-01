package herolauncher

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupSwagger configures the Swagger documentation for the API
func SetupSwagger(app *fiber.App) {
	// Use the default handler for Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "HeroLauncher API Documentation",
	}))
}
