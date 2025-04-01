package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
)

func main() {
	// Parse command line flags
	socketPath := flag.String("socket", "/tmp/herojobs.sock", "Path to the Unix domain socket")
	command := flag.String("cmd", "submit", "Command to execute (submit, list, get, delete, queuesize, queueempty, queueget, queuefetch)")
	circleID := flag.String("circle", "testcircle", "Circle ID")
	topic := flag.String("topic", "default", "Topic")
	jobID := flag.String("jobid", "", "Job ID (for get/delete commands)")
	heroscript := flag.String("heroscript", "", "HeroScript content")
	rhaiscript := flag.String("rhaiscript", "", "RhaiScript content")
	flag.Parse()

	// Create client
	client, err := herojobs.NewClient(*socketPath)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to server
	if err := client.Connect(); err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Execute command
	switch *command {
	case "submit":
		// Create and submit job
		job, err := client.CreateJob(*circleID, *topic, "test-session", *heroscript, *rhaiscript)
		if err != nil {
			fmt.Printf("Failed to submit job: %v\n", err)
			os.Exit(1)
		}

		// Print job details
		fmt.Println("Job submitted successfully:")
		fmt.Printf("  Job ID: %s\n", job.JobID)
		fmt.Printf("  Circle ID: %s\n", job.CircleID)
		fmt.Printf("  Topic: %s\n", job.Topic)
		fmt.Printf("  Status: %s\n", job.Status)
		fmt.Printf("  Time Scheduled: %d\n", job.TimeScheduled)

	case "list":
		// List jobs
		jobIDs, err := client.ListJobs(*circleID, *topic)
		if err != nil {
			fmt.Printf("Failed to list jobs: %v\n", err)
			os.Exit(1)
		}

		// Print job IDs
		fmt.Println("Jobs:")
		if len(jobIDs) > 0 {
			for _, jobID := range jobIDs {
				fmt.Printf("  %s\n", jobID)
			}
		} else {
			fmt.Println("  No jobs found")
		}

	case "get":
		if *jobID == "" {
			fmt.Println("Job ID is required for get command")
			os.Exit(1)
		}

		// Get job
		job, err := client.GetJob(*jobID)
		if err != nil {
			fmt.Printf("Failed to get job: %v\n", err)
			os.Exit(1)
		}

		// Print job details
		fmt.Println("Job details:")
		fmt.Printf("  Job ID: %s\n", job.JobID)
		fmt.Printf("  Circle ID: %s\n", job.CircleID)
		fmt.Printf("  Topic: %s\n", job.Topic)
		fmt.Printf("  Status: %s\n", job.Status)
		fmt.Printf("  Time Scheduled: %d\n", job.TimeScheduled)
		fmt.Printf("  Time Start: %d\n", job.TimeStart)
		fmt.Printf("  Time End: %d\n", job.TimeEnd)
		fmt.Printf("  Error: %s\n", job.Error)
		fmt.Printf("  Result: %s\n", job.Result)

	case "delete":
		if *jobID == "" {
			fmt.Println("Job ID is required for delete command")
			os.Exit(1)
		}

		// Delete job
		if err := client.DeleteJob(*jobID); err != nil {
			fmt.Printf("Failed to delete job: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Job deleted successfully")

	case "queuesize":
		// Get queue size
		size, err := client.QueueSize(*circleID, *topic)
		if err != nil {
			fmt.Printf("Failed to get queue size: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Queue size: %d\n", size)

	case "queueempty":
		// Empty queue
		if err := client.QueueEmpty(*circleID, *topic); err != nil {
			fmt.Printf("Failed to empty queue: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Queue emptied successfully")

	case "queueget":
		// Get job from queue
		job, err := client.QueueGet(*circleID, *topic)
		if err != nil {
			fmt.Printf("Failed to get job from queue: %v\n", err)
			os.Exit(1)
		}

		// Print job details
		fmt.Println("Job details:")
		fmt.Printf("  Job ID: %s\n", job.JobID)
		fmt.Printf("  Circle ID: %s\n", job.CircleID)
		fmt.Printf("  Topic: %s\n", job.Topic)
		fmt.Printf("  Status: %s\n", job.Status)
		fmt.Printf("  Time Scheduled: %d\n", job.TimeScheduled)

	case "queuefetch":
		// Fetch job from queue
		job, err := client.QueueFetch(*circleID, *topic)
		if err != nil {
			fmt.Printf("Failed to fetch job from queue: %v\n", err)
			os.Exit(1)
		}

		// Print job details
		fmt.Println("Job details:")
		fmt.Printf("  Job ID: %s\n", job.JobID)
		fmt.Printf("  Circle ID: %s\n", job.CircleID)
		fmt.Printf("  Topic: %s\n", job.Topic)
		fmt.Printf("  Status: %s\n", job.Status)
		fmt.Printf("  Time Scheduled: %d\n", job.TimeScheduled)

	default:
		fmt.Printf("Unknown command: %s\n", *command)
		os.Exit(1)
	}
}
