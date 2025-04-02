package herolauncher

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/executor"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/handlers"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/pages"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/redisserver"
	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/mock"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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
	HJSocketPath    string // HeroJobs socket path
}

// DefaultConfig returns a default configuration for the HeroLauncher server
func DefaultConfig() Config {
	// Get the absolute path to the project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")

	// Check for PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "9021" // Default port if not specified
	}

	return Config{
		Port:            port,
		RedisTCPPort:    "6379",
		RedisSocketPath: "/tmp/herolauncher_new.sock",
		PMSocketPath:    "/tmp/processmanager.sock", // Default ProcessManager socket path
		PMSecret:        "secret123",                // Default ProcessManager secret
		HJSocketPath:    "/tmp/herojobs.sock",       // Default HeroJobs socket path
		TemplatesPath:   filepath.Join(projectRoot, "pkg/herolauncher/web/templates"),
		StaticFilesPath: filepath.Join(projectRoot, "pkg/herolauncher/web/static"),
	}
}

// HeroLauncher represents the main application
type HeroLauncher struct {
	app             *fiber.App
	redisServer     *redisserver.Server
	executorService *executor.Executor
	pm              *processmanager.ProcessManager
	pmProcess       *os.Process           // Process for the process manager
	hjProcess       *os.Process           // Process for the HeroJobs server
	vfsManager      interfaces.VFSManager // VFS manager implementation
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

	// Initialize process manager directly
	pm := processmanager.NewProcessManager(config.PMSecret)

	// Set the shared logs path for process manager
	sharedLogsPath := filepath.Join(os.TempDir(), "herolauncher_logs")
	pm.SetLogsBasePath(sharedLogsPath)

	// Initialize VFS manager and client
	vfsManager := mock.NewMockVFSManager() // Using mock implementation for now

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
	SetupSwagger(app)

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
		pm:              pm,
		vfsManager:      vfsManager,
		config:          config,
		startTime:       time.Now(),
	}

	// Initialize and register route handlers
	hl.setupRoutes()

	return hl
}

// setupRoutes initializes and registers all route handlers
func (hl *HeroLauncher) setupRoutes() {
	// Initialize StatsManager
	statsManager, err := stats.NewStatsManagerWithDefaults()
	if err != nil {
		log.Printf("Warning: Failed to initialize StatsManager: %v\n", err)
		statsManager = nil
	}

	// Initialize API handlers
	apiAdminHandler := api.NewAdminHandler(hl, statsManager)
	apiServiceHandler := api.NewServiceHandler(hl.config.PMSocketPath, hl.config.PMSecret, log.Default())

	// Initialize Page handlers
	pageAdminHandler := pages.NewAdminHandler(hl, statsManager, hl.config.PMSocketPath, hl.config.PMSecret)
	pageServiceHandler := pages.NewServiceHandler(hl.config.PMSocketPath, hl.config.PMSecret, log.Default())
	
	// Initialize Jobs page handler
	pageJobHandler, err := pages.NewJobHandler(hl.config.HJSocketPath, log.Default())
	if err != nil {
		log.Printf("Warning: Failed to initialize Jobs page handler: %v\n", err)
	}

	// Initialize JobHandler
	jobHandler, err := handlers.NewJobHandler(hl.config.HJSocketPath, log.Default())
	if err != nil {
		log.Printf("Warning: Failed to initialize JobHandler: %v\n", err)
	} else {
		// Register Job routes
		jobHandler.RegisterRoutes(hl.app)
	}

	// Register API routes
	apiAdminHandler.RegisterRoutes(hl.app)
	apiServiceHandler.RegisterRoutes(hl.app)

	// Register Page routes
	pageAdminHandler.RegisterRoutes(hl.app)
	pageServiceHandler.RegisterRoutes(hl.app)
	
	// Register Jobs page routes if handler was initialized successfully
	if pageJobHandler != nil {
		pageJobHandler.RegisterRoutes(hl.app)
	}

	// TODO: Move these to appropriate API or pages packages
	executorHandler := api.NewExecutorHandler(hl.executorService)
	//vfsHandler := routesold.NewVFSHandler(hl.vfsClient, log.Default())

	// Create new API handlers
	redisAddr := "localhost:" + hl.config.RedisTCPPort
	redisHandler := api.NewRedisHandler(redisAddr, false)
	jetHandler := api.NewJetHandler()

	// Register legacy routes (to be migrated)
	executorHandler.RegisterRoutes(hl.app)
	//vfsHandler.RegisterRoutes(hl.app)

	// Register new API routes
	redisHandler.RegisterRoutes(hl.app)
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
		// Try to connect to the socket to verify it's working
		conn, err := net.Dial("unix", hl.config.PMSocketPath)
		if err == nil {
			// Socket is valid and we can connect to it
			conn.Close()
			log.Printf("Found existing process manager socket, using it instead of starting a new one")
			return nil
		}
		
		// If socket exists but we can't connect, assume it's stale
		log.Printf("Found existing socket, but can't connect to it: %v", err)
		log.Printf("Removing stale socket and starting a new process manager")
		_ = os.Remove(hl.config.PMSocketPath)
	}

	// Define shared logs path
	sharedLogsPath := filepath.Join(os.TempDir(), "herolauncher_logs")
	
	// Ensure the logs directory exists
	if err := os.MkdirAll(sharedLogsPath, 0755); err != nil {
		log.Printf("Warning: Failed to create logs directory: %v", err)
	}
	
	// Start the process manager with the shared logs path
	cmd := exec.Command("go", "run", processManagerPath, 
		"-socket", hl.config.PMSocketPath, 
		"-secret", hl.config.PMSecret,
		"-logs", sharedLogsPath)
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

