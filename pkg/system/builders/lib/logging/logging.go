package logging

import (
	"fmt"
	"os"
	"time"
)

// Logger represents a simple logger
type Logger struct {
	LogFile  *os.File
	LogToStd bool
}

// NewLogger creates a new logger
func NewLogger(logFilePath string) (*Logger, error) {
	// Create log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &Logger{
		LogFile:  logFile,
		LogToStd: true,
	}, nil
}

// WithLogToStd sets whether to log to stdout
func (l *Logger) WithLogToStd(logToStd bool) *Logger {
	l.LogToStd = logToStd
	return l
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.LogFile != nil {
		return l.LogFile.Close()
	}
	return nil
}

// Log logs a message
func (l *Logger) Log(message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formattedMessage := fmt.Sprintf("[%s] %s\n", timestamp, message)

	// Write to log file
	if l.LogFile != nil {
		_, err := l.LogFile.WriteString(formattedMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write to log file: %v\n", err)
		}
	}

	// Write to stdout
	if l.LogToStd {
		fmt.Print(formattedMessage)
	}
}

// LogError logs an error message
func (l *Logger) LogError(message string, err error) {
	errorMessage := fmt.Sprintf("%s: %v", message, err)
	l.Log(errorMessage)
}

// LogSuccess logs a success message
func (l *Logger) LogSuccess(message string) {
	l.Log(fmt.Sprintf("✅ %s", message))
}

// LogWarning logs a warning message
func (l *Logger) LogWarning(message string) {
	l.Log(fmt.Sprintf("⚠️ %s", message))
}

// LogFailure logs a failure message
func (l *Logger) LogFailure(message string) {
	l.Log(fmt.Sprintf("❌ %s", message))
}
