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
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/redisserver"
	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/mock"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/openrpc"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/gofiber/template/jet/v2"
)

// Config holds the configuration for the HeroLauncher server
type Config struct {
	Port            string
	RedisTCPPort    string
	RedisSocketPath string
	TemplatesPath   string
	StaticFilesPath string
	PMSocketPath    string // ProcessManager socket path
	PMSecret        string // ProcessManager authentication secret
	VFSSocketPath   string // VFS OpenRPC socket path
	VFSSecret       string // VFS OpenRPC authentication secret
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
		PMSecret:        "secret123",                // Default ProcessManager secret
		VFSSocketPath:   "/tmp/vfs.sock",            // Default VFS socket path
		VFSSecret:       "vfs_secret",               // Default VFS secret
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
	pm              *processmanager.ProcessManager
	pmProcess       *os.Process           // Process for the process manager
	vfsManager      interfaces.VFSManager // VFS manager implementation
	vfsClient       *openrpc.Client       // VFS OpenRPC client
	vfsServer       *openrpc.Server       // VFS OpenRPC server
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

	// Initialize process manager directly
	pm := processmanager.NewProcessManager(config.PMSecret)

	// Initialize VFS manager and client
	vfsManager := mock.NewMockVFSManager() // Using mock implementation for now
	vfsClient := openrpc.NewClient(config.VFSSocketPath, config.VFSSecret)

	// Initialize template engine with debugging enabled
	// Use absolute path for templates to avoid path resolution issues
	absTemplatePath, err := filepath.Abs(config.TemplatesPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for templates: %v", err)
	}

	engine := jet.New(absTemplatePath, ".jet")
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

	// Initialize VFS OpenRPC server
	vfsServer, err := openrpc.NewServer(vfsManager, config.VFSSocketPath, config.VFSSecret)
	if err != nil {
		log.Printf("Warning: Failed to initialize VFS OpenRPC server: %v\n", err)
		vfsServer = nil
	}

	// Create HeroLauncher instance
	hl := &HeroLauncher{
		app:             app,
		redisServer:     redisServer,
		executorService: executorService,
		packageManager:  packageManagerService,
		pm:              pm,
		vfsManager:      vfsManager,
		vfsClient:       vfsClient,
		vfsServer:       vfsServer,
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
	serviceHandler := routes.NewServiceHandler(hl.pm, log.Default())
	vfsHandler := routes.NewVFSHandler(hl.vfsClient, log.Default())
	jetHandler := routes.NewJetHandler()
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
	vfsHandler.RegisterRoutes(hl.app)
	jetHandler.RegisterRoutes(hl.app)
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
		// If socket exists but we can't verify it's running, assume it's stale
		log.Printf("Found existing socket, but can't verify if process manager is running")
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
				// If socket exists, assume process manager is running
				log.Printf("Process manager is up and running")
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for process manager to start")
		}
	}
}

// Start starts the HeroLauncher server
func (hl *HeroLauncher) Start() error {
	// Start VFS OpenRPC server if available
	if hl.vfsServer != nil {
		if err := hl.vfsServer.Start(); err != nil {
			log.Printf("Warning: Failed to start VFS OpenRPC server: %v\n", err)
		} else {
			log.Printf("VFS OpenRPC server started on socket: %s\n", hl.config.VFSSocketPath)
		}
	}
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
