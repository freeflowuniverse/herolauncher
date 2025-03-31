package routes

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/openrpc"
	"github.com/gofiber/fiber/v2"
)

// VFSHandler handles VFS-related routes
type VFSHandler struct {
	vfsClient *openrpc.Client
	logger    *log.Logger
}

// NewVFSHandler creates a new VFS handler
func NewVFSHandler(vfsClient *openrpc.Client, logger *log.Logger) *VFSHandler {
	return &VFSHandler{
		vfsClient: vfsClient,
		logger:    logger,
	}
}

// RegisterRoutes registers VFS routes
func (h *VFSHandler) RegisterRoutes(app *fiber.App) {
	// Create a group for VFS routes
	vfsGroup := app.Group("/api/vfs")

	// File operations
	vfsGroup.Post("/upload", h.uploadFile)
	vfsGroup.Get("/download/:vfspath", h.downloadFile)
	vfsGroup.Post("/upload-dir", h.uploadDir)
	vfsGroup.Post("/export-meta", h.exportMeta)
	vfsGroup.Post("/import-meta", h.importMeta)
	vfsGroup.Post("/export-dedupe", h.exportDedupe)
	vfsGroup.Post("/import-dedupe", h.importDedupe)
	vfsGroup.Post("/send", h.send)
	vfsGroup.Post("/send-exist", h.sendExist)
	vfsGroup.Post("/expose-webdav", h.exposeWebDAV)
	vfsGroup.Post("/expose-9p", h.expose9P)

	// OpenRPC discovery endpoint
	vfsGroup.Get("/openrpc", h.getOpenRPCSchema)
	
	// Logs endpoint
	vfsGroup.Get("/logs", h.getLogs)
}

// uploadFile handles file upload
func (h *VFSHandler) uploadFile(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	dedupePath := c.FormValue("dedupepath", "/")
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to get file: " + err.Error(),
		})
	}

	// Save the file temporarily
	tempFilePath := "/tmp/" + file.Filename
	if err := c.SaveFile(file, tempFilePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save uploaded file: " + err.Error(),
		})
	}

	// Upload file using VFS client
	result, err := h.vfsClient.UploadFile(vfsPath, dedupePath, tempFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload file: " + err.Error(),
		})
	}

	response := fiber.Map{
		"success":    result.Success,
		"message":    result.Message,
		"hash":       result.Hash,
		"vfspath":    vfsPath,
		"dedupepath": dedupePath,
		"filepath":   tempFilePath,
	}
	return c.JSON(response)
}

// downloadFile handles file download
func (h *VFSHandler) downloadFile(c *fiber.Ctx) error {
	vfsPath := c.Params("vfspath", "/")
	dedupePath := c.FormValue("dedupepath", "/")
	if vfsPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "VFS path parameter is required",
		})
	}

	// Create a temporary file to download to
	tempFilePath := "/tmp/download_" + filepath.Base(vfsPath)

	// Download file using VFS client
	_, err := h.vfsClient.DownloadFile(vfsPath, dedupePath, tempFilePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to download file: " + err.Error(),
		})
	}

	// Send the file
	// We don't need to use the result directly since the file is already downloaded to tempFilePath
	return c.SendFile(tempFilePath, true)
}

// listFiles handles directory listing
func (h *VFSHandler) listFiles(c *fiber.Ctx) error {
	// This functionality isn't directly available in the client
	// We would need to implement it using the available methods
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Directory listing not implemented in current VFS client",
		"message": "This functionality requires additional implementation",
	})
}

// uploadDir handles directory upload
func (h *VFSHandler) uploadDir(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	dedupePath := c.FormValue("dedupepath", "/")
	dirPath := c.FormValue("dirpath")
	if dirPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Directory path is required",
		})
	}

	// Upload directory using VFS client
	result, err := h.vfsClient.UploadDir(vfsPath, dedupePath, dirPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload directory: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"dedupepath": dedupePath,
		"dirpath": dirPath,
		"filesUploaded": result.FilesProcessed,
	})
}

// exportMeta handles metadata export
func (h *VFSHandler) exportMeta(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	destPath := c.FormValue("destpath")
	if destPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Destination path is required",
		})
	}

	// Export metadata using VFS client
	result, err := h.vfsClient.ExportMeta(vfsPath, destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export metadata: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"destpath": destPath,
		"metadataExported": result.Entries,
	})
}

// importMeta handles metadata import
func (h *VFSHandler) importMeta(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	sourcePath := c.FormValue("sourcepath")
	if sourcePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Source path is required",
		})
	}

	// Import metadata using VFS client
	result, err := h.vfsClient.ImportMeta(vfsPath, sourcePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to import metadata: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"sourcepath": sourcePath,
		"metadataImported": result.EntriesImported,
	})
}

// exportDedupe handles dedupe export
func (h *VFSHandler) exportDedupe(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	dedupePath := c.FormValue("dedupepath", "/")
	destPath := c.FormValue("destpath")
	if destPath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Destination path is required",
		})
	}

	// Export dedupe using VFS client
	result, err := h.vfsClient.ExportDedupe(vfsPath, dedupePath, destPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to export dedupe: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"dedupepath": dedupePath,
		"destpath": destPath,
		"hashesExported": result.HashesExported,
	})
}

// importDedupe handles dedupe import
func (h *VFSHandler) importDedupe(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	dedupePath := c.FormValue("dedupepath", "/")
	sourcePath := c.FormValue("sourcepath")
	if sourcePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Source path is required",
		})
	}

	// Import dedupe using VFS client
	result, err := h.vfsClient.ImportDedupe(vfsPath, dedupePath, sourcePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to import dedupe: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"dedupepath": dedupePath,
		"sourcepath": sourcePath,
		"hashesImported": result.HashesImported,
	})
}

