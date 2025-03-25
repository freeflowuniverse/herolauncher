package logger

import (
	"sync"
	"time"
)

// Logger represents a structured logging system
type Logger struct {
	Path        string
	LastLogTime int64 // To track when we need to write timestamps in logs
	mu          sync.Mutex
}

// LogItem represents a single log entry
type LogItem struct {
	Timestamp time.Time
	Category  string
	Message   string
	LogType   LogType
}

// LogType defines the type of log message
type LogType int

const (
	LogTypeStdout LogType = iota
	LogTypeError
)

// LogItemArgs defines parameters for creating a log entry
type LogItemArgs struct {
	Timestamp *time.Time
	Category  string
	Message   string
	LogType   LogType
}

// SearchArgs defines parameters for searching log entries
type SearchArgs struct {
	TimestampFrom *time.Time
	TimestampTo   *time.Time
	Category      string // Can be empty
	Message       string // Any content to search for
	LogType       LogType
	MaxItems      int
}
