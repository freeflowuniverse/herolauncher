package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/mcpopenapi"
)

func main() {
	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Construct the path to the OpenAPI spec file
	specPath := filepath.Join(currentDir, "petstorev3.json")
	
	// Check if the file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		log.Fatalf("OpenAPI spec file not found at %s", specPath)
	}
	
	fmt.Printf("Validating OpenAPI spec at: %s\n", specPath)
	
	// Read the OpenAPI spec file
	specContent, err := os.ReadFile(specPath)
	if err != nil {
		log.Fatalf("Failed to read OpenAPI spec file: %v", err)
	}

	// Validate the OpenAPI spec
	result, err := mcpopenapi.ValidateOpenAPISpec(specContent)
	if err != nil {
		log.Fatalf("Error validating OpenAPI spec: %v", err)
	}

	// Print the validation result
	fmt.Println("\nOpenAPI Validation Results:")
	fmt.Println("===========================")
	if result == "" {
		fmt.Println("OpenAPI specification is valid. No schemas found.")
	} else {
		fmt.Println(result)
	}
}
