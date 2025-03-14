// Package paramsparser provides functionality for parsing and manipulating parameters
// from text in a key-value format with support for multiline strings.
package paramsparser

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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
// key: 'value' anotherKey: 'another value'
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
			// Find the next key
			keyMatch := regexp.MustCompile(`([a-zA-Z0-9_]+):\s*`).FindStringSubmatchIndex(line[processedPos:])
			if keyMatch == nil {
				break
			}
			
			// Extract key
			keyStart, keyEnd := keyMatch[2]+processedPos, keyMatch[3]+processedPos
			key := line[keyStart:keyEnd]
			
			// Move position past the key and colon
			processedPos = keyMatch[1] + processedPos
			
			// Check if the value is quoted
			remaining := line[processedPos:]
			if strings.HasPrefix(remaining, "'") {
				// This is a quoted string
				if strings.Count(remaining, "'") >= 2 {
					// Single-line quoted string
					quoteEnd := strings.Index(remaining[1:], "'") + 1
					value := remaining[1:quoteEnd]
					p.params[key] = value
					processedPos += quoteEnd + 1
				} else {
					// Start of multiline string
					currentKey = key
					currentValue.WriteString(remaining[1:])
					currentValue.WriteString("\n")
					inMultilineString = true
					break
				}
			} else {
				// This is an unquoted value (number or boolean)
				numMatch := regexp.MustCompile(`^([0-9]+|true|false|yes|no)`).FindStringSubmatchIndex(remaining)
				if numMatch != nil {
					valueStart, valueEnd := numMatch[2], numMatch[3]
					value := remaining[valueStart:valueEnd]
					p.params[key] = value
					processedPos += valueEnd
				} else {
					// No valid value found, skip to next key
					break
				}
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
// key: 'value' key2: 'value2' or key: value key2: value2
// This version doesn't support multiline strings and is optimized for one-line inputs
func (p *ParamsParser) ParseString(input string) error {
	// Trim the input and normalize spaces
	input = strings.TrimSpace(input)
	
	// Process the input to extract key-value pairs
	var processedPos int
	for processedPos < len(input) {
		// Find the next key
		keyMatch := regexp.MustCompile(`([a-zA-Z0-9_]+):\s*`).FindStringSubmatchIndex(input[processedPos:])
		if keyMatch == nil {
			break
		}
		
		// Extract key
		keyStart, keyEnd := keyMatch[2]+processedPos, keyMatch[3]+processedPos
		key := input[keyStart:keyEnd]
		
		// Move position past the key and colon
		processedPos = keyMatch[1] + processedPos
		
		// Check if the value is quoted
		remaining := input[processedPos:]
		if strings.HasPrefix(remaining, "'") {
			// This is a quoted string
			quoteEnd := strings.Index(remaining[1:], "'") + 1
			if quoteEnd <= 0 {
				return errors.New("unterminated quoted string")
			}
			value := remaining[1:quoteEnd]
			p.params[key] = value
			processedPos += quoteEnd + 1
		} else {
			// This is an unquoted value (number or boolean)
			numMatch := regexp.MustCompile(`^([0-9]+|true|false|yes|no|1|0)\b`).FindStringSubmatchIndex(remaining)
			if numMatch != nil {
				valueStart, valueEnd := numMatch[2], numMatch[3]
				value := remaining[valueStart:valueEnd]
				p.params[key] = value
				processedPos += valueEnd
			} else {
				// No valid value found, skip to next key
				break
			}
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
