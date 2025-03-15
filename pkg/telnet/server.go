package telnet

import (
	"fmt"
	"io"
	"net"
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
	listener         net.Listener
	clients          map[net.Conn]*Client
	clientsMutex     sync.RWMutex
	running          bool
	authHandler      AuthHandler
	commandHandler   CommandHandler
	onClientConnect  func(*Client)
	onClientDisconnect func(*Client)
}

// AuthHandler is a function that authenticates a client
type AuthHandler func(string) bool

// CommandHandler is a function that handles a command from a client
type CommandHandler func(*Client, string) error

// NewServer creates a new telnet server
func NewServer(authHandler AuthHandler, commandHandler CommandHandler) *Server {
	return &Server{
		clients:        make(map[net.Conn]*Client),
		authHandler:    authHandler,
		commandHandler: commandHandler,
	}
}

// Start starts the telnet server on the specified address
func (s *Server) Start(address string) error {
	var err error
	s.listener, err = net.Listen("tcp", address)
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

	// Close all client connections
	s.clientsMutex.Lock()
	for conn, client := range s.clients {
		client.Close()
		delete(s.clients, conn)
	}
	s.clientsMutex.Unlock()

	return nil
}

// SetOnClientConnect sets the callback for when a client connects
func (s *Server) SetOnClientConnect(callback func(*Client)) {
	s.onClientConnect = callback
}

// SetOnClientDisconnect sets the callback for when a client disconnects
func (s *Server) SetOnClientDisconnect(callback func(*Client)) {
	s.onClientDisconnect = callback
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

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	client := NewClient(conn)
	
	// Add client to the map
	s.clientsMutex.Lock()
	s.clients[conn] = client
	s.clientsMutex.Unlock()

	// Send initial terminal negotiation sequence
	// This helps with proper terminal handling
	conn.Write([]byte{255, 251, 1})   // IAC WILL ECHO
	conn.Write([]byte{255, 251, 3})   // IAC WILL SUPPRESS GO AHEAD
	conn.Write([]byte{255, 253, 34})  // IAC DO LINEMODE
	conn.Write([]byte{255, 250, 34, 1, 0, 255, 240}) // IAC SB LINEMODE MODE 0 IAC SE

	// Call the onClientConnect callback if set
	if s.onClientConnect != nil {
		s.onClientConnect(client)
	}

	// Ensure client is removed when connection ends
	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		
		// Call the onClientDisconnect callback if set
		if s.onClientDisconnect != nil {
			s.onClientDisconnect(client)
		}
		
		client.Close()
	}()

	// Handle authentication if an auth handler is provided
	if s.authHandler != nil {
		// Don't force interactive mode during authentication
		origInteractive := client.interactive
		client.Println("Welcome: you are not authenticated, provide secret.")
		
		// Read until we get a valid secret or client disconnects
		authenticated := false
		for !authenticated {
			line, err := client.ReadLine()
			if err != nil {
				return // Client disconnected
			}
			
			// Check for exit commands or special cases
			if line == "!!exit" || line == "!!quit" || line == "q" || line == "" {
				client.Println("Goodbye!")
				return
			}
			
			if s.authHandler(line) {
				authenticated = true
				client.Println("Welcome: you are authenticated.")
			} else {
				client.Println("Invalid secret. Try again or disconnect.")
			}
		}
		
		// Restore original interactive mode
		client.interactive = origInteractive
	}

	// Main command processing loop
	for {
		line, err := client.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from client: %v\n", err)
			}
			return
		}

		// Handle built-in commands
		switch line {
		case "!!quit", "!!exit", "q":
			client.PrintlnCyan("Goodbye!")
			return
		case "!!interactive", "!!i", "i":
			client.ToggleInteractive()
			continue
		case "!!help", "?", "h":
			client.PrintHelp()
			continue
		case "":
			// Empty line, just show prompt again
			continue
		}

		// Handle custom commands via the command handler
		if s.commandHandler != nil {
			if err := s.commandHandler(client, line); err != nil {
				client.PrintlnRed(fmt.Sprintf("Error: %v", err))
			}
		} else {
			// If no command handler is set, show a message
			client.PrintlnYellow(fmt.Sprintf("Unknown command: %s", line))
		}
	}
}
