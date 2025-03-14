package doctree

import (
	"context"
	"fmt"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/tools"
	"github.com/redis/go-redis/v9"
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
	Collections map[string]*Collection
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
	currentDocTree = dt

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
	}

	return dt, nil
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
	
	// Get the page
	return collection.PageGet(pageName)
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
		"name": dt.Name,
		"path": dt.Path,
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

// Simple markdown to HTML converter
// In a real application, you would use a proper library
func markdownToHtml(markdown string) string {
	// This is a very basic implementation
	// Replace headers
	html := markdown
	html = strings.ReplaceAll(html, "# ", "<h1>")
	html = strings.ReplaceAll(html, "\n# ", "\n<h1>")
	html = strings.ReplaceAll(html, "## ", "<h2>")
	html = strings.ReplaceAll(html, "\n## ", "\n<h2>")
	html = strings.ReplaceAll(html, "### ", "<h3>")
	html = strings.ReplaceAll(html, "\n### ", "\n<h3>")

	// Add closing tags for headers
	lines := strings.Split(html, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "<h1>") {
			lines[i] = line + "</h1>"
		} else if strings.HasPrefix(line, "<h2>") {
			lines[i] = line + "</h2>"
		} else if strings.HasPrefix(line, "<h3>") {
			lines[i] = line + "</h3>"
		}
	}

	// Convert paragraphs
	html = strings.Join(lines, "\n")
	paragraphs := strings.Split(html, "\n\n")
	for i, p := range paragraphs {
		if !strings.HasPrefix(p, "<h") && p != "" {
			paragraphs[i] = "<p>" + p + "</p>"
		}
	}

	return strings.Join(paragraphs, "\n\n")
}
