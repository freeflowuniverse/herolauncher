package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
)

func main() {
	// Parse command line flags
	redisAddr := flag.String("redis", "localhost:6379", "Redis address")
	isUnixSocket := flag.Bool("unix", false, "Use Unix socket for Redis connection")
	flag.Parse()

	// Initialize Redis client
	redisClient, err := herojobs.NewRedisClient(*redisAddr, *isUnixSocket)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("Connected to Redis successfully")

	// Test with fake handler
	log.Println("Testing daemon with fake handler...")
	if err := herojobs.TestWithFakeHandler(redisClient); err != nil {
		log.Fatalf("Test failed: %v", err)
	}
	log.Println("Test with fake handler completed successfully")

	// Create and start daemon
	log.Println("Starting daemon...")
	daemon := herojobs.NewDaemon(redisClient)
	daemon.Start()

	// Create a test job
	job := herojobs.NewJob()
	job.CircleID = "test"
	job.Topic = "test"
	job.HeroScript = `
!!fake.return_success
	message: "Manual test job processed successfully"
`
	job.Timeout = 30 // 30 seconds timeout

	// Enqueue the job
	log.Printf("Enqueuing test job with ID: %s", job.JobID)
	if err := redisClient.EnqueueJob(job); err != nil {
		log.Fatalf("Failed to enqueue job: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or timeout
	select {
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
	case <-time.After(5 * time.Second):
		log.Println("Test timeout reached")
	}

	// Stop the daemon
	log.Println("Stopping daemon...")
	daemon.Stop()

	// Check the job status
	processedJob, err := redisClient.GetJob(job.JobID)
	if err != nil {
		log.Fatalf("Failed to get job: %v", err)
	}

	log.Printf("Job status: %s", processedJob.Status)
	if processedJob.Status == herojobs.JobStatusDone {
		log.Printf("Job result: %s", processedJob.Result)
	} else if processedJob.Status == herojobs.JobStatusError {
		log.Printf("Job error: %s", processedJob.Error)
	}

	fmt.Println("Daemon test completed")
}
