package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/lang/rustspecs"
)

func main() {
	// Create a new RustProcessor
	processor := rustspecs.NewRustProcessor()

	// Default path to test
	testPath := "~/code/github/freeflowuniverse/herolauncher/pkg/herojobs/rustclient"

	// Allow overriding the path from command line
	if len(os.Args) > 1 {
		testPath = os.Args[1]
	}

	// Expand ~ to the home directory
	if strings.HasPrefix(testPath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		testPath = filepath.Join(homeDir, testPath[1:])
	}

	// Get the spec for the Rust files in the specified path
	spec, err := processor.GetSpec(testPath)
	if err != nil {
		log.Fatalf("Error processing Rust files: %v", err)
	}

	// Print the spec
	fmt.Println("Rust Language Specification:")
	fmt.Println("===========================")
	fmt.Println(spec)
}
