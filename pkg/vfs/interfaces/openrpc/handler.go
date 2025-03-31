package openrpc

import (
	"encoding/json"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
)

// Handler implements the OpenRPC handlers for VFS operations
type Handler struct {
	vfsManager interfaces.VFSManager
	secret     string
}

// NewHandler creates a new RPC handler for VFS operations
func NewHandler(vfsManager interfaces.VFSManager, secret string) *Handler {
	return &Handler{
		vfsManager: vfsManager,
		secret:     secret,
	}
}

// GetHandlers returns a map of RPC handlers for the OpenRPC manager
func (h *Handler) GetHandlers() map[string]openrpcmanager.RPCHandler {
	return map[string]openrpcmanager.RPCHandler{
		"vfs.upload_file":    h.handleUploadFile,
		"vfs.upload_dir":     h.handleUploadDir,
		"vfs.download_file":  h.handleDownloadFile,
		"vfs.export_meta":    h.handleExportMeta,
		"vfs.import_meta":    h.handleImportMeta,
		"vfs.export_dedupe":  h.handleExportDedupe,
		"vfs.import_dedupe":  h.handleImportDedupe,
		"vfs.send":           h.handleSend,
		"vfs.send_exist":     h.handleSendExist,
		"vfs.expose_webdav":  h.handleExposeWebDAV,
		"vfs.expose_9p":      h.handleExpose9P,
	}
}

// handleUploadFile handles the vfs.upload_file method
func (h *Handler) handleUploadFile(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		DedupePath string `json:"dedupepath"`
		FilePath   string `json:"filepath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.UploadFile(request.VFSPath, request.DedupePath, request.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	return result, nil
}

// handleUploadDir handles the vfs.upload_dir method
func (h *Handler) handleUploadDir(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		DedupePath string `json:"dedupepath"`
		DirPath    string `json:"dirpath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.UploadDir(request.VFSPath, request.DedupePath, request.DirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to upload directory: %w", err)
	}
	return result, nil
}

// handleDownloadFile handles the vfs.download_file method
func (h *Handler) handleDownloadFile(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		DedupePath string `json:"dedupepath"`
		DestPath   string `json:"destpath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.DownloadFile(request.VFSPath, request.DedupePath, request.DestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return result, nil
}

// handleExportMeta handles the vfs.export_meta method
func (h *Handler) handleExportMeta(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath  string `json:"vfspath"`
		DestPath string `json:"destpath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.ExportMeta(request.VFSPath, request.DestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export metadata: %w", err)
	}
	return result, nil
}

// handleImportMeta handles the vfs.import_meta method
func (h *Handler) handleImportMeta(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		SourcePath string `json:"sourcepath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.ImportMeta(request.VFSPath, request.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import metadata: %w", err)
	}
	return result, nil
}

// handleExportDedupe handles the vfs.export_dedupe method
func (h *Handler) handleExportDedupe(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		DedupePath string `json:"dedupepath"`
		DestPath   string `json:"destpath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.ExportDedupe(request.VFSPath, request.DedupePath, request.DestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to export dedupe information: %w", err)
	}
	return result, nil
}

// handleImportDedupe handles the vfs.import_dedupe method
func (h *Handler) handleImportDedupe(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath    string `json:"vfspath"`
		DedupePath string `json:"dedupepath"`
		SourcePath string `json:"sourcepath"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.ImportDedupe(request.VFSPath, request.DedupePath, request.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to import dedupe information: %w", err)
	}
	return result, nil
}

// handleSend handles the vfs.send method
func (h *Handler) handleSend(params json.RawMessage) (interface{}, error) {
	var request struct {
		DedupePath string   `json:"dedupepath"`
		PubKeyDest string   `json:"pubkeydest"`
		HashList   []string `json:"hashlist"`
		Secret     string   `json:"secret"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Verify secret
	if request.Secret != h.secret {
		return nil, fmt.Errorf("authentication failed: invalid secret")
	}

	result, err := h.vfsManager.Send(request.DedupePath, request.PubKeyDest, request.HashList, request.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to send files: %w", err)
	}
	return result, nil
}

// handleSendExist handles the vfs.send_exist method
func (h *Handler) handleSendExist(params json.RawMessage) (interface{}, error) {
	var request struct {
		DedupePath string   `json:"dedupepath"`
		PubKeyDest string   `json:"pubkeydest"`
		HashList   []string `json:"hashlist"`
		Secret     string   `json:"secret"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Verify secret
	if request.Secret != h.secret {
		return nil, fmt.Errorf("authentication failed: invalid secret")
	}

	result, err := h.vfsManager.SendExist(request.DedupePath, request.PubKeyDest, request.HashList, request.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing hashes: %w", err)
	}
	return result, nil
}

// handleExposeWebDAV handles the vfs.expose_webdav method
func (h *Handler) handleExposeWebDAV(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath  string `json:"vfspath"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.ExposeWebDAV(request.VFSPath, request.Port, request.Username, request.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to expose WebDAV: %w", err)
	}
	return result, nil
}

// handleExpose9P handles the vfs.expose_9p method
func (h *Handler) handleExpose9P(params json.RawMessage) (interface{}, error) {
	var request struct {
		VFSPath  string `json:"vfspath"`
		Port     int    `json:"port"`
		ReadOnly bool   `json:"readonly"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	result, err := h.vfsManager.Expose9P(request.VFSPath, request.Port, request.ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to expose 9P: %w", err)
	}
	return result, nil
}
