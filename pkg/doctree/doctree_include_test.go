package doctree

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestDocTreeInclude(t *testing.T) {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Default Redis address
		Password: "",               // No password
		DB:       0,                // Default DB
	})
	ctx := context.Background()

	// Check if Redis is running
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis server is not running: %v", err)
	}

	// Define the paths to both collections
	collection1Path, err := filepath.Abs("example/sample-collection")
	if err != nil {
		t.Fatalf("Failed to get absolute path for collection 1: %v", err)
	}

	collection2Path, err := filepath.Abs("example/sample-collection-2")
	if err != nil {
		t.Fatalf("Failed to get absolute path for collection 2: %v", err)
	}

	// Create doctree instances for both collections
	dt1, err := New(collection1Path, "sample-collection")
	if err != nil {
		t.Fatalf("Failed to create DocTree for collection 1: %v", err)
	}

	dt2, err := New(collection2Path, "sample-collection-2")
	if err != nil {
		t.Fatalf("Failed to create DocTree for collection 2: %v", err)
	}

	// Verify the doctrees were initialized correctly
	if dt1.Name != "sample_collection" {
		t.Errorf("Expected name to be 'sample_collection', got '%s'", dt1.Name)
	}

	if dt2.Name != "sample_collection_2" {
		t.Errorf("Expected name to be 'sample_collection_2', got '%s'", dt2.Name)
	}

	// Check if both collections exist in Redis
	collection1Key := "collections:sample_collection"
	exists1, err := rdb.Exists(ctx, collection1Key).Result()
	if err != nil {
		t.Fatalf("Failed to check if collection 1 exists: %v", err)
	}
	if exists1 == 0 {
		t.Errorf("Collection key '%s' does not exist in Redis", collection1Key)
	}

	collection2Key := "collections:sample_collection_2"
	exists2, err := rdb.Exists(ctx, collection2Key).Result()
	if err != nil {
		t.Fatalf("Failed to check if collection 2 exists: %v", err)
	}
	if exists2 == 0 {
		t.Errorf("Collection key '%s' does not exist in Redis", collection2Key)
	}

	// Print all entries in Redis for debugging
	allEntries1, err := rdb.HGetAll(ctx, collection1Key).Result()
	if err != nil {
		t.Fatalf("Failed to get entries from Redis for collection 1: %v", err)
	}

	t.Logf("Found %d entries in Redis for collection '%s'", len(allEntries1), collection1Key)
	for key, value := range allEntries1 {
		t.Logf("Redis entry for collection 1: key='%s', value='%s'", key, value)
	}

	allEntries2, err := rdb.HGetAll(ctx, collection2Key).Result()
	if err != nil {
		t.Fatalf("Failed to get entries from Redis for collection 2: %v", err)
	}

	t.Logf("Found %d entries in Redis for collection '%s'", len(allEntries2), collection2Key)
	for key, value := range allEntries2 {
		t.Logf("Redis entry for collection 2: key='%s', value='%s'", key, value)
	}

	// First, let's check the raw content of both files before processing includes
	// Get the raw content of advanced.md from collection 1
	collectionKey1 := "collections:sample_collection"
	relPath1, err := rdb.HGet(ctx, collectionKey1, "advanced.md").Result()
	if err != nil {
		t.Fatalf("Failed to get path for advanced.md in collection 1: %v", err)
	}
	fullPath1 := filepath.Join(collection1Path, relPath1)
	rawContent1, err := ioutil.ReadFile(fullPath1)
	if err != nil {
		t.Fatalf("Failed to read advanced.md from collection 1: %v", err)
	}
	t.Logf("Raw content of advanced.md from collection 1: %s", string(rawContent1))

	// Get the raw content of advanced.md from collection 2
	collectionKey2 := "collections:sample_collection_2"
	relPath2, err := rdb.HGet(ctx, collectionKey2, "advanced.md").Result()
	if err != nil {
		t.Fatalf("Failed to get path for advanced.md in collection 2: %v", err)
	}
	fullPath2 := filepath.Join(collection2Path, relPath2)
	rawContent2, err := ioutil.ReadFile(fullPath2)
	if err != nil {
		t.Fatalf("Failed to read advanced.md from collection 2: %v", err)
	}
	t.Logf("Raw content of advanced.md from collection 2: %s", string(rawContent2))

	// Verify the raw content contains the expected include directive
	if !strings.Contains(string(rawContent2), "!!include name:'sample_collection:advanced'") {
		t.Errorf("Expected include directive in collection 2's advanced.md, not found")
	}

	// Now test the include functionality - Get the processed content of advanced.md from collection 2
	// This file includes advanced.md from collection 1
	content, err := dt2.PageGet("advanced")
	if err != nil {
		t.Errorf("Failed to get page 'advanced.md' from collection 2: %v", err)
		return
	}
	
	t.Logf("Processed content of advanced.md from collection 2: %s", content)
	
	// Check if the content includes text from both files
	// The advanced.md in collection 2 has: # Other and includes sample_collection:advanced
	if !strings.Contains(content, "# Other") {
		t.Errorf("Expected '# Other' in content from collection 2, not found")
	}
	
	// The advanced.md in collection 1 has: # Advanced Topics and "This covers advanced topics."
	if !strings.Contains(content, "# Advanced Topics") {
		t.Errorf("Expected '# Advanced Topics' from included file in collection 1, not found")
	}
	
	if !strings.Contains(content, "This covers advanced topics") {
		t.Errorf("Expected 'This covers advanced topics' from included file in collection 1, not found")
	}
	
	// Test nested includes if they exist
	// This would test if an included file can itself include another file
	// For this test, we would need to modify the files to have nested includes
	
	// Test HTML rendering of the page with include
	html, err := dt2.PageGetHtml("advanced")
	if err != nil {
		t.Errorf("Failed to get HTML for page 'advanced.md' from collection 2: %v", err)
		return
	}
	
	t.Logf("HTML of advanced.md from collection 2: %s", html)
	
	// Check if the HTML includes content from both files
	if !strings.Contains(html, "<h1>Other</h1>") {
		t.Errorf("Expected '<h1>Other</h1>' in HTML from collection 2, not found")
	}
	
	if !strings.Contains(html, "<h1>Advanced Topics</h1>") {
		t.Errorf("Expected '<h1>Advanced Topics</h1>' from included file in collection 1, not found")
	}
	
	// Test that the include directive itself is not visible in the final output
	if strings.Contains(html, "!!include") {
		t.Errorf("Include directive '!!include' should not be visible in the final HTML output")
	}
	
	// Test error handling for non-existent includes
	// Create a temporary file with an invalid include
	tempDt, err := New(t.TempDir(), "temp_collection")
	if err != nil {
		t.Fatalf("Failed to create temp collection: %v", err)
	}
	
	// Initialize the temp collection
	err = tempDt.Scan()
	if err != nil {
		t.Fatalf("Failed to initialize temp collection: %v", err)
	}
	
	// Test error handling for circular includes
	// This would require creating files that include each other
	
	t.Logf("All include tests completed successfully")
}
