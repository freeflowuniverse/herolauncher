package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfslocal"
	"github.com/freeflowuniverse/herolauncher/pkg/vfsdav"
)

func main() {
	// Create a temporary directory for the VFS
	tempDir, err := os.MkdirTemp("", "vfsdav-client-example")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a local VFS implementation
	vfsImpl, err := vfslocal.New(tempDir)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}

	// Create and start the WebDAV server in a goroutine
	addr := "localhost:8080"
	server := vfsdav.NewServer(vfsImpl, addr)
	
	fmt.Printf("WebDAV server started at http://%s\n", addr)
	fmt.Printf("Serving files from: %s\n", tempDir)
	
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(500 * time.Millisecond)

	// Create a WebDAV client
	client, err := webdav.NewClient(http.DefaultClient, fmt.Sprintf("http://%s", addr))
	if err != nil {
		log.Fatalf("Failed to create WebDAV client: %v", err)
	}

	// Run all the tests
	ctx := context.Background()
	
	fmt.Println("\n=== Running WebDAV Client Tests ===")
	
	// Test 1: Create a text file
	fmt.Println("\n1. Creating a text file...")
	if err := createTextFile(ctx, client, "/hello.txt", "Hello, WebDAV!"); err != nil {
		log.Fatalf("Failed to create text file: %v", err)
	}
	fmt.Println("✓ Text file created successfully")

	// Test 2: Create a binary file (small image)
	fmt.Println("\n2. Creating a binary file...")
	if err := createBinaryFile(ctx, client, "/small-image.png"); err != nil {
		log.Fatalf("Failed to create binary file: %v", err)
	}
	fmt.Println("✓ Binary file created successfully")

	// Test 3: Create files of various sizes
	fmt.Println("\n3. Creating files of various sizes...")
	
	// Create files of different sizes: 10KB, 100KB, 500KB, 1MB
	fileSizes := map[string]int{
		"/file-10kb.dat":  10 * 1024,
		"/file-100kb.dat": 100 * 1024,
		"/file-500kb.dat": 500 * 1024,
		"/file-1mb.dat":   1 * 1024 * 1024,
	}
	
	for path, size := range fileSizes {
		fmt.Printf("  Creating %s (%d bytes)...\n", path, size)
		data, err := generateRandomData(size)
		if err != nil {
			log.Fatalf("Failed to generate random data: %v", err)
		}
		
		// Save the data to verify later
		if err := createFileWithData(ctx, client, path, data); err != nil {
			log.Fatalf("Failed to create file %s: %v", path, err)
		}
		
		// Verify the file size
		fileInfo, err := client.Stat(ctx, path)
		if err != nil {
			log.Fatalf("Failed to get file info for %s: %v", path, err)
		}
		
		if fileInfo.Size != int64(size) {
			log.Fatalf("File size mismatch for %s: expected %d, got %d", path, size, fileInfo.Size)
		}
		
		fmt.Printf("    ✓ File created with correct size: %d bytes\n", fileInfo.Size)
	}
	
	fmt.Println("✓ All files of various sizes created successfully")

	// Test 4: List files in root directory
	fmt.Println("\n4. Listing files in root directory...")
	files, err := client.ReadDir(ctx, "/", false)
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}
	fmt.Println("Files in root directory:")
	for _, file := range files {
		fmt.Printf("  - %s (%d bytes, dir: %v)\n", file.Path, file.Size, file.IsDir)
	}
	fmt.Println("✓ Directory listing successful")

	// Test 5: Create a directory
	fmt.Println("\n5. Creating a directory...")
	if err := client.Mkdir(ctx, "/test-dir"); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Println("✓ Directory created successfully")

	// Test 6: Create nested directories
	fmt.Println("\n6. Creating nested directories...")
	if err := client.Mkdir(ctx, "/nested/dir/structure"); err != nil {
		log.Fatalf("Failed to create nested directories: %v", err)
	}
	fmt.Println("✓ Nested directories created successfully")

	// Test 7: Copy a file
	fmt.Println("\n7. Copying a file...")
	if err := client.Copy(ctx, "/hello.txt", "/test-dir/hello-copy.txt", nil); err != nil {
		log.Fatalf("Failed to copy file: %v", err)
	}
	fmt.Println("✓ File copied successfully")

	// Test 8: Move a file
	fmt.Println("\n8. Moving a file...")
	if err := client.Move(ctx, "/small-image.png", "/test-dir/moved-image.png", nil); err != nil {
		log.Fatalf("Failed to move file: %v", err)
	}
	fmt.Println("✓ File moved successfully")

	// Test 9: Read files and verify content integrity
	fmt.Println("\n9. Reading files and verifying content integrity...")
	
	// First, verify the simple text file
	content, err := readFile(ctx, client, "/test-dir/hello-copy.txt")
	if err != nil {
		log.Fatalf("Failed to read text file: %v", err)
	}
	
	if content != "Hello, WebDAV!" {
		log.Fatalf("Content mismatch for text file: expected 'Hello, WebDAV!', got '%s'", content)
	}
	fmt.Printf("  ✓ Text file content verified: %s\n", content)
	
	// Now verify all the files with different sizes
	for path, size := range fileSizes {
		fmt.Printf("  Verifying content integrity for %s (%d bytes)...\n", path, size)
		
		// Generate the expected data again (using the same seed)
		expectedData, err := generateRandomData(size)
		if err != nil {
			log.Fatalf("Failed to generate expected data: %v", err)
		}
		
		// Read the actual file data
		actualData, err := readFileBytes(ctx, client, path)
		if err != nil {
			log.Fatalf("Failed to read file %s: %v", path, err)
		}
		
		// Verify the data length
		if len(actualData) != size {
			log.Fatalf("Data length mismatch for %s: expected %d, got %d", path, size, len(actualData))
		}
		
		// Verify data integrity (compare checksums)
		expectedChecksum := calculateChecksum(expectedData)
		actualChecksum := calculateChecksum(actualData)
		
		if expectedChecksum != actualChecksum {
			log.Fatalf("Content integrity check failed for %s: checksums don't match", path)
		}
		
		fmt.Printf("    ✓ Content integrity verified (checksum: %d)\n", actualChecksum)
	}
	
	fmt.Println("✓ All file contents verified successfully")

	// Test 10: Stat a file
	fmt.Println("\n10. Getting file information...")
	fileInfo, err := client.Stat(ctx, "/file-1mb.dat")
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}
	fmt.Printf("File info: %s, size: %d bytes, modified: %s\n", 
		fileInfo.Path, fileInfo.Size, fileInfo.ModTime.Format(time.RFC3339))
	fmt.Println("✓ File info retrieved successfully")

	// Test 11: Create a file in a nested directory
	fmt.Println("\n11. Creating a file in a nested directory...")
	if err := createTextFile(ctx, client, "/nested/dir/structure/nested-file.txt", 
		"This is a file in a nested directory"); err != nil {
		log.Fatalf("Failed to create file in nested directory: %v", err)
	}
	fmt.Println("✓ File created in nested directory successfully")

	// Test 12: Recursive directory listing
	fmt.Println("\n12. Recursive directory listing...")
	files, err = client.ReadDir(ctx, "/", true)
	if err != nil {
		log.Fatalf("Failed to list files recursively: %v", err)
	}
	fmt.Println("All files (recursive):")
	for _, file := range files {
		fmt.Printf("  - %s (%d bytes, dir: %v)\n", file.Path, file.Size, file.IsDir)
	}
	fmt.Println("✓ Recursive directory listing successful")

	// Test 13: Delete a file
	fmt.Println("\n13. Deleting a file...")
	if err := client.RemoveAll(ctx, "/file-100kb.dat"); err != nil {
		log.Fatalf("Failed to delete file: %v", err)
	}
	fmt.Println("✓ File deleted successfully")

	// Test 14: Delete a directory with contents
	fmt.Println("\n14. Deleting a directory with contents...")
	if err := client.RemoveAll(ctx, "/test-dir"); err != nil {
		log.Fatalf("Failed to delete directory: %v", err)
	}
	fmt.Println("✓ Directory deleted successfully")

	// Test 15: Update a file
	fmt.Println("\n15. Updating a file...")
	if err := createTextFile(ctx, client, "/hello.txt", "Updated content!"); err != nil {
		log.Fatalf("Failed to update file: %v", err)
	}
	content, err = readFile(ctx, client, "/hello.txt")
	if err != nil {
		log.Fatalf("Failed to read updated file: %v", err)
	}
	fmt.Printf("Updated file content: %s\n", content)
	fmt.Println("✓ File updated successfully")

	// Test 16: Create a file with special characters
	fmt.Println("\n16. Creating a file with special characters...")
	specialFileName := "/special-chars-!@#$%^&()_+.txt"
	if err := createTextFile(ctx, client, specialFileName, "File with special characters in name"); err != nil {
		log.Fatalf("Failed to create file with special characters: %v", err)
	}
	fmt.Println("✓ File with special characters created successfully")

	// Test 17: Create files with different extensions
	fmt.Println("\n17. Creating files with different extensions...")
	extensions := []string{".html", ".css", ".js", ".json", ".xml", ".pdf", ".zip"}
	for i, ext := range extensions {
		filename := fmt.Sprintf("/file%d%s", i+1, ext)
		content := fmt.Sprintf("This is a %s file", ext)
		if err := createTextFile(ctx, client, filename, content); err != nil {
			log.Fatalf("Failed to create %s file: %v", ext, err)
		}
	}
	fmt.Println("✓ Files with different extensions created successfully")

	// Final check: List all files to verify
	fmt.Println("\n=== Final Directory Listing ===")
	files, err = client.ReadDir(ctx, "/", true)
	if err != nil {
		log.Fatalf("Failed to list files for final verification: %v", err)
	}
	for _, file := range files {
		fmt.Printf("  - %s (%d bytes, dir: %v)\n", file.Path, file.Size, file.IsDir)
	}

	fmt.Println("\n=== All WebDAV Client Tests Completed Successfully ===")
}