// startHeroJobs starts the HeroJobs server as a background process
func (hl *HeroLauncher) startHeroJobs() error {
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../..")
	heroJobsPath := filepath.Join(projectRoot, "cmd/herojobs/main.go")

	log.Printf("Starting HeroJobs from: %s", heroJobsPath)

	// Check if HeroJobs is already running by testing the socket
	if _, err := os.Stat(hl.config.HJSocketPath); err == nil {
		// Try to connect to the socket to verify it's working
		conn, err := net.Dial("unix", hl.config.HJSocketPath)
		if err == nil {
			// Socket is valid and we can connect to it
			conn.Close()
			log.Printf("Found existing HeroJobs socket, using it instead of starting a new one")
			return nil
		}
		
		// If socket exists but we can't connect, assume it's stale
		log.Printf("Found existing HeroJobs socket, but can't connect to it: %v", err)
		log.Printf("Removing stale socket and starting a new HeroJobs server")
		_ = os.Remove(hl.config.HJSocketPath)
	}

	// Define shared logs path
	sharedLogsPath := filepath.Join(os.TempDir(), "herolauncher_logs/jobs")
	
	// Ensure the logs directory exists
	if err := os.MkdirAll(sharedLogsPath, 0755); err != nil {
		log.Printf("Warning: Failed to create logs directory: %v", err)
	}
	
	// Start HeroJobs with the shared logs path
	cmd := exec.Command("go", "run", heroJobsPath, 
		"-socket", hl.config.HJSocketPath,
		"-logs", sharedLogsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start HeroJobs: %v", err)
	}

	// Store the process reference for graceful shutdown
	hl.hjProcess = cmd.Process
	log.Printf("Started HeroJobs with PID: %d", cmd.Process.Pid)

	// Wait for HeroJobs to start up
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if the socket exists
			if _, err := os.Stat(hl.config.HJSocketPath); err == nil {
				// If socket exists, assume HeroJobs is running
				log.Printf("HeroJobs is up and running")
				return nil
			}
		case <-timeout:
			return fmt.Errorf("timeout waiting for HeroJobs to start")
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
	
	// Start HeroJobs
	err = hl.startHeroJobs()
	if err != nil {
		log.Printf("Warning: Failed to start HeroJobs: %v", err)
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

		// Kill the HeroJobs server if we started it
		if hl.hjProcess != nil {
			log.Println("Stopping HeroJobs server...")
			_ = hl.hjProcess.Kill()
		}

		_ = hl.app.Shutdown()
	}()

	// Start server
	log.Printf("Starting server on :%s", hl.config.Port)
	return hl.app.Listen(":" + hl.config.Port)
}
