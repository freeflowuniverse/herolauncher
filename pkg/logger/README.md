# Logger Module (Go)

A simple logging system that provides structured logging with search capabilities, ported from V to Go.

Logs are stored in hourly files with a consistent format that makes them both human-readable and machine-parseable.

## Features

- Structured logging with categories and error types
- Automatic timestamp management
- Multi-line message support
- Search functionality with filtering options
- Human-readable log format

## Usage

```go
package main

import (
	"fmt"
	"time"
	"logger"
)

func main() {
	// Create a new logger
	l, err := logger.New("/var/logs")
	if err != nil {
		panic(err)
	}

	// Log a message
	err = l.Log(logger.LogItemArgs{
		Category: "system",
		Message:  "System started successfully",
		LogType:  logger.LogTypeStdout,
	})
	if err != nil {
		panic(err)
	}

	// Log an error
	err = l.Log(logger.LogItemArgs{
		Category: "system",
		Message:  "Failed to connect\nRetrying in 5 seconds...",
		LogType:  logger.LogTypeError,
	})
	if err != nil {
		panic(err)
	}

	// Search logs
	fromTime := time.Now().Add(-24 * time.Hour) // Last 24 hours
	results, err := l.Search(logger.SearchArgs{
		TimestampFrom: &fromTime,
		Category:      "system",     // Filter by category
		Message:       "failed",     // Search in message content
		LogType:       logger.LogTypeError, // Only error messages
		MaxItems:      100,          // Limit results
	})
	if err != nil {
		panic(err)
	}
	
	for _, item := range results {
		fmt.Printf("[%s] %s: %s\n", item.Timestamp.Format(time.RFC3339), item.Category, item.Message)
	}
}