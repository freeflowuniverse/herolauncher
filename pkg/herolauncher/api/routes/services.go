package routes

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

// ServiceHandler handles service-related routes
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

// RegisterRoutes registers service routes
func (h *ServiceHandler) RegisterRoutes(app *fiber.App) {
	services := app.Group("/admin/services")

	// Page routes
	services.Get("/", h.getServicesPage)

	// API routes
	services.Get("/data", h.getServicesData)

	// @Summary Get running services
	// @Description Get a list of all currently running services
	// @Tags services
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string][]ProcessDisplayInfo
	// @Failure 500 {object} map[string]string
	// @Router /admin/services/running [get]
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
	// @Router /admin/services/start [post]
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
	// @Router /admin/services/stop [post]
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
	// @Router /admin/services/restart [post]
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
	// @Router /admin/services/delete [post]
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
	// @Router /admin/services/logs [post]
	services.Post("/logs", h.getProcessLogs)
}

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

// startService starts a new service
func (h *ServiceHandler) startService(c *fiber.Ctx) error {
	name := c.FormValue("name")
	command := c.FormValue("command")

	if name == "" || command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and command are required",
		})
	}

	// Start the process with logging enabled by default
	err := h.pm.StartProcess(name, command, true, 0, "", "")
	if err != nil {
		h.logger.Printf("Error starting process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Process started successfully",
	})
}

// stopService stops a service
func (h *ServiceHandler) stopService(c *fiber.Ctx) error {
	// Get the process name from the form
	name := c.FormValue("name")

	if name == "" {
		// For backward compatibility, try ID field but use it as a name
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Process name is required",
			})
		}
	}

	// Stop the process directly using the process manager
	h.logger.Printf("Stopping process with name: %s", name)
	err := h.pm.StopProcess(name)
	if err != nil {
		h.logger.Printf("Error stopping process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stop process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Process stopped successfully",
	})
}

// restartService restarts a service
func (h *ServiceHandler) restartService(c *fiber.Ctx) error {
	// Get the process name from the form
	name := c.FormValue("name")

	if name == "" {
		// For backward compatibility, try ID field but use it as a name
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Process name is required",
			})
		}
	}

	// Restart the process directly using the process manager
	h.logger.Printf("Restarting process with name: %s", name) 
	err := h.pm.RestartProcess(name)
	if err != nil {
		h.logger.Printf("Error restarting process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restart process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Process restarted successfully",
	})
}

// deleteService deletes a service
func (h *ServiceHandler) deleteService(c *fiber.Ctx) error {
	// Get the service name from the form
	name := c.FormValue("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	// Debug: Log the delete request
	h.logger.Printf("Deleting process with name: %s", name)

	// Delete the service directly using the process manager
	err := h.pm.DeleteProcess(name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete service: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Service '%s' deleted successfully", name),
	})
}

// getRunningServices returns a list of running services in JSON format
func (h *ServiceHandler) getRunningServices(c *fiber.Ctx) error {
	// Get processes
	processes, err := h.getProcessList()
	if err != nil {
		h.logger.Printf("Error getting process list: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get process list: " + err.Error(),
		})
	}

	// Filter to only include running processes
	runningProcesses := make([]ProcessDisplayInfo, 0)
	for _, proc := range processes {
		if proc.Status == "running" {
			runningProcesses = append(runningProcesses, proc)
		}
	}

	return c.JSON(fiber.Map{
		"services": runningProcesses,
	})
}

// getProcessLogs retrieves logs for a specific process
func (h *ServiceHandler) getProcessLogs(c *fiber.Ctx) error {
	// Get the process name from the form
	name := c.FormValue("name")

	if name == "" {
		// For backward compatibility, try ID field but use it as a name
		name = c.FormValue("id")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Process name is required",
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

	// Get the process logs directly from the process manager
	h.logger.Printf("Getting logs for process: %s (lines: %d)", name, lines)
	logs, err := h.pm.GetProcessLogs(name, lines)
	if err != nil {
		h.logger.Printf("Error getting process logs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get process logs: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    logs,
	})
}
