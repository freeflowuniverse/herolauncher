package main

import (
	"fmt"
	"log"
	"os"

	"github.com/freeflowuniverse/herolauncher/pkg/vlang/vlangspecs"
)

func main() {
	// Create a new VlangProcessor
	processor := vlangspecs.NewVlangProcessor()

	// Default path to test
	testPath := "/Users/despiegk/code/github/freeflowuniverse/herolib/lib/circles"

	// Allow overriding the path from command line
	if len(os.Args) > 1 {
		testPath = os.Args[1]
	}

	// Get the spec for the V files in the specified path
	spec, err := processor.GetSpec(testPath)
	if err != nil {
		log.Fatalf("Error processing V files: %v", err)
	}

	// Print the spec
	fmt.Println("V Language Specification:")
	fmt.Println("=========================")
	fmt.Println(spec)
}
