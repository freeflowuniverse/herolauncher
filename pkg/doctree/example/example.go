package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/freeflowuniverse/herolauncher/pkg/doctree"
)

func main() {
	// Create a sample collection directory
	collectionPath := "sample-collection"
	err := os.MkdirAll(collectionPath, 0755)
	if err != nil {
		log.Fatalf("Failed to create collection directory: %v", err)
	}

	// Create some sample markdown files
	createFile(collectionPath, "intro.md", "# Introduction\n\nThis is the introduction page.")
	createFile(collectionPath, "getting-started.md", "# Getting Started\n\nThis is the getting started guide.\n\n!!include name:'intro'")
	createFile(collectionPath, "advanced.md", "# Advanced Topics\n\nThis covers advanced topics.")

	// Create a subdirectory with more files
	subDir := filepath.Join(collectionPath, "tutorials")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create subdirectory: %v", err)
	}
	createFile(subDir, "tutorial1.md", "# Tutorial 1\n\nThis is the first tutorial.")
	createFile(subDir, "tutorial2.md", "# Tutorial 2\n\nThis is the second tutorial.")

	// Create some image files
	createFile(collectionPath, "logo.png", "fake image content")
	createFile(subDir, "diagram.jpg", "fake diagram content")

	// Create a DocTree instance
	dt, err := doctree.New(collectionPath, "Documentation")
	if err != nil {
		log.Fatalf("Failed to create DocTree: %v", err)
	}

	// Display collection info
	info := dt.Info()
	fmt.Printf("Collection Name: %s\n", info["name"])
	fmt.Printf("Collection Path: %s\n", info["path"])

	// Get and display a page
	fmt.Println("\n--- Getting a page ---")
	introContent, err := dt.PageGet("intro")
	if err != nil {
		log.Fatalf("Failed to get intro page: %v", err)
	}
	fmt.Println(introContent)

	// Get and display a page with includes
	fmt.Println("\n--- Getting a page with includes ---")
	gettingStartedContent, err := dt.PageGet("getting-started")
	if err != nil {
		log.Fatalf("Failed to get getting-started page: %v", err)
	}
	fmt.Println(gettingStartedContent)

	// Get and display a page as HTML
	fmt.Println("\n--- Getting a page as HTML ---")
	advancedHtml, err := dt.PageGetHtml("advanced")
	if err != nil {
		log.Fatalf("Failed to get advanced page as HTML: %v", err)
	}
	fmt.Println(advancedHtml)

	// Get a file URL
	fmt.Println("\n--- Getting a file URL ---")
	logoUrl, err := dt.FileGetUrl("logo.png")
	if err != nil {
		log.Fatalf("Failed to get logo URL: %v", err)
	}
	fmt.Printf("Logo URL: %s\n", logoUrl)

	// Get a page path
	fmt.Println("\n--- Getting a page path ---")
	tutorialPath, err := dt.PageGetPath("tutorials/tutorial1")
	if err != nil {
		log.Fatalf("Failed to get tutorial path: %v", err)
	}
	fmt.Printf("Tutorial path: %s\n", tutorialPath)

	// Rescan the collection
	fmt.Println("\n--- Rescanning the collection ---")
	err = dt.Scan()
	if err != nil {
		log.Fatalf("Failed to rescan collection: %v", err)
	}
	fmt.Println("Collection rescanned successfully")

	// Clean up
	fmt.Println("\n--- Cleaning up ---")
	err = os.RemoveAll(collectionPath)
	if err != nil {
		log.Fatalf("Failed to clean up: %v", err)
	}
	fmt.Println("Cleanup completed")
}

// Helper function to create a file with content
func createFile(dir, name, content string) {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", name, err)
	}
}
