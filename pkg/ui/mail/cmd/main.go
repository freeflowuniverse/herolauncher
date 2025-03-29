package main

import (
	"flag"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/ui/mail"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Create a new mail server
	server, err := mail.NewMailServer(*configPath)
	if err != nil {
		log.Fatalf("Error creating mail server: %v", err)
	}

	// Setup routes
	server.SetupRoutes()

	// Start the server
	log.Printf("Starting mail server on port %s", *port)
	if err := server.Start(*port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
