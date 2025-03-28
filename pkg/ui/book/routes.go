package book

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all the routes for the book server
func (w *WikiServer) SetupRoutes() {
	// Home route - redirects to the first item in the sidebar if available
	w.App.Get("/", func(c *fiber.Ctx) error {
		if len(w.Config.Sidebar) == 0 || len(w.Config.Sidebar[0].Items) == 0 {
			return c.Status(fiber.StatusNotFound).Render("layout", fiber.Map{
				"content":    "<p>No markdown files found in content directory.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    w.Config.Sidebar,
				"title":      w.Config.Title,
			})
		}
		return c.Redirect("/" + w.Config.Sidebar[0].Items[0].Href)
	})

	// Wildcard route for all markdown files
	w.App.Get("/*", func(c *fiber.Ctx) error {
		// Get the path from the URL parameters
		path := c.Params("*") + ".md"
		if path == ".md" {
			return c.Redirect("/")
		}

		// Ensure path starts with a forward slash
		if path[0] != '/' {
			path = "/" + path
		}

		// Check if the file exists
		if !w.VFS.Exists(path) {
			return c.Status(fiber.StatusNotFound).Render("layout", fiber.Map{
				"content":    "<p>Markdown file not found.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    w.Config.Sidebar,
				"title":      w.Config.Title,
			})
		}

		// Read the file content
		data, err := w.VFS.FileRead(path)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).Render("layout", fiber.Map{
				"content":    "<p>Error rendering markdown content.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    w.Config.Sidebar,
				"title":      w.Config.Title,
			})
		}

		// Render the markdown content
		content, anchors, err := renderMarkdown(string(data))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).Render("layout", fiber.Map{
				"content":    "<p>Error rendering markdown content.</p>",
				"activeFile": nil,
				"safe":       true,
				"sidebar":    w.Config.Sidebar,
				"title":      w.Config.Title,
			})
		}

		// Get the active file information
		activeFile := MarkdownFile{
			Path:  path,
			Name:  path[1 : len(path)-3], // Remove leading slash and .md extension
			Title: getTitle(string(data)),
		}

		// Render the layout template with the markdown content
		return c.Render("layout", fiber.Map{
			"content":    content,
			"anchors":    anchors,
			"activeFile": activeFile,
			"safe":       true,
			"sidebar":    w.Config.Sidebar,
			"title":      w.Config.Title,
		})
	})
}

// Note: renderMarkdown and getTitle functions are defined in utils.go
