package routes

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/client"
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
	pmClient *client.Client
	logger   *log.Logger
}

// default number of log lines to retrieve
const DefaultLogLines = 50

// NewServiceHandler creates a new service handler with the provided process manager client
func NewServiceHandler(pmClient *client.Client, logger *log.Logger) *ServiceHandler {
	return &ServiceHandler{
		pmClient: pmClient,
		logger:   logger,
	}
}

// RegisterRoutes registers service routes
func (h *ServiceHandler) RegisterRoutes(app *fiber.App) {
	services := app.Group("/admin/services")

	// Page routes
	services.Get("/", h.getServicesPage)

	// API routes
	services.Get("/data", h.getServicesData)
	services.Post("/start", h.startService)
	services.Post("/stop", h.stopService)
	services.Post("/restart", h.restartService)
	services.Post("/logs", h.getProcessLogs)
}

// getServicesPage renders the services page
func (h *ServiceHandler) getServicesPage(c *fiber.Ctx) error {
	// Get processes to display on the initial page load - this will never return an error now
	processes, _ := h.getProcessList()

	// Check if the process manager socket exists and add a warning if it doesn't
	var warning string
	if _, err := os.Stat(h.pmClient.GetSocketPath()); os.IsNotExist(err) {
		warning = "Process manager is not running. To start it, run: './start_process_manager.sh'"
		h.logger.Printf("Warning: %s", warning)
	} else {
		// Socket exists but might not be accepting connections
		err := h.pmClient.Connect()
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			warning = "Process manager socket exists but is not accepting connections. The process may have crashed. Try running './start_process_manager.sh'"
			h.logger.Printf("Warning: %s", warning)
		}
		if err == nil {
			h.pmClient.Close()
		}
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

	// Check if the process manager socket exists
	if _, err := os.Stat(h.pmClient.GetSocketPath()); os.IsNotExist(err) {
		// Return empty list instead of error when process manager is not running
		h.logger.Printf("Process manager socket not found at %s, returning empty list", h.pmClient.GetSocketPath())
		return []ProcessDisplayInfo{}, nil
	}

	// Connect to the process manager
	h.logger.Printf("Attempting to connect to process manager at %s", h.pmClient.GetSocketPath())
	err := h.pmClient.Connect()
	if err != nil {
		// Also return empty list for connection errors
		h.logger.Printf("Failed to connect to process manager: %v, returning empty list", err)
		return []ProcessDisplayInfo{}, nil
	}
	defer h.pmClient.Close()

	// Get the list of processes
	h.logger.Printf("Connected successfully, fetching process list")
	// Use text format since that's what the server is returning
	response, err := h.pmClient.ListProcesses("")
	if err != nil {
		// Also return empty list for command errors
		h.logger.Printf("Failed to list processes: %v, returning empty list", err)
		return []ProcessDisplayInfo{}, nil
	}

	// Debug: Log the raw response
	h.logger.Printf("Raw response from process manager: [%s]", response)
	
	// Process the text format response
	response = strings.TrimPrefix(response, "**RESULT**\n")
	response = strings.TrimSuffix(response, "\n**ENDRESULT**")
	
	// Process each line, which represents a process
	displayInfoList := []ProcessDisplayInfo{}
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Parse the line to extract process info
		// Format: "Name: name, Status: status, PID: pid, CPU: cpu%, Memory: memory MB"
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue // Skip lines that don't have enough parts
		}
		
		processInfo := ProcessDisplayInfo{}
		
		// Extract name
		namePart := strings.TrimSpace(parts[0])
		if strings.HasPrefix(namePart, "Name: ") {
			processInfo.Name = strings.TrimPrefix(namePart, "Name: ")
		}
		
		// Extract status
		statusPart := strings.TrimSpace(parts[1])
		if strings.HasPrefix(statusPart, "Status: ") {
			processInfo.Status = strings.TrimPrefix(statusPart, "Status: ")
		}
		
		// Extract PID
		pidPart := strings.TrimSpace(parts[2])
		if strings.HasPrefix(pidPart, "PID: ") {
			processInfo.ID = strings.TrimPrefix(pidPart, "PID: ")
		}
		
		// Extract CPU if available
		if len(parts) > 3 {
			cpuPart := strings.TrimSpace(parts[3])
			if strings.HasPrefix(cpuPart, "CPU: ") {
				processInfo.CPU = strings.TrimPrefix(cpuPart, "CPU: ")
			}
		}
		
		// Extract Memory if available
		if len(parts) > 4 {
			memoryPart := strings.TrimSpace(parts[4])
			if strings.HasPrefix(memoryPart, "Memory: ") {
				processInfo.Memory = strings.TrimPrefix(memoryPart, "Memory: ")
			}
		}
		
		// Calculate uptime (since we don't have real uptime info, we'll use a placeholder)
		processInfo.Uptime = "< 1 min" // Placeholder
		processInfo.StartTime = time.Now().Format("2006-01-02 15:04:05")
		
		displayInfoList = append(displayInfoList, processInfo)
	}

	// Debug: Log the number of processes parsed
	h.logger.Printf("Parsed %d processes", len(displayInfoList))

	return displayInfoList, nil
}

