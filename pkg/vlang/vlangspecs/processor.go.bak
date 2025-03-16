package vlangspecs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// VlangProcessor processes V language files to extract public structs and methods
type VlangProcessor struct {
	// Add any fields needed for configuration or state
}

// NewVlangProcessor creates a new VlangProcessor instance
func NewVlangProcessor() *VlangProcessor {
	return &VlangProcessor{}
}

// GetSpec walks over the given path recursively, finds all .v files,
// and extracts public structs and their methods without implementation code
func (vp *VlangProcessor) GetSpec(path string) (string, error) {
	// Check if the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	var result strings.Builder

	// Find all .v files in the given path
	vFiles, err := vp.findVFiles(path)
	if err != nil {
		return "", fmt.Errorf("error finding V files: %w", err)
	}

	// Process each file
	for _, file := range vFiles {
		fileSpec, err := vp.processFile(file)
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

// findVFiles walks through the directory structure and finds all .v files
func (vp *VlangProcessor) findVFiles(root string) ([]string, error) {
	var vFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-v files
		if info.IsDir() || !strings.HasSuffix(path, ".v") {
			return nil
		}

		// Skip files starting with _ or ending with _ before .v
		baseName := filepath.Base(path)
		if strings.HasPrefix(baseName, "_") {
			return nil
		}

		// Check for files ending with _ before .v (like file_.v)
		fileNameWithoutExt := strings.TrimSuffix(baseName, ".v")
		if strings.HasSuffix(fileNameWithoutExt, "_") {
			return nil
		}

		vFiles = append(vFiles, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return vFiles, nil
}

// processFile extracts public structs, enums, and methods from a V file
func (vp *VlangProcessor) processFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)

	// Regular expressions for matching structs, methods, and enums
	structRegex := regexp.MustCompile(`^pub\s+struct\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^pub\s+fn\s+\((\w+)\s+(?:\&|\*)?([\w\.]+)\)\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^pub\s+enum\s+(\w+)`)

	inStruct := false
	inMethod := false
	inEnum := false
	bracketCount := 0
	var currentStruct string
	var currentDocComment strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Collect documentation comments
		if strings.HasPrefix(trimmedLine, "//") {
			currentDocComment.WriteString(line)
			currentDocComment.WriteString("\n")
			continue
		}

		// Only reset doc comment if we encounter a non-comment line that's not part of a struct/enum/method
		// This helps preserve comments that belong to declarations
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "//") &&
			!inStruct && !inMethod && !inEnum &&
			!structRegex.MatchString(trimmedLine) &&
			!methodRegex.MatchString(trimmedLine) &&
			!enumRegex.MatchString(trimmedLine) {
			currentDocComment.Reset()
		}

		// Check for struct declaration
		if structMatch := structRegex.FindStringSubmatch(trimmedLine); len(structMatch) > 1 && !inMethod && !inEnum {
			currentStruct = structMatch[1]
			inStruct = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Write struct declaration
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Check for enum declaration
		if enumMatch := enumRegex.FindStringSubmatch(trimmedLine); len(enumMatch) > 1 && !inMethod && !inStruct {
			inEnum = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Write enum declaration
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		// Track bracket count for struct
		if inStruct && !inMethod {
			if strings.Contains(trimmedLine, "{") {
				bracketCount++
			}
			if strings.Contains(trimmedLine, "}") {
				bracketCount--
				if bracketCount <= 0 {
					inStruct = false
					result.WriteString(line)
					result.WriteString("\n\n")
				} else {
					result.WriteString(line)
					result.WriteString("\n")
				}
				continue
			}

			// Inside struct, copy the line
			if inStruct {
				result.WriteString(line)
				result.WriteString("\n")
			}
			continue
		}

		// Track bracket count for enum
		if inEnum && !inMethod {
			if strings.Contains(trimmedLine, "{") {
				bracketCount++
			}
			if strings.Contains(trimmedLine, "}") {
				bracketCount--
				if bracketCount <= 0 {
					inEnum = false
					result.WriteString(line)
					result.WriteString("\n\n")
				} else {
					result.WriteString(line)
					result.WriteString("\n")
				}
				continue
			}

			// Inside enum, copy the line
			if inEnum {
				result.WriteString(line)
				result.WriteString("\n")
			}
			continue
		}

		// Check for method declaration
		if methodMatch := methodRegex.FindStringSubmatch(trimmedLine); len(methodMatch) > 3 {
			receiver := methodMatch[2]
			// methodName is captured but not used in the current implementation
			// we could use it for logging or filtering in the future

			// Only process methods for structs we've seen
			if receiver == currentStruct {
				inMethod = true
				bracketCount = 0

				// Write doc comment if exists
				if currentDocComment.Len() > 0 {
					result.WriteString(currentDocComment.String())
					currentDocComment.Reset()
				}

				// Write method signature
				result.WriteString(line)
				result.WriteString(" {}\n\n") // Empty implementation
				continue
			}
		}

		// Track bracket count for method
		if inMethod {
			if strings.Contains(trimmedLine, "{") {
				bracketCount++
			}
			if strings.Contains(trimmedLine, "}") {
				bracketCount--
				if bracketCount <= 0 {
					inMethod = false
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}
