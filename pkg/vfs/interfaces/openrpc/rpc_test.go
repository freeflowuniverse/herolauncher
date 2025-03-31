package openrpc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/interfaces/mock"
)

func TestVFSRPC(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "vfs-rpc-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a socket path
	socketPath := filepath.Join(tempDir, "vfs.sock")

	// Create a mock VFS manager
	mockVFS := mock.NewMockVFSManager()

	// Create a secret for authentication
	secret := "test_secret"

	// Create a server
	server, err := NewServer(mockVFS, socketPath, secret)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client := NewClient(socketPath, secret)

	// Test file operations
	t.Run("UploadAndDownloadFile", func(t *testing.T) {
		// Create a test file
		testFilePath := filepath.Join(tempDir, "test.txt")
		testContent := []byte("This is a test file")
		if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Upload the file
		uploadResult, err := client.UploadFile("/test/test.txt", "/dedupe", testFilePath)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}

		if !uploadResult.Success {
			t.Fatalf("Upload failed: %s", uploadResult.Message)
		}

		// Download the file to a new location
		downloadPath := filepath.Join(tempDir, "downloaded.txt")
		downloadResult, err := client.DownloadFile("/test/test.txt", "/dedupe", downloadPath)
		if err != nil {
			t.Fatalf("Failed to download file: %v", err)
		}

		if !downloadResult.Success {
			t.Fatalf("Download failed: %s", downloadResult.Message)
		}

		// Verify the downloaded content
		downloadedContent, err := os.ReadFile(downloadPath)
		if err != nil {
			t.Fatalf("Failed to read downloaded file: %v", err)
		}

		// In a real test, we would compare the content, but our mock just creates placeholder data
		if len(downloadedContent) == 0 {
			t.Fatalf("Downloaded file is empty")
		}
	})

	// Test metadata operations
	t.Run("ExportAndImportMeta", func(t *testing.T) {
		// Create a test file first to have some metadata
		testFilePath := filepath.Join(tempDir, "meta_test.txt")
		testContent := []byte("This is a test file for metadata")
		if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Upload the file
		_, err := client.UploadFile("/meta/test.txt", "/dedupe", testFilePath)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}

		// Export metadata
		metaPath := filepath.Join(tempDir, "meta.txt")
		exportResult, err := client.ExportMeta("/meta", metaPath)
		if err != nil {
			t.Fatalf("Failed to export metadata: %v", err)
		}

		if !exportResult.Success {
			t.Fatalf("Export metadata failed: %s", exportResult.Message)
		}

		// Import metadata to a new location
		importResult, err := client.ImportMeta("/meta2", metaPath)
		if err != nil {
			t.Fatalf("Failed to import metadata: %v", err)
		}

		if !importResult.Success {
			t.Fatalf("Import metadata failed: %s", importResult.Message)
		}
	})

	// Test dedupe operations
	t.Run("ExportAndImportDedupe", func(t *testing.T) {
		// Create a test file first to have some dedupe data
		testFilePath := filepath.Join(tempDir, "dedupe_test.txt")
		testContent := []byte("This is a test file for dedupe")
		if err := os.WriteFile(testFilePath, testContent, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Upload the file
		_, err := client.UploadFile("/dedupe/test.txt", "/dedupe", testFilePath)
		if err != nil {
			t.Fatalf("Failed to upload file: %v", err)
		}

		// Export dedupe data
		dedupePath := filepath.Join(tempDir, "dedupe.txt")
		exportResult, err := client.ExportDedupe("/dedupe", "/dedupe", dedupePath)
		if err != nil {
			t.Fatalf("Failed to export dedupe data: %v", err)
		}

		if !exportResult.Success {
			t.Fatalf("Export dedupe data failed: %s", exportResult.Message)
		}

		// Import dedupe data to a new location
		importResult, err := client.ImportDedupe("/dedupe2", "/dedupe2", dedupePath)
		if err != nil {
			t.Fatalf("Failed to import dedupe data: %v", err)
		}

		if !importResult.Success {
			t.Fatalf("Import dedupe data failed: %s", importResult.Message)
		}
	})

	// Test WebDAV exposure
	t.Run("ExposeWebDAV", func(t *testing.T) {
		result, err := client.ExposeWebDAV("/webdav", 8080, "user", "pass")
		if err != nil {
			t.Fatalf("Failed to expose WebDAV: %v", err)
		}

		if !result.Success {
			t.Fatalf("Expose WebDAV failed: %s", result.Message)
		}

		if result.URL != "http://localhost:8080" {
			t.Fatalf("Unexpected WebDAV URL: %s", result.URL)
		}
	})

	// Test 9P exposure
	t.Run("Expose9P", func(t *testing.T) {
		result, err := client.Expose9P("/9p", 9999, true)
		if err != nil {
			t.Fatalf("Failed to expose 9P: %v", err)
		}

		if !result.Success {
			t.Fatalf("Expose 9P failed: %s", result.Message)
		}

		if result.Address != "localhost:9999" {
			t.Fatalf("Unexpected 9P address: %s", result.Address)
		}
	})
}
