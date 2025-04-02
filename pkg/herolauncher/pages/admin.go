package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/handlers"
	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/host"
)

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}

// AdminHandler handles admin-related page routes
type AdminHandler struct {
	uptimeProvider UptimeProvider
	statsManager   *stats.StatsManager
	pmSocketPath   string
	pmSecret       string
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(uptimeProvider UptimeProvider, statsManager *stats.StatsManager, pmSocketPath, pmSecret string) *AdminHandler {
	// If statsManager is nil, create a new one with default settings
	if statsManager == nil {
		var err error
		statsManager, err = stats.NewStatsManagerWithDefaults()
		if err != nil {
			// Log the error but continue with nil statsManager
			fmt.Printf("Error creating StatsManager: %v\n", err)
		}
	}

	return &AdminHandler{
		uptimeProvider: uptimeProvider,
		statsManager:   statsManager,
		pmSocketPath:   pmSocketPath,
		pmSecret:       pmSecret,
	}
}

// RegisterRoutes registers all admin page routes
func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	// Admin routes
	admin := app.Group("/admin")

	// Dashboard
	admin.Get("/", h.getDashboard)

	// Create service handler with the correct socket path and secret
	serviceHandler := handlers.NewServiceHandler(h.pmSocketPath, h.pmSecret)
	// Services routes
	admin.Get("/services", serviceHandler.GetServices)
	admin.Get("/services/data", serviceHandler.GetServicesFragment)
	admin.Post("/services/start", serviceHandler.StartService)
	admin.Post("/services/stop", serviceHandler.StopService)
	admin.Post("/services/restart", serviceHandler.RestartService)
	admin.Post("/services/delete", serviceHandler.DeleteService)
	admin.Get("/services/logs", serviceHandler.GetServiceLogs)

	// System routes
	admin.Get("/system/info", h.getSystemInfo)
	admin.Get("/system/hardware-stats", h.getHardwareStats)
	
	// Create process handler
	processHandler := handlers.NewProcessHandler(h.statsManager)
	admin.Get("/system/processes", processHandler.GetProcesses)
	admin.Get("/system/processes-data", processHandler.GetProcessesData)
	
	// Create log handler
	// Ensure log directory exists
	// Using the same shared logs path as process manager
	logDir := filepath.Join(os.TempDir(), "herolauncher_logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Error creating log directory: %v\n", err)
	}
	
	logHandler, err := handlers.NewLogHandler(logDir)
	if err != nil {
		fmt.Printf("Error creating log handler: %v\n", err)
		// Fallback to old implementation if log handler creation failed
		admin.Get("/system/logs", h.getSystemLogs)
		admin.Get("/system/logs-test", h.getSystemLogsTest)
	} else {
		fmt.Printf("Log handler created successfully\n")
		// Use the log handler for log routes
		admin.Get("/system/logs", logHandler.GetLogs)
		// Keep the fragment endpoint for backward compatibility
		// but it now just redirects to the main logs endpoint
		admin.Get("/system/logs-fragment", logHandler.GetLogsFragment)
		admin.Get("/system/logs-test", h.getSystemLogsTest) // Keep the test logs route
		
		// Log API endpoints
		app.Get("/api/logs", logHandler.GetLogsAPI)
	}
	
	admin.Get("/system/settings", h.getSystemSettings)

	// OpenRPC routes
	admin.Get("/openrpc", h.getOpenRPCManager)
	admin.Get("/openrpc/vfs", h.getOpenRPCVFS)
	admin.Get("/openrpc/vfs/logs", h.getOpenRPCVFSLogs)

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

