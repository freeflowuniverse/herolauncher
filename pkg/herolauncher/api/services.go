package api

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/gofiber/fiber/v2"
)

// ProcessDisplayInfo represents information about a process for display purposes
type ProcessDisplayInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	StartTime string `json:"start_time"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
}

// ConvertToDisplayInfo converts a ProcessInfo from the processmanager package to ProcessDisplayInfo
func ConvertToDisplayInfo(info *processmanager.ProcessInfo) ProcessDisplayInfo {
	// Calculate uptime from start time
	uptime := formatUptime(time.Since(info.StartTime))
	
	return ProcessDisplayInfo{
		ID:        fmt.Sprintf("%d", info.PID),
		Name:      info.Name,
		Status:    string(info.Status),
		Uptime:    uptime,
		StartTime: info.StartTime.Format("2006-01-02 15:04:05"),
		CPU:       fmt.Sprintf("%.2f%%", info.CPUPercent),
		Memory:    fmt.Sprintf("%.2f MB", info.MemoryMB),
	}
}

// ServiceHandler handles service-related API routes
type ServiceHandler struct {
	pm     *processmanager.ProcessManager
	logger *log.Logger
}

// default number of log lines to retrieve - use a high value to essentially show all logs
const DefaultLogLines = 10000

// NewServiceHandler creates a new service handler with the provided process manager
func NewServiceHandler(pm *processmanager.ProcessManager, logger *log.Logger) *ServiceHandler {
	return &ServiceHandler{
		pm:     pm,
		logger: logger,
	}
}

// RegisterRoutes registers service API routes
func (h *ServiceHandler) RegisterRoutes(app *fiber.App) {
	services := app.Group("/api/services")

	// @Summary Get running services
	// @Description Get a list of all currently running services
	// @Tags services
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string][]ProcessDisplayInfo
	// @Failure 500 {object} map[string]string
	// @Router /api/services/running [get]
	services.Get("/running", h.getRunningServices)

	// @Summary Start a service
	// @Description Start a new service with the given name and command
	// @Tags services
	// @Accept x-www-form-urlencoded
	// @Produce json
	// @Param name formData string true "Service name"
	// @Param command formData string true "Command to run"
	// @Success 200 {object} map[string]interface{}
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /api/services/start [post]
	services.Post("/start", h.startService)

	// @Summary Stop a service
	// @Description Stop a running service by name
	// @Tags services
	// @Accept x-www-form-urlencoded
	// @Produce json
	// @Param name formData string true "Service name"
	// @Success 200 {object} map[string]interface{}
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /api/services/stop [post]
	services.Post("/stop", h.stopService)

	// @Summary Restart a service
	// @Description Restart a running service by name
	// @Tags services
	// @Accept x-www-form-urlencoded
	// @Produce json
	// @Param name formData string true "Service name"
	// @Success 200 {object} map[string]interface{}
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /api/services/restart [post]
	services.Post("/restart", h.restartService)

	// @Summary Delete a service
	// @Description Delete a service by name
	// @Tags services
	// @Accept x-www-form-urlencoded
	// @Produce json
	// @Param name formData string true "Service name"
	// @Success 200 {object} map[string]interface{}
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /api/services/delete [post]
	services.Post("/delete", h.deleteService)

	// @Summary Get process logs
	// @Description Get logs for a specific process
	// @Tags services
	// @Accept x-www-form-urlencoded
	// @Produce json
	// @Param name formData string true "Service name"
	// @Param lines formData integer false "Number of log lines to retrieve"
	// @Success 200 {object} map[string]string
	// @Failure 400 {object} map[string]string
	// @Failure 500 {object} map[string]string
	// @Router /api/services/logs [post]
	services.Post("/logs", h.getProcessLogs)
}

// getProcessList gets a list of processes from the process manager
func (h *ServiceHandler) getProcessList() ([]ProcessDisplayInfo, error) {
	// Debug: Log the function entry
	h.logger.Printf("Entering getProcessList() function")

	// Get the list of processes directly from the process manager
	processes := h.pm.ListProcesses()

	// Convert to display info format
	displayInfoList := make([]ProcessDisplayInfo, 0, len(processes))
	for _, procInfo := range processes {
		displayInfo := ConvertToDisplayInfo(procInfo)
		displayInfoList = append(displayInfoList, displayInfo)
	}

	// Debug: Log the number of processes
	h.logger.Printf("Found %d processes", len(displayInfoList))

	return displayInfoList, nil
}

// formatUptime formats a duration as a human-readable uptime string
func formatUptime(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())
	days := totalSeconds / (24 * 3600)
	hours := (totalSeconds % (24 * 3600)) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

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

// startService starts a new service
func (h *ServiceHandler) startService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")
	command := c.FormValue("command")

	// Validate inputs
	if name == "" || command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name and command are required",
		})
	}

	// Start the process
	pid, err := h.pm.StartProcess(name, command)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to start service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' started with PID %d", name, pid),
		"pid":     pid,
	})
}

// stopService stops a service
func (h *ServiceHandler) stopService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// Validate inputs
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name is required",
		})
	}

	// Stop the process
	err := h.pm.StopProcess(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to stop service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' stopped", name),
	})
}

// restartService restarts a service
func (h *ServiceHandler) restartService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// Validate inputs
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name is required",
		})
	}

	// Restart the process
	err := h.pm.RestartProcess(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to restart service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' restarted", name),
	})
}

// deleteService deletes a service
func (h *ServiceHandler) deleteService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// Validate inputs
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name is required",
		})
	}

	// Delete the process
	err := h.pm.DeleteProcess(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to delete service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' deleted", name),
	})
}

// getRunningServices returns a list of running services in JSON format
func (h *ServiceHandler) getRunningServices(c *fiber.Ctx) error {
	// Get the list of processes
	processes, err := h.getProcessList()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to get process list: %v", err),
		})
	}

	// Return the processes as JSON
	return c.JSON(fiber.Map{
		"success":   true,
		"processes": processes,
	})
}

// getProcessLogs retrieves logs for a specific process
func (h *ServiceHandler) getProcessLogs(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")
	linesStr := c.FormValue("lines", fmt.Sprintf("%d", DefaultLogLines))

	// Validate inputs
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Name is required",
		})
	}

	// Parse lines
	lines, err := strconv.Atoi(linesStr)
	if err != nil {
		lines = DefaultLogLines
	}

	// Get logs
	logs, err := h.pm.GetProcessLogs(name, lines)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to get logs: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    logs,
	})
}
