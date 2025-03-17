package herolauncher

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/executor"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api/routes"
	"github.com/freeflowuniverse/herolauncher/pkg/packagemanager"
	"github.com/freeflowuniverse/herolauncher/pkg/redisserver"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/pug/v2"
)

// Config holds the configuration for the HeroLauncher server
type Config struct {
	Port            string
	RedisTCPPort    string
	RedisSocketPath string
	TemplatesPath   string
	StaticFilesPath string
}

// DefaultConfig returns a default configuration for the HeroLauncher server
func DefaultConfig() Config {
	// Get the absolute path to the project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..") // go up two levels from factory.go

	return Config{
		Port:            "9020",
		RedisTCPPort:    "6379",
		RedisSocketPath: "/tmp/herolauncher_new.sock",
		TemplatesPath:   filepath.Join(projectRoot, "pkg/herolauncher/web/templates"),
		StaticFilesPath: filepath.Join(projectRoot, "pkg/herolauncher/web/static"),
	}
}

// HeroLauncher represents the main application
type HeroLauncher struct {
	app             *fiber.App
	redisServer     *redisserver.Server
	executorService *executor.Executor
	packageManager  *packagemanager.PackageManager
	config          Config
	startTime       time.Time
}

// New creates a new instance of HeroLauncher with the provided configuration
func New(config Config) *HeroLauncher {
	// Initialize modules
	redisServer := redisserver.NewServer(redisserver.ServerConfig{
		TCPPort:        config.RedisTCPPort,
		UnixSocketPath: config.RedisSocketPath,
	})
	executorService := executor.NewExecutor()
	packageManagerService := packagemanager.NewPackageManager()

	// Initialize template engine with debugging enabled
	// Use absolute path for templates to avoid path resolution issues
	absTemplatePath, err := filepath.Abs(config.TemplatesPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for templates: %v", err)
	}

	engine := pug.New(absTemplatePath, ".pug")
	engine.Debug(true) // Enable debug mode to see template errors
	// Reload templates on each render in development
	engine.Reload(true)

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

	// Static files - serve all directories with proper paths
	app.Static("/", config.StaticFilesPath)
	app.Static("/css", config.StaticFilesPath+"/css")
	app.Static("/js", config.StaticFilesPath+"/js")
	app.Static("/img", config.StaticFilesPath+"/img")
	app.Static("/favicon.ico", config.StaticFilesPath+"/favicon.ico")

	// Create HeroLauncher instance
	hl := &HeroLauncher{
		app:             app,
		redisServer:     redisServer,
		executorService: executorService,
		packageManager:  packageManagerService,
		config:          config,
		startTime:       time.Now(),
	}

	// Initialize and register route handlers
	hl.setupRoutes()

	return hl
}

// setupRoutes initializes and registers all route handlers
func (hl *HeroLauncher) setupRoutes() {
	// Initialize route handlers
	executorHandler := routes.NewExecutorHandler(hl.executorService)
	packageManagerHandler := routes.NewPackageManagerHandler(hl.packageManager)
	redisHandler := routes.NewRedisHandler(hl.redisServer)
	// Pass HeroLauncher as an UptimeProvider
	adminHandler := routes.NewAdminHandler(hl)

	// Register routes
	executorHandler.RegisterRoutes(hl.app)
	packageManagerHandler.RegisterRoutes(hl.app)
	redisHandler.RegisterRoutes(hl.app)
	adminHandler.RegisterRoutes(hl.app)
}

// GetUptime returns the uptime of the HeroLauncher server as a formatted string
func (hl *HeroLauncher) GetUptime() string {
	// Calculate uptime based on the server's start time
	uptimeDuration := time.Since(hl.startTime)
	
	// Extract days and hours for a more readable format
	days := int(uptimeDuration.Hours() / 24)
	hours := int(uptimeDuration.Hours()) % 24
	
	return fmt.Sprintf("%d days, %d hours", days, hours)
}

// Start starts the HeroLauncher server
func (hl *HeroLauncher) Start() error {
	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		_ = hl.app.Shutdown()
	}()

	// Start server
	log.Printf("Starting server on :%s", hl.config.Port)
	return hl.app.Listen(":" + hl.config.Port)
}
