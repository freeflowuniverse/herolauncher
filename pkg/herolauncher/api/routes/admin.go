package routes

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/openrpc"
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
	statsManager   *stats.StatsManager
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(uptimeProvider UptimeProvider, statsManager *stats.StatsManager) *AdminHandler {
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
			upload, download := "Unknown", "Unknown"
			if uploadSpeed, ok := v["upload_speed"].(string); ok {
				upload = uploadSpeed
			}
			if downloadSpeed, ok := v["download_speed"].(string); ok {
				download = downloadSpeed
			}
			networkInfo = fmt.Sprintf("Upload: %s, Download: %s", upload, download)
		}
	}

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
	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStats()
	}

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
	// Get process stats from the StatsManager (limit to top 30 processes)
	var processData *stats.ProcessStats
	var err error
	if h.statsManager != nil {
		processData, err = h.statsManager.GetProcessStats(30)
	} else {
		// Fallback to direct function call if StatsManager is not available
		processData, err = stats.GetProcessStats(30)
	}
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
	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStatsJSON()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStatsJSON()
	}

	// Convert to fiber.Map for JSON response
	response := fiber.Map{}
	for k, v := range hardwareStats {
		response[k] = v
	}

	// Return JSON response
	return c.JSON(response)
}

// getProcesses renders the processes page with initial process data
func (h *AdminHandler) getProcesses(c *fiber.Ctx) error {
	// Get process data from the StatsManager
	processData, err := h.statsManager.GetProcessStats(0) // Get all processes
	if err != nil {
		// If there's an error, still render the page but with empty data
		return c.Render("admin/system/processes", fiber.Map{
			"processes": []fiber.Map{},
			"error":     "Failed to load process data: " + err.Error(),
		})
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

	// Render the full page with initial process data
	return c.Render("admin/system/processes", fiber.Map{
		"processes": processStats,
	})
}

// getOpenRPCManager renders the OpenRPC Manager view page
func (h *AdminHandler) getOpenRPCManager(c *fiber.Ctx) error {
	return c.Render("admin/openrpc/index", fiber.Map{
		"title": "OpenRPC Manager",
	})
}

// getOpenRPCVFS renders the OpenRPC VFS view page
func (h *AdminHandler) getOpenRPCVFS(c *fiber.Ctx) error {
	return c.Render("admin/openrpc/vfs", fiber.Map{
		"title": "Virtual File System API",
	})
}

// getOpenRPCVFSLogs renders the OpenRPC logs content for Unpoly or direct access
func (h *AdminHandler) getOpenRPCVFSLogs(c *fiber.Ctx) error {
	// Get the OpenRPC manager name and endpoint from query parameters or use defaults
	managerName := c.Query("manager", "Virtual Filesystem")
	managerEndpoint := c.Query("endpoint", "/api/vfs/logs")

	// Get VFS client
	vfsClient := openrpc.NewClient("/tmp/vfs.sock", "")
	defer vfsClient.Close()

	// Get available methods from the OpenRPC schema
	methods := []string{}
	schema, err := vfsClient.Discover()
	if err == nil {
		for _, method := range schema.Methods {
			// Skip internal RPC methods
			if method.Name != "rpc.discover" && method.Name != "rpc.introspect" {
				methods = append(methods, method.Name)
			}
		}
	} else {
		// If we couldn't get the schema, add some default methods
		servicePrefix := "vfs"
		methods = append(methods, 
			servicePrefix+".list", 
			servicePrefix+".get", 
			servicePrefix+".create", 
			servicePrefix+".update", 
			servicePrefix+".delete")
	}
	
	// Create a map of method names and their display names for the template
	methodDisplayNames := make(map[string]string)
	for _, method := range methods {
		// Extract the method name after the last dot and replace underscores with spaces
		parts := strings.Split(method, ".")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			methodDisplayNames[method] = strings.ReplaceAll(name, "_", " ")
		} else {
			methodDisplayNames[method] = method
		}
	}

	// Create the data map with required variables
	data := fiber.Map{
		"title":             "OpenRPC VFS Logs",
		"managerName":       managerName,
		"managerEndpoint":   managerEndpoint,
		"methods":           methods,
		"methodDisplayNames": methodDisplayNames,
	}

	// Check if this is an Unpoly request
	if c.Get("X-Up-Target", "") != "" {
		// Only render the content of the logs tab without the layout
		data["layout"] = false
	}
	
	// If accessed directly, render with the full layout
	data["title"] = managerName + " Logs"
	
	// Create a temporary HTML file with the template variables replaced
	tmplContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s - HeroLauncher</title>
  <link rel="stylesheet" href="/static/css/bootstrap.min.css">
  <link rel="stylesheet" href="/static/css/custom.css">
  <script src="/static/js/jquery.min.js"></script>
  <script src="/static/js/bootstrap.bundle.min.js"></script>
  <script src="/static/js/unpoly.min.js"></script>
</head>
<body>
  <div class="container-fluid p-4">
    <div class="row mb-4">
      <div class="col">
        <h2>%s Logs</h2>
        <p>View and filter logs from the %s service.</p>
      </div>
    </div>

    <div class="row mb-4">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <h5 class="mb-0">Filter Logs</h5>
          </div>
          <div class="card-body">
            <form id="filter-form">
              <input type="hidden" id="manager" name="manager" value="%s">
              <input type="hidden" id="endpoint" name="endpoint" value="%s">
              
              <div class="row">
                <div class="col-md-3">
                  <label for="method-filter">Filter by Method:</label>
                  <select class="form-control" id="method-filter">
                    <option value="">All Methods</option>
                    %s
                  </select>
                </div>
                
                <div class="col-md-3">
                  <label for="status-filter">Filter by Status:</label>
                  <select class="form-control" id="status-filter">
                    <option value="">All Statuses</option>
                    <option value="success">Success</option>
                    <option value="error">Error</option>
                  </select>
                </div>
                
                <div class="col-md-3">
                  <label for="date-filter">Filter by Date:</label>
                  <input type="date" class="form-control" id="date-filter">
                </div>
                
                <div class="col-md-3">
                  <label for="limit-filter">Limit Results:</label>
                  <select class="form-control" id="limit-filter">
                    <option value="50">50</option>
                    <option value="100">100</option>
                    <option value="200">200</option>
                    <option value="500">500</option>
                  </select>
                </div>
              </div>
              
              <div class="row mt-3">
                <div class="col">
                  <button type="button" id="apply-filters" class="btn btn-primary">Apply Filters</button>
                  <button type="button" id="reset-filters" class="btn btn-secondary">Reset Filters</button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <div class="row">
      <div class="col">
        <div class="card">
          <div class="card-header">
            <h5 class="mb-0">Logs</h5>
          </div>
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-striped">
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Method</th>
                    <th>Status</th>
                    <th>Duration</th>
                    <th>Details</th>
                  </tr>
                </thead>
                <tbody id="logs-table-body">
                  <!-- Logs will be loaded here -->
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>

  <script>
    $(document).ready(function() {
      // Apply filters when the button is clicked
      $('#apply-filters').click(function() {
        const queryParams = new URLSearchParams();
        
        // Get filter values
        const methodFilter = $('#method-filter').val();
        const statusFilter = $('#status-filter').val();
        const dateFilter = $('#date-filter').val();
        const limitFilter = $('#limit-filter').val();
        
        // Add filters to query parameters if they are set
        if (methodFilter) queryParams.append('method', methodFilter);
        if (statusFilter) queryParams.append('status', statusFilter);
        if (dateFilter) queryParams.append('date', dateFilter);
        if (limitFilter) queryParams.append('limit', limitFilter);
        
        // Add the manager and endpoint parameters to preserve them when reloading
        queryParams.append('manager', document.getElementById('manager').value);
        queryParams.append('endpoint', document.getElementById('endpoint').value);
        
        // Redirect to the same page with new query parameters
        window.location.href = '/admin/openrpc/vfs/logs?' + queryParams.toString();
      });
      
      // Reset filters when the button is clicked
      $('#reset-filters').click(function() {
        // Clear all filter inputs
        $('#method-filter').val('');
        $('#status-filter').val('');
        $('#date-filter').val('');
        $('#limit-filter').val('50');
        
        // Redirect to the base URL with only manager and endpoint parameters
        const queryParams = new URLSearchParams();
        queryParams.append('manager', document.getElementById('manager').value);
        queryParams.append('endpoint', document.getElementById('endpoint').value);
        window.location.href = '/admin/openrpc/vfs/logs?' + queryParams.toString();
      });
    });
  </script>
</body>
</html>`, 
		data["title"], managerName, managerName, managerName, managerEndpoint, generateMethodOptions(methods, methodDisplayNames))
	
	return c.Type("html").SendString(tmplContent)
}

// generateMethodOptions generates HTML option tags for method dropdown
func generateMethodOptions(methods []string, methodDisplayNames map[string]string) string {
	options := ""
	for _, method := range methods {
		displayName := method
		if name, ok := methodDisplayNames[method]; ok {
			displayName = name
		}
		options += fmt.Sprintf("<option value=\"%s\">%s</option>\n", method, displayName)
	}
	return options
}

// getProcessesData returns the HTML fragment for processes data
func (h *AdminHandler) getProcessesData(c *fiber.Ctx) error {
	// Check if this is a manual refresh request (with X-Requested-With header set)
	isManualRefresh := c.Get("X-Requested-With") == "XMLHttpRequest"
	
	// For manual refresh, always get fresh data by forcing cache invalidation
	var processData *stats.ProcessStats
	var err error
	if isManualRefresh {
		// Force bypass cache for manual refresh by using fresh data
		processData, err = h.statsManager.GetProcessStatsFresh(0)
	} else {
		// Use cached data for auto-polling
		processData, err = h.statsManager.GetProcessStats(0)
	}
	if err != nil {
		// Handle AJAX requests differently from regular requests
		isAjax := c.Get("X-Requested-With") == "XMLHttpRequest"
		if isAjax {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to get process data: " + err.Error())
		}
		// For regular requests, render the error within the fragment
		return c.Render("admin/system/processes_fragment", fiber.Map{
			"error":   "Failed to get process data: " + err.Error(),
			"layout":  "",
		})
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

	// Return only the table HTML content directly to be injected into the processes-table-content div
	return c.Render("admin/system/processes_fragment", fiber.Map{
		"processes": processStats,
		"title":     "System Processes", // Adding title to ensure variable scope is working
		"layout":    "",                 // Disable layout for partial template
	})
}
