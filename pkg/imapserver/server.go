package imapserver

import (
	"log"
	"os"

	"github.com/emersion/go-imap/server"
	"github.com/redis/go-redis/v9"
)

// Server represents an IMAP server
type Server struct {
	imapServer *server.Server
	backend    *Backend
	addr       string
	debugMode  bool
}

// NewServer creates a new IMAP server
func NewServer(redisClient *redis.Client, addr string, debugMode bool) *Server {
	backend := NewBackend(redisClient, debugMode)
	s := &Server{
		backend:   backend,
		addr:      addr,
		debugMode: debugMode,
	}

	// Create a new IMAP server
	s.imapServer = server.New(backend)
	s.imapServer.Addr = addr
	s.imapServer.AllowInsecureAuth = true // Allow insecure authentication for testing

	// Set up logging
	s.imapServer.ErrorLog = log.New(os.Stderr, "IMAP SERVER ERROR: ", log.LstdFlags)
	// Debug logger is not set as it requires an io.Writer and log.Logger doesn't implement it

	return s
}

// Start starts the IMAP server
func (s *Server) Start() error {
	log.Printf("Starting IMAP server on %s", s.addr)
	return s.imapServer.ListenAndServe()
}

// StartTLS starts the IMAP server with TLS
// Note: This is not implemented correctly yet
func (s *Server) StartTLS(certFile, keyFile string) error {
	log.Printf("TLS not supported yet, starting without TLS on %s", s.addr)
	return s.imapServer.ListenAndServe()
}

// Close stops the IMAP server
func (s *Server) Close() error {
	return s.imapServer.Close()
}

// ListenAndServe starts the IMAP server and listens for connections
func ListenAndServe(redisAddr, imapAddr string, debugMode bool) error {
	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Create and start the IMAP server
	server := NewServer(redisClient, imapAddr, debugMode)
	return server.Start()
}
