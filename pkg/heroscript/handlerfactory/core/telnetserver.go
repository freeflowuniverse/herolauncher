package core

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/heroscript/playbook"
)

// ANSI color codes for terminal output
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

// TelnetServer represents a telnet server for processing HeroScript commands
type TelnetServer struct {
	factory      *HandlerFactory
	secrets      []string
	unixListener net.Listener
	tcpListener  net.Listener
	clients      map[net.Conn]bool // map of client connections to authentication status
	clientsMutex sync.RWMutex
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	sigCh        chan os.Signal
	onShutdown   func()
	// Map to store client preferences (like json formatting)
	clientPrefs  map[net.Conn]map[string]bool
	prefsMutex   sync.RWMutex
}

// NewTelnetServer creates a new telnet server
func NewTelnetServer(factory *HandlerFactory, secrets ...string) *TelnetServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &TelnetServer{
		factory:    factory,
		secrets:    secrets,
		clients:    make(map[net.Conn]bool),
		clientPrefs: make(map[net.Conn]map[string]bool),
		running:    false,
		ctx:        ctx,
		cancel:     cancel,
		sigCh:      make(chan os.Signal, 1),
		onShutdown: func() {},
	}
}

// Start starts the telnet server on a Unix socket
func (ts *TelnetServer) Start(socketPath string) error {
	// Remove existing socket file if it exists
	if err := os.Remove(socketPath); err != nil {
		// Ignore error if the file doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing socket: %v", err)
		}
	}

	// Create Unix domain socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %v", err)
	}

	ts.unixListener = listener
	ts.running = true

	// Accept connections in a goroutine
	ts.wg.Add(1)
	go ts.acceptConnections(listener)

	// Setup signal handling if this is the first listener
	if ts.unixListener != nil && ts.tcpListener == nil {
		ts.setupSignalHandling()
	}

	return nil
}

// StartTCP starts the telnet server on a TCP port
func (ts *TelnetServer) StartTCP(address string) error {
	// Create TCP listener
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen on TCP address: %v", err)
	}

	ts.tcpListener = listener
	ts.running = true

	// Accept connections in a goroutine
	ts.wg.Add(1)
	go ts.acceptConnections(listener)

	// Setup signal handling if this is the first listener
	if ts.tcpListener != nil && ts.unixListener == nil {
		ts.setupSignalHandling()
	}

	return nil
}

// Stop stops the telnet server
func (ts *TelnetServer) Stop() error {
	if !ts.running {
		return nil
	}

	ts.running = false

	// Signal all goroutines to stop
	ts.cancel()

	// Close the listeners
	if ts.unixListener != nil {
		if err := ts.unixListener.Close(); err != nil {
			return fmt.Errorf("failed to close Unix listener: %v", err)
		}
	}

	if ts.tcpListener != nil {
		if err := ts.tcpListener.Close(); err != nil {
			return fmt.Errorf("failed to close TCP listener: %v", err)
		}
	}

	// Close all client connections
	ts.clientsMutex.Lock()
	for conn := range ts.clients {
		conn.Close()
		delete(ts.clients, conn)
	}
	ts.clientsMutex.Unlock()

	// Wait for all goroutines to finish
	ts.wg.Wait()

	// Call the onShutdown callback if set
	if ts.onShutdown != nil {
		ts.onShutdown()
	}

	return nil
}

// acceptConnections accepts incoming connections
func (ts *TelnetServer) acceptConnections(listener net.Listener) {
	defer ts.wg.Done()

	for {
		// Use a separate goroutine to accept connections so we can check for context cancellation
		connCh := make(chan net.Conn)
		errCh := make(chan error)

		go func() {
			conn, err := listener.Accept()
			if err != nil {
				errCh <- err
				return
			}
			connCh <- conn
		}()

		select {
		case <-ts.ctx.Done():
			// Context was canceled, exit the loop
			return
		case conn := <-connCh:
			// Handle the connection in a goroutine
			ts.wg.Add(1)
			go ts.handleConnection(conn)
		case err := <-errCh:
			if ts.running {
				fmt.Printf("Failed to accept connection: %v\n", err)
			} else {
				// If we're not running, this is expected during shutdown
				return
			}
		}
	}
}