// send handles sending files based on dedupe hashes
func (h *VFSHandler) send(c *fiber.Ctx) error {
	dedupePath := c.FormValue("dedupepath", "/")
	pubKeyDest := c.FormValue("pubkeydest")
	if pubKeyDest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Public key destination is required",
		})
	}

	// Parse hash list from request body
	type SendRequest struct {
		HashList []string `json:"hashlist"`
	}

	var req SendRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if len(req.HashList) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Hash list cannot be empty",
		})
	}

	// Send files using VFS client
	result, err := h.vfsClient.Send(dedupePath, pubKeyDest, req.HashList, h.vfsClient.Secret())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send files: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"dedupepath": dedupePath,
		"pubkeydest": pubKeyDest,
		"hashesSent": result.HashesSent,
	})
}

// sendExist checks which dedupe hashes exist
func (h *VFSHandler) sendExist(c *fiber.Ctx) error {
	dedupePath := c.FormValue("dedupepath", "/")
	pubKeyDest := c.FormValue("pubkeydest")
	if pubKeyDest == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Public key destination is required",
		})
	}

	// Parse hash list from request body
	type SendExistRequest struct {
		HashList []string `json:"hashlist"`
	}

	var req SendExistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body: " + err.Error(),
		})
	}

	if len(req.HashList) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Hash list cannot be empty",
		})
	}

	// Check existing hashes using VFS client
	result, err := h.vfsClient.SendExist(dedupePath, pubKeyDest, req.HashList, h.vfsClient.Secret())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing hashes: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"dedupepath": dedupePath,
		"pubkeydest": pubKeyDest,
		"existingHashes": result.ExistingHashes,
	})
}

// exposeWebDAV exposes the VFS via WebDAV
func (h *VFSHandler) exposeWebDAV(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	portStr := c.FormValue("port", "8080")
	username := c.FormValue("username", "")
	password := c.FormValue("password", "")

	// Convert port string to int
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid port number: " + err.Error(),
		})
	}

	// Expose WebDAV using VFS client
	result, err := h.vfsClient.ExposeWebDAV(vfsPath, port, username, password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to expose WebDAV: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"port": port,
		"url": result.URL,
	})
}

// expose9P exposes the VFS via 9P protocol
func (h *VFSHandler) expose9P(c *fiber.Ctx) error {
	vfsPath := c.FormValue("vfspath", "/")
	portStr := c.FormValue("port", "5640")
	readonlyStr := c.FormValue("readonly", "false")

	// Convert port string to int
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid port number: " + err.Error(),
		})
	}

	// Convert readonly string to bool
	readonly := false
	if readonlyStr == "true" {
		readonly = true
	}

	// Expose 9P using VFS client
	result, err := h.vfsClient.Expose9P(vfsPath, port, readonly)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to expose 9P: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"vfspath": vfsPath,
		"port": port,
		"readonly": readonly,
		"address": result.Address,
	})
}

// getOpenRPCSchema returns the OpenRPC schema for VFS
func (h *VFSHandler) getOpenRPCSchema(c *fiber.Ctx) error {
	// Create a schema based on the VFSManager interface
	schema := map[string]interface{}{
		"openrpc": "1.2.6",
		"info": map[string]interface{}{
			"title": "VFS API",
			"description": "Virtual File System API for HeroLauncher",
			"version": "1.0.0",
		},
		"methods": []map[string]interface{}{
			createMethodSchema("UploadFile", "Uploads a file to the virtual file system", 
				[]map[string]interface{}{
					{"name": "vfspath", "description": "Path in the virtual file system", "schema": map[string]string{"type": "string"}},
					{"name": "dedupepath", "description": "Path for deduplication", "schema": map[string]string{"type": "string"}},
					{"name": "filepath", "description": "Local file path to upload", "schema": map[string]string{"type": "string"}},
				}),
			createMethodSchema("UploadDir", "Uploads a directory to the virtual file system", 
				[]map[string]interface{}{
					{"name": "vfspath", "description": "Path in the virtual file system", "schema": map[string]string{"type": "string"}},
					{"name": "dedupepath", "description": "Path for deduplication", "schema": map[string]string{"type": "string"}},
					{"name": "dirpath", "description": "Local directory path to upload", "schema": map[string]string{"type": "string"}},
				}),
			createMethodSchema("DownloadFile", "Downloads a file from the virtual file system", 
				[]map[string]interface{}{
					{"name": "vfspath", "description": "Path in the virtual file system", "schema": map[string]string{"type": "string"}},
					{"name": "dedupepath", "description": "Path for deduplication", "schema": map[string]string{"type": "string"}},
					{"name": "destpath", "description": "Local destination path", "schema": map[string]string{"type": "string"}},
				}),
			// Add other methods here...
		},
	}

	// Set the content type to application/json
	c.Set("Content-Type", "application/json")
	return c.JSON(schema)
}

// getLogs returns logs of recent RPC calls to the VFS service
func (h *VFSHandler) getLogs(c *fiber.Ctx) error {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	method := c.Query("method", "")
	status := c.Query("status", "")
	
	// Get logs from VFS client
	logs, err := h.vfsClient.GetLogs(limit, method, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get logs: " + err.Error(),
		})
	}
	
	// Return logs as JSON
	return c.JSON(logs)
}

// Helper function to create method schema
func createMethodSchema(name, description string, params []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"description": description,
		"params": params,
		"result": map[string]interface{}{
			"name": "result",
			"description": "Result of the operation",
			"schema": map[string]interface{}{
				"type": "object",
			},
		},
	}
}
