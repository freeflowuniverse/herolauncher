package doctree

import (
	"bytes"
	"context"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/tools"
	"github.com/redis/go-redis/v9"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// Redis client for the doctree package
var redisClient *redis.Client
var ctx = context.Background()
var currentCollection *Collection

// Initialize the Redis client
func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

// DocTree represents a manager for multiple collections
type DocTree struct {
	Collections       map[string]*Collection
	defaultCollection string
	// For backward compatibility
	Name string
	Path string
}

// New creates a new DocTree instance
// For backward compatibility, it also accepts path and name parameters
// to create a DocTree with a single collection
func New(args ...string) (*DocTree, error) {
	dt := &DocTree{
		Collections: make(map[string]*Collection),
	}

	// Set the global currentDocTree variable
	// This ensures that all DocTree instances can access each other's collections
	if currentDocTree == nil {
		currentDocTree = dt
	}

	// For backward compatibility with existing code
	if len(args) == 2 {
		path, name := args[0], args[1]
		// Apply namefix for compatibility with tests
		nameFixed := tools.NameFix(name)

		// Use the fixed name for the collection
		_, err := dt.AddCollection(path, nameFixed)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize DocTree: %w", err)
		}

		// For backward compatibility
		dt.defaultCollection = nameFixed
		dt.Path = path
		dt.Name = nameFixed

		// Register this collection in the global currentDocTree as well
		// This ensures that includes can find collections across different DocTree instances
		if currentDocTree != dt && !containsCollection(currentDocTree.Collections, nameFixed) {
			currentDocTree.Collections[nameFixed] = dt.Collections[nameFixed]
		}
	}

	return dt, nil
}

// Helper function to check if a collection exists in a map
func containsCollection(collections map[string]*Collection, name string) bool {
	_, exists := collections[name]
	return exists
}

// AddCollection adds a new collection to the DocTree
func (dt *DocTree) AddCollection(path string, name string) (*Collection, error) {
	// Create a new collection
	collection := NewCollection(path, name)

	// Scan the collection
	err := collection.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to scan collection: %w", err)
	}

	// Add to the collections map
	dt.Collections[collection.Name] = collection

	return collection, nil
}

// GetCollection retrieves a collection by name
func (dt *DocTree) GetCollection(name string) (*Collection, error) {
	// For compatibility with tests, apply namefix
	namefixed := tools.NameFix(name)

	// Check if the collection exists
	collection, exists := dt.Collections[namefixed]
	if !exists {
		return nil, fmt.Errorf("collection not found: %s", name)
	}

	return collection, nil
}

// DeleteCollection removes a collection from the DocTree
func (dt *DocTree) DeleteCollection(name string) error {
	// For compatibility with tests, apply namefix
	namefixed := tools.NameFix(name)

	// Check if the collection exists
	_, exists := dt.Collections[namefixed]
	if !exists {
		return fmt.Errorf("collection not found: %s", name)
	}

	// Delete from Redis
	collectionKey := fmt.Sprintf("collections:%s", namefixed)
	redisClient.Del(ctx, collectionKey)

	// Remove from the collections map
	delete(dt.Collections, namefixed)

	return nil
}

// ListCollections returns a list of all collections
func (dt *DocTree) ListCollections() []string {
	collections := make([]string, 0, len(dt.Collections))
	for name := range dt.Collections {
		collections = append(collections, name)
	}
	return collections
}

// PageGet gets a page by name from a specific collection
// For backward compatibility, if only one argument is provided, it uses the default collection
func (dt *DocTree) PageGet(args ...string) (string, error) {
	var collectionName, pageName string

	if len(args) == 1 {
		// Backward compatibility mode
		if dt.defaultCollection == "" {
			return "", fmt.Errorf("no default collection set")
		}
		collectionName = dt.defaultCollection
		pageName = args[0]
	} else if len(args) == 2 {
		collectionName = args[0]
		pageName = args[1]
	} else {
		return "", fmt.Errorf("invalid number of arguments")
	}

	// Get the collection
	collection, err := dt.GetCollection(collectionName)
	if err != nil {
		return "", err
	}

	// Set the current collection for include processing
	currentCollection = collection

	// Get the page content
	content, err := collection.PageGet(pageName)
	if err != nil {
		return "", err
	}

	// Process includes for PageGet as well
	// This is needed for the tests that check the content directly
	processedContent := processIncludes(content, collectionName, dt)

	return processedContent, nil
}

// PageGetHtml gets a page by name from a specific collection and returns its HTML content
// For backward compatibility, if only one argument is provided, it uses the default collection
func (dt *DocTree) PageGetHtml(args ...string) (string, error) {
	var collectionName, pageName string

	if len(args) == 1 {
		// Backward compatibility mode
		if dt.defaultCollection == "" {
			return "", fmt.Errorf("no default collection set")
		}
		collectionName = dt.defaultCollection
		pageName = args[0]
	} else if len(args) == 2 {
		collectionName = args[0]
		pageName = args[1]
	} else {
		return "", fmt.Errorf("invalid number of arguments")
	}

	// Get the collection
	collection, err := dt.GetCollection(collectionName)
	if err != nil {
		return "", err
	}

	// Get the HTML
	return collection.PageGetHtml(pageName)
}

// FileGetUrl returns the URL for a file in a specific collection
// For backward compatibility, if only one argument is provided, it uses the default collection
func (dt *DocTree) FileGetUrl(args ...string) (string, error) {
	var collectionName, fileName string

	if len(args) == 1 {
		// Backward compatibility mode
		if dt.defaultCollection == "" {
			return "", fmt.Errorf("no default collection set")
		}
		collectionName = dt.defaultCollection
		fileName = args[0]
	} else if len(args) == 2 {
		collectionName = args[0]
		fileName = args[1]
	} else {
		return "", fmt.Errorf("invalid number of arguments")
	}

	// Get the collection
	collection, err := dt.GetCollection(collectionName)
	if err != nil {
		return "", err
	}

	// Get the URL
	return collection.FileGetUrl(fileName)
}

// PageGetPath returns the path to a page in the default collection
// For backward compatibility
func (dt *DocTree) PageGetPath(pageName string) (string, error) {
	if dt.defaultCollection == "" {
		return "", fmt.Errorf("no default collection set")
	}

	collection, err := dt.GetCollection(dt.defaultCollection)
	if err != nil {
		return "", err
	}

	return collection.PageGetPath(pageName)
}

// Info returns information about the DocTree
// For backward compatibility
func (dt *DocTree) Info() map[string]string {
	return map[string]string{
		"name":        dt.Name,
		"path":        dt.Path,
		"collections": fmt.Sprintf("%d", len(dt.Collections)),
	}
}

// Scan scans the default collection
// For backward compatibility
func (dt *DocTree) Scan() error {
	if dt.defaultCollection == "" {
		return fmt.Errorf("no default collection set")
	}

	collection, err := dt.GetCollection(dt.defaultCollection)
	if err != nil {
		return err
	}

	return collection.Scan()
}

// markdownToHtml converts markdown content to HTML using the goldmark library
func markdownToHtml(markdown string) string {
	var buf bytes.Buffer
	// Create a new goldmark instance with default extensions
	converter := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// Convert markdown to HTML
	if err := converter.Convert([]byte(markdown), &buf); err != nil {
		// If conversion fails, return the original markdown
		return markdown
	}

	return buf.String()
}
