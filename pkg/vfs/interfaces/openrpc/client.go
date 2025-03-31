package openrpc

import (
	"encoding/json"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager/client"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
)

// Client provides a client for interacting with VFS operations via RPC
type Client struct {
	client.BaseClient
	secret string
}

// NewClient creates a new client for the VFS API
func NewClient(socketPath, secret string) *Client {
	return &Client{
		BaseClient: *client.NewClient(socketPath, secret),
		secret:     secret,
	}
}

// Secret returns the client's secret
func (c *Client) Secret() string {
	return c.secret
}

// UploadFile uploads a file to the virtual file system
func (c *Client) UploadFile(vfspath, dedupepath, filepath string) (interfaces.UploadResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"dedupepath": dedupepath,
		"filepath":   filepath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.UploadResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.upload_file", paramsJSON, "")
	if err != nil {
		return interfaces.UploadResult{}, fmt.Errorf("failed to upload file: %w", err)
	}

	// Convert result to UploadResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.UploadResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var uploadResult interfaces.UploadResult
	if err := json.Unmarshal(resultJSON, &uploadResult); err != nil {
		return interfaces.UploadResult{}, fmt.Errorf("failed to unmarshal upload result: %w", err)
	}

	return uploadResult, nil
}

// UploadDir uploads a directory to the virtual file system
func (c *Client) UploadDir(vfspath, dedupepath, dirpath string) (interfaces.UploadDirResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"dedupepath": dedupepath,
		"dirpath":    dirpath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.UploadDirResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.upload_dir", paramsJSON, "")
	if err != nil {
		return interfaces.UploadDirResult{}, fmt.Errorf("failed to upload directory: %w", err)
	}

	// Convert result to UploadDirResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.UploadDirResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var uploadDirResult interfaces.UploadDirResult
	if err := json.Unmarshal(resultJSON, &uploadDirResult); err != nil {
		return interfaces.UploadDirResult{}, fmt.Errorf("failed to unmarshal upload directory result: %w", err)
	}

	return uploadDirResult, nil
}

// DownloadFile downloads a file from the virtual file system
func (c *Client) DownloadFile(vfspath, dedupepath, destpath string) (interfaces.DownloadResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"dedupepath": dedupepath,
		"destpath":   destpath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.DownloadResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.download_file", paramsJSON, "")
	if err != nil {
		return interfaces.DownloadResult{}, fmt.Errorf("failed to download file: %w", err)
	}

	// Convert result to DownloadResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.DownloadResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var downloadResult interfaces.DownloadResult
	if err := json.Unmarshal(resultJSON, &downloadResult); err != nil {
		return interfaces.DownloadResult{}, fmt.Errorf("failed to unmarshal download result: %w", err)
	}

	return downloadResult, nil
}