// handleConnection handles a client connection
func (ts *TelnetServer) handleConnection(conn net.Conn) {
	defer ts.wg.Done()

	// Add client to the map (not authenticated yet)
	ts.clientsMutex.Lock()
	ts.clients[conn] = false
	ts.clientsMutex.Unlock()
	
	// Initialize client preferences
	ts.prefsMutex.Lock()
	ts.clientPrefs[conn] = make(map[string]bool)
	ts.prefsMutex.Unlock()

	// Ensure client is removed when connection closes
	defer func() {
		conn.Close()
		ts.clientsMutex.Lock()
		delete(ts.clients, conn)
		ts.clientsMutex.Unlock()
		// Also remove client preferences
		ts.prefsMutex.Lock()
		delete(ts.clientPrefs, conn)
		ts.prefsMutex.Unlock()
	}()

	// Welcome message
	if len(ts.secrets) > 0 {
		conn.Write([]byte(" ** Welcome: you are not authenticated, please authenticate with !!core.auth secret:'your_secret'\n"))
	} else {
		conn.Write([]byte(" ** Welcome to HeroLauncher Telnet Server\n ** Note: Press Enter twice after sending heroscript to execute\n"))
	}

	// Create a scanner for reading input
	scanner := bufio.NewScanner(conn)
	var heroscriptBuffer strings.Builder
	commandHistory := []string{}
	historyPos := 0
	interactiveMode := true

	// Process client input
	for scanner.Scan() {
		line := scanner.Text()

		// Check for Ctrl+C (ASCII value 3)
		if line == "\x03" {
			conn.Write([]byte("Goodbye!\n"))
			return
		}

		// Check for arrow up (ANSI escape sequence for up arrow: "\x1b[A")
		if line == "\x1b[A" && len(commandHistory) > 0 {
			if historyPos > 0 {
				historyPos--
			}
			if historyPos < len(commandHistory) {
				conn.Write([]byte(commandHistory[historyPos]))
				line = commandHistory[historyPos]
			}
		}

		// Handle quit/exit commands
		if line == "!!quit" || line == "!!exit" || line == "q" {
			conn.Write([]byte("Goodbye!\n"))
			return
		}

		// Handle help command
		if line == "!!help" || line == "h" || line == "?" {
			helpText := ts.generateHelpText(interactiveMode)
			conn.Write([]byte(helpText))
			continue
		}

		// Handle interactive mode toggle
		if line == "!!interactive" || line == "!!i" || line == "i" {
			interactiveMode = !interactiveMode
			if interactiveMode {
				// Only use colors in terminal output, not in telnet
				fmt.Println(ColorGreen + "Interactive mode enabled for client. Using colors for console output." + ColorReset)
				conn.Write([]byte("Interactive mode enabled. Using formatted output.\n"))
			} else {
				fmt.Println("Interactive mode disabled for client. Plain text console output.")
				conn.Write([]byte("Interactive mode disabled. Plain text output.\n"))
			}
			continue
		}
		
		// Handle JSON format toggle
		if line == "!!json" {
			ts.prefsMutex.Lock()
			prefs, exists := ts.clientPrefs[conn]
			if !exists {
				prefs = make(map[string]bool)
				ts.clientPrefs[conn] = prefs
			}
			
			// Toggle JSON format preference
			currentSetting := prefs["json"]
			prefs["json"] = !currentSetting
			ts.prefsMutex.Unlock()
			
			if prefs["json"] {
				conn.Write([]byte("JSON format will be automatically added to all heroscripts.\n"))
			} else {
				conn.Write([]byte("JSON format will no longer be automatically added to heroscripts.\n"))
			}
			continue
		}

		// Check authentication
		isAuthenticated := ts.isClientAuthenticated(conn)

		// Handle authentication
		if !isAuthenticated {
			// Check if this is an auth command
			if strings.HasPrefix(strings.TrimSpace(line), "!!core.auth") || strings.HasPrefix(strings.TrimSpace(line), "!!auth") {
				pb, err := playbook.NewFromText(line)
				if err != nil {
					conn.Write([]byte("Authentication syntax error. Use !!core.auth secret:'your_secret'\n"))
					continue
				}

				if len(pb.Actions) > 0 {
					action := pb.Actions[0]
					// Support both auth.auth and core.auth patterns
					validActor := action.Actor == "auth" || action.Actor == "core"
					validAction := action.Name == "auth"

					if validActor && validAction {
						secret := action.Params.Get("secret")
						if ts.isValidSecret(secret) {
							ts.clientsMutex.Lock()
							ts.clients[conn] = true
							ts.clientsMutex.Unlock()
							conn.Write([]byte(" ** Authentication successful. You can now send commands.\n"))
							continue
						} else {
							conn.Write([]byte("Authentication failed: Invalid secret provided.\n"))
							continue
						}
					}
				}
				conn.Write([]byte("Invalid authentication format. Use !!core.auth secret:'your_secret'\n"))
			} else {
				conn.Write([]byte("You must authenticate first. Use !!core.auth secret:'your_secret'\n"))
			}
			continue
		}

		// Empty line executes pending command but does not repeat last command
		if line == "" {
			if heroscriptBuffer.Len() > 0 {
				// Execute pending command
				commandText := heroscriptBuffer.String()
				result := ts.executeHeroscript(commandText, conn, interactiveMode)
				conn.Write([]byte(result))

				// Add to history
				commandHistory = append(commandHistory, commandText)
				historyPos = len(commandHistory)

				// Reset buffer
				heroscriptBuffer.Reset()
			}
			continue
		}

		// Add line to heroscript buffer
		if heroscriptBuffer.Len() > 0 {
			heroscriptBuffer.WriteString("\n")
		}
		heroscriptBuffer.WriteString(line)

	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading from connection: %v\n", err)
	}
}

