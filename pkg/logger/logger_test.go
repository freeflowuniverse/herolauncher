package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestLogger creates a temporary directory and a new logger instance for testing
func setupTestLogger(t *testing.T) (*Logger, string) {
	t.Helper()
	
	// Create temporary directory for logs
	tempDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Create a new logger
	logger, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	return logger, tempDir
}

// cleanupTestLogger removes the temporary directory used for testing
func cleanupTestLogger(tempDir string) {
	os.RemoveAll(tempDir)
}

// TestLoggerCreation tests the creation of a new logger instance
func TestLoggerCreation(t *testing.T) {
	// Create temporary directory for logs
	tempDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test successful logger creation
	logger, err := New(tempDir)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}
	
	if logger.Path != tempDir {
		t.Errorf("Expected logger path to be %q, got %q", tempDir, logger.Path)
	}
	
	if logger.LastLogTime != 0 {
		t.Errorf("Expected LastLogTime to be 0, got %d", logger.LastLogTime)
	}
	
	// Test logger creation with non-existent directory
	nonExistentDir := filepath.Join(tempDir, "non_existent")
	logger, err = New(nonExistentDir)
	if err != nil {
		t.Fatalf("Failed to create logger with non-existent directory: %v", err)
	}
	
	// Verify the directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", nonExistentDir)
	}
}

// TestBasicLogging tests basic logging functionality
func TestBasicLogging(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Test standard output logging
	err := logger.Log(LogItemArgs{
		Category: "system",
		Message:  "System started successfully",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}
	
	// Verify log file exists
	currentTime := time.Now()
	expectedFileName := fmt.Sprintf("%s.log", formatDayHour(currentTime))
	logFilePath := filepath.Join(tempDir, expectedFileName)
	
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Fatalf("Log file was not created: %s", logFilePath)
	}
	
	// Read log file content
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Verify log content
	logContent := string(content)
	
	if !strings.Contains(logContent, "system     - System started successfully") {
		t.Errorf("Log file doesn't contain expected stdout message")
	}
}

// TestMultiLineLogging tests logging of multi-line messages
func TestMultiLineLogging(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Test multi-line logging
	err := logger.Log(LogItemArgs{
		Category: "system",
		Message:  "Failed to connect\nRetrying in 5 seconds...",
		LogType:  LogTypeError,
	})
	if err != nil {
		t.Fatalf("Failed to log multi-line: %v", err)
	}
	
	// Verify log file exists
	currentTime := time.Now()
	expectedFileName := fmt.Sprintf("%s.log", formatDayHour(currentTime))
	logFilePath := filepath.Join(tempDir, expectedFileName)
	
	// Read log file content
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Verify log content
	logContent := string(content)
	
	if !strings.Contains(logContent, "E system     - Failed to connect") {
		t.Errorf("Log file doesn't contain expected error message")
	}
	
	if !strings.Contains(logContent, "E              Retrying in 5 seconds...") {
		t.Errorf("Log file doesn't contain expected multi-line formatting")
	}
}

