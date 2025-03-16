package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/imapserver"
)

func main() {
	// Parse command line flags
	redisAddr := flag.String("redis-addr", "localhost:6378", "Redis server address")
	imapAddr := flag.String("imap-addr", ":1143", "IMAP server address")
	debugMode := flag.Bool("debug", false, "Enable debug mode with verbose logging")
	flag.Parse()

	// Set up signal handling for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Start the IMAP server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Starting IMAP server on %s with Redis at %s (Debug mode: %v)", *imapAddr, *redisAddr, *debugMode)
		if err := imapserver.ListenAndServe(*redisAddr, *imapAddr, *debugMode); err != nil {
			errCh <- err
		}
	}()

	// Wait for either an error or a signal
	select {
	case err := <-errCh:
		log.Fatalf("IMAP server error: %v", err)
	case sig := <-sigs:
		log.Printf("Received signal %v, shutting down", sig)
	}
}
