package main

import (
	"os"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/client"
	"github.com/stretchr/testify/assert"
)

func TestVFSDB9P(t *testing.T) {
	// Create a temporary directory for the database
	tempDir, err := os.MkdirTemp("", "vfsdb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the vfsdb backend
	vfsImpl, err := vfsdb.NewFromPath(tempDir)
	if err != nil {
		t.Fatalf("Failed to create VFS: %v", err)
	}
	defer vfsImpl.Destroy()

	// Create a new 9p filesystem
	fsys, _ := fs.NewFS("glenda", "glenda", 0777,
		fs.WithCreateFile(createVFSDBFile(vfsImpl)),
		fs.WithCreateDir(createVFSDBDir(vfsImpl)),
		fs.WithRemoveFile(removeVFSDBFile(vfsImpl)),
	)

	// Start a test server
	addr := "localhost:9998"
	serverErrCh := make(chan error, 1)
	go func() {
		if err := go9p.Serve(addr, fsys.Server()); err != nil {
			serverErrCh <- err
		}
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Connect a client to the server
	conn, err := client.Dial("tcp", addr, "user", "")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	// No need to close the client as it doesn't have a Close method

	// Test basic operations
	// Create a test directory - handle the case where it might already exist
	dir, err := conn.Create("/testdir", 0755|os.ModeDir)
	if err != nil {
		// If the directory already exists, try to open it instead
		// Try to open the directory to check if it exists
		dir, err = conn.Open("/testdir", 0)
		if err != nil {
			// If we can't open it either, then there's a real problem
			assert.NoError(t, err, "Failed to create or open directory")
		}
	}
	dir.Close()

	// Create a test file with a unique name to avoid conflicts
	testFileName := "/testdir/testfile_" + time.Now().Format("20060102150405")
	f, err := conn.Create(testFileName, 0644)
	assert.NoError(t, err, "Failed to create file")

	// Write to the file
	testData := []byte("Hello, 9p2000!")
	var n int
	n, err = f.Write(testData)
	assert.NoError(t, err, "Failed to write to file")
	assert.Equal(t, len(testData), n, "Write returned wrong byte count")
	f.Close()

	// We're reusing the testData from above
	// testData = []byte("Hello, 9p2000!")
	
	// Read the file back
	f, err = conn.Open(testFileName, 0)
	assert.NoError(t, err, "Failed to open file for reading")
	readData := make([]byte, len(testData)+10) // Extra space to check if more data is read
	// Reuse n variable from above
	n, err = f.Read(readData)
	assert.NoError(t, err, "Failed to read from file")
	assert.Equal(t, len(testData), n, "Read returned wrong byte count")
	assert.Equal(t, testData, readData[:n], "Read data doesn't match written data")
	f.Close()

	// List directory contents
	entries, err := conn.Readdir("/testdir")
	assert.NoError(t, err, "Failed to read directory")
	assert.GreaterOrEqual(t, len(entries), 1, "Directory should have at least one entry")
	
	// Extract just the filename from testFileName (remove the directory part)
	expectedFileName := testFileName[len("/testdir/"):]
	
	// Check if our file exists in the directory listing
	foundFile := false
	for _, entry := range entries {
		if entry.Name == expectedFileName {
			foundFile = true
			break
		}
	}
	assert.True(t, foundFile, "Could not find our test file in directory listing")

	// Remove the file
	err = conn.Remove(testFileName)
	assert.NoError(t, err, "Failed to remove file")

	// Check that the file is gone
	entries, err = conn.Readdir("/testdir")
	assert.NoError(t, err, "Failed to read directory after file removal")
	// The directory might not be empty if other tests have created files in it
	// So we'll just check that our file is gone
	foundFile = false
	for _, entry := range entries {
		if entry.Name == expectedFileName {
			foundFile = true
			break
		}
	}
	assert.False(t, foundFile, "Test file should be gone from directory listing")

	// Remove the directory
	err = conn.Remove("/testdir")
	assert.NoError(t, err, "Failed to remove directory")

	// Check that the server is still running
	select {
	case err := <-serverErrCh:
		t.Fatalf("Server stopped unexpectedly: %v", err)
	default:
		t.Log("Server is still running as expected")
	}
}