// ExportMeta exports metadata from the virtual file system
func (c *Client) ExportMeta(vfspath, destpath string) (interfaces.ExportMetaResult, error) {
	params := map[string]string{
		"vfspath":  vfspath,
		"destpath": destpath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ExportMetaResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.export_meta", paramsJSON, "")
	if err != nil {
		return interfaces.ExportMetaResult{}, fmt.Errorf("failed to export metadata: %w", err)
	}

	// Convert result to ExportMetaResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ExportMetaResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var exportMetaResult interfaces.ExportMetaResult
	if err := json.Unmarshal(resultJSON, &exportMetaResult); err != nil {
		return interfaces.ExportMetaResult{}, fmt.Errorf("failed to unmarshal export metadata result: %w", err)
	}

	return exportMetaResult, nil
}

// ImportMeta imports metadata to the virtual file system
func (c *Client) ImportMeta(vfspath, sourcepath string) (interfaces.ImportMetaResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"sourcepath": sourcepath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ImportMetaResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.import_meta", paramsJSON, "")
	if err != nil {
		return interfaces.ImportMetaResult{}, fmt.Errorf("failed to import metadata: %w", err)
	}

	// Convert result to ImportMetaResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ImportMetaResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var importMetaResult interfaces.ImportMetaResult
	if err := json.Unmarshal(resultJSON, &importMetaResult); err != nil {
		return interfaces.ImportMetaResult{}, fmt.Errorf("failed to unmarshal import metadata result: %w", err)
	}

	return importMetaResult, nil
}

// ExportDedupe exports dedupe information from the virtual file system
func (c *Client) ExportDedupe(vfspath, dedupepath, destpath string) (interfaces.ExportDedupeResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"dedupepath": dedupepath,
		"destpath":   destpath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ExportDedupeResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.export_dedupe", paramsJSON, "")
	if err != nil {
		return interfaces.ExportDedupeResult{}, fmt.Errorf("failed to export dedupe information: %w", err)
	}

	// Convert result to ExportDedupeResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ExportDedupeResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var exportDedupeResult interfaces.ExportDedupeResult
	if err := json.Unmarshal(resultJSON, &exportDedupeResult); err != nil {
		return interfaces.ExportDedupeResult{}, fmt.Errorf("failed to unmarshal export dedupe result: %w", err)
	}

	return exportDedupeResult, nil
}

// ImportDedupe imports dedupe information to the virtual file system
func (c *Client) ImportDedupe(vfspath, dedupepath, sourcepath string) (interfaces.ImportDedupeResult, error) {
	params := map[string]string{
		"vfspath":    vfspath,
		"dedupepath": dedupepath,
		"sourcepath": sourcepath,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ImportDedupeResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.import_dedupe", paramsJSON, "")
	if err != nil {
		return interfaces.ImportDedupeResult{}, fmt.Errorf("failed to import dedupe information: %w", err)
	}

	// Convert result to ImportDedupeResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ImportDedupeResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var importDedupeResult interfaces.ImportDedupeResult
	if err := json.Unmarshal(resultJSON, &importDedupeResult); err != nil {
		return interfaces.ImportDedupeResult{}, fmt.Errorf("failed to unmarshal import dedupe result: %w", err)
	}

	return importDedupeResult, nil
}

// Send sends files based on dedupe hashes to a destination
func (c *Client) Send(dedupepath, pubkeydest string, hashlist []string, secret string) (interfaces.SendResult, error) {
	params := map[string]interface{}{
		"dedupepath": dedupepath,
		"pubkeydest": pubkeydest,
		"hashlist":   hashlist,
		"secret":     secret, // Use the provided secret for authentication
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.SendResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.send", paramsJSON, c.secret)
	if err != nil {
		return interfaces.SendResult{}, fmt.Errorf("failed to send files: %w", err)
	}

	// Convert result to SendResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.SendResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var sendResult interfaces.SendResult
	if err := json.Unmarshal(resultJSON, &sendResult); err != nil {
		return interfaces.SendResult{}, fmt.Errorf("failed to unmarshal send result: %w", err)
	}

	return sendResult, nil
}

// SendExist checks which dedupe hashes exist and returns a list
func (c *Client) SendExist(dedupepath, pubkeydest string, hashlist []string, secret string) (interfaces.SendExistResult, error) {
	params := map[string]interface{}{
		"dedupepath": dedupepath,
		"pubkeydest": pubkeydest,
		"hashlist":   hashlist,
		"secret":     secret, // Use the provided secret for authentication
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.SendExistResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.send_exist", paramsJSON, c.secret)
	if err != nil {
		return interfaces.SendExistResult{}, fmt.Errorf("failed to check existing hashes: %w", err)
	}

	// Convert result to SendExistResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.SendExistResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var sendExistResult interfaces.SendExistResult
	if err := json.Unmarshal(resultJSON, &sendExistResult); err != nil {
		return interfaces.SendExistResult{}, fmt.Errorf("failed to unmarshal send exist result: %w", err)
	}

	return sendExistResult, nil
}

// ExposeWebDAV exposes the virtual file system via WebDAV
func (c *Client) ExposeWebDAV(vfspath string, port int, username, password string) (interfaces.ExposeWebDAVResult, error) {
	params := map[string]interface{}{
		"vfspath":  vfspath,
		"port":     port,
		"username": username,
		"password": password,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ExposeWebDAVResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.expose_webdav", paramsJSON, "")
	if err != nil {
		return interfaces.ExposeWebDAVResult{}, fmt.Errorf("failed to expose WebDAV: %w", err)
	}

	// Convert result to ExposeWebDAVResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ExposeWebDAVResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var exposeWebDAVResult interfaces.ExposeWebDAVResult
	if err := json.Unmarshal(resultJSON, &exposeWebDAVResult); err != nil {
		return interfaces.ExposeWebDAVResult{}, fmt.Errorf("failed to unmarshal expose WebDAV result: %w", err)
	}

	return exposeWebDAVResult, nil
}

// GetLogs returns logs of recent RPC calls to the VFS service
func (c *Client) GetLogs(limit int, method string, status string) (client.IntrospectionResponse, error) {
	return c.Introspect(limit, method, status)
}

// Expose9P exposes the virtual file system via 9P protocol
func (c *Client) Expose9P(vfspath string, port int, readonly bool) (interfaces.Expose9PResult, error) {
	params := map[string]interface{}{
		"vfspath":  vfspath,
		"port":     port,
		"readonly": readonly,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.Expose9PResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("vfs.expose_9p", paramsJSON, "")
	if err != nil {
		return interfaces.Expose9PResult{}, fmt.Errorf("failed to expose 9P: %w", err)
	}

	// Convert result to Expose9PResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.Expose9PResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var expose9PResult interfaces.Expose9PResult
	if err := json.Unmarshal(resultJSON, &expose9PResult); err != nil {
		return interfaces.Expose9PResult{}, fmt.Errorf("failed to unmarshal expose 9P result: %w", err)
	}

	return expose9PResult, nil
}
