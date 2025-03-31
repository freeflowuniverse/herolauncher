package routes

import (
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/gofiber/fiber/v2"
)

// JetTemplateRequest represents the request body for the checkjet endpoint
type JetTemplateRequest struct {
	Template string `json:"template"`
}

// JetTemplateResponse represents the response for the checkjet endpoint
type JetTemplateResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// JetHandler handles Jet template-related API endpoints
type JetHandler struct {
	// No dependencies needed for this handler
}

// NewJetHandler creates a new Jet template handler
func NewJetHandler() *JetHandler {
	return &JetHandler{}
}

// RegisterRoutes registers Jet template routes to the fiber app
func (h *JetHandler) RegisterRoutes(app *fiber.App) {
	// Register the checkjet endpoint directly at the root level
	app.Post("/checkjet", h.validateTemplate)
}

// @Summary Validate a Jet template
// @Description Validates a Jet template and returns detailed error information if invalid
// @Tags jet
// @Accept json
// @Produce json
// @Param template body JetTemplateRequest true "Jet template to validate"
// @Success 200 {object} JetTemplateResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/jet/checkjet [post]
func (h *JetHandler) validateTemplate(c *fiber.Ctx) error {
	var req JetTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	if req.Template == "" {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Template cannot be empty",
		})
	}

	// Create a temporary in-memory loader for the template
	loader := jet.NewInMemLoader()

	// Add the template to the loader
	loader.Set("test.jet", req.Template)

	// Create a new Jet set with the loader and enable development mode for better error reporting
	set := jet.NewSet(loader, jet.InDevelopmentMode())

	// Get the template to parse it
	_, err := set.GetTemplate("test.jet")

	// Check if the template is valid
	if err != nil {
		// Extract meaningful error information
		errMsg := err.Error()

		// Ignore errors related to extended or included files not found
		// These aren't syntax errors but dependency errors we want to ignore
		if strings.Contains(errMsg, "no template") ||
			strings.Contains(errMsg, "unable to locate template") ||
			strings.Contains(errMsg, "template not found") ||
			strings.Contains(errMsg, "extends|import") ||
			strings.Contains(errMsg, "could not be found") ||
			strings.Contains(errMsg, "template /") {
			// Still valid since it's only a dependency error, not a syntax error
			return c.JSON(JetTemplateResponse{
				Valid:   true,
				Message: "Template syntax is valid (ignoring extends/include errors)",
			})
		}

		return c.JSON(JetTemplateResponse{
			Valid: false,
			Error: errMsg,
		})
	}

	// If no error, the template is valid
	return c.JSON(JetTemplateResponse{
		Valid:   true,
		Message: "Template is valid",
	})
}