// isClientAuthenticated checks if a client is authenticated
func (ts *TelnetServer) isClientAuthenticated(conn net.Conn) bool {
	// If no secrets are configured, authentication is not required
	if len(ts.secrets) == 0 {
		return true
	}

	ts.clientsMutex.RLock()
	defer ts.clientsMutex.RUnlock()

	authenticated, exists := ts.clients[conn]
	return exists && authenticated
}

// isValidSecret checks if a secret is valid
func (ts *TelnetServer) isValidSecret(secret string) bool {
	for _, validSecret := range ts.secrets {
		if secret == validSecret {
			return true
		}
	}
	return false
}

// connKey is a type for context keys
type connKey struct{}

// connKeyValue is the key for storing the connection in context
var connKeyValue = connKey{}

// executeHeroscript executes a heroscript and returns the result
func (ts *TelnetServer) executeHeroscript(script string, conn net.Conn, interactive bool) string {
	// Check if this connection has JSON formatting enabled
	if conn != nil {
		ts.prefsMutex.RLock()
		prefs, exists := ts.clientPrefs[conn]
		ts.prefsMutex.RUnlock()
		
		if exists && prefs["json"] {
			// Add format:json if not already present
			if !strings.Contains(script, "format:json") {
				script = ts.addJsonFormat(script)
			}
		}
	}
	
	if interactive {
		// Format the script with colors
		formattedScript := formatHeroscript(script)
		fmt.Println("Executing heroscript:\n" + formattedScript)
	} else {
		fmt.Println("Executing heroscript:\n" + script)
	}

	// Process the heroscript
	result, err := ts.factory.ProcessHeroscript(script)
	if err != nil {
		errorMsg := fmt.Sprintf("Error: %v", err)
		if interactive {
			// Only use colors in terminal output, not in telnet response
			fmt.Println(ColorRed + errorMsg + ColorReset)
		}
		return errorMsg
	}

	if interactive {
		// Only use colors in terminal output, not in telnet response
		fmt.Println(ColorGreen + "Result: " + result + ColorReset)
	}
	return result
}

// addJsonFormat adds format:json to a heroscript if not already present
func (ts *TelnetServer) addJsonFormat(script string) string {
	lines := strings.Split(script, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "!!") {
			// Found action line, add format:json if not present
			if !strings.Contains(line, "format:") {
				lines[i] = line + " format:json"
			}
		}
	}
	return strings.Join(lines, "\n")
}