// formatUptime formats a duration as a human-readable uptime string
func formatUptime(duration time.Duration) string {
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else {
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// getServicesData returns only the services fragment for AJAX updates
func (h *ServiceHandler) getServicesData(c *fiber.Ctx) error {
	// Get processes - this will never return an error now
	processes, _ := h.getProcessList()

	// Check if the process manager socket exists and add a warning if it doesn't
	var warning string
	if _, err := os.Stat(h.pmClient.GetSocketPath()); os.IsNotExist(err) {
		warning = "Process manager is not running. To start it, run: './start_process_manager.sh'"
		h.logger.Printf("Warning: %s", warning)
	} else {
		// Socket exists but might not be accepting connections
		err := h.pmClient.Connect()
		if err != nil && strings.Contains(err.Error(), "connection refused") {
			warning = "Process manager socket exists but is not accepting connections. The process may have crashed. Try running './start_process_manager.sh'"
			h.logger.Printf("Warning: %s", warning)
		}
		if err == nil {
			h.pmClient.Close()
		}
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

	// Connect to the process manager
	err := h.pmClient.Connect()
	if err != nil {
		h.logger.Printf("Error connecting to process manager: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to process manager: " + err.Error(),
		})
	}
	defer h.pmClient.Close()

	// Start the process with logging enabled by default
	response, err := h.pmClient.StartProcess(name, command, true, 0, "", "")
	if err != nil {
		h.logger.Printf("Error starting process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Process started successfully",
		"response": response,
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

	// Connect to the process manager
	err := h.pmClient.Connect()
	if err != nil {
		h.logger.Printf("Error connecting to process manager: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to process manager: " + err.Error(),
		})
	}
	defer h.pmClient.Close()

	// Stop the process - the client method expects a name, not an ID
	h.logger.Printf("Stopping process with name: %s", name)
	response, err := h.pmClient.StopProcess(name)
	if err != nil {
		h.logger.Printf("Error stopping process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stop process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Process stopped successfully",
		"response": response,
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

	// Connect to the process manager
	err := h.pmClient.Connect()
	if err != nil {
		h.logger.Printf("Error connecting to process manager: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to process manager: " + err.Error(),
		})
	}
	defer h.pmClient.Close()

	// Restart the process - the client method expects a name, not an ID
	h.logger.Printf("Restarting process with name: %s", name) 
	response, err := h.pmClient.RestartProcess(name)
	if err != nil {
		h.logger.Printf("Error restarting process: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restart process: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"message":  "Process restarted successfully",
		"response": response,
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

	// Connect to the process manager
	err := h.pmClient.Connect()
	if err != nil {
		h.logger.Printf("Error connecting to process manager: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to connect to process manager: " + err.Error(),
		})
	}
	defer h.pmClient.Close()

	// Get the process logs
	h.logger.Printf("Getting logs for process: %s (lines: %d)", name, lines)
	logs, err := h.pmClient.GetProcessLogs(name, lines)
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
