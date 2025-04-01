package herolauncher

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	// Import generated docs
	_ "github.com/freeflowuniverse/herolauncher/pkg/herolauncher/docs"
)

// SetupSwagger configures the Swagger documentation for the API
func SetupSwagger(app *fiber.App) {
	// Add global CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:9001,http://127.0.0.1:9001",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
	}))

	// Use the default handler for Swagger
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
		Title:       "HeroLauncher API Documentation",
		Layout:      "BaseLayout",
		DocExpansion: "list",
	}))

	// Add a redirect from /swagger to /swagger/index.html
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html")
	})
}
