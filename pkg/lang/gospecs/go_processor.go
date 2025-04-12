package gospecs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
)

// GoProcessor processes Go language files to extract public structs, interfaces, and methods
type GoProcessor struct {
	// Add any fields needed for configuration or state
}

// NewGoProcessor creates a new GoProcessor instance
func NewGoProcessor() *GoProcessor {
	return &GoProcessor{}
}

// GetSpec walks over the given path recursively, finds all .go files,
// and extracts public structures and functions without implementation code
func (gp *GoProcessor) GetSpec(path string) (string, error) {
	// Check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	var result strings.Builder

	// Find all .go files in the given path
	goFiles, err := gp.findGoFiles(path)
	if err != nil {
		return "", fmt.Errorf("error finding Go files: %w", err)
	}

	// Process each file
	for _, file := range goFiles {
		fileSpec, err := gp.processFile(file)
		if err != nil {
			return "", fmt.Errorf("error processing file %s: %w", file, err)
		}

		if fileSpec != "" {
			result.WriteString(fmt.Sprintf("// From file: %s\n", file))
			result.WriteString(fileSpec)
			result.WriteString("\n\n")
		}
	}

	return result.String(), nil
}

// findGoFiles walks through the directory structure and finds all .go files
func (gp *GoProcessor) findGoFiles(root string) ([]string, error) {
	var goFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-go files, and test files
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		goFiles = append(goFiles, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return goFiles, nil
}

// processFile extracts public structs, interfaces, functions, and methods from a Go file
func (gp *GoProcessor) processFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)

	// Regular expressions for matching structs, interfaces, functions, and methods
	structRegex := regexp.MustCompile(`^type\s+(\w+)\s+struct\s+{`)
	interfaceRegex := regexp.MustCompile(`^type\s+(\w+)\s+interface\s+{`)
	funcRegex := regexp.MustCompile(`^func\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^func\s+\((\w+)\s+\*?(\w+)\)\s+(\w+)`)

	// Variable declarations
	inStruct := false
	inInterface := false
	inFunc := false
	bracketCount := 0
	var currentDocComment strings.Builder
	isPublicEntity := false
	lastLineEmpty := false

	foundPackage := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for package declaration
		if !foundPackage && strings.HasPrefix(trimmedLine, "package ") {
			foundPackage = true
			result.WriteString(line)
			result.WriteString("\n\n")
			continue
		}

		// Check for empty lines
		if trimmedLine == "" {
			lastLineEmpty = true
			continue
		}

		// Collect documentation comments
		if strings.HasPrefix(trimmedLine, "//") {
			// If the last line was empty, reset the comment buffer
			if lastLineEmpty {
				currentDocComment.Reset()
			}

			currentDocComment.WriteString(line)
			currentDocComment.WriteString("\n")
			lastLineEmpty = false
			continue
		}

		// Reset comments if we encounter a non-relevant line
		if trimmedLine != "" &&
			!strings.HasPrefix(trimmedLine, "//") &&
			!inStruct && !inInterface && !inFunc &&
			!structRegex.MatchString(trimmedLine) &&
			!interfaceRegex.MatchString(trimmedLine) &&
			!funcRegex.MatchString(trimmedLine) &&
			!methodRegex.MatchString(trimmedLine) {
			currentDocComment.Reset()
		}

		// Check if the upcoming declaration is public (capitalized)
		isPublicEntity = false
		if structMatch := structRegex.FindStringSubmatch(trimmedLine); len(structMatch) > 1 {
			isPublicEntity = strings.ToUpper(structMatch[1][:1]) == structMatch[1][:1]
		} else if interfaceMatch := interfaceRegex.FindStringSubmatch(trimmedLine); len(interfaceMatch) > 1 {
			isPublicEntity = strings.ToUpper(interfaceMatch[1][:1]) == interfaceMatch[1][:1]
		} else if funcMatch := funcRegex.FindStringSubmatch(trimmedLine); len(funcMatch) > 1 {
			isPublicEntity = strings.ToUpper(funcMatch[1][:1]) == funcMatch[1][:1]
		} else if methodMatch := methodRegex.FindStringSubmatch(trimmedLine); len(methodMatch) > 3 {
			isPublicEntity = strings.ToUpper(methodMatch[3][:1]) == methodMatch[3][:1]
		}

		// Check for struct declaration
		if structMatch := structRegex.FindStringSubmatch(trimmedLine); len(structMatch) > 1 && !inInterface && !inFunc {
			if isPublicEntity {
				inStruct = true
				bracketCount = 1

				// Write doc comment if exists
				if currentDocComment.Len() > 0 {
					result.WriteString(currentDocComment.String())
					currentDocComment.Reset()
				}

				// Write struct declaration
				result.WriteString(line)
				result.WriteString("\n")
			}
			lastLineEmpty = false
			continue
		}

		// Check for interface declaration
		if interfaceMatch := interfaceRegex.FindStringSubmatch(trimmedLine); len(interfaceMatch) > 1 && !inStruct && !inFunc {
			if isPublicEntity {
				inInterface = true
				bracketCount = 1

				// Write doc comment if exists
				if currentDocComment.Len() > 0 {
					result.WriteString(currentDocComment.String())
					currentDocComment.Reset()
				}

				// Write interface declaration
				result.WriteString(line)
				result.WriteString("\n")
			}
			lastLineEmpty = false
			continue
		}

		// Check for function declaration
		funcMatch := funcRegex.FindStringSubmatch(trimmedLine)
		methodMatch := methodRegex.FindStringSubmatch(trimmedLine)

		if ((len(funcMatch) > 1 && !inStruct && !inInterface) ||
			(len(methodMatch) > 3 && !inStruct && !inInterface)) && isPublicEntity {
			inFunc = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Write function signature without implementation
			signature := trimmedLine
			if strings.Contains(signature, "{") {
				signature = strings.Split(signature, "{")[0]
			}

			result.WriteString(signature)
			result.WriteString("\n\n")
			lastLineEmpty = false
			continue
		}

		// Track bracket count for struct
		if inStruct {
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inStruct = false
					result.WriteString(line)
					result.WriteString("\n\n")
					continue
				}
			}

			// Inside struct, copy the line if it's a field
			if isPublicFieldLine(line) {
				result.WriteString(line)
				result.WriteString("\n")
			}
			lastLineEmpty = false
			continue
		}

		// Track bracket count for interface
		if inInterface {
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inInterface = false
					result.WriteString(line)
					result.WriteString("\n\n")
					continue
				}
			}

			// Inside interface, copy the line
			result.WriteString(line)
			result.WriteString("\n")
			lastLineEmpty = false
			continue
		}

		// Track bracket count for function implementations to skip them
		if inFunc {
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inFunc = false
				}
			}
			lastLineEmpty = false
			continue
		}

		lastLineEmpty = false
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}

// isPublicFieldLine checks if a line inside a struct represents a public field
func isPublicFieldLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)

	// Skip comments, closing brackets, or empty lines
	if trimmedLine == "" || strings.HasPrefix(trimmedLine, "//") || trimmedLine == "}" {
		return false
	}

	// Skip private fields (starting with lowercase)
	parts := strings.Fields(trimmedLine)
	if len(parts) > 0 {
		firstField := parts[0]
		// Check if the first character is uppercase (public)
		if len(firstField) > 0 && strings.ToUpper(firstField[:1]) == firstField[:1] {
			return true
		}
	}

	return false
}

// HeroHandler is the main handler factory that manages all registered handlers
type HeroHandler struct {
	factory      *core.HandlerFactory
	telnetServer *core.TelnetServer
}