// Helper function to create a text file
func createTextFile(ctx context.Context, client *webdav.Client, path, content string) error {
	w, err := client.Create(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	_, err = io.Copy(w, strings.NewReader(content))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write content: %w", err)
	}
	return w.Close()
}

// Helper function to create a binary file (a small PNG image)
func createBinaryFile(ctx context.Context, client *webdav.Client, path string) error {
	// Simple 1x1 transparent PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	w, err := client.Create(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to create binary file: %w", err)
	}
	_, err = io.Copy(w, bytes.NewReader(pngData))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write binary content: %w", err)
	}
	return w.Close()
}

// Helper function to create a large file of specified size
func createLargeFile(ctx context.Context, client *webdav.Client, path string, size int) error {
	w, err := client.Create(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to create large file: %w", err)
	}

	// Create a pattern of data to repeat
	pattern := []byte("0123456789ABCDEF")
	buffer := make([]byte, 8192) // 8KB buffer for efficiency
	
	for i := 0; i < len(buffer); i += len(pattern) {
		copy(buffer[i:], pattern)
	}

	remaining := size
	for remaining > 0 {
		writeSize := remaining
		if writeSize > len(buffer) {
			writeSize = len(buffer)
		}
		
		n, err := w.Write(buffer[:writeSize])
		if err != nil {
			w.Close()
			return fmt.Errorf("failed to write large file data: %w", err)
		}
		
		remaining -= n
	}

	return w.Close()
}

