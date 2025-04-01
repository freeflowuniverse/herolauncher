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

	apiServices := app.Group("/api/services")
	apiServices.Get("/running", h.getRunningServices)
	apiServices.Post("/start", h.startService)
	apiServices.Post("/stop", h.stopService)
	apiServices.Post("/restart", h.restartService)
	apiServices.Post("/delete", h.deleteService)
	apiServices.Post("/logs", h.getProcessLogs)

	adminServices := app.Group("/admin/services")
	adminServices.Get("/", h.getServicesPage)
	adminServices.Get("/data", h.getServicesData)
	adminServices.Get("/running", h.getRunningServices)
	adminServices.Post("/start", h.startService)
	adminServices.Post("/stop", h.stopService)
	adminServices.Post("/restart", h.restartService)
	adminServices.Post("/delete", h.deleteService)
	adminServices.Post("/logs", h.getProcessLogs)
}

// getProcessList gets a list of processes from the process manager
// TODO: add swagger annotations
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
// @Router /admin/services/start [post]
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

	// Start the process with default values
	// logEnabled=true, deadline=0 (no deadline), no cron, no jobID
	err := h.pm.StartProcess(name, command, true, 0, "", "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to start service: %v", err),
		})
	}

	// Get the process info to return the PID
	processes := h.pm.ListProcesses()
	var pid int
	for _, proc := range processes {
		if proc.Name == name {
			pid = int(proc.PID) // Convert int32 to int
			break
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' started with PID %d", name, pid),
		"pid":     pid,
	})
}

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
// @Router /admin/services/stop [post]
// stopService stops a service
func (h *ServiceHandler) stopService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// For backward compatibility, try ID field if name is empty
	if name == "" {
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Process name is required",
			})
		}
	}

	// Log the stop request
	h.logger.Printf("Stopping process with name: %s", name)

	// Stop the process
	err := h.pm.StopProcess(name)
	if err != nil {
		h.logger.Printf("Error stopping process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to stop service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' stopped successfully", name),
	})
}

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
// @Router /admin/services/restart [post]
// restartService restarts a service
func (h *ServiceHandler) restartService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// For backward compatibility, try ID field if name is empty
	if name == "" {
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Process name is required",
			})
		}
	}

	// Log the restart request
	h.logger.Printf("Restarting process with name: %s", name)

	// Restart the process
	err := h.pm.RestartProcess(name)
	if err != nil {
		h.logger.Printf("Error restarting process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to restart service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Service '%s' restarted successfully", name),
	})
}

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
// @Router /admin/services/delete [post]
// deleteService deletes a service
func (h *ServiceHandler) deleteService(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// Validate inputs
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Service name is required",
		})
	}

	// Debug: Log the delete request
	h.logger.Printf("Deleting process with name: %s", name)

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
		"message": fmt.Sprintf("Service '%s' deleted successfully", name),
	})
}

// @Summary Get running services
// @Description Get a list of all currently running services
// @Tags services
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]ProcessDisplayInfo
// @Failure 500 {object} map[string]string
// @Router /api/services/running [get]
// @Router /admin/services/running [get]
func (h *ServiceHandler) getRunningServices(c *fiber.Ctx) error {
	// Get the list of processes
	processes, err := h.getProcessList()
	if err != nil {
		h.logger.Printf("Error getting process list: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to get process list: %v", err),
		})
	}

	// Filter to only include running processes
	runningProcesses := make([]ProcessDisplayInfo, 0)
	for _, proc := range processes {
		if proc.Status == "running" {
			runningProcesses = append(runningProcesses, proc)
		}
	}

	// Return the processes as JSON
	return c.JSON(fiber.Map{
		"success":   true,
		"services":  runningProcesses,
		"processes": processes, // Keep for backward compatibility
	})
}

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
// @Router /admin/services/logs [post]
// getProcessLogs retrieves logs for a specific process
func (h *ServiceHandler) getProcessLogs(c *fiber.Ctx) error {
	// Get form values
	name := c.FormValue("name")

	// For backward compatibility, try ID field if name is empty
	if name == "" {
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Process name is required",
			})
		}
	}

	// Get the number of lines to retrieve
	linesStr := c.FormValue("lines")
	lines := DefaultLogLines
	if linesStr != "" {
		if parsedLines, err := strconv.Atoi(linesStr); err == nil && parsedLines > 0 {
			lines = parsedLines
		}
	}

	// Log the request
	h.logger.Printf("Getting logs for process: %s (lines: %d)", name, lines)

	// Get logs
	logs, err := h.pm.GetProcessLogs(name, lines)
	if err != nil {
		h.logger.Printf("Error getting process logs: %v", err)
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

// @Summary Get services page
// @Description Get the services management page
// @Tags admin
// @Produce html
// @Success 200 {string} string "HTML content"
// @Failure 500 {object} map[string]string
// @Router /admin/services/ [get]
// getServicesPage renders the services page
func (h *ServiceHandler) getServicesPage(c *fiber.Ctx) error {
	// Get processes to display on the initial page load
	processes, _ := h.getProcessList()

	// No need to check for socket existence since we're using the process manager directly
	var warning string
	if h.pm == nil {
		warning = "Process manager is not properly initialized."
		h.logger.Printf("Warning: %s", warning)
	}

	return c.Render("admin/services", fiber.Map{
		"title":     "Services",
		"processes": processes,
		"warning":   warning,
	})
}

// @Summary Get services data
// @Description Get services data for AJAX updates
// @Tags admin
// @Produce html
// @Success 200 {string} string "HTML content"
// @Failure 500 {object} map[string]string
// @Router /admin/services/data [get]
// getServicesData returns only the services fragment for AJAX updates
func (h *ServiceHandler) getServicesData(c *fiber.Ctx) error {
	// Get processes
	processes, _ := h.getProcessList()

	// No need to check for socket existence since we're using the process manager directly
	var warning string
	if h.pm == nil {
		warning = "Process manager is not properly initialized."
		h.logger.Printf("Warning: %s", warning)
	}

	// Return the fragment with process data and optional warning
	return c.Render("admin/services_fragment", fiber.Map{
		"processes": processes,
		"warning":   warning,
		"layout":    "",
	})
}
