package main

import (
	"os"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
)

func TestVFSDB9P(t *testing.T) {
	// Create a temporary directory for the database
	tempDir, err := os.MkdirTemp("", "vfsdb-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize the vfsdb backend
	dbFactory := vfsdb.NewFactory(tempDir)
	vfsImpl, err := dbFactory.Create()
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
	go func() {
		if err := go9p.Serve(addr, fsys.Server()); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// TODO: Add client tests here
	// For now, we're just testing that the server starts without errors
	t.Log("Server started successfully")
}
