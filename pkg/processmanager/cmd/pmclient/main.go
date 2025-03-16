package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

func main() {
	// Define common flags
	socketPath := flag.String("socket", "/tmp/processmanager.sock", "Path to the Unix domain socket")
	secret := flag.String("secret", "", "Authentication secret for the telnet server")

	// Define command-specific flags
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startName := startCmd.String("name", "", "Name of the process")
	startCommand := startCmd.String("command", "", "Command to run")
	startLog := startCmd.Bool("log", false, "Enable logging")
	startDeadline := startCmd.Int("deadline", 0, "Deadline in seconds (0 for no deadline)")
	startCron := startCmd.String("cron", "", "Cron schedule")
	startJobID := startCmd.String("jobid", "", "Job ID")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listFormat := listCmd.String("format", "", "Output format (json or empty for text)")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteName := deleteCmd.String("name", "", "Name of the process")

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	statusName := statusCmd.String("name", "", "Name of the process")
	statusFormat := statusCmd.String("format", "", "Output format (json or empty for text)")

	restartCmd := flag.NewFlagSet("restart", flag.ExitOnError)
	restartName := restartCmd.String("name", "", "Name of the process")

	stopCmd := flag.NewFlagSet("stop", flag.ExitOnError)
	stopName := stopCmd.String("name", "", "Name of the process")

	// Parse common flags
	flag.Parse()

	// Check if secret is provided
	if *secret == "" {
		log.Fatal("Error: secret is required")
	}

	// Check if a command is provided
	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
	}

	// Create client
	client := processmanager.NewClient(*socketPath, *secret)

	// Connect to the process manager
	err := client.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to process manager: %v", err)
	}
	defer client.Close()

	// Process command
	switch flag.Arg(0) {
	case "start":
		startCmd.Parse(flag.Args()[1:])
		if *startName == "" || *startCommand == "" {
			log.Fatal("Error: name and command are required for start")
		}
		result, err := client.StartProcess(*startName, *startCommand, *startLog, *startDeadline, *startCron, *startJobID)
		if err != nil {
			log.Fatalf("Failed to start process: %v", err)
		}
		fmt.Println(result)

	case "list":
		listCmd.Parse(flag.Args()[1:])
		result, err := client.ListProcesses(*listFormat)
		if err != nil {
			log.Fatalf("Failed to list processes: %v", err)
		}
		fmt.Println(result)

	case "delete":
		deleteCmd.Parse(flag.Args()[1:])
		if *deleteName == "" {
			log.Fatal("Error: name is required for delete")
		}
		result, err := client.DeleteProcess(*deleteName)
		if err != nil {
			log.Fatalf("Failed to delete process: %v", err)
		}
		fmt.Println(result)

	case "status":
		statusCmd.Parse(flag.Args()[1:])
		if *statusName == "" {
			log.Fatal("Error: name is required for status")
		}
		result, err := client.GetProcessStatus(*statusName, *statusFormat)
		if err != nil {
			log.Fatalf("Failed to get process status: %v", err)
		}
		fmt.Println(result)

	case "restart":
		restartCmd.Parse(flag.Args()[1:])
		if *restartName == "" {
			log.Fatal("Error: name is required for restart")
		}
		result, err := client.RestartProcess(*restartName)
		if err != nil {
			log.Fatalf("Failed to restart process: %v", err)
		}
		fmt.Println(result)

	case "stop":
		stopCmd.Parse(flag.Args()[1:])
		if *stopName == "" {
			log.Fatal("Error: name is required for stop")
		}
		result, err := client.StopProcess(*stopName)
		if err != nil {
			log.Fatalf("Failed to stop process: %v", err)
		}
		fmt.Println(result)

	default:
		fmt.Printf("Unknown command: %s\n", flag.Arg(0))
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: pmclient [global flags] command [command flags]")
	fmt.Println("\nGlobal flags:")
	fmt.Println("  -socket string   Path to the Unix domain socket (default \"/tmp/processmanager.sock\")")
	fmt.Println("  -secret string   Authentication secret for the telnet server")
	
	fmt.Println("\nCommands:")
	fmt.Println("  start    Start a new process")
	fmt.Println("    -name string      Name of the process")
	fmt.Println("    -command string   Command to run")
	fmt.Println("    -log              Enable logging")
	fmt.Println("    -deadline int     Deadline in seconds (0 for no deadline)")
	fmt.Println("    -cron string      Cron schedule")
	fmt.Println("    -jobid string     Job ID")
	fmt.Println("  list     List all processes")
	fmt.Println("    -format string    Output format (json or empty for text)")
	fmt.Println("  delete   Delete a process")
	fmt.Println("    -name string      Name of the process")
	fmt.Println("  status   Get the status of a process")
	fmt.Println("    -name string      Name of the process")
	fmt.Println("    -format string    Output format (json or empty for text)")
	fmt.Println("  restart  Restart a process")
	fmt.Println("    -name string      Name of the process")
	fmt.Println("  stop     Stop a process")
	fmt.Println("    -name string      Name of the process")
}
