package rustspecs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RustProcessor struct{}

func NewRustProcessor() *RustProcessor {
	return &RustProcessor{}
}

func (rp *RustProcessor) GetSpec(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path does not exist: %s", path)
	}

	var result strings.Builder

	// Find all Rust files in the given path
	rustFiles, err := rp.findRustFiles(path)
	if err != nil {
		return "", fmt.Errorf("error finding Rust files: %w", err)
	}

	// Process each file
	for _, file := range rustFiles {
		fileSpec, err := rp.processFile(file)
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

// findRustFiles walks through the directory structure and finds all .rs files
func (rp *RustProcessor) findRustFiles(root string) ([]string, error) {
	var rustFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Rust files
		if info.IsDir() || !strings.HasSuffix(path, ".rs") {
			return nil
		}

		rustFiles = append(rustFiles, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rustFiles, nil
}

// processFile extracts public functions, structs, and enums from a Rust file
func (rp *RustProcessor) processFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)

	// Regular expressions for matching structs, functions, and enums
	structRegex := regexp.MustCompile(`^\s*pub\s+struct\s+(\w+)`)
	fnRegex := regexp.MustCompile(`^\s*pub\s+fn\s+(\w+)`)
	implRegex := regexp.MustCompile(`^\s*impl(\s+\w+)?\s+for\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*pub\s+fn\s+(\w+)\s*\(`)
	enumRegex := regexp.MustCompile(`^\s*pub\s+enum\s+(\w+)`)
	attributeRegex := regexp.MustCompile(`^\s*#\[.*\]`)
	fieldRegex := regexp.MustCompile(`^\s*pub\s+(\w+):\s+(.*)$`)

	inStruct := false
	inEnum := false
	inImpl := false
	inMethod := false
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

		// Collect documentation comments (Rust uses ///, //, or /** ... */)
		if strings.HasPrefix(trimmedLine, "///") || strings.HasPrefix(trimmedLine, "//") ||
			strings.HasPrefix(trimmedLine, "/**") || strings.HasPrefix(trimmedLine, " *") {

			// If the last line was empty, reset the comment buffer
			if lastLineEmpty {
				currentDocComment.Reset()
			}

			currentDocComment.WriteString(line)
			currentDocComment.WriteString("\n")
			lastLineEmpty = false
			continue
		}

		// Skip attribute annotations but don't reset comments
		if attributeRegex.MatchString(trimmedLine) {
			lastLineEmpty = false
			continue
		}

		// Reset comments if we encounter a non-relevant line
		if trimmedLine != "" &&
			!attributeRegex.MatchString(trimmedLine) &&
			!strings.HasPrefix(trimmedLine, "///") &&
			!strings.HasPrefix(trimmedLine, "//") &&
			!inStruct && !inEnum && !inImpl && !inMethod &&
			!structRegex.MatchString(trimmedLine) &&
			!fnRegex.MatchString(trimmedLine) &&
			!implRegex.MatchString(trimmedLine) &&
			!enumRegex.MatchString(trimmedLine) {
			currentDocComment.Reset()
		}

		// Check for struct declaration
		if structMatch := structRegex.FindStringSubmatch(trimmedLine); len(structMatch) > 1 && !inEnum && !inMethod {
			currentStruct = structMatch[1]
			inStruct = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Write struct declaration
			result.WriteString(fmt.Sprintf("pub struct %s", currentStruct))
			if strings.Contains(line, "{") {
				result.WriteString(" {\n")
				bracketCount++
			} else {
				result.WriteString("\n")
			}

			lastLineEmpty = false
			continue
		}

		// Check for enum declaration
		if enumMatch := enumRegex.FindStringSubmatch(trimmedLine); len(enumMatch) > 1 && !inStruct && !inMethod {
			inEnum = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Write enum declaration
			result.WriteString(fmt.Sprintf("pub enum %s {", enumMatch[1]))
			result.WriteString("\n")
			if strings.Contains(line, "{") {
				bracketCount++
			}

			lastLineEmpty = false
			continue
		}

		// Check for impl blocks
		if implMatch := implRegex.FindStringSubmatch(trimmedLine); len(implMatch) > 1 && !inStruct && !inEnum {
			inImpl = true
			bracketCount = 0

			// Write impl declaration
			result.WriteString(line)
			result.WriteString("\n")
			if strings.Contains(line, "{") {
				bracketCount++
			}

			lastLineEmpty = false
			continue
		}

		// Check for methods inside impl blocks
		if inImpl && methodRegex.MatchString(trimmedLine) && !inMethod {
			inMethod = true
			bracketCount = 0

			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Extract the method signature without the implementation
			signature := trimmedLine
			if strings.Contains(signature, "{") {
				signature = strings.Split(signature, "{")[0]
				bracketCount++
			}

			result.WriteString("    ")
			result.WriteString(signature)
			result.WriteString("\n")
			lastLineEmpty = false
			continue
		}

		// Check for standalone functions
		if fnMatch := fnRegex.FindStringSubmatch(trimmedLine); len(fnMatch) > 1 && !inStruct && !inEnum && !inImpl && !inMethod {
			// Write doc comment if exists
			if currentDocComment.Len() > 0 {
				result.WriteString(currentDocComment.String())
				currentDocComment.Reset()
			}

			// Extract the function signature without the implementation
			signature := trimmedLine
			if strings.Contains(signature, "{") {
				signature = strings.Split(signature, "{")[0]
			}

			result.WriteString(signature)
			result.WriteString("\n\n")
			lastLineEmpty = false
			continue
		}

		// Handle struct fields
		if inStruct {
			// Count brackets to determine when the struct definition ends
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inStruct = false
					result.WriteString("}\n\n")
					continue
				}
			}

			// Match and include struct fields
			if fieldMatch := fieldRegex.FindStringSubmatch(trimmedLine); len(fieldMatch) > 1 {
				result.WriteString(fmt.Sprintf("    pub %s: %s\n", fieldMatch[1], fieldMatch[2]))
			} else if !strings.Contains(trimmedLine, "{") && !strings.Contains(trimmedLine, "}") && trimmedLine != "" {
				// If not a field match but still inside the struct, include the line
				result.WriteString(fmt.Sprintf("    %s\n", trimmedLine))
			}

			lastLineEmpty = false
			continue
		}

		// Handle enum variants
		if inEnum {
			// Count brackets to determine when the enum definition ends
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inEnum = false
					result.WriteString("}\n\n")
					continue
				}
			}

			// Include enum variants
			if bracketCount > 0 && !strings.Contains(trimmedLine, "{") && !strings.Contains(trimmedLine, "}") && trimmedLine != "" {
				result.WriteString(fmt.Sprintf("    %s\n", trimmedLine))
			}

			lastLineEmpty = false
			continue
		}

		// Handle impl blocks and methods
		if inImpl {
			// Count brackets for impl blocks
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inImpl = false
					result.WriteString("}\n\n")
					continue
				}
			}
		}

		// Handle method bodies (skip them)
		if inMethod {
			// Count brackets for method implementations
			if strings.Contains(line, "{") {
				bracketCount += strings.Count(line, "{")
			}
			if strings.Contains(line, "}") {
				bracketCount -= strings.Count(line, "}")
				if bracketCount <= 0 {
					inMethod = false
					continue
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
