package main

import (
	"fmt"
	"log"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

func main() {
	// Create a new process manager
	pm := processmanager.NewProcessManager("test123")
	
	// Set logs base path
	pm.SetLogsBasePath("/tmp/herolauncher/process_logs")
	
	fmt.Println("Starting test process...")
	
	// Start a simple process
	err := pm.StartProcess("test-echo", "echo 'Hello, World!' && sleep 2", true, 0, "", "")
	if err != nil {
		log.Fatalf("Failed to start process: %v", err)
	}
	
	// List processes
	fmt.Println("Listing processes:")
	processes := pm.ListProcesses()
	for _, proc := range processes {
		fmt.Printf("- %s (PID: %d, Status: %s)\n", proc.Name, proc.PID, proc.Status)
	}
	
	// Wait for the process to complete
	time.Sleep(3 * time.Second)
	
	// Get process status
	fmt.Println("Getting process status:")
	procInfo, err := pm.GetProcessStatus("test-echo")
	if err != nil {
		log.Fatalf("Failed to get process status: %v", err)
	}
	fmt.Printf("Process status: %s\n", procInfo.Status)
	
	// Get process logs
	fmt.Println("Getting process logs:")
	logs, err := pm.GetProcessLogs("test-echo", 10)
	if err != nil {
		log.Fatalf("Failed to get process logs: %v", err)
	}
	fmt.Printf("Process logs:\n%s\n", logs)
	
	// Start a long-running process
	fmt.Println("Starting long-running process...")
	err = pm.StartProcess("test-sleep", "sleep 10", true, 0, "", "")
	if err != nil {
		log.Fatalf("Failed to start long-running process: %v", err)
	}
	
	// List processes again
	fmt.Println("Listing processes again:")
	processes = pm.ListProcesses()
	for _, proc := range processes {
		fmt.Printf("- %s (PID: %d, Status: %s)\n", proc.Name, proc.PID, proc.Status)
	}
	
	// Stop the long-running process
	fmt.Println("Stopping long-running process...")
	err = pm.StopProcess("test-sleep")
	if err != nil {
		log.Fatalf("Failed to stop process: %v", err)
	}
	
	// Get process status after stopping
	fmt.Println("Getting process status after stopping:")
	procInfo, err = pm.GetProcessStatus("test-sleep")
	if err != nil {
		log.Fatalf("Failed to get process status: %v", err)
	}
	fmt.Printf("Process status: %s\n", procInfo.Status)
	
	// Delete the processes
	fmt.Println("Deleting processes...")
	err = pm.DeleteProcess("test-echo")
	if err != nil {
		log.Printf("Failed to delete test-echo: %v", err)
	}
	
	err = pm.DeleteProcess("test-sleep")
	if err != nil {
		log.Printf("Failed to delete test-sleep: %v", err)
	}
	
	fmt.Println("Test completed successfully!")
}
