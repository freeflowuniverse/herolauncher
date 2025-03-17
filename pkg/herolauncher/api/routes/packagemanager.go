package routes

import (
	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/pkg/packagemanager"
	"github.com/gofiber/fiber/v2"
)

// PackageManagerHandler handles package manager API endpoints
type PackageManagerHandler struct {
	packageManager *packagemanager.PackageManager
}

// NewPackageManagerHandler creates a new package manager handler
func NewPackageManagerHandler(pm *packagemanager.PackageManager) *PackageManagerHandler {
	return &PackageManagerHandler{
		packageManager: pm,
	}
}

// RegisterRoutes registers package manager routes to the fiber app
func (h *PackageManagerHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/packages")

	group.Post("/install", h.installPackage)
	group.Post("/uninstall", h.uninstallPackage)
	group.Get("/list", h.listPackages)
	group.Get("/search/:query", h.searchPackages)
}

// @Summary Install a package
// @Description Install a package using the appropriate package manager
// @Tags packages
// @Accept json
// @Produce json
// @Param package body api.InstallPackageRequest true "Package to install"
// @Success 200 {object} api.InstallPackageResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/packages/install [post]
func (h *PackageManagerHandler) installPackage(c *fiber.Ctx) error {
	var req api.InstallPackageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	output, err := h.packageManager.InstallPackage(req.PackageName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: "Failed to install package: " + err.Error(),
		})
	}

	return c.JSON(api.InstallPackageResponse{
		Success: true,
		Output:  output,
	})
}

// @Summary Uninstall a package
// @Description Uninstall a package using the appropriate package manager
// @Tags packages
// @Accept json
// @Produce json
// @Param package body api.UninstallPackageRequest true "Package to uninstall"
// @Success 200 {object} api.UninstallPackageResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/packages/uninstall [post]
func (h *PackageManagerHandler) uninstallPackage(c *fiber.Ctx) error {
	var req api.UninstallPackageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	output, err := h.packageManager.UninstallPackage(req.PackageName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: "Failed to uninstall package: " + err.Error(),
		})
	}

	return c.JSON(api.UninstallPackageResponse{
		Success: true,
		Output:  output,
	})
}

// @Summary List installed packages
// @Description Get a list of all installed packages
// @Tags packages
// @Produce json
// @Success 200 {object} api.ListPackagesResponse
// @Router /api/packages/list [get]
func (h *PackageManagerHandler) listPackages(c *fiber.Ctx) error {
	packages, err := h.packageManager.ListInstalledPackages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: "Failed to list packages: " + err.Error(),
		})
	}

	return c.JSON(api.ListPackagesResponse{
		Packages: packages,
	})
}

// @Summary Search for packages
// @Description Search for packages using the appropriate package manager
// @Tags packages
// @Produce json
// @Param query path string true "Search query"
// @Success 200 {object} api.SearchPackagesResponse
// @Router /api/packages/search/{query} [get]
func (h *PackageManagerHandler) searchPackages(c *fiber.Ctx) error {
	query := c.Params("query")
	
	packages, err := h.packageManager.SearchPackage(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.ErrorResponse{
			Error: "Failed to search packages: " + err.Error(),
		})
	}

	return c.JSON(api.SearchPackagesResponse{
		Packages: packages,
	})
}
