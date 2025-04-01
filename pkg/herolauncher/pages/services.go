package pages

import (
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/gofiber/fiber/v2"
)

// ServiceHandler handles service-related page routes
type ServiceHandler struct {
	pm     *processmanager.ProcessManager
	logger *log.Logger
}

// NewServiceHandler creates a new service handler with the provided process manager
func NewServiceHandler(pm *processmanager.ProcessManager, logger *log.Logger) *ServiceHandler {
	return &ServiceHandler{
		pm:     pm,
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
