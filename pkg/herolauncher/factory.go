package herolauncher

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/executor"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api/routes"
	"github.com/freeflowuniverse/herolauncher/pkg/packagemanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/client"
	"github.com/freeflowuniverse/herolauncher/pkg/redisserver"
	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/pug/v2"
)

// Config holds the configuration for the HeroLauncher server
type Config struct {
	Port                 string
	RedisTCPPort         string
	RedisSocketPath      string
	TemplatesPath        string
	StaticFilesPath      string
	PMSocketPath         string // ProcessManager socket path
	PMSecret             string // ProcessManager authentication secret
}

// DefaultConfig returns a default configuration for the HeroLauncher server
func DefaultConfig() Config {
	// Get the absolute path to the project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	
	// Check for PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "9020" // Default port if not specified
	}

	return Config{
		Port:            port,
		RedisTCPPort:    "6379",
		RedisSocketPath: "/tmp/herolauncher_new.sock",
		PMSocketPath:    "/tmp/processmanager.sock", // Default ProcessManager socket path
		PMSecret:        "secret123", // Default ProcessManager secret
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
	pmClient        *client.Client
	pmProcess       *os.Process    // Process for the process manager
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
	
	// Initialize process manager client
	pmClient := client.New(config.PMSocketPath, config.PMSecret)

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
		pmClient:        pmClient,
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
	serviceHandler := routes.NewServiceHandler(hl.pmClient, log.Default())
	// Initialize StatsManager
	statsManager, err := stats.NewStatsManagerWithDefaults()
	if err != nil {
		log.Printf("Warning: Failed to initialize StatsManager: %v\n", err)
		statsManager = nil
	}

	// Pass HeroLauncher as an UptimeProvider and StatsManager
	adminHandler := routes.NewAdminHandler(hl, statsManager)

	// Register routes
	executorHandler.RegisterRoutes(hl.app)
	packageManagerHandler.RegisterRoutes(hl.app)
	redisHandler.RegisterRoutes(hl.app)
	adminHandler.RegisterRoutes(hl.app)
	serviceHandler.RegisterRoutes(hl.app)
}

// GetUptime returns the uptime of the HeroLauncher server as a formatted string
func (hl *HeroLauncher) GetUptime() string {
	// Calculate uptime based on the server's start time
	uptimeDuration := time.Since(hl.startTime)

	// Use more precise calculation for the uptime
	totalSeconds := int(uptimeDuration.Seconds())
	days := totalSeconds / (24 * 3600)
	hours := (totalSeconds % (24 * 3600)) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	// Format the uptime string based on the duration
	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}

// startProcessManager starts the process manager as a background process
func (hl *HeroLauncher) startProcessManager() error {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	processManagerPath := filepath.Join(projectRoot, "cmd/processmanager/main.go")

	log.Printf("Starting process manager from: %s", processManagerPath)

	// Check if processmanager is already running by testing the socket
	if _, err := os.Stat(hl.config.PMSocketPath); err == nil {
		// Try to connect to test if it's actually running
		pmClient := client.New(hl.config.PMSocketPath, hl.config.PMSecret)
		err := pmClient.Connect()
		if err == nil {
			pmClient.Close()
			log.Printf("Process manager already running, using existing instance")
			return nil
		}
		// If socket exists but we can't connect, remove it as it's stale
		_ = os.Remove(hl.config.PMSocketPath)
	}

	// Start the process manager
	cmd := exec.Command("go", "run", processManagerPath, "-socket", hl.config.PMSocketPath, "-secret", hl.config.PMSecret)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start process manager: %v", err)
	}

	hl.pmProcess = cmd.Process
	log.Printf("Started process manager with PID: %d", cmd.Process.Pid)

	// Wait for the process manager to start up
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if the socket exists
			if _, err := os.Stat(hl.config.PMSocketPath); err == nil {
				// Test connection
				pmClient := client.New(hl.config.PMSocketPath, hl.config.PMSecret)
				err := pmClient.Connect()
				if err == nil {
					pmClient.Close()
					log.Printf("Process manager is up and running")
					return nil
				}
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for process manager to start")
		}
	}
}

// Start starts the HeroLauncher server
func (hl *HeroLauncher) Start() error {
	// Start the process manager first
	err := hl.startProcessManager()
	if err != nil {
		log.Printf("Warning: Failed to start process manager: %v", err)
		// Continue anyway, we'll just show warnings in the UI
	}

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down server...")
		
		// Kill the process manager if we started it
		if hl.pmProcess != nil {
			log.Println("Stopping process manager...")
			_ = hl.pmProcess.Kill()
		}
		
		_ = hl.app.Shutdown()
	}()

	// Start server
	log.Printf("Starting server on :%s", hl.config.Port)
	return hl.app.Listen(":" + hl.config.Port)
}
