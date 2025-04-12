package smtpserver

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-smtp"
	mailmodel "github.com/freeflowuniverse/herolauncher/pkg/mail"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/net/context"
)

// Config holds the configuration for the SMTP server
type Config struct {
	Host              string
	Port              int
	Domain            string
	AllowInsecureAuth bool
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxMessageBytes   int
	MaxRecipients     int
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
}

// Server represents the SMTP server
type Server struct {
	config      Config
	smtpServer  *smtp.Server
	redisClient *redis.Client
}

// GetRedisClient returns the Redis client
func (s *Server) GetRedisClient() *redis.Client {
	return s.redisClient
}

// Backend implements the SMTP server backend
type Backend struct {
	redisClient *redis.Client
}

// Session represents an SMTP session
type Session struct {
	from        string
	to          []string
	redisClient *redis.Client
}

// NewServer creates a new SMTP server
func NewServer(config Config) (*Server, error) {
	log.Printf("Creating new SMTP server with config: host=%s, port=%d, domain=%s, redis=%s",
		config.Host, config.Port, config.Domain, config.RedisAddr)

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Printf("ERROR: Failed to connect to Redis at %s: %v", config.RedisAddr, err)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Printf("Successfully connected to Redis at %s", config.RedisAddr)

	// Create backend
	be := &Backend{
		redisClient: redisClient,
	}

	// Create SMTP server
	smtpServer := smtp.NewServer(be)
	smtpServer.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	smtpServer.Domain = config.Domain
	smtpServer.ReadTimeout = config.ReadTimeout
	smtpServer.WriteTimeout = config.WriteTimeout
	smtpServer.MaxMessageBytes = int64(config.MaxMessageBytes)
	smtpServer.MaxRecipients = config.MaxRecipients
	smtpServer.AllowInsecureAuth = config.AllowInsecureAuth

	return &Server{
		config:      config,
		smtpServer:  smtpServer,
		redisClient: redisClient,
	}, nil
}

// Start starts the SMTP server
func (s *Server) Start() error {
	log.Printf("Starting SMTP server at %s with domain %s", s.smtpServer.Addr, s.smtpServer.Domain)
	err := s.smtpServer.ListenAndServe()
	if err != nil {
		log.Printf("ERROR: SMTP server failed to start: %v", err)
	}
	return err
}

// Stop stops the SMTP server
func (s *Server) Stop() error {
	log.Printf("Stopping SMTP server at %s", s.smtpServer.Addr)
	err := s.smtpServer.Close()
	if err != nil {
		log.Printf("ERROR: Failed to stop SMTP server: %v", err)
	}
	return err
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	log.Printf("New SMTP session from %s", c.Conn().RemoteAddr())
	return &Session{
		redisClient: b.redisClient,
	}, nil
}

// Mail handles the MAIL FROM command
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Printf("MAIL FROM: %s", from)
	s.from = from
	return nil
}

// Rcpt handles the RCPT TO command
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Printf("RCPT TO: %s", to)
	s.to = append(s.to, to)
	return nil
}

// Data handles the DATA command
func (s *Session) Data(r io.Reader) error {
	log.Printf("DATA command received from %s to %v", s.from, s.to)

	// Read email data
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("ERROR: Failed to read email data: %v", err)
		return err
	}
	log.Printf("Received %d bytes of email data", len(data))

	// Convert to Unicode if needed
	unicodeData := data
	// We already read all data from r, so we can't read from it again
	// Instead, create a new reader from the data we already read

	// Parse email
	log.Printf("Parsing email from %s to %v", s.from, s.to)
	// parseEmail now returns a mailmodel.Email type
	var email *mailmodel.Email
	email, err = parseEmail(s.from, s.to, unicodeData)
	if err != nil {
		log.Printf("ERROR: Failed to parse email: %v", err)
		return err
	}
	log.Printf("Successfully parsed email with subject: %s", email.Subject)

	// Convert email to JSON
	emailJSON, err := json.Marshal(email)
	if err != nil {
		fmt.Printf("Failed to marshal email: %v\n", err)
		return err
	}

	// Generate unique ID for the email using Blake2b-192 hash
	ctx := context.Background()
	log.Printf("Generating unique ID for email using Blake2b-192 hash")

	// Create a Blake2b-192 hash (24 bytes) from the email JSON
	hash, err := blake2b.New(24, nil)
	if err != nil {
		log.Printf("ERROR: Failed to create Blake2b hash: %v", err)
		return err
	}

	// Add timestamp to ensure uniqueness even for identical emails
	timestamp := time.Now().UnixNano()
	hashInput := fmt.Sprintf("%s:%d", string(emailJSON), timestamp)

	// Write the data to the hash
	_, err = hash.Write([]byte(hashInput))
	if err != nil {
		log.Printf("ERROR: Failed to write to hash: %v", err)
		return err
	}

	// Get the hash sum and convert to hex string
	hashSum := hash.Sum(nil)
	hashHex := hex.EncodeToString(hashSum)

	mailID := fmt.Sprintf("mail:out:%s", hashHex)
	log.Printf("Generated mail ID: %s", mailID)

	// Store email in Redis
	log.Printf("Storing email in Redis with ID: %s", mailID)
	if err := s.redisClient.HSet(ctx, mailID, "data", string(emailJSON)).Err(); err != nil {
		log.Printf("ERROR: Failed to store email in Redis: %v", err)
		return err
	}

	// Add to mail queue
	log.Printf("Adding email to mail:out queue")
	if err := s.redisClient.RPush(ctx, "mail:out", mailID).Err(); err != nil {
		log.Printf("ERROR: Failed to add email to queue: %v", err)
		return err
	}

	log.Printf("Email stored with ID: %s", mailID)
	return nil
}

// Reset resets the session
func (s *Session) Reset() {
	log.Printf("Resetting SMTP session")
	s.from = ""
	s.to = []string{}
}

// Logout handles the QUIT command
func (s *Session) Logout() error {
	log.Printf("SMTP session logout")
	return nil
}

// DefaultConfig returns the default configuration for the SMTP server
func DefaultConfig() Config {
	return Config{
		Host:              "0.0.0.0",
		Port:              25,
		Domain:            "localhost",
		AllowInsecureAuth: true,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxMessageBytes:   10 * 1024 * 1024, // 10 MB
		MaxRecipients:     50,
		RedisAddr:         "localhost:6379",
		RedisPassword:     "",
		RedisDB:           0,
	}
}
