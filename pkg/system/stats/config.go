package stats

import (
	"time"
)

// Config contains configuration options for the StatsManager
type Config struct {
	// Redis connection settings
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	
	// Default expiration times for different types of stats (in seconds)
	ExpirationTimes map[string]time.Duration
	
	// Debug mode - if true, requests are direct without caching
	Debug bool
	
	// Default timeout for waiting for stats (in seconds)
	DefaultTimeout time.Duration
	
	// Maximum queue size for update requests
	QueueSize int
}

// DefaultConfig returns the default configuration for StatsManager
func DefaultConfig() *Config {
	return &Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		ExpirationTimes: map[string]time.Duration{
			"system":   60 * time.Second,  // System info expires after 60 seconds
			"disk":     300 * time.Second, // Disk info expires after 5 minutes
			"process":  30 * time.Second,  // Process info expires after 30 seconds
			"network":  30 * time.Second,  // Network info expires after 30 seconds
			"hardware": 120 * time.Second, // Hardware stats expire after 2 minutes
		},
		Debug:          false,
		DefaultTimeout: 60 * time.Second, // 1 minute default timeout
		QueueSize:      100,
	}
}
