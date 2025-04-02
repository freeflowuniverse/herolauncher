package handlers

import (
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces/openrpc"
	"github.com/gofiber/fiber/v2"
)

// ServiceHandler handles service-related routes
type ServiceHandler struct {
	client *openrpc.Client
}

// NewServiceHandler creates a new ServiceHandler
func NewServiceHandler(socketPath, secret string) *ServiceHandler {
	fmt.Printf("DEBUG: Creating new ServiceHandler with socket path: %s and secret: %s\n", socketPath, secret)
	return &ServiceHandler{
		client: openrpc.NewClient(socketPath, secret),
	}
}

// GetServices renders the services page
func (h *ServiceHandler) GetServices(c *fiber.Ctx) error {
	return c.Render("admin/services", fiber.Map{
		"title":   "Services",
		"error":   c.Query("error", ""),
		"warning": c.Query("warning", ""),
	})
}

// GetServicesFragment returns the services table fragment for Unpoly updates
func (h *ServiceHandler) GetServicesFragment(c *fiber.Ctx) error {
	processes, err := h.getProcessList()
	if err != nil {
		return c.Render("admin/services_fragment", fiber.Map{
			"error": fmt.Sprintf("Failed to fetch services: %v", err),
		})
	}

	return c.Render("admin/services_fragment", fiber.Map{
		"processes": processes,
	})
}

// StartService handles the request to start a new service
func (h *ServiceHandler) StartService(c *fiber.Ctx) error {
	name := c.FormValue("name")
	command := c.FormValue("command")

	if name == "" || command == "" {
		return c.JSON(fiber.Map{
			"error": "Service name and command are required",
		})
	}

	// Default to enabling logs
	logEnabled := true

	// Start the process with no deadline, no cron, and no job ID
	fmt.Printf("DEBUG: StartService called for '%s' using client: %p\n", name, h.client)
	result, err := h.client.StartProcess(name, command, logEnabled, 0, "", "")
	if err != nil {
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to start service: %v", err),
		})
	}

	if !result.Success {
		return c.JSON(fiber.Map{
			"error": result.Message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": result.Message,
		"pid":     result.PID,
	})
}

// StopService handles the request to stop a service
func (h *ServiceHandler) StopService(c *fiber.Ctx) error {
	name := c.FormValue("name")

	if name == "" {
		return c.JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	result, err := h.client.StopProcess(name)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to stop service: %v", err),
		})
	}

	if !result.Success {
		return c.JSON(fiber.Map{
			"error": result.Message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": result.Message,
	})
}

// RestartService handles the request to restart a service
func (h *ServiceHandler) RestartService(c *fiber.Ctx) error {
	name := c.FormValue("name")

	if name == "" {
		return c.JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	result, err := h.client.RestartProcess(name)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to restart service: %v", err),
		})
	}

	if !result.Success {
		return c.JSON(fiber.Map{
			"error": result.Message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": result.Message,
		"pid":     result.PID,
	})
}

// DeleteService handles the request to delete a service
func (h *ServiceHandler) DeleteService(c *fiber.Ctx) error {
	name := c.FormValue("name")

	if name == "" {
		return c.JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	result, err := h.client.DeleteProcess(name)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete service: %v", err),
		})
	}

	if !result.Success {
		return c.JSON(fiber.Map{
			"error": result.Message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": result.Message,
	})
}

// GetServiceLogs handles the request to get logs for a service
func (h *ServiceHandler) GetServiceLogs(c *fiber.Ctx) error {
	name := c.Query("name")
	lines := c.QueryInt("lines", 100)

	fmt.Printf("DEBUG: GetServiceLogs called for service '%s' using client: %p\n", name, h.client)

	if name == "" {
		return c.JSON(fiber.Map{
			"error": "Service name is required",
		})
	}

	// Debug: List all processes before getting logs
	processes, listErr := h.getProcessList()
	if listErr == nil {
		fmt.Println("DEBUG: Current processes in service handler:")
		for _, proc := range processes {
			fmt.Printf("DEBUG: - '%v' (PID: %v, Status: %v)\n", proc["Name"], proc["ID"], proc["Status"])
		}
	} else {
		fmt.Printf("DEBUG: Error listing processes: %v\n", listErr)
	}

	result, err := h.client.GetProcessLogs(name, lines)
	if err != nil {
		return c.JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get service logs: %v", err),
		})
	}

	if !result.Success {
		return c.JSON(fiber.Map{
			"error": result.Message,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"logs":    result.Logs,
	})
}

// Helper function to get the list of processes and format them for the UI
func (h *ServiceHandler) getProcessList() ([]fiber.Map, error) {
	// Get the list of processes
	result, err := h.client.ListProcesses("json")
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	// Convert the result to a slice of ProcessStatus
	processList, ok := result.([]interfaces.ProcessStatus)
	if !ok {
		return nil, fmt.Errorf("unexpected result type from ListProcesses")
	}

	// Format the processes for the UI
	formattedProcesses := make([]fiber.Map, 0, len(processList))
	for _, proc := range processList {
		// Calculate uptime
		uptime := "N/A"
		if proc.Status == "running" {
			duration := time.Since(proc.StartTime)
			if duration.Hours() >= 24 {
				days := int(duration.Hours() / 24)
				hours := int(duration.Hours()) % 24
				uptime = fmt.Sprintf("%dd %dh", days, hours)
			} else if duration.Hours() >= 1 {
				hours := int(duration.Hours())
				minutes := int(duration.Minutes()) % 60
				uptime = fmt.Sprintf("%dh %dm", hours, minutes)
			} else {
				minutes := int(duration.Minutes())
				seconds := int(duration.Seconds()) % 60
				uptime = fmt.Sprintf("%dm %ds", minutes, seconds)
			}
		}

		// Format CPU and memory usage
		cpuUsage := fmt.Sprintf("%.1f%%", proc.CPUPercent)
		memoryUsage := fmt.Sprintf("%.1f MB", proc.MemoryMB)

		formattedProcesses = append(formattedProcesses, fiber.Map{
			"Name":   proc.Name,
			"Status": string(proc.Status),
			"ID":     proc.PID,
			"CPU":    cpuUsage,
			"Memory": memoryUsage,
			"Uptime": uptime,
		})
	}

	return formattedProcesses, nil
}