// getSystemInfo renders the system info page
func (h *AdminHandler) getSystemInfo(c *fiber.Ctx) error {
	// Initialize default values
	cpuInfo := "Unknown"
	memoryInfo := "Unknown"
	diskInfo := "Unknown"
	networkInfo := "Unknown"
	osInfo := "Unknown"
	uptimeInfo := "Unknown"

	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStats()
	}

	// Extract the formatted strings - safely handle different return types
	if cpuVal, ok := hardwareStats["cpu"]; ok {
		switch v := cpuVal.(type) {
		case string:
			cpuInfo = v
		case map[string]interface{}:
			// Format the map into a string
			if model, ok := v["model"].(string); ok {
				usage := 0.0
				if usagePercent, ok := v["usage_percent"].(float64); ok {
					usage = usagePercent
				}
				cpuInfo = fmt.Sprintf("%s (Usage: %.1f%%)", model, usage)
			}
		}
	}

	if memVal, ok := hardwareStats["memory"]; ok {
		switch v := memVal.(type) {
		case string:
			memoryInfo = v
		case map[string]interface{}:
			// Format the map into a string
			total, used := 0.0, 0.0
			if totalGB, ok := v["total_gb"].(float64); ok {
				total = totalGB
			}
			if usedGB, ok := v["used_gb"].(float64); ok {
				used = usedGB
			}
			usedPercent := 0.0
			if percent, ok := v["used_percent"].(float64); ok {
				usedPercent = percent
			}
			memoryInfo = fmt.Sprintf("%.1f GB / %.1f GB (%.1f%% used)", used, total, usedPercent)
		}
	}

	if diskVal, ok := hardwareStats["disk"]; ok {
		switch v := diskVal.(type) {
		case string:
			diskInfo = v
		case map[string]interface{}:
			// Format the map into a string
			total, used := 0.0, 0.0
			if totalGB, ok := v["total_gb"].(float64); ok {
				total = totalGB
			}
			if usedGB, ok := v["used_gb"].(float64); ok {
				used = usedGB
			}
			usedPercent := 0.0
			if percent, ok := v["used_percent"].(float64); ok {
				usedPercent = percent
			}
			diskInfo = fmt.Sprintf("%.1f GB / %.1f GB (%.1f%% used)", used, total, usedPercent)
		}
	}

	if netVal, ok := hardwareStats["network"]; ok {
		switch v := netVal.(type) {
		case string:
			networkInfo = v
		case map[string]interface{}:
			// Format the map into a string
			var interfaces []string
			if ifaces, ok := v["interfaces"].([]interface{}); ok {
				for _, iface := range ifaces {
					if ifaceMap, ok := iface.(map[string]interface{}); ok {
						name := ifaceMap["name"].(string)
						ip := ifaceMap["ip"].(string)
						interfaces = append(interfaces, fmt.Sprintf("%s: %s", name, ip))
					}
				}
				networkInfo = strings.Join(interfaces, ", ")
			}
		}
	}

	// Get OS info
	hostInfo, err := host.Info()
	if err == nil {
		osInfo = fmt.Sprintf("%s %s (%s)", hostInfo.Platform, hostInfo.PlatformVersion, hostInfo.KernelVersion)
	}

	// Get uptime
	if h.uptimeProvider != nil {
		uptimeInfo = h.uptimeProvider.GetUptime()
	}

	// Render the template with the system info
	return c.Render("admin/system/info", fiber.Map{
		"title":       "System Information",
		"cpuInfo":     cpuInfo,
		"memoryInfo":  memoryInfo,
		"diskInfo":    diskInfo,
		"networkInfo": networkInfo,
		"osInfo":      osInfo,
		"uptimeInfo":  uptimeInfo,
	})
}

// getSystemLogs renders the system logs page
func (h *AdminHandler) getSystemLogs(c *fiber.Ctx) error {
	return c.Render("admin/system/logs", fiber.Map{
		"title": "System Logs",
	})
}

// getSystemLogsTest renders the test logs page
func (h *AdminHandler) getSystemLogsTest(c *fiber.Ctx) error {
	return c.Render("admin/system/logs_test", fiber.Map{
		"title": "Test Logs",
	})
}

