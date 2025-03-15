package vfs9p

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/knusbaum/go9p"
)

func TestVFS9PAdapter(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "vfs9p-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create database path
	dbPath := filepath.Join(tempDir, "test.db")

	// Create VFS
	vfsImpl, err := vfsdb.NewFromPath(dbPath)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}

	// Create some test data
	_, err = vfsImpl.RootGet()
	if err != nil {
		t.Fatalf("Failed to get root directory: %v", err)
	}

	// Create a test file
	_, err = vfsImpl.FileCreate("/test.txt")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Write some data to the test file
	err = vfsImpl.FileWrite("/test.txt", []byte("Hello, 9P world!"))
	if err != nil {
		t.Fatalf("Failed to write to test file: %v", err)
	}

	// Create a test directory
	_, err = vfsImpl.DirCreate("/testdir")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a file in the test directory
	_, err = vfsImpl.FileCreate("/testdir/nested.txt")
	if err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Write some data to the nested file
	err = vfsImpl.FileWrite("/testdir/nested.txt", []byte("Nested file content"))
	if err != nil {
		t.Fatalf("Failed to write to nested file: %v", err)
	}

	// Create VFS9P adapter
	adapter := NewVFS9P(vfsImpl)

	// This is where you would normally start a 9P server and connect a client
	// For a real test, you would need to:
	// 1. Start a 9P server in a goroutine
	// 2. Connect a 9P client
	// 3. Perform operations through the client
	// 4. Verify the results

	// For this simple test, we'll just verify that the adapter was created successfully
	if adapter == nil {
		t.Fatal("Failed to create VFS9P adapter")
	}

	// Test that the adapter has the correct VFS implementation
	if adapter.vfsImpl != vfsImpl {
		t.Fatal("Adapter has incorrect VFS implementation")
	}

	t.Log("VFS9P adapter created successfully")
}

// Example of how to use the VFS9P adapter in a real application
func ExampleNewVFS9P() {
	// Create a VFS instance
	vfsImpl, err := vfsdb.NewFromPath("/path/to/database")
	if err != nil {
		// Handle error
		return
	}

	// Create VFS9P adapter
	adapter := NewVFS9P(vfsImpl)

	// Serve 9p on port 5640
	// This would normally block, so in a real application you might want to run it in a goroutine
	go func() {
		err := go9p.Serve("0.0.0.0:5640", adapter)
		if err != nil {
			// Handle error
		}
	}()

	// Your application can continue running here
	// ...
}
