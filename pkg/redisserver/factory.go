package redisserver

import (
	"sync"
	"time"
)

// entry represents a stored value. For strings, value is stored as a string.
// For hashes, value is stored as a map[string]string.
type entry struct {
	value      interface{}
	expiration time.Time // zero means no expiration
}

// Server holds the in-memory datastore and provides thread-safe access.
// It implements a Redis-compatible server using redcon.
type Server struct {
	mu   sync.RWMutex
	data map[string]*entry
}

type ServerConfig struct {
	TCPPort        string
	UnixSocketPath string
}

// NewCustomServer creates a new server instance with custom TCP port and Unix socket path.
// It starts a cleanup goroutine and Redis-compatible servers on the specified addresses.
func NewServer(config ServerConfig) *Server {

	if config.UnixSocketPath == "" {
		config.UnixSocketPath = "/tmp/redis.sock"
	}

	s := &Server{
		data: make(map[string]*entry),
	}
	go s.cleanupExpiredKeys()

	// Start TCP server if port is provided
	if config.TCPPort != "" {
		tcpAddr := ":" + config.TCPPort
		go s.startRedisServer(tcpAddr, "")
	}

	// Start Unix socket server if path is provided
	if config.UnixSocketPath != "" {
		go s.startRedisServer(config.UnixSocketPath, "unix")
	}

	return s
}

// cleanupExpiredKeys periodically removes expired keys.
func (s *Server) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, ent := range s.data {
			if !ent.expiration.IsZero() && now.After(ent.expiration) {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}
