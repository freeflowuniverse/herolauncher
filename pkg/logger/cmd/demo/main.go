package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/logger"
)

func main() {
	// Create logs directory in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	logDir := filepath.Join(homeDir, "herolauncher_logs")
	fmt.Printf("Logs will be stored in: %s\n", logDir)

	// Create a new logger
	log, err := logger.New(logDir)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Log regular messages
	fmt.Println("Logging standard messages...")
	log.Log(logger.LogItemArgs{
		Category: "system",
		Message:  "Application started",
		LogType:  logger.LogTypeStdout,
	})

	log.Log(logger.LogItemArgs{
		Category: "config",
		Message:  "Configuration loaded successfully",
		LogType:  logger.LogTypeStdout,
	})

	// Log error messages
	fmt.Println("Logging error messages...")
	log.Log(logger.LogItemArgs{
		Category: "network",
		Message:  "Connection failed\nRetrying in 5 seconds...",
		LogType:  logger.LogTypeError,
	})

	log.Log(logger.LogItemArgs{
		Category: "database",
		Message:  "Query timeout\nTrying fallback connection",
		LogType:  logger.LogTypeError,
	})

	// Wait a moment to ensure logs are written
	time.Sleep(500 * time.Millisecond)

	// Now search for logs
	fmt.Println("\nSearching for all logs:")
	results, err := log.Search(logger.SearchArgs{
		MaxItems: 100,
	})
	if err != nil {
		fmt.Printf("Error searching logs: %v\n", err)
		return
	}
	printSearchResults(results)

	// Search for error logs only
	fmt.Println("\nSearching for error logs only:")
	errorResults, err := log.Search(logger.SearchArgs{
		LogType:  logger.LogTypeError,
		MaxItems: 100,
	})
	if err != nil {
		fmt.Printf("Error searching for error logs: %v\n", err)
		return
	}
	printSearchResults(errorResults)

	// Search by category
	fmt.Println("\nSearching for 'network' category logs:")
	networkResults, err := log.Search(logger.SearchArgs{
		Category: "network",
		MaxItems: 100,
	})
	if err != nil {
		fmt.Printf("Error searching for network logs: %v\n", err)
		return
	}
	printSearchResults(networkResults)

	fmt.Println("\nLog file can be found at:", logDir)
}

func printSearchResults(results []logger.LogItem) {
	fmt.Printf("Found %d log items:\n", len(results))
	for i, item := range results {
		logType := "STDOUT"
		if item.LogType == logger.LogTypeError {
			logType = "ERROR"
		}

		fmt.Printf("%d. [%s] [%s] %s: %s\n",
			i+1,
			item.Timestamp.Format("15:04:05"),
			logType,
			item.Category,
			item.Message)
	}
}
