package pages

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces/openrpc"
	"github.com/gofiber/fiber/v2"
)

// ServiceHandler handles service-related page routes
type ServiceHandler struct {
	client  *openrpc.Client
	logger  *log.Logger
}

// NewServiceHandler creates a new service handler with the provided socket path and secret
func NewServiceHandler(socketPath, secret string, logger *log.Logger) *ServiceHandler {
	fmt.Printf("DEBUG: Creating new pages.ServiceHandler with socket path: %s and secret: %s\n", socketPath, secret)
	return &ServiceHandler{
		client: openrpc.NewClient(socketPath, secret),
		logger: logger,
	}
}

// RegisterRoutes registers service page routes
func (h *ServiceHandler) RegisterRoutes(app *fiber.App) {
	services := app.Group("/services")

	// Page routes
	services.Get("/", h.getServicesPage)
	services.Get("/data", h.getServicesData)
}

// getServicesPage renders the services page
func (h *ServiceHandler) getServicesPage(c *fiber.Ctx) error {
	// Get processes to display on the initial page load
	processes, _ := h.getProcessList()

	// Check if we can connect to the process manager
	var warning string
	_, err := h.client.ListProcesses("json")
	if err != nil {
		warning = "Could not connect to process manager: " + err.Error()
		h.logger.Printf("Warning: %s", warning)
	}

	return c.Render("admin/services", fiber.Map{
		"title":     "Services",
		"processes": processes,
		"warning":   warning,
	})
}

// getServicesData returns only the services fragment for AJAX updates
func (h *ServiceHandler) getServicesData(c *fiber.Ctx) error {
	// Get processes
	processes, _ := h.getProcessList()

	// Render only the services fragment
	return c.Render("admin/services_fragment", fiber.Map{
		"processes": processes,
	})
}

// getProcessList gets a list of processes from the process manager
func (h *ServiceHandler) getProcessList() ([]ProcessDisplayInfo, error) {
	// Debug: Log the function entry
	h.logger.Printf("Entering getProcessList() function")
	fmt.Printf("DEBUG: getProcessList called using client: %p\n", h.client)

	// Get the list of processes via the client
	result, err := h.client.ListProcesses("json")
	if err != nil {
		h.logger.Printf("Error listing processes: %v", err)
		return nil, err
	}

	// Convert the result to a slice of ProcessStatus
	listResult, ok := result.([]interface{})
	if !ok {
		h.logger.Printf("Error: unexpected result type from ListProcesses")
		return nil, fmt.Errorf("unexpected result type from ListProcesses")
	}

	// Convert to display info format
	displayInfoList := make([]ProcessDisplayInfo, 0, len(listResult))
	for _, item := range listResult {
		procMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Create a ProcessDisplayInfo from the map
		displayInfo := ProcessDisplayInfo{
			ID:        fmt.Sprintf("%v", procMap["pid"]),
			Name:      fmt.Sprintf("%v", procMap["name"]),
			Status:    fmt.Sprintf("%v", procMap["status"]),
			Uptime:    fmt.Sprintf("%v", procMap["uptime"]),
			StartTime: fmt.Sprintf("%v", procMap["start_time"]),
			CPU:       fmt.Sprintf("%v%%", procMap["cpu"]),
			Memory:    fmt.Sprintf("%v MB", procMap["memory"]),
		}
		displayInfoList = append(displayInfoList, displayInfo)
	}

	// Debug: Log the number of processes
	h.logger.Printf("Found %d processes", len(displayInfoList))

	return displayInfoList, nil
}
