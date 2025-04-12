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
	// Primary method regex for standard V method syntax: pub fn (receiver Type) method_name
	methodRegex := regexp.MustCompile(`^pub\s+fn\s+\((\w+)\s+(?:\&|\*)?([\\w\.]+)\)\s+(\w+)`)
	// Secondary method regex for other variations of V method syntax
	methodRegex2 := regexp.MustCompile(`^pub\s+fn\s+\([^)]+\)\s+(\w+)`)
	// Standalone function regex to find module functions
	funcRegex := regexp.MustCompile(`^pub\s+fn\s+(\w+)`)
	enumRegex := regexp.MustCompile(`^pub\s+enum\s+(\w+)`)
	// Import regex to detect commented imports
	importRegex := regexp.MustCompile(`^\s*//\s*import\s+`)

	inStruct := false
	inMethod := false
	inEnum := false
	bracketCount := 0
	var currentStruct string
	var currentDocComment strings.Builder
	lastLineEmpty := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		
		// Check for empty lines
		if trimmedLine == "" {
			lastLineEmpty = true
			continue
		}

		// Collect documentation comments
		if strings.HasPrefix(trimmedLine, "//") {
			// If the last line was empty, reset the comment buffer
			// This helps separate unrelated comment blocks
			if lastLineEmpty {
				currentDocComment.Reset()
			}
			
			// Skip commented imports
			if importRegex.MatchString(trimmedLine) {
				lastLineEmpty = false
				continue
			}
			
			currentDocComment.WriteString(line)
			currentDocComment.WriteString("\n")
			lastLineEmpty = false
			continue
		}

		// Only reset doc comment if we encounter a non-comment line that's not part of a struct/enum/method
// This helps preserve comments that belong to declarations
if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "//") &&
!inStruct && !inMethod && !inEnum &&
!structRegex.MatchString(trimmedLine) &&
!methodRegex.MatchString(trimmedLine) &&
!methodRegex2.MatchString(trimmedLine) &&
!funcRegex.MatchString(trimmedLine) &&
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
lastLineEmpty = false
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
lastLineEmpty = false
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
lastLineEmpty = false
continue
}

// Inside struct, copy the line
if inStruct {
result.WriteString(line)
result.WriteString("\n")
}
lastLineEmpty = false
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
lastLineEmpty = false
continue
}

// Inside enum, copy the line
if inEnum {
result.WriteString(line)
result.WriteString("\n")
}
lastLineEmpty = false
continue
}

// Check for method declaration using primary regex
if methodMatch := methodRegex.FindStringSubmatch(trimmedLine); len(methodMatch) > 3 {
receiver := methodMatch[2]
// methodName := methodMatch[3] - Not using this variable to avoid unused variable warning

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
				lastLineEmpty = false
				continue
			}
		}

		// Check for method declaration using secondary regex
		// This is a fallback for methods that might have different syntax
		if methodMatch := methodRegex2.FindStringSubmatch(trimmedLine); len(methodMatch) > 0 && !inMethod {
			// methodName := methodMatch[1] - Not using this variable to avoid unused variable warning
			
			// Look for struct name in the line
			receiverLine := strings.TrimSpace(line)
			if strings.Contains(receiverLine, currentStruct) {
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
				lastLineEmpty = false
				continue
			}
		}

		// Check for standalone functions that might be related to the current struct
		if funcMatch := funcRegex.FindStringSubmatch(trimmedLine); len(funcMatch) > 0 && !inMethod && !inStruct && !inEnum {
			funcName := funcMatch[1]
			
			// If function name contains the struct name, it might be related
			if strings.Contains(strings.ToLower(funcName), strings.ToLower(currentStruct)) {
				inMethod = true
				bracketCount = 0

				// Write doc comment if exists
				if currentDocComment.Len() > 0 {
					result.WriteString(currentDocComment.String())
					currentDocComment.Reset()
				}

				// Write function signature
				result.WriteString(line)
				result.WriteString(" {}\n\n") // Empty implementation
				lastLineEmpty = false
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
		
		lastLineEmpty = false
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return result.String(), nil
}