// formatHeroscript formats heroscript with colors for console output only
// This is not used for telnet responses, only for server-side logging
func formatHeroscript(script string) string {
	var formatted strings.Builder
	lines := strings.Split(script, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Comments
		if strings.HasPrefix(trimmed, "//") {
			formatted.WriteString(ColorBlue + line + ColorReset + "\n")
			continue
		}

		// Action lines
		if strings.HasPrefix(trimmed, "!") {
			parts := strings.SplitN(trimmed, " ", 2)
			actionPart := parts[0]

			// Highlight actor.action
			formatted.WriteString(Bold + ColorYellow + actionPart + ColorReset)

			// Add the rest of the line
			if len(parts) > 1 {
				formatted.WriteString(" " + parts[1])
			}
			formatted.WriteString("\n")
			continue
		}

		// Parameter lines
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				// Parameter name
				formatted.WriteString(parts[0] + ":")

				// Parameter value
				value := parts[1]
				if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
					formatted.WriteString(ColorCyan + value + ColorReset + "\n")
				} else {
					formatted.WriteString(ColorPurple + value + ColorReset + "\n")
				}
				continue
			}
		}

		// Default formatting
		formatted.WriteString(line + "\n")
	}

	return formatted.String()
}

// generateHelpText generates help text for available commands
// EnableSignalHandling sets up signal handling for graceful shutdown
// This is now deprecated as signal handling is automatically set up when the server starts
// It's kept for backward compatibility
func (ts *TelnetServer) EnableSignalHandling(onShutdown func()) {
	// Set the onShutdown callback
	ts.onShutdown = onShutdown

	// Setup the signal handling
	ts.setupSignalHandling()
}

// setupSignalHandling sets up signal handling for graceful shutdown
func (ts *TelnetServer) setupSignalHandling() {
	// Reset any previous signal notification
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	// Register for SIGINT and SIGTERM signals
	signal.Notify(ts.sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	ts.wg.Add(1)
	go func() {
		defer ts.wg.Done()

		// Wait for signal
		sig := <-ts.sigCh

		// Log that we received a signal
		fmt.Printf("Received %s signal, shutting down telnet server...\n", sig)

		// Stop the telnet server
		if err := ts.Stop(); err != nil {
			fmt.Printf("Error stopping telnet server: %v\n", err)
		} else {
			fmt.Println("Telnet server stopped successfully")
		}

		// Call the onShutdown callback if set
		if ts.onShutdown != nil {
			ts.onShutdown()
		}

		// Exit the program if this was triggered by a signal
		os.Exit(0)
	}()
}

func (ts *TelnetServer) generateHelpText(interactive bool) string {
	var help strings.Builder

	// Only use colors in console output, not in telnet
	if interactive {
		fmt.Println(Bold + ColorCyan + "Generating help text for client" + ColorReset)
	}

	help.WriteString("Available Commands:\n")

	// System commands
	help.WriteString("  System Commands:\n")
	help.WriteString("    !!help, h, ?      - Show this help\n")
	help.WriteString("    !!interactive, i  - Toggle interactive mode\n")
	help.WriteString("    !!json            - Toggle automatic JSON formatting for heroscripts\n")
	help.WriteString("    !!quit, q         - Disconnect\n")
	help.WriteString("    !!exit            - Disconnect\n")
	help.WriteString("\n")

	// Authentication
	help.WriteString("  Authentication:\n")
	help.WriteString("    !!core.auth secret:'your_secret'  - Authenticate with a secret\n")
	help.WriteString("\n")

	// Usage tips
	help.WriteString("  Usage Tips:\n")
	help.WriteString("    - Enter an empty line to execute a command\n")
	help.WriteString("    - Commands can span multiple lines\n")
	help.WriteString("    - Use arrow up to access command history\n")
	help.WriteString("------------------------------------------------\n\n")

	// Handler help sections
	help.WriteString("Handler Documentation:\n\n")

	// Get all registered handlers
	for actorName, handler := range ts.factory.handlers {
		// Try to call the Help method on each handler using reflection
		handlerValue := reflect.ValueOf(handler)
		helpMethod := handlerValue.MethodByName("Help")
		
		if helpMethod.IsValid() {
			// Call the Help method
			args := []reflect.Value{reflect.ValueOf("")}
			result := helpMethod.Call(args)
			
			// Get the result
			if len(result) > 0 && result[0].Kind() == reflect.String {
				helpText := result[0].String()
				help.WriteString(fmt.Sprintf("  %s Handler (%s):\n", strings.Title(actorName), actorName))
				help.WriteString(fmt.Sprintf("  %s\n", helpText))
				help.WriteString("\n")
			}
		}
	}

	return help.String()
}
