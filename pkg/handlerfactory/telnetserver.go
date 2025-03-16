package handlerfactory

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

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
}

// NewTelnetServer creates a new telnet server
func NewTelnetServer(factory *HandlerFactory, secrets ...string) *TelnetServer {
	return &TelnetServer{
		factory: factory,
		secrets: secrets,
		clients: make(map[net.Conn]bool),
		running: false,
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
	go ts.acceptConnections(listener)

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
	go ts.acceptConnections(listener)

	return nil
}

// Stop stops the telnet server
func (ts *TelnetServer) Stop() error {
	if !ts.running {
		return nil
	}

	ts.running = false

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

	return nil
}

// acceptConnections accepts incoming connections
func (ts *TelnetServer) acceptConnections(listener net.Listener) {
	for ts.running {
		conn, err := listener.Accept()
		if err != nil {
			if ts.running {
				fmt.Printf("Failed to accept connection: %v\n", err)
			}
			continue
		}

		// Handle the connection in a goroutine
		go ts.handleConnection(conn)
	}
}

// handleConnection handles a client connection
func (ts *TelnetServer) handleConnection(conn net.Conn) {
	// Add client to the map (not authenticated yet)
	ts.clientsMutex.Lock()
	ts.clients[conn] = false
	ts.clientsMutex.Unlock()

	// Ensure client is removed when connection closes
	defer func() {
		conn.Close()
		ts.clientsMutex.Lock()
		delete(ts.clients, conn)
		ts.clientsMutex.Unlock()
	}()

	// Welcome message
	conn.Write([]byte(" ** Welcome: you are not authenticated, please authenticate with !!core.auth secret:1234\n"))

	// Create a scanner for reading input
	scanner := bufio.NewScanner(conn)
	var heroscriptBuffer strings.Builder
	var lastCommand string
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

		// Empty line executes pending command or repeats last command
		if line == "" {
			if heroscriptBuffer.Len() > 0 {
				// Execute pending command
				commandText := heroscriptBuffer.String()
				result := ts.executeHeroscript(commandText, interactiveMode)
				conn.Write([]byte(result + "\n"))

				// Add to history
				commandHistory = append(commandHistory, commandText)
				historyPos = len(commandHistory)

				// Reset buffer
				heroscriptBuffer.Reset()
				lastCommand = commandText
			} else if lastCommand != "" {
				// Repeat last command
				result := ts.executeHeroscript(lastCommand, interactiveMode)
				conn.Write([]byte(result + "\n"))
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

// executeHeroscript executes a heroscript and returns the result
func (ts *TelnetServer) executeHeroscript(script string, interactive bool) string {
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
	help.WriteString("    !!quit, q         - Disconnect\n")
	help.WriteString("    !!exit            - Disconnect\n")
	help.WriteString("\n")

	// Authentication
	help.WriteString("  Authentication:\n")
	help.WriteString("    !!core.auth secret:'your_secret'  - Authenticate with a secret\n")
	help.WriteString("\n")

	// Handler actions
	help.WriteString("  Supported Actions:\n")
	actions := ts.factory.GetSupportedActions()
	for actor, actorActions := range actions {
		help.WriteString(fmt.Sprintf("    %s:\n", actor))
		for _, action := range actorActions {
			help.WriteString(fmt.Sprintf("      !!%s.%s\n", actor, action))
		}
	}
	help.WriteString("\n")

	// Usage tips
	help.WriteString("  Usage Tips:\n")
	help.WriteString("    - Enter an empty line to execute a command\n")
	help.WriteString("    - Commands can span multiple lines\n")
	help.WriteString("    - Use arrow up to access command history\n")

	return help.String()
}
