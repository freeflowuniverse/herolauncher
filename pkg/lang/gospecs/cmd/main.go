package main

import (
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/lang/gospecs"
)

func main() {
	// Create a new GoProcessor
	processor := gospecs.NewGoProcessor()

	// Default path to test
	testPath := "/Users/despiegk/code/github/freeflowuniverse/herolauncher/pkg/herojobs"

	// Allow overriding the path from command line
	if len(os.Args) > 1 {
		testPath = os.Args[1]
	}

	// Get the spec for the Go files in the specified path
	spec, err := processor.GetSpec(testPath)
	if err != nil {
		log.Fatalf("Error processing Go files: %v", err)
	}

	// Print the spec
	fmt.Println("Go Language Specification:")
	fmt.Println("=========================")
	fmt.Println(spec)
}
