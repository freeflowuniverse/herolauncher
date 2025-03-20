package routes

import (
	"fmt"
	"runtime"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/host"
)

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}



// AdminHandler handles admin-related routes
type AdminHandler struct {
	uptimeProvider UptimeProvider
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(uptimeProvider UptimeProvider) *AdminHandler {
	return &AdminHandler{
		uptimeProvider: uptimeProvider,
	}
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	// Admin routes
	admin := app.Group("/admin")

	// Dashboard
	admin.Get("/", h.getDashboard)

	// Services
	admin.Get("/services", h.getServices)

	// Packages
	admin.Get("/packages", h.getPackages)

	// System routes
	admin.Get("/system/info", h.getSystemInfo)
	admin.Get("/system/hardware-stats", h.getHardwareStats)
	admin.Get("/system/processes", h.getProcesses)
	admin.Get("/system/processes-data", h.getProcessesData)
	admin.Get("/system/logs", h.getSystemLogs)
	admin.Get("/system/logs-test", h.getSystemLogsTest)

	// API endpoints
	admin.Get("/api/hardware-stats", h.getHardwareStatsJSON)
	admin.Get("/api/process-stats", h.getProcessStatsJSON)
	admin.Get("/system/settings", h.getSystemSettings)

	// Redirect root to admin
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/admin")
	})
}

// getDashboard renders the admin dashboard
func (h *AdminHandler) getDashboard(c *fiber.Ctx) error {
	return c.Render("admin/index", fiber.Map{
		"title": "Dashboard",
	})
}

// getServices renders the services page
func (h *AdminHandler) getServices(c *fiber.Ctx) error {
	return c.Render("admin/services", fiber.Map{
		"title": "Services",
	})
}

// getPackages renders the packages page
func (h *AdminHandler) getPackages(c *fiber.Ctx) error {
	return c.Render("admin/packages", fiber.Map{
		"title": "Packages",
	})
}

// getSystemInfo renders the system info page
func (h *AdminHandler) getSystemInfo(c *fiber.Ctx) error {
	// Initialize default values
	cpuInfo := "Unknown"
	memoryInfo := "Unknown"
	diskInfo := "Unknown"
	networkInfo := "Unknown"
	osInfo := "Unknown"
	uptimeInfo := "Unknown"

	// Get hardware stats from the stats package
	hardwareStats := stats.GetHardwareStats()
	
	// Extract the formatted strings
	cpuInfo = hardwareStats["cpu"].(string)
	memoryInfo = hardwareStats["memory"].(string)
	diskInfo = hardwareStats["disk"].(string)
	networkInfo = hardwareStats["network"].(string)

	// Software information
	// OS and Uptime
	try := func() {
		hostInfo, err := host.Info()
		if err == nil {
			osInfo = fmt.Sprintf("%s %s", hostInfo.Platform, hostInfo.PlatformVersion)

			// Format uptime from seconds to days and hours
			uptime := hostInfo.Uptime
			days := uptime / (60 * 60 * 24)
			hours := (uptime % (60 * 60 * 24)) / (60 * 60)
			uptimeInfo = fmt.Sprintf("%d days, %d hours", days, hours)
		}
	}
	try()

	// If OS info couldn't be retrieved, use runtime info
	if osInfo == "Unknown" {
		osInfo = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	}

	// Go version is always available through runtime
	goVersion := runtime.Version()

	// HeroLauncher version
	heroLauncherVersion := "v0.1.0" // This should be fetched from a version constant

	// Always use the uptimeProvider when available
	if h.uptimeProvider != nil {
		uptimeInfo = h.uptimeProvider.GetUptime()
	} else if uptimeInfo == "Unknown" {
		// If uptimeProvider is not available and system uptime couldn't be retrieved, use a default value
		// Calculate a simulated uptime based on current time
		startTime := time.Now().Add(-72 * time.Hour) // Simulate 3 days uptime
		uptimeDuration := time.Since(startTime)
		days := int(uptimeDuration.Hours() / 24)
		hours := int(uptimeDuration.Hours()) % 24
		uptimeInfo = fmt.Sprintf("%d days, %d hours", days, hours)
	}

	// Create hardware info map
	hardware := fiber.Map{
		"cpu":     cpuInfo,
		"memory":  memoryInfo,
		"disk":    diskInfo,
		"network": networkInfo,
	}

	// Create software info map
	software := fiber.Map{
		"os":           osInfo,
		"go_version":   goVersion,
		"herolauncher": heroLauncherVersion,
		"uptime":       uptimeInfo,
	}

	print(hardware)
	print(software)
	return c.Render("admin/system/info", fiber.Map{
		"title": "System Info",
		"system": fiber.Map{
			"hardware": hardware,
			"software": software,
		},
	})
}

