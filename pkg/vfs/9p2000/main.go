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
	var verbose bool
	flag.StringVar(&listenAddr, "listen", "0.0.0.0:9999", "Address to listen on")
	flag.StringVar(&dbPath, "db", "./vfsdb", "Path to database")
	flag.BoolVar(&verbose, "verbose", true, "Enable verbose logging")
	flag.Parse()
	
	// Set up logging
	if verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Initialize the vfsdb backend
	vfsImpl, err := vfsdb.NewFromPath(dbPath)
	if err != nil {
		log.Fatalf("Failed to create VFS: %v", err)
	}
	defer vfsImpl.Destroy()

	// Create a new 9p filesystem
	// Use "nobody" as the default user for better compatibility with Linux 9p mounts
	fsys, root := fs.NewFS("nobody", "nobody", 0777,
		fs.WithCreateFile(createVFSDBFile(vfsImpl)),
		fs.WithCreateDir(createVFSDBDir(vfsImpl)),
		fs.WithRemoveFile(removeVFSDBFile(vfsImpl)),
	)

	// Start serving the 9p filesystem with enhanced logging
	log.Printf("Starting 9p server on %s with root directory: %s", listenAddr, root.Stat().Name)
	log.Printf("Server configuration: verbose=%v", verbose)
	log.Printf("IMPORTANT: When mounting this 9p filesystem from Linux, use: mount -t 9p -o version=9p2000,trans=tcp,uname=nobody <server-ip>:9999 /mnt/myvfs")
	log.Printf("For debugging, you can add ,debug=0x8000 to the mount options")
	
	go func() {
		log.Printf("Server listening on %s", listenAddr)
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
