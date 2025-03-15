package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/vfs/vfsdb"
	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
)

func main() {
	// Parse command line arguments
	var listenAddr string
	var dbPath string
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:9999", "Address to listen on")
	flag.StringVar(&dbPath, "db", "./vfsdb", "Path to database")
	flag.Parse()

	// Initialize the vfsdb backend
	vfsImpl, err := vfsdb.NewFromPath(dbPath)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}
	defer vfsImpl.Destroy()

	// Create a new 9p filesystem
	fsys, root := fs.NewFS("glenda", "glenda", 0777,
		fs.WithCreateFile(createVFSDBFile(vfsImpl)),
		fs.WithCreateDir(createVFSDBDir(vfsImpl)),
		fs.WithRemoveFile(removeVFSDBFile(vfsImpl)),
	)

	// Start serving the 9p filesystem
	log.Printf("Starting 9p server on %s with root directory: %s", listenAddr, root.Stat().Name)
	go func() {
		if err := go9p.Serve(listenAddr, fsys.Server()); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down...")
}
