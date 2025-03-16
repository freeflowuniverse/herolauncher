package tools

import (
	"strings"
	"unicode"
)

// NameFix converts a string to a standardized format:
// - Removes non-ASCII characters
// - Replaces spaces, hyphens, commas, and other non-alphanumeric characters with underscores
// - Preserves periods (.)
// - Converts all characters to lowercase
// - Ensures multiple underscores are replaced with a single underscore
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

	lowered = TrimSpacesAndQuotes(lowered)

	// Process each character
	var result strings.Builder
	lastWasUnderscore := false
	for _, r := range lowered {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == ',' {
			// Keep alphanumeric characters, periods, and commas
			result.WriteRune(r)
			lastWasUnderscore = false
		} else {
			// Replace any other character with underscore, but avoid consecutive underscores
			if !lastWasUnderscore {
				result.WriteRune('_')
				lastWasUnderscore = true
			}
		}
	}

	// Only trim leading underscores, keep trailing ones
	return strings.TrimPrefix(result.String(), "_")
}

// trimSpacesAndQuotes removes leading and trailing spaces and single quotes from a string
func TrimSpacesAndQuotes(s string) string {
	// Trim leading spaces and quotes
	for strings.HasPrefix(s, " ") || strings.HasPrefix(s, "'") {
		s = strings.TrimPrefix(s, " ")
		s = strings.TrimPrefix(s, "'")
	}

	// Trim trailing spaces and quotes
	for strings.HasSuffix(s, " ") || strings.HasSuffix(s, "'") {
		s = strings.TrimSuffix(s, " ")
		s = strings.TrimSuffix(s, "'")
	}

	return s
}
