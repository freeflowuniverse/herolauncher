package handlers

import (
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
)

// ProcessHandler handles process-related routes
type ProcessHandler struct {
	statsManager *stats.StatsManager
}

// NewProcessHandler creates a new ProcessHandler
func NewProcessHandler(statsManager *stats.StatsManager) *ProcessHandler {
	return &ProcessHandler{
		statsManager: statsManager,
	}
}

// GetProcessStatsJSON returns process stats in JSON format for API consumption
func (h *ProcessHandler) GetProcessStatsJSON(c *fiber.Ctx) error {
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
		"total": processData.Total,
		"filtered": processData.Filtered,
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

// GetProcesses renders the processes page with initial process data
func (h *ProcessHandler) GetProcesses(c *fiber.Ctx) error {
	// Check if StatsManager is properly initialized
	if h.statsManager == nil {
		return c.Render("admin/system/processes", fiber.Map{
			"processes": []fiber.Map{},
			"error":     "System error: Stats manager not initialized",
			"warning":   "The process manager is not properly initialized.",
		})
	}

	// Force cache refresh for process stats
	h.statsManager.ForceUpdate("process")

	// Get process data from the StatsManager
	processData, err := h.statsManager.GetProcessStatsFresh(0) // Get all processes with fresh data
	if err != nil {
		// Try getting cached data as fallback
		processData, err = h.statsManager.GetProcessStats(0)
		if err != nil {
			// If there's an error, still render the page but with empty data
			return c.Render("admin/system/processes", fiber.Map{
				"processes": []fiber.Map{},
				"error":     "Failed to load process data: " + err.Error(),
				"warning":   "System attempted both fresh and cached data retrieval but failed.",
			})
		}
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
			"cpu_percent_str": fmt.Sprintf("%.1f%%", proc.CPUPercent),
			"memory_mb_str":   fmt.Sprintf("%.1f MB", proc.MemoryMB),
		}
	}

	// Render the full page with initial process data
	return c.Render("admin/system/processes", fiber.Map{
		"processes": processStats,
	})
}

// GetProcessesData returns the HTML fragment for processes data
func (h *ProcessHandler) GetProcessesData(c *fiber.Ctx) error {
	// Check if this is a manual refresh request (with X-Requested-With header set)
	isManualRefresh := c.Get("X-Requested-With") == "XMLHttpRequest"

	// Check if StatsManager is properly initialized
	if h.statsManager == nil {
		return c.Render("admin/system/processes_data", fiber.Map{
			"error":   "System error: Stats manager not initialized",
			"layout":  "",
		})
	}

	// For manual refresh, always get fresh data by forcing cache invalidation
	var processData *stats.ProcessStats
	var err error

	// Force cache refresh for process stats on manual refresh
	if isManualRefresh {
		h.statsManager.ForceUpdate("process")
	}

	if isManualRefresh {
		// Force bypass cache for manual refresh by using fresh data
		processData, err = h.statsManager.GetProcessStatsFresh(0)
	} else {
		// Use cached data for auto-polling
		processData, err = h.statsManager.GetProcessStats(0)
	}

	if err != nil {
		// Try alternative method if the primary method fails
		if isManualRefresh {
			processData, err = h.statsManager.GetProcessStats(0)
		} else {
			processData, err = h.statsManager.GetProcessStatsFresh(0)
		}
		
		if err != nil {
			// Handle AJAX requests differently from regular requests
			isAjax := c.Get("X-Requested-With") == "XMLHttpRequest"
			if isAjax {
				return c.Status(fiber.StatusInternalServerError).SendString("Failed to get process data: " + err.Error())
			}
			// For regular requests, render the error within the fragment
			return c.Render("admin/system/processes_data", fiber.Map{
				"error":   "Failed to get process data: " + err.Error(),
				"layout":  "",
			})
		}
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
			"cpu_percent_str": fmt.Sprintf("%.1f%%", proc.CPUPercent),
			"memory_mb_str":   fmt.Sprintf("%.1f MB", proc.MemoryMB),
		}
	}

	// Create a boolean to indicate if we have processes
	hasProcesses := len(processStats) > 0

	// Create template data with fiber.Map
	templateData := fiber.Map{
		"hasProcesses": hasProcesses,
		"processCount": len(processStats),
		"processStats": processStats,
		"layout":       "", // Disable layout for partial template
	}
	
	// Return only the table HTML content directly to be injected into the processes-table-content div
	return c.Render("admin/system/processes_data", templateData)
}