// getSystemSettings renders the system settings page
func (h *AdminHandler) getSystemSettings(c *fiber.Ctx) error {
	// Get system settings
	// This is a placeholder - in a real app, you would fetch settings from a database or config file
	settings := map[string]interface{}{
		"logLevel":        "info",
		"enableDebugMode": false,
		"dataDirectory":   "/var/lib/herolauncher",
		"maxLogSize":      "100MB",
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

	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStats()
	}

	// Extract the formatted strings - safely handle different return types
	if cpuVal, ok := hardwareStats["cpu"]; ok {
		switch v := cpuVal.(type) {
		case string:
			cpuInfo = v
		case map[string]interface{}:
			// Format the map into a string
			if model, ok := v["model"].(string); ok {
				cpuInfo = model
			}
		}
	}

	if memVal, ok := hardwareStats["memory"]; ok {
		switch v := memVal.(type) {
		case string:
			memoryInfo = v
		case map[string]interface{}:
			// Format the map into a string
			total, used := 0.0, 0.0
			if totalGB, ok := v["total_gb"].(float64); ok {
				total = totalGB
			}
			if usedGB, ok := v["used_gb"].(float64); ok {
				used = usedGB
			}
			memoryInfo = fmt.Sprintf("%.1f GB / %.1f GB", used, total)
		}
	}

	if diskVal, ok := hardwareStats["disk"]; ok {
		switch v := diskVal.(type) {
		case string:
			diskInfo = v
		case map[string]interface{}:
			// Format the map into a string
			total, used := 0.0, 0.0
			if totalGB, ok := v["total_gb"].(float64); ok {
				total = totalGB
			}
			if usedGB, ok := v["used_gb"].(float64); ok {
				used = usedGB
			}
			diskInfo = fmt.Sprintf("%.1f GB / %.1f GB", used, total)
		}
	}

	if netVal, ok := hardwareStats["network"]; ok {
		switch v := netVal.(type) {
		case string:
			networkInfo = v
		case map[string]interface{}:
			// Format the map into a string
			var interfaces []string
			if ifaces, ok := v["interfaces"].([]interface{}); ok {
				for _, iface := range ifaces {
					if ifaceMap, ok := iface.(map[string]interface{}); ok {
						name := ifaceMap["name"].(string)
						ip := ifaceMap["ip"].(string)
						interfaces = append(interfaces, fmt.Sprintf("%s: %s", name, ip))
					}
				}
				networkInfo = strings.Join(interfaces, ", ")
			}
		}
	}

	// Format for display
	cpuUsage := "0.0%"
	memUsage := "0.0%"
	diskUsage := "0.0%"

	// Safely extract usage percentages
	if cpuVal, ok := hardwareStats["cpu"].(map[string]interface{}); ok {
		if usagePercent, ok := cpuVal["usage_percent"].(float64); ok {
			cpuUsage = fmt.Sprintf("%.1f%%", usagePercent)
		}
	}

	if memVal, ok := hardwareStats["memory"].(map[string]interface{}); ok {
		if usedPercent, ok := memVal["used_percent"].(float64); ok {
			memUsage = fmt.Sprintf("%.1f%%", usedPercent)
		}
	}

	if diskVal, ok := hardwareStats["disk"].(map[string]interface{}); ok {
		if usedPercent, ok := diskVal["used_percent"].(float64); ok {
			diskUsage = fmt.Sprintf("%.1f%%", usedPercent)
		}
	}

	// Render only the hardware stats fragment
	return c.Render("admin/system/hardware_stats_fragment", fiber.Map{
		"cpuInfo":     cpuInfo,
		"memoryInfo":  memoryInfo,
		"diskInfo":    diskInfo,
		"networkInfo": networkInfo,
		"cpuUsage":    cpuUsage,
		"memUsage":    memUsage,
		"diskUsage":   diskUsage,
	})
}

// getProcesses has been moved to the handlers package
// See handlers.ProcessHandler.GetProcesses

// getOpenRPCManager renders the OpenRPC Manager view page
func (h *AdminHandler) getOpenRPCManager(c *fiber.Ctx) error {
	return c.Render("admin/openrpc/index", fiber.Map{
		"title": "OpenRPC Manager",
	})
}

// getOpenRPCVFS renders the OpenRPC VFS view page
func (h *AdminHandler) getOpenRPCVFS(c *fiber.Ctx) error {
	return c.Render("admin/openrpc/vfs", fiber.Map{
		"title": "VFS OpenRPC Interface",
	})
}

