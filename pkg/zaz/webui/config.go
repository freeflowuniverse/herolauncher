package webui

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config holds the configuration for the Freezone Manager UI server
type Config struct {
	Port            string
	TemplatesPath   string
	StaticFilesPath string
	DatabasePath    string
}

// DefaultConfig returns a default configuration for the Freezone Manager UI server
func DefaultConfig() Config {
	// Get the absolute path to the project root
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "../../..")
	
	// Check for PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "9030" // Default port if not specified
	}

	return Config{
		Port:            port,
		TemplatesPath:   filepath.Join(projectRoot, "pkg/zaz/webui/templates"),
		StaticFilesPath: filepath.Join(projectRoot, "pkg/zaz/webui/static"),
		DatabasePath:    filepath.Join(projectRoot, "data/zaz/freezone.db"),
	}
}
