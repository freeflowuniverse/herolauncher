package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/host"
)

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}

// SystemHandler handles system-related page routes
type SystemHandler struct {
	uptimeProvider UptimeProvider
	statsManager   *stats.StatsManager
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(uptimeProvider UptimeProvider, statsManager *stats.StatsManager) *SystemHandler {
	// If statsManager is nil, create a new one with default settings
	if statsManager == nil {
		var err error
		statsManager, err = stats.NewStatsManagerWithDefaults()
		if err != nil {
			// Log the error but continue with nil statsManager
			fmt.Printf("Error creating StatsManager: %v\n", err)
		}
	}

	return &SystemHandler{
		uptimeProvider: uptimeProvider,
		statsManager:   statsManager,
	}
}

// GetSystemInfo renders the system info page
func (h *SystemHandler) GetSystemInfo(c *fiber.Ctx) error {
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

// GetHardwareStats returns only the hardware stats for Unpoly polling
func (h *SystemHandler) GetHardwareStats(c *fiber.Ctx) error {
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

// GetHardwareStatsAPI returns hardware stats in JSON format
func (h *SystemHandler) GetHardwareStatsAPI(c *fiber.Ctx) error {
	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStats()
	}

	return c.JSON(hardwareStats)
}

// GetProcessStatsAPI returns process stats in JSON format for API consumption
func (h *SystemHandler) GetProcessStatsAPI(c *fiber.Ctx) error {
	// Check if StatsManager is properly initialized
	if h.statsManager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "System error: Stats manager not initialized",
		})
	}

	// Get process data from the StatsManager
	processData, err := h.statsManager.GetProcessStatsFresh(100) // Limit to 100 processes
	if err != nil {
		// Try getting cached data as fallback
		processData, err = h.statsManager.GetProcessStats(100)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get process data: " + err.Error(),
			})
		}
	}

	// Convert to fiber.Map for JSON response
	response := fiber.Map{
		"total":     processData.Total,
		"filtered":  processData.Filtered,
		"timestamp": time.Now().Unix(),
	}

	// Convert processes to a slice of maps
	processes := make([]fiber.Map, len(processData.Processes))
	for i, proc := range processData.Processes {
		processes[i] = fiber.Map{
			"pid":             proc.PID,
			"name":            proc.Name,
			"status":          proc.Status,
			"cpu_percent":     proc.CPUPercent,
			"memory_mb":       proc.MemoryMB,
			"create_time_str": proc.CreateTime,
			"is_current":      proc.IsCurrent,
		}
	}
	
	response["processes"] = processes

	// Return JSON response
	return c.JSON(response)
}

// GetSystemLogs renders the system logs page
func (h *SystemHandler) GetSystemLogs(c *fiber.Ctx) error {
	return c.Render("admin/system/logs", fiber.Map{
		"title": "System Logs",
	})
}

// GetSystemLogsTest renders the test logs page
func (h *SystemHandler) GetSystemLogsTest(c *fiber.Ctx) error {
	return c.Render("admin/system/logs_test", fiber.Map{
		"title": "Test Logs",
	})
}

// GetSystemSettings renders the system settings page
func (h *SystemHandler) GetSystemSettings(c *fiber.Ctx) error {
	// Get the current time
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Render the template with the system settings
	return c.Render("admin/system/settings", fiber.Map{
		"title":       "System Settings",
		"currentTime": currentTime,
		"settings": map[string]interface{}{
			"autoUpdate":      true,
			"logLevel":        "info",
			"maxLogSize":      "100MB",
			"backupFrequency": "Daily",
		},
	})
}
