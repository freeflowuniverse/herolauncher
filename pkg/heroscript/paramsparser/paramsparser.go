// Package paramsparser provides functionality for parsing and manipulating parameters
// from text in a key-value format with support for multiline strings.
package paramsparser

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/freeflowuniverse/herolauncher/pkg/tools"
)

// ParamsParser represents a parameter parser that can handle various parameter sources
type ParamsParser struct {
	params        map[string]string
	defaultParams map[string]string
}

// New creates a new ParamsParser instance
func New() *ParamsParser {
	return &ParamsParser{
		params:        make(map[string]string),
		defaultParams: make(map[string]string),
	}
}

// Parse parses a string containing key-value pairs in the format:
// key:value or key:'value'
// It supports multiline string values.
func (p *ParamsParser) Parse(input string) error {
	// Normalize line endings
	input = strings.ReplaceAll(input, "\r\n", "\n")
	
	// Track the current state
	var currentKey string
	var currentValue strings.Builder
	var inMultilineString bool
	
	// Process each line
	lines := strings.Split(input, "\n")
	for i := 0; i < len(lines); i++ {
		// Only trim space for non-multiline string processing
		var line string
		if !inMultilineString {
			line = strings.TrimSpace(lines[i])
		} else {
			line = lines[i]
		}
		
		// Skip empty lines unless we're in a multiline string
		if line == "" && !inMultilineString {
			continue
		}
		
		// If we're in a multiline string
		if inMultilineString {
			// Check if this line ends the multiline string
			if strings.HasSuffix(line, "'") && !strings.HasSuffix(line, "\\'") {
				// Add the line without the closing quote
				currentValue.WriteString(line[:len(line)-1])
				p.params[currentKey] = currentValue.String()
				inMultilineString = false
				currentKey = ""
				currentValue.Reset()
			} else {
				// Continue the multiline string
				currentValue.WriteString(line)
				currentValue.WriteString("\n")
			}
			continue
		}
		
		// Process the line to extract key-value pairs
		var processedPos int
		for processedPos < len(line) {
			// Skip leading whitespace
			for processedPos < len(line) && (line[processedPos] == ' ' || line[processedPos] == '\t') {
				processedPos++
			}
			
			if processedPos >= len(line) {
				break
			}
			
			// Find the next key by looking for a colon
			keyStart := processedPos
			colonPos := -1
			
			for j := processedPos; j < len(line); j++ {
				if line[j] == ':' {
					colonPos = j
					break
				}
			}
			
			if colonPos == -1 {
				// No colon found, skip this part
				break
			}
			
			// Extract key and use NameFix to standardize it
			rawKey := strings.TrimSpace(line[keyStart:colonPos])
			key := tools.NameFix(rawKey)
			
			if key == "" {
				// Invalid key, move past the colon and continue
				processedPos = colonPos + 1
				continue
			}
			
			// Move position past the colon
			processedPos = colonPos + 1
			
			if processedPos >= len(line) {
				// End of line reached, store empty value
				p.params[key] = ""
				break
			}
			
			// Skip whitespace after the colon
			for processedPos < len(line) && (line[processedPos] == ' ' || line[processedPos] == '\t') {
				processedPos++
			}
			
			if processedPos >= len(line) {
				// End of line reached after whitespace, store empty value
				p.params[key] = ""
				break
			}
			
			// Check if the value is quoted
			if line[processedPos] == '\'' {
				// This is a quoted string
				processedPos++ // Skip the opening quote
				
				// Look for the closing quote
				quoteEnd := -1
				for j := processedPos; j < len(line); j++ {
					// Check for escaped quote
					if line[j] == '\'' && (j == 0 || line[j-1] != '\\') {
						quoteEnd = j
						break
					}
				}
				
				if quoteEnd != -1 {
					// Single-line quoted string
					value := line[processedPos:quoteEnd]
					// For quoted values, we preserve the original formatting
					// But for single-line values, we can apply NameFix if needed
					if key != "description" {
						value = tools.NameFix(value)
					}
					p.params[key] = value
					processedPos = quoteEnd + 1 // Move past the closing quote
				} else {
					// Start of multiline string
					currentKey = key
					currentValue.WriteString(line[processedPos:])
					currentValue.WriteString("\n")
					inMultilineString = true
					break
				}
			} else {
				// This is an unquoted value
				valueStart := processedPos
				valueEnd := valueStart
				
				// Find the end of the value (space or end of line)
				for valueEnd < len(line) && line[valueEnd] != ' ' && line[valueEnd] != '\t' {
					valueEnd++
				}
				
				value := line[valueStart:valueEnd]
				// For unquoted values, use NameFix to standardize them
				// This handles the 'without' keyword and other special cases
				p.params[key] = tools.NameFix(value)
				processedPos = valueEnd
			}
		}
	}
	
	// If we're still in a multiline string at the end, that's an error
	if inMultilineString {
		return errors.New("unterminated multiline string")
	}
	
	return nil
}

