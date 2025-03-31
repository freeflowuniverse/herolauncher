package mock

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces"
)

// MockVFSManager provides a mock implementation of the VFSManager interface for testing
type MockVFSManager struct {
	files       map[string][]byte
	dedupeStore map[string][]byte
	exposedDirs map[string]string
}

// NewMockVFSManager creates a new mock VFS manager
func NewMockVFSManager() *MockVFSManager {
	return &MockVFSManager{
		files:       make(map[string][]byte),
		dedupeStore: make(map[string][]byte),
		exposedDirs: make(map[string]string),
	}
}

// UploadFile implements the VFSManager.UploadFile method
func (m *MockVFSManager) UploadFile(vfspath, dedupepath, filepath string) (interfaces.UploadResult, error) {
	// Check if the file exists
	data, err := os.ReadFile(filepath)
	if err != nil {
		return interfaces.UploadResult{
			Success: false,
			Message: fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	// Generate a mock hash
	hash := fmt.Sprintf("hash_%x", data[:min(len(data), 8)])

	// Store the file in the VFS
	m.files[vfspath] = data

	// Store the file in the dedupe store
	m.dedupeStore[hash] = data

	return interfaces.UploadResult{
		Success: true,
		Message: "file uploaded successfully",
		Hash:    hash,
	}, nil
}

// UploadDir implements the VFSManager.UploadDir method
func (m *MockVFSManager) UploadDir(vfspath, dedupepath, dirpath string) (interfaces.UploadDirResult, error) {
	// Check if the directory exists
	info, err := os.Stat(dirpath)
	if err != nil {
		return interfaces.UploadDirResult{
			Success: false,
			Message: fmt.Sprintf("failed to access directory: %v", err),
		}, nil
	}

	if !info.IsDir() {
		return interfaces.UploadDirResult{
			Success: false,
			Message: "path is not a directory",
		}, nil
	}

	// Mock uploading the directory
	var filesProcessed int
	var totalSize int64

	err = filepath.Walk(dirpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, err := filepath.Rel(dirpath, path)
			if err != nil {
				return err
			}

			targetPath := filepath.Join(vfspath, relPath)
			result, err := m.UploadFile(targetPath, dedupepath, path)
			if err != nil {
				return err
			}

			if result.Success {
				filesProcessed++
				totalSize += info.Size()
			}
		}

		return nil
	})

	if err != nil {
		return interfaces.UploadDirResult{
			Success: false,
			Message: fmt.Sprintf("failed to upload directory: %v", err),
		}, nil
	}

	return interfaces.UploadDirResult{
		Success:       true,
		Message:       "directory uploaded successfully",
		FilesProcessed: filesProcessed,
		TotalSize:     totalSize,
	}, nil
}

