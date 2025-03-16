package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
)

func main() {
	// Define command-line flags
	specFile := flag.String("spec", "", "Path to OpenAPI specification file (JSON or YAML)")
	port := flag.Int("port", 8080, "Port to run the server on")
	generateOnly := flag.Bool("generate-only", false, "Only generate the server code, don't run the server")
	outputFile := flag.String("output", "", "Output file for generated server code (only used with --generate-only)")
	
	flag.Parse()

	// Check if spec file is provided
	if *specFile == "" {
		fmt.Println("Error: OpenAPI specification file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the OpenAPI specification
	fmt.Printf("Parsing OpenAPI specification from %s...\n", *specFile)
	spec, err := openapi.ParseFromFile(*specFile)
	if err != nil {
		log.Fatalf("Failed to parse OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Print summary of the API
	fmt.Println("\nAPI Summary:")
	operations := spec.GetOperations()
	fmt.Printf("Found %d operations across %d paths\n", len(operations), len(spec.GetPaths()))

	for operationKey, operation := range operations {
		operationID := operation.OperationId
		if operationID == "" {
			operationID = "unknown"
		}
		fmt.Printf("- %s (OperationID: %s)\n", operationKey, operationID)
	}

	// Generate server code
	if *generateOnly {
		serverCode := generator.GenerateServerCode()
		
		if *outputFile != "" {
			// Ensure directory exists
			dir := filepath.Dir(*outputFile)
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory for output file: %v", err)
			}
			
			// Write server code to file
			if err := os.WriteFile(*outputFile, []byte(serverCode), 0644); err != nil {
				log.Fatalf("Failed to write server code to file: %v", err)
			}
			fmt.Printf("\nServer code written to %s\n", *outputFile)
		} else {
			// Print server code to stdout
			fmt.Println("\nGenerated Server Code:")
			fmt.Println(serverCode)
		}
	} else {
		// Generate and run the server
		app := generator.GenerateServer()
		
		// Start the server
		fmt.Printf("\nStarting server on port %d...\n", *port)
		fmt.Printf("Press Ctrl+C to stop the server\n")
		
		if err := app.Listen(fmt.Sprintf(":%d", *port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}
