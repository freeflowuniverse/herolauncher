package doctree

import (
	"fmt"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// Global variable to track the current DocTree instance
var currentDocTree *DocTree

// processIncludeLine processes a single line for include directives
// Returns collectionName and pageName if found, or empty strings if not an include directive
//
// Supports:
// !!include collectionname:'pagename'
// !!include collectionname:'pagename.md'
// !!include 'pagename'
// !!include collectionname:pagename
// !!include collectionname:pagename.md
// !!include name:'pagename'
// !!include pagename
func parseIncludeLine(line string) (string, string, error) {
	// Check if the line contains an include directive
	if !strings.Contains(line, "!!include") {
		return "", "", nil
	}

	// Extract the part after !!include
	parts := strings.SplitN(line, "!!include", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed include directive: %s", line)
	}

	// Trim spaces and check if the include part is empty
	includeText := tools.TrimSpacesAndQuotes(parts[1])
	if includeText == "" {
		return "", "", fmt.Errorf("empty include directive: %s", line)
	}

	// Remove name: prefix if present
	if strings.HasPrefix(includeText, "name:") {
		includeText = strings.TrimSpace(strings.TrimPrefix(includeText, "name:"))
		if includeText == "" {
			return "", "", fmt.Errorf("empty page name after 'name:' prefix: %s", line)
		}
	}

	// Check if it contains a collection reference (has a colon)
	if strings.Contains(includeText, ":") {
		parts := strings.SplitN(includeText, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("malformed collection reference: %s", includeText)
		}

		collectionName := tools.NameFix(parts[0])
		pageName := tools.NameFix(parts[1])

		if collectionName == "" {
			return "", "", fmt.Errorf("empty collection name in include directive: %s", line)
		}

		if pageName == "" {
			return "", "", fmt.Errorf("empty page name in include directive: %s", line)
		}

		return collectionName, pageName, nil
	}

	return "", includeText, nil
}

// processIncludes handles all the different include directive formats in markdown
func processIncludes(content string, currentCollectionName string, dt *DocTree) string {

	// Find all include directives
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		collectionName, pageName, err := parseIncludeLine(line)
		if err != nil {
			errorMsg := fmt.Sprintf(">>ERROR: Failed to process include directive: %v", err)
			result = append(result, errorMsg)
			continue
		}

		if collectionName == "" && pageName == "" {
			// Not an include directive, keep the line
			result = append(result, line)
		} else {
			includeContent := ""
			var includeErr error

			// If no collection specified, use the current collection
			if collectionName == "" {
				collectionName = currentCollectionName
			}

			// Process the include
			includeContent, includeErr = handleInclude(pageName, collectionName, dt)

			if includeErr != nil {
				errorMsg := fmt.Sprintf(">>ERROR: %v", includeErr)
				result = append(result, errorMsg)
			} else {
				// Process any nested includes in the included content
				processedIncludeContent := processIncludes(includeContent, collectionName, dt)
				result = append(result, processedIncludeContent)
			}
		}
	}

	return strings.Join(result, "\n")
}

// handleInclude processes the include directive with the given page name and optional collection name
func handleInclude(pageName, collectionName string, dt *DocTree) (string, error) {
	// Check if it's from another collection
	if collectionName != "" {
		// Format: othercollection:pagename
		namefixedCollectionName := tools.NameFix(collectionName)

		// Remove .md extension if present for the API call
		namefixedPageName := tools.NameFix(pageName)
		namefixedPageName = strings.TrimSuffix(namefixedPageName, ".md")

		// Try to get the collection from the DocTree
		// First check if the collection exists in the current DocTree
		otherCollection, err := dt.GetCollection(namefixedCollectionName)
		if err != nil {
			// If not found in the current DocTree, check the global currentDocTree
			if currentDocTree != nil && currentDocTree != dt {
				otherCollection, err = currentDocTree.GetCollection(namefixedCollectionName)
				if err != nil {
					return "", fmt.Errorf("cannot include from non-existent collection: %s", collectionName)
				}
			} else {
				return "", fmt.Errorf("cannot include from non-existent collection: %s", collectionName)
			}
		}

		// Get the page content using the collection's PageGet method
		content, err := otherCollection.PageGet(namefixedPageName)
		if err != nil {
			return "", fmt.Errorf("cannot include non-existent page: %s from collection: %s", pageName, collectionName)
		}

		return content, nil
	} else {
		// For same collection includes, we need to get the current collection
		currentCollection, err := dt.GetCollection(dt.defaultCollection)
		if err != nil {
			return "", fmt.Errorf("failed to get current collection: %w", err)
		}

		// Include from the same collection
		// Remove .md extension if present for the API call
		namefixedPageName := tools.NameFix(pageName)
		namefixedPageName = strings.TrimSuffix(namefixedPageName, ".md")

		// Use the current collection to get the page content
		content, err := currentCollection.PageGet(namefixedPageName)
		if err != nil {
			return "", fmt.Errorf("cannot include non-existent page: %s", pageName)
		}

		return content, nil
	}
}