// Helper function to read a file as string
func readFile(ctx context.Context, client *webdav.Client, path string) (string, error) {
	r, err := client.Open(ctx, path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(data), nil
}

// Helper function to read a file as bytes
func readFileBytes(ctx context.Context, client *webdav.Client, path string) ([]byte, error) {
	r, err := client.Open(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// Helper function to create a file with the provided data
func createFileWithData(ctx context.Context, client *webdav.Client, path string, data []byte) error {
	w, err := client.Create(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	_, err = io.Copy(w, bytes.NewReader(data))
	if err != nil {
		w.Close()
		return fmt.Errorf("failed to write content: %w", err)
	}
	return w.Close()
}

// Helper function to generate random data of specified size
func generateRandomData(size int) ([]byte, error) {
	// Use a fixed seed for reproducibility when verifying
	data := make([]byte, size)
	
	// Generate deterministic pattern for verification
	for i := 0; i < size; i++ {
		data[i] = byte((i * 37) % 256) // Simple pattern that's not just sequential
	}
	
	return data, nil
}

// Helper function to calculate a simple checksum for data verification
func calculateChecksum(data []byte) int64 {
	var sum int64
	for i, b := range data {
		sum += int64(b) * int64(i%1000+1) // Simple weighted sum
	}
	return sum
}
