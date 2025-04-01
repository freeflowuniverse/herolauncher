package main

import (
	"log"
	"sync"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/herohandler"
)

func main() {
	// Initialize the herohandler.DefaultInstance
	if err := herohandler.Init(); err != nil {
		log.Fatalf("Failed to initialize herohandler: %v", err)
	}

	// Start the telnet server on both Unix socket and TCP
	socketPath := "/tmp/hero.sock"
	tcpAddress := "localhost:8023"

	log.Println("Starting telnet server...")
	//if err := herohandler.DefaultInstance.StartTelnet(socketPath, tcpAddress, "1234");
	if err := herohandler.DefaultInstance.StartTelnet(socketPath, tcpAddress); err != nil {
		log.Fatalf("Failed to start telnet server: %v", err)
	}
	log.Println("Telnet server started successfully")

	// Register a callback for when the server shuts down
	herohandler.DefaultInstance.EnableSignalHandling(func() {
		log.Println("Server shutdown complete")
	})

	log.Println("Press Ctrl+C to stop the server")

	// Create a WaitGroup that never completes to keep the program running
	// The signal handling in the telnet server will handle the shutdown
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
