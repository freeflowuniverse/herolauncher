package main

import (
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

func main() {
	// Example of using the process manager handler through heroscript

	// Create a new playbook
	pb := playbook.New()

	// Start a simple process
	startAction := pb.NewAction(1, "start", "process", 0, playbook.ActionTypeUnknown)
	startAction.Params.Set("name", "example_process")
	startAction.Params.Set("command", "ping -c 60 localhost")
	startAction.Params.Set("log", "true")

	// List all processes
	listAction := pb.NewAction(2, "list", "process", 0, playbook.ActionTypeUnknown)
	listAction.Params.Set("format", "table")

	// Get status of a specific process
	statusAction := pb.NewAction(3, "status", "process", 0, playbook.ActionTypeUnknown)
	statusAction.Params.Set("name", "example_process")

	// Get logs of a specific process
	logsAction := pb.NewAction(4, "logs", "process", 0, playbook.ActionTypeUnknown)
	logsAction.Params.Set("name", "example_process")
	logsAction.Params.Set("lines", "10")

	// Stop a process
	stopAction := pb.NewAction(5, "stop", "process", 0, playbook.ActionTypeUnknown)
	stopAction.Params.Set("name", "example_process")

	// Generate the heroscript
	script := pb.HeroScript(true)

	// Print the script
	fmt.Println("=== Example HeroScript for Process Manager ===")
	fmt.Println(script)
	fmt.Println("============================================")
	fmt.Println("To use this script:")
	fmt.Println("1. Start the process manager handler server")
	fmt.Println("2. Connect to it using: telnet localhost 8025")
	fmt.Println("3. Authenticate with: !!auth 1234")
	fmt.Println("4. Copy and paste the above script")
	fmt.Println("5. Or use individual commands like: !!process.start name:myprocess command:\"sleep 60\"")
}
