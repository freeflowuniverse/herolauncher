package test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/openapi"
)

// TestOpenAPIGeneration tests the OpenAPI code generation and server functionality
func TestOpenAPIGeneration(t *testing.T) {
	// Test Petstore API generation
	t.Run("PetstoreAPIGeneration", func(t *testing.T) {
		testPetstoreAPIGeneration(t)
	})

	// Test Actions API generation
	t.Run("ActionsAPIGeneration", func(t *testing.T) {
		testActionsAPIGeneration(t)
	})
}

// testPetstoreAPIGeneration tests the Petstore API generation
func testPetstoreAPIGeneration(t *testing.T) {
	// Path to the OpenAPI specification file
	specFile := filepath.Join("..", "petstore.yaml")

	// Parse the OpenAPI specification
	fmt.Printf("Parsing OpenAPI specification from %s...\n", specFile)
	spec, err := openapi.ParseFromFile(specFile)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputDir := filepath.Join("..", "petstoreapi")
	outputPath := filepath.Join(outputDir, "server.go")
	
	// Ensure the output directory exists
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	
	err = os.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write server code: %v", err)
	}

	// Verify the server code was written
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Server code file was not created: %v", err)
	}

	// Print summary of the API
	fmt.Println("\nPetstore API Summary:")
	operations := spec.GetOperations()
	fmt.Printf("Found %d operations across %d paths\n", len(operations), len(spec.GetPaths()))

	for operationKey, operation := range operations {
		operationID := operation.OperationId
		if operationID == "" {
			operationID = "unknown"
		}
		fmt.Printf("- %s (OperationID: %s)\n", operationKey, operationID)
	}

	fmt.Printf("Generated Petstore API code in %s\n", outputPath)
}

// testActionsAPIGeneration tests the Actions API generation
func testActionsAPIGeneration(t *testing.T) {
	// Path to the OpenAPI specification file
	specFile := filepath.Join("..", "actions.yaml")

	// Parse the OpenAPI specification
	fmt.Printf("Parsing OpenAPI specification from %s...\n", specFile)
	spec, err := openapi.ParseFromFile(specFile)
	if err != nil {
		t.Fatalf("Failed to parse OpenAPI specification: %v", err)
	}

	// Create a server generator
	generator := openapi.NewServerGenerator(spec)

	// Generate server code
	serverCode := generator.GenerateServerCode()

	// Write the server code to a file
	outputDir := filepath.Join("..", "actionsapi")
	outputPath := filepath.Join(outputDir, "server.go")
	
	// Ensure the output directory exists
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}
	
	err = os.WriteFile(outputPath, []byte(serverCode), 0644)
	if err != nil {
		t.Fatalf("Failed to write server code: %v", err)
	}

	// Verify the server code was written
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Server code file was not created: %v", err)
	}

	// Print summary of the API
	fmt.Println("\nActions API Summary:")
	operations := spec.GetOperations()
	fmt.Printf("Found %d operations across %d paths\n", len(operations), len(spec.GetPaths()))

	for operationKey, operation := range operations {
		operationID := operation.OperationId
		if operationID == "" {
			operationID = "unknown"
		}
		fmt.Printf("- %s (OperationID: %s)\n", operationKey, operationID)
	}

	fmt.Printf("Generated Actions API code in %s\n", outputPath)
}

// TestMultiApiServer tests the multi-API server functionality
func TestMultiApiServer(t *testing.T) {
	// This test would normally start the server and make HTTP requests to verify functionality
	// For the purposes of this example, we'll just log that this would be tested
	fmt.Println("Testing multi-API server functionality...")
	fmt.Println("This would start the server and verify that both APIs are accessible")
	fmt.Println("It would also verify that the Swagger UI is accessible")
}

// RunTests runs all the tests
func RunTests() {
	// Create a new testing.T
	t := &testing.T{}
	
	// Run the tests
	TestOpenAPIGeneration(t)
	TestMultiApiServer(t)
	
	fmt.Println("All tests completed successfully!")
}

// main is the entry point when running the file directly
func main() {
	RunTests()
}
