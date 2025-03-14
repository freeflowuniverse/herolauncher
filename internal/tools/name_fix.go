package tools

import (
	"strings"
	"unicode"
)

// NameFix converts a string to a standardized format:
// - Removes non-ASCII characters
// - Replaces spaces, hyphens, commas, and other non-alphanumeric characters with underscores
// - Converts all characters to lowercase
func NameFix(input string) string {
	// Make ASCII only
	var asciiOnly strings.Builder
	for _, r := range input {
		if r <= unicode.MaxASCII {
			asciiOnly.WriteRune(r)
		}
	}

	// Convert to lowercase
	lowered := strings.ToLower(asciiOnly.String())

	// Process each character
	var result strings.Builder
	for _, r := range lowered {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// Keep alphanumeric characters
			result.WriteRune(r)
		} else {
			// Replace any other character with underscore
			result.WriteRune('_')
		}
	}

	return result.String()
}
