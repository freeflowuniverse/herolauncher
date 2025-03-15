package telnet

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

// Color constants
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	Bold        = "\033[1m"
)

// Server represents a telnet server
type Server struct {
	listener            net.Listener
	sessions            map[net.Conn]*Session
	sessionsMutex       sync.RWMutex
	running             bool
	authHandler         AuthHandler
	commandHandler      CommandHandler
	onSessionConnect    func(*Session)
	onSessionDisconnect func(*Session)
	debugMode           bool // Whether to print debug messages
}

// AuthHandler is a function that authenticates a session
type AuthHandler func(string) bool

// CommandHandler is a function that handles a command from a session
type CommandHandler func(*Session, string) error

func NewServer(authHandler AuthHandler, commandHandler CommandHandler, debugMode bool) *Server {
	return &Server{
		sessions:       make(map[net.Conn]*Session),
		authHandler:    authHandler,
		commandHandler: commandHandler,
		debugMode:      debugMode,
	}
}

// Start starts the telnet server on the specified address
func (s *Server) Start(address string) error {
	var err error

	// Determine if this is a Unix socket path or a TCP address
	networkType := "tcp"
	if !strings.Contains(address, ":") {
		// If there's no colon, assume it's a Unix socket path
		networkType = "unix"

		// Remove the socket file if it already exists
		if _, err := os.Stat(address); err == nil {
			if err := os.Remove(address); err != nil {
				return fmt.Errorf("failed to remove existing socket file: %w", err)
			}
		}
	}

	s.listener, err = net.Listen(networkType, address)
	if err != nil {
		return fmt.Errorf("failed to start telnet server: %w", err)
	}

	s.running = true
	go s.acceptConnections()
	return nil
}

// Stop stops the telnet server
func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	s.running = false
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Close all session connections
	s.sessionsMutex.Lock()
	for conn, session := range s.sessions {
		session.Close()
		delete(s.sessions, conn)
	}
	s.sessionsMutex.Unlock()

	return nil
}

// SetOnSessionConnect sets the callback for when a session connects
func (s *Server) SetOnSessionConnect(callback func(*Session)) {
	s.onSessionConnect = callback
}

// SetOnSessionDisconnect sets the callback for when a session disconnects
func (s *Server) SetOnSessionDisconnect(callback func(*Session)) {
	s.onSessionDisconnect = callback
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				fmt.Printf("Error accepting connection: %v\n", err)
			}
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a session connection
func (s *Server) handleConnection(conn net.Conn) {
	// Log connection information if debug mode is enabled
	if s.debugMode {
		fmt.Printf("DEBUG: New connection from %s\n", conn.RemoteAddr())
	}
	session := NewSession(conn)

	// Add session to the map
	s.sessionsMutex.Lock()
	s.sessions[conn] = session
	s.sessionsMutex.Unlock()

	// Send initial terminal negotiation sequence
	// This helps with proper terminal handling
	conn.Write([]byte{255, 251, 1})                  // IAC WILL ECHO
	conn.Write([]byte{255, 251, 3})                  // IAC WILL SUPPRESS GO AHEAD
	conn.Write([]byte{255, 253, 34})                 // IAC DO LINEMODE
	conn.Write([]byte{255, 250, 34, 1, 0, 255, 240}) // IAC SB LINEMODE MODE 0 IAC SE

	// Call the onSessionConnect callback if set
	if s.onSessionConnect != nil {
		s.onSessionConnect(session)
	}

	// Ensure session is removed when connection ends
	defer func() {
		s.sessionsMutex.Lock()
		delete(s.sessions, conn)
		s.sessionsMutex.Unlock()

		// Call the onSessionDisconnect callback if set
		if s.onSessionDisconnect != nil {
			s.onSessionDisconnect(session)
		}

		session.Close()
	}()

	// Handle authentication if an auth handler is provided
	if s.authHandler != nil {
		// Don't force interactive mode during authentication
		origInteractive := session.interactive
		session.PrintlnBold("Welcome to the telnet server!")
		session.PrintlnBold("=== Authentication Required ===")
		session.Println("To authenticate, type 'auth:YOUR_SECRET'")
		session.Println("Welcome to the telnet server!")
		session.Println("Please enter your password to authenticate.")
		session.Println("You can also type q to quit.")

		// Authentication loop
		authenticated := false

		// Read all input during authentication (echo is enabled)
		for !authenticated {
			// Use ReadLine with suppressEcho parameter, but we've modified it to show typing
			line, err := session.ReadLine(true)
			if err != nil {
				fmt.Printf("Error during authentication: %v\n", err)
				return // Session disconnected
			}

			// Check for exit commands
			if line == "!!exit" || line == "!!quit" || line == "q" {
				session.Println("Goodbye!")
				return
			}

			// Try to authenticate with the provided password directly
			// First, check if it has the old auth: prefix and extract it if needed
			secret := line
			if strings.Contains(line, "auth:") {
				parts := strings.Split(line, "auth:")
				if len(parts) > 1 {
					secret = parts[1]
				}
			}

			// Trim any whitespace from the secret
			secret = strings.TrimSpace(secret)

			if s.debugMode {
				fmt.Printf("DEBUG: Authentication attempt with secret: '%s'\n", secret)
			}

			// Validate the secret
			if s.authHandler(secret) {
				authenticated = true
				// Authentication successful
				if s.debugMode {
					fmt.Printf("DEBUG: Authentication successful with secret: '%s'\n", secret)
				}
				session.PrintlnGreen("Authentication successful!")
				session.Println("Type 'help' or '?' for available commands.")
			} else {
				session.PrintlnRed("Invalid password. Try again or type q to quit.")
				if s.debugMode {
					fmt.Printf("DEBUG: Authentication failed with secret: '%s'\n", secret)
				}
			}
		}

		// Restore original interactive mode
		session.interactive = origInteractive
	}

	// Main command processing loop
	for {
		line, err := session.ReadLine(false)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from session: %v\n", err)
			}
			return
		}

		// Handle built-in commands - use strings.TrimSpace to handle any whitespace
		cleanLine := strings.TrimSpace(line)

		// Debug output to see what command is being processed
		if s.debugMode {
			fmt.Printf("DEBUG: Processing command: '%s'\n", cleanLine)
		}

		switch cleanLine {
		case "!!quit", "!!exit", "q":
			session.PrintlnCyan("Goodbye!")
			return
		case "!!interactive", "!!i", "i":
			session.ToggleInteractive()
			continue
		case "!!help", "?", "h", "help":
			session.PrintHelp()
			continue
		case "":
			// Empty line, just show prompt again
			continue
		}

		// Handle custom commands via the command handler
		if s.commandHandler != nil {
			// We already trimmed the line above
			if s.debugMode {
				fmt.Printf("DEBUG: Passing command to handler: '%s'\n", cleanLine)
			}

			if err := s.commandHandler(session, cleanLine); err != nil {
				session.PrintlnRed(fmt.Sprintf("Error: %v", err))
			}
		} else {
			// If no command handler is set, show a message
			session.PrintlnYellow(fmt.Sprintf("Unknown command: %s", cleanLine))
		}
	}
}
