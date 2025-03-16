package main

import (
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/mcpopenapi"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: validate_openapi <path-to-openapi-spec>")
		os.Exit(1)
	}

	// Get the OpenAPI spec file path from command line arguments
	specPath := os.Args[1]
	
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
	if result == "" {
		fmt.Println("OpenAPI specification is valid. No schemas found.")
	} else {
		fmt.Println("OpenAPI validation result:")
		fmt.Println(result)
	}
}
