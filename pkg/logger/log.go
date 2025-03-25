package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Log writes a log entry to the appropriate log file
func (l *Logger) Log(args LogItemArgs) error {

	// Protect concurrent use
	l.mu.Lock()
	defer l.mu.Unlock()

	// Use current time if not provided
	timestamp := time.Now()
	if args.Timestamp != nil {
		timestamp = *args.Timestamp
	}

	// Format category (max 10 chars, ASCII only)
	category := formatName(args.Category)
	if len(category) > 10 {
		return fmt.Errorf("category cannot be longer than 10 chars")
	}
	category = expandString(category, 10, ' ')

	// Clean up the message
	message := strings.TrimSpace(dedent(args.Message))

	// Determine log file path based on date and hour
	logFilePath := filepath.Join(l.Path, fmt.Sprintf("%s.log", formatDayHour(timestamp)))

	// Create log file if it doesn't exist
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		if err := os.WriteFile(logFilePath, []byte{}, 0644); err != nil {
			return err
		}
		l.LastLogTime = 0 // Make sure we put time again
	}

	// Open file for appending
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	var content strings.Builder

	// Add timestamp if we're in a new second
	currentUnix := timestamp.Unix()
	if currentUnix > l.LastLogTime {
		// If not the first entry in the file, add an extra newline for separation
		if l.LastLogTime != 0 {
			content.WriteString("\n")
		}
		content.WriteString(fmt.Sprintf("%s\n", timestamp.Format("15:04:05")))
		l.LastLogTime = currentUnix
	}

	// Format log lines
	errorPrefix := " " // Default for stdout
	if args.LogType == LogTypeError {
		errorPrefix = "E"
	}

	lines := strings.Split(message, "\n")
	for i, line := range lines {
		if i == 0 {
			content.WriteString(fmt.Sprintf("%s %s - %s\n", errorPrefix, category, line))
		} else {
			content.WriteString(fmt.Sprintf("%s              %s\n", errorPrefix, line))
		}
	}

	// Write to file
	// Don't trim the trailing newline to ensure proper line separation
	_, err = f.WriteString(content.String())
	if err != nil {
		return err
	}

	return nil
}

// Helper functions
func formatName(name string) string {
	// Replace hyphens with underscores and remove non-alphanumeric chars
	result := strings.Map(func(r rune) rune {
		if r == '-' {
			return '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, name)
	return result
}

func expandString(s string, length int, char byte) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(string(char), length-len(s))
}

func dedent(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= 1 {
		return s
	}

	// Find minimum indentation
	minIndent := -1
	for _, line := range lines[1:] {
		trimmed := strings.TrimLeft(line, " \t")
		if len(trimmed) == 0 {
			continue // Skip empty lines
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	// No indentation found
	if minIndent <= 0 {
		return s
	}

	// Remove common indentation
	result := []string{lines[0]}
	for _, line := range lines[1:] {
		if len(line) > minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func formatDayHour(t time.Time) string {
	return t.Format("2006-01-02-15")
}
