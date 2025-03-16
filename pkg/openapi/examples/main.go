package main

import (
	"fmt"
	"log"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
)

func main() {
	// Path to the OpenAPI specification file
	specFile := "petstore.json"

	// Parse the OpenAPI specification
	fmt.Printf("Parsing OpenAPI specification from %s...\n", specFile)
	spec, err := openapi.ParseFromFile(specFile)
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

	// Generate and run the server
	app := generator.GenerateServer()
	
	// Start the server
	port := 8080
	fmt.Printf("\nStarting server on port %d...\n", port)
	fmt.Printf("Press Ctrl+C to stop the server\n")
	
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
