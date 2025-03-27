package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/srtrecorder"
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", 8090, "Port to listen on for SRT connections")
	outputDir := flag.String("output", "./recordings", "Directory to save recordings")
	flag.Parse()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Get absolute path for output directory
	absOutputDir, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path for output directory: %v", err)
	}

	// Create and start the SRT recorder
	recorder := srtrecorder.NewSRTRecorder(*port, absOutputDir)
	if err := recorder.Start(); err != nil {
		log.Fatalf("Failed to start SRT recorder: %v", err)
	}

	fmt.Printf("SRT recorder started on port %d\n", *port)
	fmt.Printf("Recordings will be saved to: %s\n", absOutputDir)
	fmt.Println("Press Ctrl+C to stop")

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nStopping SRT recorder...")
	if err := recorder.Stop(); err != nil {
		log.Fatalf("Failed to stop SRT recorder: %v", err)
	}

	fmt.Println("SRT recorder stopped")
}
