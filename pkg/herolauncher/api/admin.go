package api

import (
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
)

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}

// AdminHandler handles admin-related API routes
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

// RegisterRoutes registers all admin API routes
func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	// API endpoints
	admin := app.Group("/api")

	// @Summary Get hardware stats
	// @Description Get hardware statistics in JSON format
	// @Tags admin
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 500 {object} ErrorResponse
	// @Router /api/hardware-stats [get]
	admin.Get("/hardware-stats", h.getHardwareStatsJSON)

	// @Summary Get process stats
	// @Description Get process statistics in JSON format
	// @Tags admin
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Failure 500 {object} ErrorResponse
	// @Router /api/process-stats [get]
	admin.Get("/process-stats", h.getProcessStatsJSON)
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
			"success": false,
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
		"success":   true,
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
	response := fiber.Map{
		"success": true,
	}
	for k, v := range hardwareStats {
		response[k] = v
	}

	// Return JSON response
	return c.JSON(response)
}
