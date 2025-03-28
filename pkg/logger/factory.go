package logger

import (
	"os"
)

// New creates a new Logger instance
func New(path string) (*Logger, error) {
	// Create directory if it doesn't exist
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	return &Logger{
		Path:        path,
		LastLogTime: 0,
	}, nil
}
