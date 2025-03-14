package doctree

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// Collection represents a collection of markdown pages and files
type Collection struct {
	Path string // Base path of the collection
	Name string // Name of the collection (namefixed)
}

// NewCollection creates a new Collection instance
func NewCollection(path string, name string) *Collection {
	// For compatibility with tests, apply namefix
	namefixed := tools.NameFix(name)

	return &Collection{
		Path: path,
		Name: namefixed,
	}
}

// Scan walks over the path and finds all files and .md files
// It stores the relative positions in Redis
func (c *Collection) Scan() error {
	// Key for the collection in Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)

	// Delete existing collection data if any
	redisClient.Del(ctx, collectionKey)

	// Walk through the directory
	err := filepath.Walk(c.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path from the base path
		relPath, err := filepath.Rel(c.Path, path)
		if err != nil {
			return err
		}

		// Get the filename and apply namefix
		filename := filepath.Base(path)
		namefixedFilename := tools.NameFix(filename)

		// Store in Redis using the namefixed filename as the key
		redisClient.HSet(ctx, collectionKey, namefixedFilename, relPath)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	return nil
}

// PageGet gets a page by name and returns its markdown content
func (c *Collection) PageGet(pageName string) (string, error) {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	relPath, err := redisClient.HGet(ctx, collectionKey, namefixedPageName).Result()
	if err != nil {
		return "", fmt.Errorf("page not found: %s", pageName)
	}

	// Read the file
	fullPath := filepath.Join(c.Path, relPath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read page: %w", err)
	}

	// Process includes
	markdown := string(content)
	// Skip include processing at this level to avoid infinite recursion
	// Include processing will be done at the higher level

	return markdown, nil
}

// PageSet creates or updates a page in the collection
func (c *Collection) PageSet(pageName string, content string) error {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Create the full path
	fullPath := filepath.Join(c.Path, namefixedPageName)

	// Create directories if needed
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write content to file
	err = os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	// Update Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	redisClient.HSet(ctx, collectionKey, namefixedPageName, namefixedPageName)

	return nil
}

// PageDelete deletes a page from the collection
func (c *Collection) PageDelete(pageName string) error {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	relPath, err := redisClient.HGet(ctx, collectionKey, namefixedPageName).Result()
	if err != nil {
		return fmt.Errorf("page not found: %s", pageName)
	}

	// Delete the file
	fullPath := filepath.Join(c.Path, relPath)
	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete page: %w", err)
	}

	// Remove from Redis
	redisClient.HDel(ctx, collectionKey, namefixedPageName)

	return nil
}

// PageList returns a list of all pages in the collection
func (c *Collection) PageList() ([]string, error) {
	// Get all keys from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	keys, err := redisClient.HKeys(ctx, collectionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list pages: %w", err)
	}

	// Filter to only include .md files
	pages := make([]string, 0)
	for _, key := range keys {
		if strings.HasSuffix(key, ".md") {
			pages = append(pages, key)
		}
	}

	return pages, nil
}

// FileGetUrl returns the URL for a file
func (c *Collection) FileGetUrl(fileName string) (string, error) {
	// Apply namefix to the file name
	namefixedFileName := tools.NameFix(fileName)

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	relPath, err := redisClient.HGet(ctx, collectionKey, namefixedFileName).Result()
	if err != nil {
		return "", fmt.Errorf("file not found: %s", fileName)
	}

	// Construct a URL for the file
	url := fmt.Sprintf("/collections/%s/files/%s", c.Name, relPath)

	return url, nil
}

// FileSet adds or updates a file in the collection
func (c *Collection) FileSet(fileName string, content []byte) error {
	// Apply namefix to the file name
	namefixedFileName := tools.NameFix(fileName)

	// Create the full path
	fullPath := filepath.Join(c.Path, namefixedFileName)

	// Create directories if needed
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write content to file
	err = ioutil.WriteFile(fullPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	redisClient.HSet(ctx, collectionKey, namefixedFileName, namefixedFileName)

	return nil
}

// FileDelete deletes a file from the collection
func (c *Collection) FileDelete(fileName string) error {
	// Apply namefix to the file name
	namefixedFileName := tools.NameFix(fileName)

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	relPath, err := redisClient.HGet(ctx, collectionKey, namefixedFileName).Result()
	if err != nil {
		return fmt.Errorf("file not found: %s", fileName)
	}

	// Delete the file
	fullPath := filepath.Join(c.Path, relPath)
	err = os.Remove(fullPath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Remove from Redis
	redisClient.HDel(ctx, collectionKey, namefixedFileName)

	return nil
}

// FileList returns a list of all files (non-markdown) in the collection
func (c *Collection) FileList() ([]string, error) {
	// Get all keys from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	keys, err := redisClient.HKeys(ctx, collectionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Filter to exclude .md files
	files := make([]string, 0)
	for _, key := range keys {
		if !strings.HasSuffix(key, ".md") {
			files = append(files, key)
		}
	}

	return files, nil
}

// PageGetPath returns the relative path of a page in the collection
func (c *Collection) PageGetPath(pageName string) (string, error) {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", c.Name)
	relPath, err := redisClient.HGet(ctx, collectionKey, namefixedPageName).Result()
	if err != nil {
		return "", fmt.Errorf("page not found: %s", pageName)
	}

	return relPath, nil
}

// PageGetHtml gets a page by name and returns its HTML content
func (c *Collection) PageGetHtml(pageName string) (string, error) {
	// Get the markdown content
	markdown, err := c.PageGet(pageName)
	if err != nil {
		return "", err
	}

	// Process includes
	processedMarkdown := processIncludes(markdown, c.Name, currentDocTree)

	// Convert markdown to HTML
	html := markdownToHtml(processedMarkdown)

	return html, nil
}

// Info returns information about the Collection
func (c *Collection) Info() map[string]string {
	return map[string]string{
		"name": c.Name,
		"path": c.Path,
	}
}