// getOpenRPCVFSLogs renders the OpenRPC logs content for Unpoly or direct access
func (h *AdminHandler) getOpenRPCVFSLogs(c *fiber.Ctx) error {
	// Get query parameters
	method := c.Query("method", "")
	params := c.Query("params", "")

	// Define available methods and their display names
	methods := []string{
		"vfs_ls",
		"vfs_read",
		"vfs_write",
		"vfs_mkdir",
		"vfs_rm",
		"vfs_mv",
		"vfs_cp",
		"vfs_exists",
		"vfs_isdir",
		"vfs_isfile",
	}

	methodDisplayNames := map[string]string{
		"vfs_ls":     "List Directory",
		"vfs_read":   "Read File",
		"vfs_write":  "Write File",
		"vfs_mkdir":  "Create Directory",
		"vfs_rm":     "Remove File/Directory",
		"vfs_mv":     "Move/Rename",
		"vfs_cp":     "Copy",
		"vfs_exists": "Check Exists",
		"vfs_isdir":  "Is Directory",
		"vfs_isfile": "Is File",
	}

	// Generate method options HTML
	methodOptions := generateMethodOptions(methods, methodDisplayNames)

	// Initialize variables
	var requestJSON, responseJSON, responseTime string
	var hasResponse bool

	// If a method is selected, make the OpenRPC call
	if method != "" {
		// Prepare the request
		requestJSON = fmt.Sprintf(`{
			"jsonrpc": "2.0",
			"method": "%s",
			"params": %s,
			"id": 1
		}`, method, params)

		// In a real implementation, we would make the actual OpenRPC call here
		// For now, we'll just simulate a response

		// Simulate response time (would be real in production)
		time.Sleep(100 * time.Millisecond)
		responseTime = "100ms"

		// Simulate a response based on the method
		switch method {
		case "vfs_ls":
			responseJSON = `{
				"jsonrpc": "2.0",
				"result": [
					{"name": "file1.txt", "size": 1024, "isDir": false, "modTime": "2023-01-01T12:00:00Z"},
					{"name": "dir1", "size": 0, "isDir": true, "modTime": "2023-01-01T12:00:00Z"}
				],
				"id": 1
			}`
		case "vfs_read":
			responseJSON = `{
				"jsonrpc": "2.0",
				"result": "File content would be here",
				"id": 1
			}`
		default:
			responseJSON = `{
				"jsonrpc": "2.0",
				"result": "Operation completed successfully",
				"id": 1
			}`
		}

		hasResponse = true
	}

	// Determine if this is an Unpoly request
	isUnpoly := c.Get("X-Up-Target") != ""

	// If it's an Unpoly request, render just the logs fragment
	if isUnpoly {
		return c.Render("admin/openrpc/vfs_logs", fiber.Map{
			"methodOptions":  methodOptions,
			"selectedMethod": method,
			"params":         params,
			"requestJSON":    requestJSON,
			"responseJSON":   responseJSON,
			"responseTime":   responseTime,
			"hasResponse":    hasResponse,
		})
	}

	// Otherwise render the full page
	return c.Render("admin/openrpc/vfs_overview", fiber.Map{
		"title":          "VFS OpenRPC Logs",
		"methodOptions":  methodOptions,
		"selectedMethod": method,
		"params":         params,
		"requestJSON":    requestJSON,
		"responseJSON":   responseJSON,
		"responseTime":   responseTime,
		"hasResponse":    hasResponse,
	})
}

// generateMethodOptions generates HTML option tags for method dropdown
func generateMethodOptions(methods []string, methodDisplayNames map[string]string) string {
	var options []string
	for _, method := range methods {
		displayName, ok := methodDisplayNames[method]
		if !ok {
			displayName = method
		}
		options = append(options, fmt.Sprintf(`<option value="%s">%s</option>`, method, displayName))
	}
	return strings.Join(options, "\n")
}

// Note: getProcessesData has been consolidated in the API routes file
// to avoid duplication and ensure consistent behavior