// getSystemLogs renders the system logs page
func (h *AdminHandler) getSystemLogs(c *fiber.Ctx) error {
	// Get recent logs
	logs := []fiber.Map{
		{"timestamp": "2025-03-14T06:30:00Z", "level": "info", "message": "System started"},
		{"timestamp": "2025-03-14T06:35:12Z", "level": "info", "message": "Service 'redis' started"},
		{"timestamp": "2025-03-14T07:15:45Z", "level": "warning", "message": "High memory usage detected"},
		{"timestamp": "2025-03-14T07:25:30Z", "level": "info", "message": "Package 'web-ui' updated"},
	}

	return c.Render("admin/system/logs", fiber.Map{
		"title": "System Logs",
		"logs":  logs,
	})
}

// getSystemLogsTest renders the test logs page
func (h *AdminHandler) getSystemLogsTest(c *fiber.Ctx) error {
	return c.Render("admin/system/logs_test", fiber.Map{
		"title": "Test Logs Page",
	})
}

// getSystemSettings renders the system settings page
func (h *AdminHandler) getSystemSettings(c *fiber.Ctx) error {
	// Get current settings
	settings := fiber.Map{
		"debug_mode":      true,
		"auto_update":     false,
		"backup_enabled":  true,
		"backup_interval": "daily",
		"log_level":       "info",
	}

	return c.Render("admin/system/settings", fiber.Map{
		"title":    "System Settings",
		"settings": settings,
	})
}

// getHardwareStats returns only the hardware stats for Unpoly polling
func (h *AdminHandler) getHardwareStats(c *fiber.Ctx) error {
	// Get hardware stats from the stats package
	hardwareStats := stats.GetHardwareStats()
	
	// Convert to fiber.Map for template rendering
	hardware := fiber.Map{}
	for k, v := range hardwareStats {
		hardware[k] = v
	}

	return c.Render("admin/system/hardware_stats", fiber.Map{
		"hardware": hardware,
	})
}

// getProcessStatsJSON returns process statistics in JSON format for API consumption
func (h *AdminHandler) getProcessStatsJSON(c *fiber.Ctx) error {
	// Get process stats from the stats package (limit to top 30 processes)
	processData, err := stats.GetProcessStats(30)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get process stats: " + err.Error(),
		})
	}
	
	// Convert to []fiber.Map for JSON response
	processStats := make([]fiber.Map, len(processData.Processes))
	for i, proc := range processData.Processes {
		processStats[i] = fiber.Map{
			"pid":             proc.PID,
			"name":            proc.Name,
			"status":          proc.Status,
			"cpu_percent":     proc.CPUPercent,
			"memory_mb":       proc.MemoryMB,
			"create_time_str": proc.CreateTime,
			"is_current":      proc.IsCurrent,
		}
	}

	// Return JSON response
	return c.JSON(fiber.Map{
		"processes": processStats,
		"timestamp": time.Now().Unix(),
	})
}

// getHardwareStatsJSON returns hardware stats in JSON format for API consumption
func (h *AdminHandler) getHardwareStatsJSON(c *fiber.Ctx) error {
	// Get hardware stats from the stats package
	hardwareStats := stats.GetHardwareStatsJSON()
	
	// Convert to fiber.Map for JSON response
	response := fiber.Map{}
	for k, v := range hardwareStats {
		response[k] = v
	}

	// Return JSON response
	return c.JSON(response)
}

// getProcesses renders the processes page without waiting for process data
func (h *AdminHandler) getProcesses(c *fiber.Ctx) error {
	// Initialize with an empty processes array to ensure the variable exists
	return c.Render("admin/system/processes", fiber.Map{
		"processes": []fiber.Map{},
	})
}

// getProcessesData returns the HTML fragment for processes data
func (h *AdminHandler) getProcessesData(c *fiber.Ctx) error {
	// Get process data from the stats package
	processData, err := stats.GetProcessStats(0) // Get all processes
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to get process data: " + err.Error())
	}
	
	// Convert to []fiber.Map for template rendering
	processStats := make([]fiber.Map, len(processData.Processes))
	for i, proc := range processData.Processes {
		processStats[i] = fiber.Map{
			"pid":             proc.PID,
			"name":            proc.Name,
			"status":          proc.Status,
			"cpu_percent":     proc.CPUPercent,
			"memory_mb":       proc.MemoryMB,
			"create_time_str": proc.CreateTime,
			"is_current":      proc.IsCurrent,
		}
	}

	// Return only the table fragment with process data
	return c.Render("admin/system/processes_table", fiber.Map{
		"processes": processStats,
		"title": "System Processes", // Adding title to ensure variable scope is working
		"layout": "", // Disable layout for partial template
	})
}
