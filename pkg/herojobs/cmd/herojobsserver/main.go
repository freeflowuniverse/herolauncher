package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/herojobs.sock", "Path to the Unix domain socket")
	redisAddr := flag.String("redis", "localhost:6379", "Redis server address")
	redisUnixSocket := flag.Bool("redis-unix", false, "Use Unix domain socket for Redis connection")
	flag.Parse()

	// Create and start the server
	server, err := herojobs.NewServer(*socketPath, *redisAddr, *redisUnixSocket)
	if err != nil {
		fmt.Printf("Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Start the server
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("HeroJobs server started on socket: %s\n", *socketPath)
	fmt.Printf("Connected to Redis at: %s\n", *redisAddr)
	fmt.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Stop the server
	if err := server.Stop(); err != nil {
		fmt.Printf("Failed to stop server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server stopped")
}
