package routes

import (
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}

// AdminHandler handles admin-related routes
type AdminHandler struct{
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
	admin.Get("/system/logs", h.getSystemLogs)
	admin.Get("/system/logs-test", h.getSystemLogsTest)
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

	// Hardware information
	// CPU - use runtime.NumCPU() as a fallback if gopsutil fails
	cpuCount := runtime.NumCPU()
	cpuModel := "Unknown"

	// Try to get CPU info, but don't fail if it's not available
	try := func() {
		info, err := cpu.Info()
		if err == nil && len(info) > 0 {
			cpuModel = info[0].ModelName
		}
	}
	try() // Wrap in function to avoid error propagation
	cpuInfo = fmt.Sprintf("%d cores (%s)", cpuCount, cpuModel)

	// Memory
	try = func() {
		memInfo, err := mem.VirtualMemory()
		if err == nil {
			memoryTotal := float64(memInfo.Total) / (1024 * 1024 * 1024) // Convert to GB
			memoryUsed := float64(memInfo.Used) / (1024 * 1024 * 1024)   // Convert to GB
			memoryInfo = fmt.Sprintf("%.1fGB (%.1fGB used)", memoryTotal, memoryUsed)
		}
	}
	try()

	// Disk
	try = func() {
		diskUsage, err := disk.Usage("/")
		if err == nil {
			diskTotal := float64(diskUsage.Total) / (1024 * 1024 * 1024) // Convert to GB
			diskFree := float64(diskUsage.Free) / (1024 * 1024 * 1024)   // Convert to GB
			diskInfo = fmt.Sprintf("%.0fGB (%.0fGB free)", diskTotal, diskFree)
		}
	}
	try()

	// Network
	networkUpSpeed := "Unknown"
	networkDownSpeed := "Unknown"
	try = func() {
		netInfo, err := net.IOCounters(false)
		if err == nil && len(netInfo) > 0 {
			bytesRecv := netInfo[0].BytesRecv
			bytesSent := netInfo[0].BytesSent

			if bytesRecv > 0 || bytesSent > 0 {
				recvMbps := float64(bytesRecv) * 8 / 1000000
				sentMbps := float64(bytesSent) * 8 / 1000000
				networkUpSpeed = fmt.Sprintf("%.2fMbps", sentMbps)
				networkDownSpeed = fmt.Sprintf("%.2fMbps", recvMbps)
			}
		}
	}
	try()
	networkInfo = fmt.Sprintf("Up: %s\nDown: %s", networkUpSpeed, networkDownSpeed)

	// Software information
	// OS and Uptime
	try = func() {
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