// DownloadFile implements the VFSManager.DownloadFile method
func (m *MockVFSManager) DownloadFile(vfspath, dedupepath, destpath string) (interfaces.DownloadResult, error) {
	// Check if the file exists in VFS
	data, ok := m.files[vfspath]
	if !ok {
		return interfaces.DownloadResult{
			Success: false,
			Message: "file not found in VFS",
		}, nil
	}

	// Create the destination directory if it doesn't exist
	destDir := filepath.Dir(destpath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return interfaces.DownloadResult{
			Success: false,
			Message: fmt.Sprintf("failed to create destination directory: %v", err),
		}, nil
	}

	// Write the file to the destination
	if err := os.WriteFile(destpath, data, 0644); err != nil {
		return interfaces.DownloadResult{
			Success: false,
			Message: fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	return interfaces.DownloadResult{
		Success: true,
		Message: "file downloaded successfully",
		Size:    int64(len(data)),
	}, nil
}

// ExportMeta implements the VFSManager.ExportMeta method
func (m *MockVFSManager) ExportMeta(vfspath, destpath string) (interfaces.ExportMetaResult, error) {
	// Create a simple metadata file
	var metaData strings.Builder
	entries := 0

	for path := range m.files {
		if strings.HasPrefix(path, vfspath) {
			metaData.WriteString(fmt.Sprintf("%s:%d\n", path, len(m.files[path])))
			entries++
		}
	}

	// Write metadata to the destination file
	if err := os.WriteFile(destpath, []byte(metaData.String()), 0644); err != nil {
		return interfaces.ExportMetaResult{
			Success: false,
			Message: fmt.Sprintf("failed to write metadata: %v", err),
		}, nil
	}

	return interfaces.ExportMetaResult{
		Success: true,
		Message: "metadata exported successfully",
		Entries: entries,
	}, nil
}

// ImportMeta implements the VFSManager.ImportMeta method
func (m *MockVFSManager) ImportMeta(vfspath, sourcepath string) (interfaces.ImportMetaResult, error) {
	// Read the metadata file
	data, err := os.ReadFile(sourcepath)
	if err != nil {
		return interfaces.ImportMetaResult{
			Success: false,
			Message: fmt.Sprintf("failed to read metadata: %v", err),
		}, nil
	}

	// Parse the metadata
	lines := strings.Split(string(data), "\n")
	entriesImported := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		path := parts[0]
		// Create a mock file with the specified size
		m.files[path] = []byte(fmt.Sprintf("mock data for %s", path))
		entriesImported++
	}

	return interfaces.ImportMetaResult{
		Success:        true,
		Message:        "metadata imported successfully",
		EntriesImported: entriesImported,
	}, nil
}

// ExportDedupe implements the VFSManager.ExportDedupe method
func (m *MockVFSManager) ExportDedupe(vfspath, dedupepath, destpath string) (interfaces.ExportDedupeResult, error) {
	// Create a simple dedupe export file
	var dedupeData strings.Builder
	hashesExported := 0

	for hash, data := range m.dedupeStore {
		dedupeData.WriteString(fmt.Sprintf("%s:%d\n", hash, len(data)))
		hashesExported++
	}

	// Write dedupe data to the destination file
	if err := os.WriteFile(destpath, []byte(dedupeData.String()), 0644); err != nil {
		return interfaces.ExportDedupeResult{
			Success: false,
			Message: fmt.Sprintf("failed to write dedupe data: %v", err),
		}, nil
	}

	return interfaces.ExportDedupeResult{
		Success:       true,
		Message:       "dedupe data exported successfully",
		HashesExported: hashesExported,
	}, nil
}

// ImportDedupe implements the VFSManager.ImportDedupe method
func (m *MockVFSManager) ImportDedupe(vfspath, dedupepath, sourcepath string) (interfaces.ImportDedupeResult, error) {
	// Read the dedupe file
	data, err := os.ReadFile(sourcepath)
	if err != nil {
		return interfaces.ImportDedupeResult{
			Success: false,
			Message: fmt.Sprintf("failed to read dedupe data: %v", err),
		}, nil
	}

	// Parse the dedupe data
	lines := strings.Split(string(data), "\n")
	hashesImported := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}

		hash := parts[0]
		// Create a mock dedupe entry
		m.dedupeStore[hash] = []byte(fmt.Sprintf("mock data for hash %s", hash))
		hashesImported++
	}

	return interfaces.ImportDedupeResult{
		Success:       true,
		Message:       "dedupe data imported successfully",
		HashesImported: hashesImported,
	}, nil
}

// Send implements the VFSManager.Send method
func (m *MockVFSManager) Send(dedupepath, pubkeydest string, hashlist []string, secret string) (interfaces.SendResult, error) {
	// Verify the secret
	if secret != "mock_secret" {
		return interfaces.SendResult{
			Success: false,
			Message: "authentication failed: invalid secret",
		}, nil
	}

	// Mock sending the files
	hashesSent := 0
	var totalSize int64

	for _, hash := range hashlist {
		if data, ok := m.dedupeStore[hash]; ok {
			hashesSent++
			totalSize += int64(len(data))
		}
	}

	return interfaces.SendResult{
		Success:   true,
		Message:   "files sent successfully",
		HashesSent: hashesSent,
		TotalSize: totalSize,
	}, nil
}

// SendExist implements the VFSManager.SendExist method
func (m *MockVFSManager) SendExist(dedupepath, pubkeydest string, hashlist []string, secret string) (interfaces.SendExistResult, error) {
	// Verify the secret
	if secret != "mock_secret" {
		return interfaces.SendExistResult{
			Success: false,
			Message: "authentication failed: invalid secret",
		}, nil
	}

	// Check which hashes exist
	var existingHashes []string

	for _, hash := range hashlist {
		if _, ok := m.dedupeStore[hash]; ok {
			existingHashes = append(existingHashes, hash)
		}
	}

	return interfaces.SendExistResult{
		Success:       true,
		Message:       "hash check completed successfully",
		ExistingHashes: existingHashes,
	}, nil
}

// ExposeWebDAV implements the VFSManager.ExposeWebDAV method
func (m *MockVFSManager) ExposeWebDAV(vfspath string, port int, username, password string) (interfaces.ExposeWebDAVResult, error) {
	// Mock exposing the directory via WebDAV
	m.exposedDirs[vfspath] = fmt.Sprintf("webdav://localhost:%d", port)

	return interfaces.ExposeWebDAVResult{
		Success: true,
		Message: "WebDAV server started successfully",
		URL:     fmt.Sprintf("http://localhost:%d", port),
	}, nil
}

// Expose9P implements the VFSManager.Expose9P method
func (m *MockVFSManager) Expose9P(vfspath string, port int, readonly bool) (interfaces.Expose9PResult, error) {
	// Mock exposing the directory via 9P
	m.exposedDirs[vfspath] = fmt.Sprintf("9p://localhost:%d", port)

	return interfaces.Expose9PResult{
		Success: true,
		Message: "9P server started successfully",
		Address: fmt.Sprintf("localhost:%d", port),
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
