package webhandlers

import (
	"github.com/gofiber/fiber/v2"
)

// UserData represents the user information to be passed to templates
type UserData struct {
	ID    int
	Name  string
	Email string
	Role  string
}

// RenderWithDefaults renders a template with default values common to all templates
func RenderWithDefaults(c *fiber.Ctx, templateName string, data fiber.Map) error {
	// Add default values
	if data == nil {
		data = fiber.Map{}
	}
	
	// Add a user value (would come from authentication in a real app)
	if _, exists := data["user"]; !exists {
		// Pass an empty user struct instead of nil
		// For demo purposes we'll create a mock user
		data["user"] = &UserData{
			ID:    1,
			Name:  "Demo User",
			Email: "demo@example.com",
			Role:  "Admin",
		}
	}
	
	// Custom CSS is now in stylesheets
	
	if _, exists := data["bodyClass"]; !exists {
		data["bodyClass"] = "default-body"
	}
	
	if _, exists := data["flashMessage"]; !exists {
		data["flashMessage"] = nil
	}

	return c.Render(templateName, data)
}

// GetCSRFToken returns a CSRF token for the current request
func GetCSRFToken(c *fiber.Ctx) string {
	// In a real application, this would generate a secure token
	// For demo purposes, we'll return a fixed string
	return "csrf-token-12345"
}
