package api

import (
	"fmt"
	"runtime"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/gofiber/fiber/v2"
	"github.com/shirou/gopsutil/v3/host"
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
	// @Tags api
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /api/hardware-stats [get]
	admin.Get("/hardware-stats", h.getHardwareStatsJSON)

	// @Summary Get process stats
	// @Description Get process statistics in JSON format
	// @Tags api
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /api/process-stats [get]
	admin.Get("/process-stats", h.getProcessStatsJSON)
}

// getProcessStatsJSON returns process statistics in JSON format for API consumption
func (h *AdminHandler) getProcessStatsJSON(c *fiber.Ctx) error {
	// Get process stats from the StatsManager
	var processStats map[string]interface{}
	if h.statsManager != nil {
		processStats = h.statsManager.GetProcessStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		processStats = stats.GetProcessStats()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    processStats,
	})
}

// getHardwareStatsJSON returns hardware stats in JSON format for API consumption
func (h *AdminHandler) getHardwareStatsJSON(c *fiber.Ctx) error {
	// Get hardware stats from the StatsManager
	var hardwareStats map[string]interface{}
	if h.statsManager != nil {
		hardwareStats = h.statsManager.GetHardwareStats()
	} else {
		// Fallback to direct function call if StatsManager is not available
		hardwareStats = stats.GetHardwareStats()
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    hardwareStats,
	})
}