// ParseString is a simpler version that parses a string with the format:
// key:value or key:'value'
// This version doesn't support multiline strings and is optimized for one-line inputs
func (p *ParamsParser) ParseString(input string) error {
	// Trim the input
	input = strings.TrimSpace(input)
	
	// Process the input to extract key-value pairs
	var processedPos int
	for processedPos < len(input) {
		// Skip leading whitespace
		for processedPos < len(input) && (input[processedPos] == ' ' || input[processedPos] == '\t') {
			processedPos++
		}
		
		if processedPos >= len(input) {
			break
		}
		
		// Find the next key by looking for a colon
		keyStart := processedPos
		colonPos := -1
		
		for j := processedPos; j < len(input); j++ {
			if input[j] == ':' {
				colonPos = j
				break
			}
		}
		
		if colonPos == -1 {
			// No colon found, skip this part
			break
		}
		
		// Extract key and use NameFix to standardize it
		rawKey := strings.TrimSpace(input[keyStart:colonPos])
		key := tools.NameFix(rawKey)
		
		if key == "" {
			// Invalid key, move past the colon and continue
			processedPos = colonPos + 1
			continue
		}
		
		// Move position past the colon
		processedPos = colonPos + 1
		
		if processedPos >= len(input) {
			// End of input reached, store empty value
			p.params[key] = ""
			break
		}
		
		// Skip whitespace after the colon
		for processedPos < len(input) && (input[processedPos] == ' ' || input[processedPos] == '\t') {
			processedPos++
		}
		
		if processedPos >= len(input) {
			// End of input reached after whitespace, store empty value
			p.params[key] = ""
			break
		}
		
		// Check if the value is quoted
		if input[processedPos] == '\'' {
			// This is a quoted string
			processedPos++ // Skip the opening quote
			
			// Look for the closing quote
			quoteEnd := -1
			for j := processedPos; j < len(input); j++ {
				// Check for escaped quote
				if input[j] == '\'' && (j == 0 || input[j-1] != '\\') {
					quoteEnd = j
					break
				}
			}
			
			if quoteEnd == -1 {
				return errors.New("unterminated quoted string")
			}
			
			value := input[processedPos:quoteEnd]
			// For quoted values in ParseString, we can apply NameFix
			// since this method doesn't handle multiline strings
			if key != "description" {
				value = tools.NameFix(value)
			}
			p.params[key] = value
			processedPos = quoteEnd + 1 // Move past the closing quote
		} else {
			// This is an unquoted value
			valueStart := processedPos
			valueEnd := valueStart
			
			// Find the end of the value (space or end of input)
			for valueEnd < len(input) && input[valueEnd] != ' ' && input[valueEnd] != '\t' {
				valueEnd++
			}
			
			value := input[valueStart:valueEnd]
			// For unquoted values, use NameFix to standardize them
			// This handles the 'without' keyword and other special cases
			p.params[key] = tools.NameFix(value)
			processedPos = valueEnd
		}
	}
	
	return nil
}

// ParseFile parses a file containing key-value pairs
func (p *ParamsParser) ParseFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return p.Parse(string(data))
}

// SetDefault sets a default value for a parameter
func (p *ParamsParser) SetDefault(key, value string) {
	p.defaultParams[key] = value
}

// SetDefaults sets multiple default values at once
func (p *ParamsParser) SetDefaults(defaults map[string]string) {
	for k, v := range defaults {
		p.defaultParams[k] = v
	}
}

// Set explicitly sets a parameter value
func (p *ParamsParser) Set(key, value string) {
	p.params[key] = value
}

// Get retrieves a parameter value, returning the default if not found
func (p *ParamsParser) Get(key string) string {
	if value, exists := p.params[key]; exists {
		return value
	}
	if defaultValue, exists := p.defaultParams[key]; exists {
		return defaultValue
	}
	return ""
}

// GetInt retrieves a parameter as an integer
func (p *ParamsParser) GetInt(key string) (int, error) {
	value := p.Get(key)
	if value == "" {
		return 0, errors.New("parameter not found")
	}
	return strconv.Atoi(value)
}

// GetIntDefault retrieves a parameter as an integer with a default value
func (p *ParamsParser) GetIntDefault(key string, defaultValue int) int {
	value, err := p.GetInt(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBool retrieves a parameter as a boolean
func (p *ParamsParser) GetBool(key string) bool {
	value := p.Get(key)
	if value == "" {
		return false
	}
	
	// Check for common boolean string representations
	value = strings.ToLower(value)
	return value == "true" || value == "yes" || value == "1" || value == "on"
}

// GetBoolDefault retrieves a parameter as a boolean with a default value
func (p *ParamsParser) GetBoolDefault(key string, defaultValue bool) bool {
	if !p.Has(key) {
		return defaultValue
	}
	return p.GetBool(key)
}

// GetFloat retrieves a parameter as a float64
func (p *ParamsParser) GetFloat(key string) (float64, error) {
	value := p.Get(key)
	if value == "" {
		return 0, errors.New("parameter not found")
	}
	return strconv.ParseFloat(value, 64)
}

// GetFloatDefault retrieves a parameter as a float64 with a default value
func (p *ParamsParser) GetFloatDefault(key string, defaultValue float64) float64 {
	value, err := p.GetFloat(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Has checks if a parameter exists
func (p *ParamsParser) Has(key string) bool {
	_, exists := p.params[key]
	return exists
}

// GetAll returns all parameters as a map
func (p *ParamsParser) GetAll() map[string]string {
	result := make(map[string]string)
	
	// First add defaults
	for k, v := range p.defaultParams {
		result[k] = v
	}
	
	// Then override with actual params
	for k, v := range p.params {
		result[k] = v
	}
	
	return result
}

// MustGet retrieves a parameter value, panicking if not found
func (p *ParamsParser) MustGet(key string) string {
	value := p.Get(key)
	if value == "" {
		panic(fmt.Sprintf("required parameter '%s' not found", key))
	}
	return value
}

// MustGetInt retrieves a parameter as an integer, panicking if not found or invalid
func (p *ParamsParser) MustGetInt(key string) int {
	value, err := p.GetInt(key)
	if err != nil {
		panic(fmt.Sprintf("required integer parameter '%s' not found or invalid", key))
	}
	return value
}

// MustGetFloat retrieves a parameter as a float64, panicking if not found or invalid
func (p *ParamsParser) MustGetFloat(key string) float64 {
	value, err := p.GetFloat(key)
	if err != nil {
		panic(fmt.Sprintf("required float parameter '%s' not found or invalid", key))
	}
	return value
}
