package doctree

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestDocTree(t *testing.T) {
	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Default Redis address
		Password: "",               // No password
		DB:       0,                  // Default DB
	})
	ctx := context.Background()

	// Check if Redis is running
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Redis server is not running: %v", err)
	}

	// Define the path to the sample collection
	collectionPath, err := filepath.Abs("example/sample-collection")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create doctree instance
	dt, err := New(collectionPath, "sample-collection")
	if err != nil {
		t.Fatalf("Failed to create DocTree: %v", err)
	}

	// Verify the doctree was initialized correctly
	if dt.Name != "sample_collection" {
		t.Errorf("Expected name to be 'sample_collection', got '%s'", dt.Name)
	}

	// Check if the collection exists in Redis
	collectionKey := "collections:sample_collection"
	exists, err := rdb.Exists(ctx, collectionKey).Result()
	if err != nil {
		t.Fatalf("Failed to check if collection exists: %v", err)
	}
	if exists == 0 {
		t.Errorf("Collection key '%s' does not exist in Redis", collectionKey)
	}

	// Print all entries in Redis for debugging
	allEntries, err := rdb.HGetAll(ctx, collectionKey).Result()
	if err != nil {
		t.Fatalf("Failed to get entries from Redis: %v", err)
	}

	t.Logf("Found %d entries in Redis for collection '%s'", len(allEntries), collectionKey)
	for key, value := range allEntries {
		t.Logf("Redis entry: key='%s', value='%s'", key, value)
	}

	// Check that the expected files are stored in Redis
	// The keys in Redis are the namefixed filenames without path structure
	expectedFilesMap := map[string]string{
		"advanced.md":        "advanced.md",
		"getting_started.md": "Getting- starteD.md",
		"intro.md":           "intro.md",
		"logo.png":           "logo.png",
		"diagram.jpg":        "tutorials/diagram.jpg",
		"tutorial1.md":       "tutorials/tutorial1.md",
		"tutorial2.md":       "tutorials/tutorial2.md",
	}

	// Check each expected file
	for key, expectedPath := range expectedFilesMap {
		// Get the relative path from Redis
		relPath, err := rdb.HGet(ctx, collectionKey, key).Result()
		if err != nil {
			t.Errorf("File with key '%s' not found in Redis: %v", key, err)
			continue
		}

		t.Logf("Found file '%s' in Redis with path '%s'", key, relPath)

		// Verify the path is correct
		if relPath != expectedPath {
			t.Errorf("Expected path '%s' for key '%s', got '%s'", expectedPath, key, relPath)
		}
	}

	// Directly check if we can get the intro.md key from Redis
	introContent, err := rdb.HGet(ctx, collectionKey, "intro.md").Result()
	if err != nil {
		t.Errorf("Failed to get 'intro.md' directly from Redis: %v", err)
	} else {
		t.Logf("Successfully got 'intro.md' directly from Redis: %s", introContent)
	}

	// Test PageGet function
	content, err := dt.PageGet("intro")
	if err != nil {
		t.Errorf("Failed to get page 'intro': %v", err)
	} else {
		if !strings.Contains(content, "Introduction") {
			t.Errorf("Expected 'Introduction' in content, got '%s'", content)
		}
	}

	// Test PageGetHtml function
	html, err := dt.PageGetHtml("intro")
	if err != nil {
		t.Errorf("Failed to get HTML for page 'intro': %v", err)
	} else {
		if !strings.Contains(html, "<h1>Introduction") {
			t.Errorf("Expected '<h1>Introduction' in HTML, got '%s'", html)
		}
	}

	// Test FileGetUrl function
	url, err := dt.FileGetUrl("logo.png")
	if err != nil {
		t.Errorf("Failed to get URL for file 'logo.png': %v", err)
	} else {
		if !strings.Contains(url, "sample_collection") || !strings.Contains(url, "logo.png") {
			t.Errorf("Expected URL to contain 'sample_collection' and 'logo.png', got '%s'", url)
		}
	}

	// Test PageGetPath function
	path, err := dt.PageGetPath("intro")
	if err != nil {
		t.Errorf("Failed to get path for page 'intro': %v", err)
	} else {
		if path != "intro.md" {
			t.Errorf("Expected path to be 'intro.md', got '%s'", path)
		}
	}

	// Test Info function
	info := dt.Info()
	if info["name"] != "sample_collection" {
		t.Errorf("Expected name to be 'sample_collection', got '%s'", info["name"])
	}
	if info["path"] != collectionPath {
		t.Errorf("Expected path to be '%s', got '%s'", collectionPath, info["path"])
	}
}
