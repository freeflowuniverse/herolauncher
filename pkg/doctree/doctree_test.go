package doctree

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocTree(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "doctree-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test collection structure
	collectionPath := filepath.Join(tempDir, "test-collection")
	err = os.Mkdir(collectionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create collection directory: %v", err)
	}

	// Create some markdown files
	createTestFile(t, collectionPath, "page1.md", "# Page 1\n\nThis is page 1 content.")
	createTestFile(t, collectionPath, "page2.md", "# Page 2\n\nThis is page 2 content.\n\n!!include name:'page3'")
	createTestFile(t, collectionPath, "page3.md", "# Page 3\n\nThis is page 3 content.")
	createTestFile(t, collectionPath, "special-chars-page.md", "# Special Chars Page\n\nThis page has special characters in its name.")

	// Create some other files
	createTestFile(t, collectionPath, "image.png", "fake image content")
	createTestFile(t, collectionPath, "document.pdf", "fake PDF content")

	// Create a subdirectory with more files
	subDir := filepath.Join(collectionPath, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	createTestFile(t, subDir, "subpage.md", "# Subpage\n\nThis is a page in a subdirectory.")
	createTestFile(t, subDir, "subimage.jpg", "fake JPG content")

	// Create another collection for testing cross-collection includes
	otherCollectionPath := filepath.Join(tempDir, "other-collection")
	err = os.Mkdir(otherCollectionPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create other collection directory: %v", err)
	}
	createTestFile(t, otherCollectionPath, "other-page.md", "# Other Page\n\nThis is a page from another collection.")
	createTestFile(t, collectionPath, "cross-include.md", "# Cross Include\n\n!!include name:'other-collection:other-page'")

	// Create a new DocTree instance
	dt, err := New(collectionPath, "Test Collection")
	if err != nil {
		t.Fatalf("Failed to create DocTree: %v", err)
	}

	// Test Info method
	info := dt.Info()
	if info["name"] != "test_collection" {
		t.Errorf("Expected name to be 'test_collection', got '%s'", info["name"])
	}
	if info["path"] != collectionPath {
		t.Errorf("Expected path to be '%s', got '%s'", collectionPath, info["path"])
	}

	// Test PageGet method
	page1Content, err := dt.PageGet("page1")
	if err != nil {
		t.Errorf("Failed to get page1: %v", err)
	}
	if !strings.Contains(page1Content, "This is page 1 content") {
		t.Errorf("Page1 content doesn't match expected: %s", page1Content)
	}

	// Test PageGet with special characters in name
	specialPageContent, err := dt.PageGet("special-chars-page")
	if err != nil {
		t.Errorf("Failed to get special-chars-page: %v", err)
	}
	if !strings.Contains(specialPageContent, "This page has special characters") {
		t.Errorf("Special page content doesn't match expected: %s", specialPageContent)
	}

	// Test PageGet with include directive
	page2Content, err := dt.PageGet("page2")
	if err != nil {
		t.Errorf("Failed to get page2: %v", err)
	}
	if !strings.Contains(page2Content, "This is page 2 content") || !strings.Contains(page2Content, "This is page 3 content") {
		t.Errorf("Page2 content doesn't include page3 content: %s", page2Content)
	}

	// Test PageGetHtml method
	page1Html, err := dt.PageGetHtml("page1")
	if err != nil {
		t.Errorf("Failed to get page1 HTML: %v", err)
	}
	if !strings.Contains(page1Html, "<h1>Page 1</h1>") || !strings.Contains(page1Html, "<p>This is page 1 content.</p>") {
		t.Errorf("Page1 HTML doesn't match expected: %s", page1Html)
	}

	// Test FileGetUrl method
	imageUrl, err := dt.FileGetUrl("image.png")
	if err != nil {
		t.Errorf("Failed to get image URL: %v", err)
	}
	if !strings.Contains(imageUrl, "test_collection") || !strings.Contains(imageUrl, "image.png") {
		t.Errorf("Image URL doesn't match expected: %s", imageUrl)
	}

	// Test PageGetPath method
	page1Path, err := dt.PageGetPath("page1")
	if err != nil {
		t.Errorf("Failed to get page1 path: %v", err)
	}
	if page1Path != "page1.md" {
		t.Errorf("Page1 path doesn't match expected: %s", page1Path)
	}

	// Test cross-collection include
	_, err = New(otherCollectionPath, "Other Collection")
	if err != nil {
		t.Fatalf("Failed to create other DocTree: %v", err)
	}

	crossIncludeContent, err := dt.PageGet("cross-include")
	if err != nil {
		t.Errorf("Failed to get cross-include page: %v", err)
	}
	if !strings.Contains(crossIncludeContent, "This is a page from another collection") {
		t.Errorf("Cross-include content doesn't include other page content: %s", crossIncludeContent)
	}
}

// Helper function to create a test file with content
func createTestFile(t *testing.T, dir, name, content string) {
	path := filepath.Join(dir, name)
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", name, err)
	}
}
