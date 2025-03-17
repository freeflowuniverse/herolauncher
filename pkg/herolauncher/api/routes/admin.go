package routes

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
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

// getNetworkSpeed returns the current network speed in Mbps
func getNetworkSpeed() (string, string) {
	networkUpSpeed := "Unknown"
	networkDownSpeed := "Unknown"
	
	// Get initial counters
	initNetInfo, err := net.IOCounters(false)
	if err == nil && len(initNetInfo) > 0 {
		initBytesRecv := initNetInfo[0].BytesRecv
		initBytesSent := initNetInfo[0].BytesSent

		// Wait a short time to measure throughput
		time.Sleep(200 * time.Millisecond)

		// Get updated counters
		updatedNetInfo, err := net.IOCounters(false)
		if err == nil && len(updatedNetInfo) > 0 {
			// Calculate bytes transferred during the interval
			bytesRecvDelta := updatedNetInfo[0].BytesRecv - initBytesRecv
			bytesSentDelta := updatedNetInfo[0].BytesSent - initBytesSent

			// Convert to Mbps (megabits per second)
			// Multiply by 8 to convert bytes to bits, divide by time in seconds (0.2), then by 1,000,000 for Mbps
			recvMbps := float64(bytesRecvDelta) * 8 / 0.2 / 1000000
			sentMbps := float64(bytesSentDelta) * 8 / 0.2 / 1000000

			networkUpSpeed = fmt.Sprintf("%.2fMbps", sentMbps)
			networkDownSpeed = fmt.Sprintf("%.2fMbps", recvMbps)
		}
	}
	
	return networkUpSpeed, networkDownSpeed
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
	admin.Get("/system/logs", h.getSystemLogs)
	admin.Get("/system/logs-test", h.getSystemLogsTest)
	
	// API endpoints
	admin.Get("/api/hardware-stats", h.getHardwareStatsJSON)
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
	networkUpSpeed, networkDownSpeed := getNetworkSpeed()
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

// getHardwareStats returns only the hardware stats for Unpoly polling
func (h *AdminHandler) getHardwareStats(c *fiber.Ctx) error {
	// Initialize default values
	cpuInfo := "Unknown"
	memoryInfo := "Unknown"
	diskInfo := "Unknown"
	networkInfo := "Unknown"

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
	networkUpSpeed, networkDownSpeed := getNetworkSpeed()
	networkInfo = fmt.Sprintf("Up: %s\nDown: %s", networkUpSpeed, networkDownSpeed)

	// Create hardware info map
	hardware := fiber.Map{
		"cpu":     cpuInfo,
		"memory":  memoryInfo,
		"disk":    diskInfo,
		"network": networkInfo,
	}

	return c.Render("admin/system/hardware_stats", fiber.Map{
		"hardware": hardware,
	})
}

// getHardwareStatsJSON returns hardware stats in JSON format for API consumption
func (h *AdminHandler) getHardwareStatsJSON(c *fiber.Ctx) error {
	// CPU - use runtime.NumCPU() as a fallback if gopsutil fails
	cpuCount := runtime.NumCPU()
	cpuModel := "Unknown"
	
	// CPU usage percentage
	cpuUsage := 0.0
	
	// Try to get CPU info, but don't fail if it's not available
	try := func() {
		info, err := cpu.Info()
		if err == nil && len(info) > 0 {
			cpuModel = info[0].ModelName
		}
		
		// Get CPU usage percentage
		percentages, err := cpu.Percent(0, false)
		if err == nil && len(percentages) > 0 {
			cpuUsage = percentages[0]
		}
	}
	try()

	// Memory
	memoryTotal := 0.0
	memoryUsed := 0.0
	memoryUsedPercent := 0.0
	
	try = func() {
		memInfo, err := mem.VirtualMemory()
		if err == nil {
			memoryTotal = float64(memInfo.Total) / (1024 * 1024 * 1024) // Convert to GB
			memoryUsed = float64(memInfo.Used) / (1024 * 1024 * 1024)   // Convert to GB
			memoryUsedPercent = memInfo.UsedPercent
		}
	}
	try()

	// Disk
	diskTotal := 0.0
	diskFree := 0.0
	diskUsedPercent := 0.0
	
	try = func() {
		diskUsage, err := disk.Usage("/")
		if err == nil {
			diskTotal = float64(diskUsage.Total) / (1024 * 1024 * 1024) // Convert to GB
			diskFree = float64(diskUsage.Free) / (1024 * 1024 * 1024)   // Convert to GB
			diskUsedPercent = diskUsage.UsedPercent
		}
	}
	try()

	// Network
	networkUpSpeed, networkDownSpeed := getNetworkSpeed()
	
	// Parse the network speeds to get numeric values
	parseSpeed := func(speed string) float64 {
		if speed == "Unknown" {
			return 0.0
		}
		val, err := strconv.ParseFloat(strings.TrimSuffix(speed, "Mbps"), 64)
		if err != nil {
			return 0.0
		}
		return val
	}
	
	networkUp := parseSpeed(networkUpSpeed)
	networkDown := parseSpeed(networkDownSpeed)

	// Create hardware stats JSON response
	return c.JSON(fiber.Map{
		"cpu": fiber.Map{
			"cores": cpuCount,
			"model": cpuModel,
			"usage": cpuUsage,
		},
		"memory": fiber.Map{
			"total": memoryTotal,
			"used": memoryUsed,
			"usedPercent": memoryUsedPercent,
		},
		"disk": fiber.Map{
			"total": diskTotal,
			"free": diskFree,
			"usedPercent": diskUsedPercent,
		},
		"network": fiber.Map{
			"upload": networkUp,
			"download": networkDown,
		},
		"timestamp": time.Now().Unix(),
	})
}