// TestTimestampFormatting tests that timestamps are properly formatted in log files
func TestTimestampFormatting(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Log a message
	err := logger.Log(LogItemArgs{
		Category: "time",
		Message:  "Testing timestamp",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}
	
	// Wait a second and log another message to ensure a new timestamp is written
	time.Sleep(1 * time.Second)
	
	err = logger.Log(LogItemArgs{
		Category: "time",
		Message:  "Another timestamp",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log second message: %v", err)
	}
	
	// Verify log file exists
	currentTime := time.Now()
	expectedFileName := fmt.Sprintf("%s.log", formatDayHour(currentTime))
	logFilePath := filepath.Join(tempDir, expectedFileName)
	
	// Read log file content
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Verify log content contains timestamps
	logContent := string(content)
	
	// Check for timestamp format (HH:MM:SS)
	timestampPattern := time.Now().Format("15:04:")
	if !strings.Contains(logContent, timestampPattern) {
		t.Errorf("Log file doesn't contain expected timestamp format. Content: %s", logContent)
	}
}

// TestCustomTimestamp tests logging with a custom timestamp
func TestCustomTimestamp(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Create a custom timestamp
	customTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.Local)
	
	// Log with custom timestamp
	err := logger.Log(LogItemArgs{
		Timestamp: &customTime,
		Category:  "custom",
		Message:   "Custom timestamp test",
		LogType:   LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log with custom timestamp: %v", err)
	}
	
	// Verify log file exists with custom date
	expectedFileName := fmt.Sprintf("%s.log", formatDayHour(customTime))
	logFilePath := filepath.Join(tempDir, expectedFileName)
	
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Fatalf("Log file with custom timestamp was not created: %s", logFilePath)
	}
	
	// Read log file content
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Verify log content
	logContent := string(content)
	
	expectedTimestamp := customTime.Format("15:04:05")
	if !strings.Contains(logContent, expectedTimestamp) {
		t.Errorf("Log file doesn't contain expected custom timestamp. Expected: %s, Content: %s", 
			expectedTimestamp, logContent)
	}
}

// TestCategoryValidation tests category name validation
func TestCategoryValidation(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Test with valid category
	err := logger.Log(LogItemArgs{
		Category: "valid-cat",
		Message:  "Valid category test",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Errorf("Failed to log with valid category: %v", err)
	}
	
	// Test with too long category
	err = logger.Log(LogItemArgs{
		Category: "category-too-long-for-logger",
		Message:  "Should fail",
		LogType:  LogTypeStdout,
	})
	if err == nil {
		t.Error("Expected error for category longer than 10 chars, but got none")
	}
	
	// Test with special characters in category
	err = logger.Log(LogItemArgs{
		Category: "spec!@l",
		Message:  "Special chars test",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Errorf("Failed to log with special characters in category: %v", err)
	}
	
	// Verify log file exists
	currentTime := time.Now()
	expectedFileName := fmt.Sprintf("%s.log", formatDayHour(currentTime))
	logFilePath := filepath.Join(tempDir, expectedFileName)
	
	// Read log file content
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	
	// Verify special characters were properly handled
	logContent := string(content)
	
	if !strings.Contains(logContent, "spec__l    - Special chars test") {
		t.Errorf("Log file doesn't contain properly formatted special characters category")
	}
}

// TestSearchFunctionality tests the search functionality
func TestSearchFunctionality(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Log some test messages
	err := logger.Log(LogItemArgs{
		Category: "system",
		Message:  "System started successfully",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}
	
	err = logger.Log(LogItemArgs{
		Category: "system",
		Message:  "Failed to connect\nRetrying in 5 seconds...",
		LogType:  LogTypeError,
	})
	if err != nil {
		t.Fatalf("Failed to log multi-line: %v", err)
	}
	
	err = logger.Log(LogItemArgs{
		Category: "network",
		Message:  "Network connection established",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log: %v", err)
	}
	
	// Ensure logs have been written
	time.Sleep(1 * time.Second)
	
	// Test search by category
	results, err := logger.Search(SearchArgs{
		Category: "system",
		MaxItems: 100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs by category: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 system logs, got %d", len(results))
	}
	
	// Test search by log type
	results, err = logger.Search(SearchArgs{
		LogType:  LogTypeError,
		MaxItems: 100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs by type: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 error log, got %d", len(results))
	}
	
	if len(results) > 0 && !strings.Contains(results[0].Message, "Failed to connect") {
		t.Errorf("Search result doesn't contain expected message")
	}
	
	// Test search by message content
	results, err = logger.Search(SearchArgs{
		Message:  "connect",
		MaxItems: 100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs by message content: %v", err)
	}
	
	// We expect either the original error log entry containing 'connect' or both that entry
	// and potentially the continuation line, depending on how the search parses multi-line entries
	if len(results) < 1 {
		t.Errorf("Expected at least 1 log containing 'connect', got %d", len(results))
	}
	
	// Verify that the first result contains the expected content
	if len(results) > 0 && !strings.Contains(results[0].Message, "Failed to connect") {
		t.Errorf("Expected message to contain 'Failed to connect', got: %s", results[0].Message)
	}
	
	// Test search with multiple criteria
	results, err = logger.Search(SearchArgs{
		Category: "network",
		Message:  "established",
		LogType:  LogTypeStdout,
		MaxItems: 100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs with multiple criteria: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 log matching multiple criteria, got %d", len(results))
	}
	
	// Test max items limit
	results, err = logger.Search(SearchArgs{
		MaxItems: 1,
	})
	if err != nil {
		t.Fatalf("Failed to search logs with max items limit: %v", err)
	}
	
	if len(results) > 1 {
		t.Errorf("Expected at most 1 log due to MaxItems, got %d", len(results))
	}
}

// TestSearchTimeRange tests searching logs within a specific time range
func TestSearchTimeRange(t *testing.T) {
	logger, tempDir := setupTestLogger(t)
	defer cleanupTestLogger(tempDir)
	
	// Create timestamps for testing
	now := time.Now()
	pastTime := now.Add(-1 * time.Hour)
	futureTime := now.Add(1 * time.Hour)
	
	// Log with custom timestamps
	err := logger.Log(LogItemArgs{
		Timestamp: &pastTime,
		Category:  "past",
		Message:   "Past message",
		LogType:   LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log past message: %v", err)
	}
	
	err = logger.Log(LogItemArgs{
		Category: "present",
		Message:  "Present message",
		LogType:  LogTypeStdout,
	})
	if err != nil {
		t.Fatalf("Failed to log present message: %v", err)
	}
	
	// Search with time range from past to now
	from := pastTime.Add(-1 * time.Minute) // Just before past time
	to := now.Add(1 * time.Minute)         // Just after now
	
	results, err := logger.Search(SearchArgs{
		TimestampFrom: &from,
		TimestampTo:   &to,
		MaxItems:      100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs with time range: %v", err)
	}
	
	if len(results) != 2 {
		t.Errorf("Expected 2 logs in time range, got %d", len(results))
	}
	
	// Search with future time range
	// When searching with a future start time, we need to ensure the end time is also in the future
	farFutureTime := futureTime.Add(1 * time.Hour) // 2 hours in the future
	results, err = logger.Search(SearchArgs{
		TimestampFrom: &futureTime,
		TimestampTo:   &farFutureTime,
		MaxItems:      100,
	})
	if err != nil {
		t.Fatalf("Failed to search logs with future time range: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected 0 logs in future time range, got %d", len(results))
	}
	
	// Test invalid time range (from after to)
	results, err = logger.Search(SearchArgs{
		TimestampFrom: &futureTime,
		TimestampTo:   &pastTime,
		MaxItems:      100,
	})
	if err == nil {
		t.Error("Expected error for invalid time range, but got none")
	}
}

// TestHelperFunctions tests the helper functions in the logger package
func TestHelperFunctions(t *testing.T) {
	// Test formatName function
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with-hyphen", "with_hyphen"},
		{"special!@#", "special___"},
		{"mixed-123!@#", "mixed_123___"},
	}
	
	for _, tc := range testCases {
		result := formatName(tc.input)
		if result != tc.expected {
			t.Errorf("formatName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
	
	// Test expandString function
	if expandString("test", 8, ' ') != "test    " {
		t.Errorf("expandString failed to pad correctly")
	}
	
	if expandString("toolong", 4, ' ') != "toolong" {
		t.Errorf("expandString failed to handle strings longer than length")
	}
	
	// Test dedent function
	indentedText := "First line\n  Second line\n    Third line"
	expected := "First line\nSecond line\n  Third line"
	if dedent(indentedText) != expected {
		t.Errorf("dedent failed to remove common indentation")
	}
	
	// Test formatDayHour function
	testTime := time.Date(2023, 1, 15, 14, 30, 0, 0, time.Local)
	expected = "2023-01-15-14"
	if formatDayHour(testTime) != expected {
		t.Errorf("formatDayHour(%v) = %q, expected %q", testTime, formatDayHour(testTime), expected)
	}
}
