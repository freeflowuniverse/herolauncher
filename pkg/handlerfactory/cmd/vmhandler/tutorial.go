package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory"
)

// runTutorial runs an interactive tutorial demonstrating the VM handler
func runTutorial() {
	fmt.Println("=== VM Handler Tutorial ===")
	fmt.Println("This tutorial will demonstrate how to use the VM handler with heroscript commands.")
	fmt.Println("Press Enter to continue through each step...")
	waitForEnter()

	// Create a new handler factory
	fmt.Println("\nStep 1: Create a new HandlerFactory")
	fmt.Println("factory := handlerfactory.NewHandlerFactory()")
	factory := handlerfactory.NewHandlerFactory()
	waitForEnter()

	// Create a VM handler
	fmt.Println("\nStep 2: Create a VM handler")
	fmt.Println("vmHandler := NewVMHandler()")
	vmHandler := NewVMHandler()
	waitForEnter()

	// Register the VM handler with the factory
	fmt.Println("\nStep 3: Register the VM handler with the factory")
	fmt.Println("factory.RegisterHandler(vmHandler)")
	err := factory.RegisterHandler(vmHandler)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println("Handler registered successfully!")
	waitForEnter()

	// Show available actions
	fmt.Println("\nStep 4: List available actions for the VM handler")
	actions := factory.GetSupportedActions()
	fmt.Println("Supported actions for 'vm' actor:")
	for _, action := range actions["vm"] {
		fmt.Printf("- %s\n", action)
	}
	waitForEnter()

	// Process heroscript commands
	fmt.Println("\nStep 5: Process heroscript commands")
	
	// Define a VM
	defineScript := `!!vm.define name:'tutorial_vm' cpu:2 memory:'4GB' storage:'50GB'
    description: 'A tutorial VM for demonstration purposes'`
	fmt.Println("\nCommand:")
	fmt.Println(defineScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err := factory.ProcessHeroscript(defineScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Start the VM
	startScript := `!!vm.start name:'tutorial_vm'`
	fmt.Println("\nCommand:")
	fmt.Println(startScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(startScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Add a disk
	diskAddScript := `!!vm.disk_add name:'tutorial_vm' size:'20GB' type:'SSD'`
	fmt.Println("\nCommand:")
	fmt.Println(diskAddScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(diskAddScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Check VM status
	statusScript := `!!vm.status name:'tutorial_vm'`
	fmt.Println("\nCommand:")
	fmt.Println(statusScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(statusScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// List all VMs
	listScript := `!!vm.list`
	fmt.Println("\nCommand:")
	fmt.Println(listScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(listScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Stop the VM
	stopScript := `!!vm.stop name:'tutorial_vm'`
	fmt.Println("\nCommand:")
	fmt.Println(stopScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(stopScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Delete the VM
	deleteScript := `!!vm.delete name:'tutorial_vm'`
	fmt.Println("\nCommand:")
	fmt.Println(deleteScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(deleteScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Try an invalid command
	invalidScript := `!!vm.invalid name:'tutorial_vm'`
	fmt.Println("\nInvalid Command:")
	fmt.Println(invalidScript)
	fmt.Println("\nProcessing...")
	time.Sleep(1 * time.Second)
	result, err = factory.ProcessHeroscript(invalidScript)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %s\n", result)
	}
	waitForEnter()

	// Conclusion
	fmt.Println("\nTutorial Complete!")
	fmt.Println("You've seen how to:")
	fmt.Println("1. Create a HandlerFactory")
	fmt.Println("2. Register a VM handler")
	fmt.Println("3. Process various heroscript commands")
	fmt.Println("\nTo run the full telnet server example, execute:")
	fmt.Println("go run main.go vm_handler.go")
	fmt.Println("\nPress Enter to exit the tutorial...")
	waitForEnter()
}

// waitForEnter waits for the user to press Enter
func waitForEnter() {
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
}

// addTutorialCommand adds the tutorial command to the main function
func addTutorialCommand() {
	// Check command line arguments
	if len(os.Args) > 1 && os.Args[1] == "tutorial" {
		runTutorial()
		os.Exit(0)
	}
}
