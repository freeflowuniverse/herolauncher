package doctree

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/freeflowuniverse/herolauncher/internal/tools"
)

// DocTree represents a collection of markdown pages and files
type DocTree struct {
	Path string // Base path of the collection
	Name string // Name of the collection (namefixed)
}

// New creates a new DocTree instance and initializes it
func New(path string, name string) (*DocTree, error) {
	// Apply namefix to the collection name
	namefixed := tools.NameFix(name)

	dt := &DocTree{
		Path: path,
		Name: namefixed,
	}

	// Initialize the collection by scanning the path
	err := dt.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize DocTree: %w", err)
	}

	return dt, nil
}

// Scan walks over the path and finds all files and .md files
// It stores the relative positions in Redis
func (dt *DocTree) Scan() error {
	// Key for the collection in Redis
	collectionKey := fmt.Sprintf("collections:%s", dt.Name)

	// Delete existing collection data if any
	redisClient.Del(collectionKey)

	// Walk through the directory
	err := filepath.Walk(dt.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path from the base path
		relPath, err := filepath.Rel(dt.Path, path)
		if err != nil {
			return err
		}

		// Get the filename
		filename := filepath.Base(path)

		// Apply namefix to the filename
		namefixedFilename := tools.NameFix(filename)

		// Store in Redis
		redisClient.HSet(collectionKey, namefixedFilename, relPath)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	return nil
}

// PageGet gets a page by name and returns its markdown content
func (dt *DocTree) PageGet(pageName string) (string, error) {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", dt.Name)
	relPath, ok := redisClient.HGet(collectionKey, namefixedPageName)
	if !ok {
		return "", fmt.Errorf("page not found: %s", pageName)
	}

	// Read the file
	fullPath := filepath.Join(dt.Path, relPath)
	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read page: %w", err)
	}

	// Process includes
	markdown := string(content)
	markdown = dt.processIncludes(markdown)

	return markdown, nil
}

// processIncludes handles the !!include directives in markdown
func (dt *DocTree) processIncludes(content string) string {
	// Find all !!include directives
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		if strings.Contains(line, "!!include name:") {
			// Extract the page name
			startIdx := strings.Index(line, "!!include name:'") + len("!!include name:'")
			endIdx := strings.LastIndex(line, "'")
			if startIdx > 0 && endIdx > startIdx {
				includeName := line[startIdx:endIdx]

				// Check if it's from another collection
				var includeContent string
				var err error

				if strings.Contains(includeName, ":") {
					// Format: othercollection:pagename
					parts := strings.SplitN(includeName, ":", 2)
					if len(parts) == 2 {
						collectionName := parts[0]
						pageName := parts[1]

						// Create a temporary DocTree for the other collection
						otherDT := &DocTree{
							Path: dt.Path, // Assuming same base path
							Name: tools.NameFix(collectionName),
						}
						includeContent, err = otherDT.PageGet(pageName)
					} else {
						err = fmt.Errorf("invalid include format: %s", includeName)
					}
				} else {
					// Include from the same collection
					includeContent, err = dt.PageGet(includeName)
				}

				if err == nil {
					// Add the included content
					result = append(result, includeContent)
				} else {
					// Keep the original line if there was an error
					result = append(result, line)
				}
			} else {
				// Invalid format, keep the original line
				result = append(result, line)
			}
		} else {
			// Not an include directive, keep the line
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// PageGetHtml gets a page by name and returns its HTML content
func (dt *DocTree) PageGetHtml(pageName string) (string, error) {
	// Get the markdown content
	markdown, err := dt.PageGet(pageName)
	if err != nil {
		return "", err
	}

	// Convert markdown to HTML
	// This is a simple implementation - in a real application,
	// you would use a proper markdown to HTML converter
	html := markdownToHtml(markdown)

	return html, nil
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

// FileGetUrl returns the URL for a file
func (dt *DocTree) FileGetUrl(fileName string) (string, error) {
	// Apply namefix to the file name
	namefixedFileName := tools.NameFix(fileName)

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", dt.Name)
	relPath, ok := redisClient.HGet(collectionKey, namefixedFileName)
	if !ok {
		return "", fmt.Errorf("file not found: %s", fileName)
	}

	// Construct a URL for the file
	// This is a simple implementation - in a real application,
	// you would use a proper URL construction based on your web server
	url := fmt.Sprintf("/collections/%s/files/%s", dt.Name, relPath)

	return url, nil
}

// PageGetPath returns the relative path of a page in the collection
func (dt *DocTree) PageGetPath(pageName string) (string, error) {
	// Apply namefix to the page name
	namefixedPageName := tools.NameFix(pageName)

	// Ensure it has .md extension
	if !strings.HasSuffix(namefixedPageName, ".md") {
		namefixedPageName += ".md"
	}

	// Get the relative path from Redis
	collectionKey := fmt.Sprintf("collections:%s", dt.Name)
	relPath, ok := redisClient.HGet(collectionKey, namefixedPageName)
	if !ok {
		return "", fmt.Errorf("page not found: %s", pageName)
	}

	return relPath, nil
}

// Info returns information about the DocTree
func (dt *DocTree) Info() map[string]string {
	return map[string]string{
		"name": dt.Name,
		"path": dt.Path,
	}
}
