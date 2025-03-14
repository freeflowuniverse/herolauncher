package routes

import (
	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin-related routes
type AdminHandler struct{}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// RegisterRoutes registers all admin routes
func (h *AdminHandler) RegisterRoutes(app *fiber.App) {
	// Admin routes
	admin := app.Group("/admin")

	// Dashboard
	admin.Get("/", h.getDashboard)
	
	// Services
	admin.Get("/services", h.getServices)
	
	// Packages
	admin.Get("/packages", h.getPackages)
	
	// System routes
	admin.Get("/system/info", h.getSystemInfo)
	admin.Get("/system/logs", h.getSystemLogs)
	admin.Get("/system/settings", h.getSystemSettings)
	
	// Redirect root to admin
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/admin")
	})
}

// getDashboard renders the admin dashboard
func (h *AdminHandler) getDashboard(c *fiber.Ctx) error {
	return c.Render("admin/index", fiber.Map{
		"title": "Dashboard",
	})
}

// getServices renders the services page
func (h *AdminHandler) getServices(c *fiber.Ctx) error {
	return c.Render("admin/services", fiber.Map{
		"title": "Services",
	})
}

// getPackages renders the packages page
func (h *AdminHandler) getPackages(c *fiber.Ctx) error {
	return c.Render("admin/packages", fiber.Map{
		"title": "Packages",
	})
}

// getSystemInfo renders the system info page
func (h *AdminHandler) getSystemInfo(c *fiber.Ctx) error {
	// Get system information
	systemInfo := fiber.Map{
		"hostname": "herolauncher-server",
		"os": "Linux",
		"version": "1.0.0",
		"uptime": "2 days, 3 hours",
		"cpu_usage": 24.5,
		"memory_usage": 1.2,
		"memory_total": 8.0,
	}

	return c.Render("admin/system/info", fiber.Map{
		"title": "System Info",
		"system": systemInfo,
	})
}

// getSystemLogs renders the system logs page
func (h *AdminHandler) getSystemLogs(c *fiber.Ctx) error {
	// Get recent logs
	logs := []fiber.Map{
		{"timestamp": "2025-03-14T06:30:00Z", "level": "info", "message": "System started"},
		{"timestamp": "2025-03-14T06:35:12Z", "level": "info", "message": "Service 'redis' started"},
		{"timestamp": "2025-03-14T07:15:45Z", "level": "warning", "message": "High memory usage detected"},
		{"timestamp": "2025-03-14T07:25:30Z", "level": "info", "message": "Package 'web-ui' updated"},
	}

	return c.Render("admin/system/logs", fiber.Map{
		"title": "System Logs",
		"logs": logs,
	})
}

// getSystemSettings renders the system settings page
func (h *AdminHandler) getSystemSettings(c *fiber.Ctx) error {
	// Get current settings
	settings := fiber.Map{
		"debug_mode": true,
		"auto_update": false,
		"backup_enabled": true,
		"backup_interval": "daily",
		"log_level": "info",
	}

	return c.Render("admin/system/settings", fiber.Map{
		"title": "System Settings",
		"settings": settings,
	})
}
