package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/api/routes"
	"github.com/freeflowuniverse/herolauncher/pkg/executor"
	"github.com/freeflowuniverse/herolauncher/pkg/packagemanager"
	"github.com/freeflowuniverse/herolauncher/pkg/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/pug/v2"
)

// @title HeroLauncher API
// @version 1.0
// @description API for HeroLauncher - a modular service manager
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@freeflowuniverse.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9001
// @BasePath /api
func main() {
	// Initialize modules
	redisServer := redis.NewServer()
	executorService := executor.NewExecutor()
	packageManagerService := packagemanager.NewPackageManager()

	// Initialize template engine
	engine := pug.New("./web/templates", ".pug")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
				Error: err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Static files
	app.Static("/", "./web/static")
	app.Static("/css", "./web/static/css")
	app.Static("/js", "./web/static/js")

	// Initialize route handlers
	executorHandler := routes.NewExecutorHandler(executorService)
	packageManagerHandler := routes.NewPackageManagerHandler(packageManagerService)
	redisHandler := routes.NewRedisHandler(redisServer)
	adminHandler := routes.NewAdminHandler()

	// Register routes
	executorHandler.RegisterRoutes(app)
	packageManagerHandler.RegisterRoutes(app)
	redisHandler.RegisterRoutes(app)
	adminHandler.RegisterRoutes(app)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		_ = app.Shutdown()
	}()

	// Start server
	log.Println("Starting server on :9001")
	if err := app.Listen(":9001"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
