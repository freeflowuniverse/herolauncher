package main

import (
	"fmt"
	"log"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces/openrpc"
)

// RunClientExample runs a complete example of using the process manager OpenRPC client
func RunClientExample(socketPath, secret string) error {
	// Create a new client
	client := openrpc.NewClient(socketPath, secret)

	log.Println("🚀 Starting example process...")
	// Start a process
	result, err := client.StartProcess("example-process", "sleep 60", true, 0, "", "")
	if err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}
	log.Printf("Start result: success=%v, message=%s, PID=%d", result.Success, result.Message, result.PID)

	// Wait a bit for the process to start
	time.Sleep(500 * time.Millisecond)

	log.Println("📊 Getting process status...")
	// Get the process status
	status, err := client.GetProcessStatus("example-process", "json")
	if err != nil {
		return fmt.Errorf("failed to get process status: %w", err)
	}
	printProcessStatus(status.(interfaces.ProcessStatus))

	log.Println("📋 Listing all processes...")
	// List all processes
	processList, err := client.ListProcesses("json")
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}
	
	// For simplicity in this example, just log that we got a response
	log.Printf("Got process list response: %T", processList)
	
	// Try to handle the response in a more robust way
	switch v := processList.(type) {
	case []interface{}:
		log.Printf("Found %d processes", len(v))
		for i, p := range v {
			log.Printf("Process %d: %T", i, p)
			if processMap, ok := p.(map[string]interface{}); ok {
				log.Printf("  Name: %v", processMap["name"])
				log.Printf("  Command: %v", processMap["command"])
				log.Printf("  Status: %v", processMap["status"])
			}
		}
	case map[string]interface{}:
		log.Printf("Process list is a map with %d entries", len(v))
		for k, val := range v {
			log.Printf("  %s: %T", k, val)
		}
	default:
		log.Printf("Process list is of unexpected type: %T", processList)
	}

	log.Println("📜 Getting process logs...")
	// Get process logs
	logResult, err := client.GetProcessLogs("example-process", 10)
	if err != nil {
		return fmt.Errorf("failed to get process logs: %w", err)
	}
	log.Printf("Process logs: success=%v, message=%s", logResult.Success, logResult.Message)
	log.Printf("Logs:\n%s", logResult.Logs)

	log.Println("🔄 Restarting process...")
	// Restart the process
	restartResult, err := client.RestartProcess("example-process")
	if err != nil {
		return fmt.Errorf("failed to restart process: %w", err)
	}
	log.Printf("Restart result: success=%v, message=%s, PID=%d", restartResult.Success, restartResult.Message, restartResult.PID)

	// Wait a bit for the process to restart
	time.Sleep(500 * time.Millisecond)

	// Get the process status after restart
	status, err = client.GetProcessStatus("example-process", "json")
	if err != nil {
		return fmt.Errorf("failed to get process status after restart: %w", err)
	}
	log.Println("Process status after restart:")
	printProcessStatus(status.(interfaces.ProcessStatus))

	log.Println("⏹️ Stopping process...")
	// Stop the process
	stopResult, err := client.StopProcess("example-process")
	if err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}
	log.Printf("Stop result: success=%v, message=%s", stopResult.Success, stopResult.Message)

	// Wait a bit for the process to stop
	time.Sleep(500 * time.Millisecond)

	// Get the process status after stop
	status, err = client.GetProcessStatus("example-process", "json")
	if err != nil {
		return fmt.Errorf("failed to get process status after stop: %w", err)
	}
	log.Println("Process status after stop:")
	printProcessStatus(status.(interfaces.ProcessStatus))

	log.Println("🗑️ Deleting process...")
	// Delete the process
	deleteResult, err := client.DeleteProcess("example-process")
	if err != nil {
		return fmt.Errorf("failed to delete process: %w", err)
	}
	log.Printf("Delete result: success=%v, message=%s", deleteResult.Success, deleteResult.Message)

	// Try to get the process status after delete (should fail)
	_, err = client.GetProcessStatus("example-process", "json")
	if err != nil {
		log.Printf("Expected error after deletion: %v", err)
	} else {
		return fmt.Errorf("process still exists after deletion")
	}

	log.Println("✅ Example completed successfully!")
	return nil
}

// printProcessStatus prints the status of a process
func printProcessStatus(status interfaces.ProcessStatus) {
	log.Printf("Process: %s", status.Name)
	log.Printf("  Command: %s", status.Command)
	log.Printf("  Status: %s", status.Status)
	log.Printf("  PID: %d", status.PID)
	if status.CPUPercent > 0 {
		log.Printf("  CPU: %.2f%%", status.CPUPercent)
	}
	if status.MemoryMB > 0 {
		log.Printf("  Memory: %.2f MB", status.MemoryMB)
	}
	if !status.StartTime.IsZero() {
		log.Printf("  Started: %s", status.StartTime.Format(time.RFC3339))
	}
}
