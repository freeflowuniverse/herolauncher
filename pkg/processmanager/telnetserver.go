package processmanager

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

// TelnetServer represents a telnet server for interacting with the process manager
type TelnetServer struct {
	processManager *ProcessManager
	listener       net.Listener
	clients        map[net.Conn]bool
	clientsMutex   sync.RWMutex
	running        bool
}

// NewTelnetServer creates a new telnet server
func NewTelnetServer(processManager *ProcessManager) *TelnetServer {
	return &TelnetServer{
		processManager: processManager,
		clients:        make(map[net.Conn]bool),
	}
}

// Start starts the telnet server on the specified socket path
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

	ts.listener = listener
	ts.running = true

	// Accept connections in a goroutine
	go ts.acceptConnections()

	return nil
}

// Stop stops the telnet server
func (ts *TelnetServer) Stop() error {
	if !ts.running {
		return nil
	}

	ts.running = false

	// Close the listener
	if ts.listener != nil {
		if err := ts.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %v", err)
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
func (ts *TelnetServer) acceptConnections() {
	for ts.running {
		conn, err := ts.listener.Accept()
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
	// Add client to the map
	ts.clientsMutex.Lock()
	ts.clients[conn] = false // Not authenticated yet
	ts.clientsMutex.Unlock()

	// Ensure client is removed when connection closes
	defer func() {
		conn.Close()
		ts.clientsMutex.Lock()
		delete(ts.clients, conn)
		ts.clientsMutex.Unlock()
	}()

	// Welcome message
	conn.Write([]byte(" ** Welcome: you are not authenticated, provide secret.\n"))

	// Create a scanner for reading input
	scanner := bufio.NewScanner(conn)
	authenticated := false
	var heroscriptBuffer strings.Builder
	var lastCommand string
	commandHistory := []string{}
	historyPos := 0
	interactiveMode := false

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

		// Handle authentication
		if !authenticated {
			if line == ts.processManager.GetSecret() {
				authenticated = true
				ts.clientsMutex.Lock()
				ts.clients[conn] = true // Mark as authenticated
				ts.clientsMutex.Unlock()
				conn.Write([]byte(" ** Welcome: you are authenticated.\n"))
			} else {
				conn.Write([]byte("Invalid secret. Try again or disconnect.\n"))
			}
			continue
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
				conn.Write([]byte(ColorGreen + "Interactive mode enabled. Using colors for output." + ColorReset + "\n"))
			} else {
				conn.Write([]byte("Interactive mode disabled. Plain text output.\n"))
			}
			continue
		}

		// Empty line executes previous command if there's no pending command
		if line == "" {
			if heroscriptBuffer.Len() > 0 {
				// Execute pending command
				result := ts.executeHeroscript(heroscriptBuffer.String(), interactiveMode)
				lastCommand = heroscriptBuffer.String()
				conn.Write([]byte(result))
				heroscriptBuffer.Reset()
			} else if lastCommand != "" {
				// Execute last command
				result := ts.executeHeroscript(lastCommand, interactiveMode)
				conn.Write([]byte(result))
			}
			continue
		}

		// Process heroscript commands
		if (strings.HasPrefix(line, "!!") || strings.HasPrefix(line, "#")) && heroscriptBuffer.Len() > 0 {
			// Execute previous heroscript if there's any
			result := ts.executeHeroscript(heroscriptBuffer.String(), interactiveMode)
			lastCommand = heroscriptBuffer.String()
			// Add to command history
			commandHistory = append([]string{lastCommand}, commandHistory...)
			if len(commandHistory) > 50 { // Limit history size
				commandHistory = commandHistory[:50]
			}
			historyPos = 0
			conn.Write([]byte(result))
			heroscriptBuffer.Reset()
		}

		// Append the line to the heroscript buffer
		heroscriptBuffer.WriteString(line + "\n")
	}

	// Execute any remaining heroscript
	if authenticated && heroscriptBuffer.Len() > 0 {
		result := ts.executeHeroscript(heroscriptBuffer.String(), interactiveMode)
		lastCommand = heroscriptBuffer.String()
		conn.Write([]byte(result))
	}
}

// executeHeroscript executes a heroscript and returns the result
func (ts *TelnetServer) executeHeroscript(script string, interactive bool) string {
	// Parse the heroscript
	pb, err := playbook.NewFromText(script)
	if err != nil {
		return fmt.Sprintf("Error parsing heroscript: %v\n", err)
	}

	// Find the job ID if any
	var jobID string
	for _, action := range pb.Actions {
		if action.Params != nil {
			jobID = action.Params.Get("jobid")
			if jobID == "" {
				// Try alternative casing
				jobID = action.Params.Get("jobId")
			}
			break
		}
	}

	// Process each action
	var result strings.Builder
	if interactive {
		result.WriteString(fmt.Sprintf(ColorCyan+Bold+"**RESULT** %s"+ColorReset+"\n", jobID))
	} else {
		result.WriteString(fmt.Sprintf("**RESULT** %s\n", jobID))
	}

	for _, action := range pb.Actions {
		// Process the action based on actor and name
		if action.Actor == "process" {
			switch action.Name {
			case "start":
				result.WriteString(ts.handleProcessStart(action))
			case "list":
				result.WriteString(ts.handleProcessList(action))
			case "delete":
				result.WriteString(ts.handleProcessDelete(action))
			case "status":
				result.WriteString(ts.handleProcessStatus(action))
			case "restart":
				result.WriteString(ts.handleProcessRestart(action))
			case "stop":
				result.WriteString(ts.handleProcessStop(action))
			default:
				result.WriteString(fmt.Sprintf("Unknown action: %s.%s\n", action.Actor, action.Name))
			}
		} else {
			result.WriteString(fmt.Sprintf("Unknown actor: %s\n", action.Actor))
		}
	}

	if interactive {
		result.WriteString(ColorCyan + Bold + "**ENDRESULT**" + ColorReset + "\n")
	} else {
		result.WriteString("**ENDRESULT**\n")
	}
	return result.String()
}

// handleProcessStart handles the process.start action
func (ts *TelnetServer) handleProcessStart(action *playbook.Action) string {
	// Format the heroscript if in interactive mode
	if action.Params != nil && action.Params.GetBool("interactive") {
		return formatHeroscript(action.HeroScript())
	}
	name := action.Params.Get("name")
	if name == "" {
		return "Error: name parameter is required\n"
	}

	command := action.Params.Get("command")
	if command == "" {
		return "Error: command parameter is required\n"
	}

	logEnabled := action.Params.GetBool("log")
	deadline, _ := action.Params.GetInt("deadline")
	cron := action.Params.Get("cron")
	jobID := action.Params.Get("jobid")
	if jobID == "" {
		jobID = action.Params.Get("jobId")
	}

	err := ts.processManager.StartProcess(name, command, logEnabled, deadline, cron, jobID)
	if err != nil {
		return fmt.Sprintf("Error starting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' started successfully\n", name)
}

// handleProcessList handles the process.list action
func (ts *TelnetServer) handleProcessList(action *playbook.Action) string {
	format := action.Params.Get("format")
	processes := ts.processManager.ListProcesses()

	result, err := FormatProcessList(processes, format)
	if err != nil {
		return fmt.Sprintf("Error formatting process list: %v\n", err)
	}

	return result
}

// handleProcessDelete handles the process.delete action
func (ts *TelnetServer) handleProcessDelete(action *playbook.Action) string {
	name := action.Params.Get("name")
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ts.processManager.DeleteProcess(name)
	if err != nil {
		return fmt.Sprintf("Error deleting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' deleted successfully\n", name)
}

// handleProcessStatus handles the process.status action
func (ts *TelnetServer) handleProcessStatus(action *playbook.Action) string {
	name := action.Params.Get("name")
	if name == "" {
		return "Error: name parameter is required\n"
	}

	format := action.Params.Get("format")

	procInfo, err := ts.processManager.GetProcessStatus(name)
	if err != nil {
		return fmt.Sprintf("Error getting process status: %v\n", err)
	}

	result, err := FormatProcessInfo(procInfo, format)
	if err != nil {
		return fmt.Sprintf("Error formatting process info: %v\n", err)
	}

	return result
}

// handleProcessRestart handles the process.restart action
func (ts *TelnetServer) handleProcessRestart(action *playbook.Action) string {
	name := action.Params.Get("name")
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ts.processManager.RestartProcess(name)
	if err != nil {
		return fmt.Sprintf("Error restarting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' restarted successfully\n", name)
}

// handleProcessStop handles the process.stop action
func (ts *TelnetServer) handleProcessStop(action *playbook.Action) string {
	name := action.Params.Get("name")
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ts.processManager.StopProcess(name)
	if err != nil {
		return fmt.Sprintf("Error stopping process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' stopped successfully\n", name)
}

// formatHeroscript formats heroscript with colors for interactive mode
func formatHeroscript(script string) string {
	lines := strings.Split(script, "\n")
	var formatted strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			formatted.WriteString("\n")
			continue
		}

		// Format comments
		if strings.HasPrefix(line, "//") {
			formatted.WriteString(ColorGreen + line + ColorReset + "\n")
			continue
		}

		// Format action lines
		if strings.HasPrefix(line, "!!!") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) > 0 {
				formatted.WriteString(ColorPurple + Bold + parts[0] + ColorReset + " ")
				if len(parts) > 1 {
					formatted.WriteString(parts[1])
				}
				formatted.WriteString("\n")
			}
			continue
		}

		if strings.HasPrefix(line, "!!") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) > 0 {
				formatted.WriteString(ColorBlue + Bold + parts[0] + ColorReset + " ")
				if len(parts) > 1 {
					formatted.WriteString(parts[1])
				}
				formatted.WriteString("\n")
			}
			continue
		}

		if strings.HasPrefix(line, "!") {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) > 0 {
				formatted.WriteString(ColorYellow + Bold + parts[0] + ColorReset + " ")
				if len(parts) > 1 {
					formatted.WriteString(parts[1])
				}
				formatted.WriteString("\n")
			}
			continue
		}

		// Format parameter lines
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				formatted.WriteString("    " + ColorCyan + parts[0] + ColorReset + ":" + ColorYellow + parts[1] + ColorReset + "\n")
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
	var helpText string
	if interactive {
		helpText = ColorCyan + Bold + "**RESULT**" + ColorReset + "\n" + Bold + "Available commands:" + ColorReset + "\n\n"
	} else {
		helpText = "**RESULT** \nAvailable commands:\n\n"
	}

	// Process commands
	if interactive {
		helpText += Bold + ColorBlue + "Process management commands:" + ColorReset + "\n"
	} else {
		helpText += "Process management commands:\n"
	}
	helpText += "  !!process.start name:'<name>' command:'<command>' [log:true|false] [deadline:<seconds>] [cron:'<schedule>'] [jobid:'<id>']\n"
	helpText += "  !!process.list [format:'json']\n"
	helpText += "  !!process.delete name:'<name>'\n"
	helpText += "  !!process.status name:'<name>' [format:'json']\n"
	helpText += "  !!process.restart name:'<name>'\n"
	helpText += "  !!process.stop name:'<name>'\n\n"

	// Special commands
	if interactive {
		helpText += Bold + ColorBlue + "Special commands:" + ColorReset + "\n"
	} else {
		helpText += "Special commands:\n"
	}
	if interactive {
		helpText += "  " + ColorGreen + "!!help" + ColorReset + ", " + ColorGreen + "?" + ColorReset + " or " + ColorGreen + "h" + ColorReset + " - Show this help text\n"
		helpText += "  " + ColorGreen + "!!interactive" + ColorReset + " or " + ColorGreen + "!!i" + ColorReset + " - Toggle interactive mode with colors\n"
	} else {
		helpText += "  !!help, ? or h - Show this help text\n"
		helpText += "  !!interactive or !!i - Toggle interactive mode with colors\n"
	}
	if interactive {
		helpText += "  " + ColorGreen + "!!exit" + ColorReset + " or " + ColorGreen + "!!quit" + ColorReset + " or " + ColorGreen + "q" + ColorReset + " or " + ColorGreen + "Ctrl+C" + ColorReset + " - Close the connection\n"
	} else {
		helpText += "  !!exit or !!quit or q or Ctrl+C - Close the connection\n"
	}
	if interactive {
		helpText += "  " + ColorGreen + "<empty line>" + ColorReset + " - Execute previous command or pending command\n"
		helpText += "  " + ColorGreen + "Up arrow" + ColorReset + " - Navigate to previous commands\n\n"
	} else {
		helpText += "  <empty line> - Execute previous command or pending command\n"
		helpText += "  Up arrow - Navigate to previous commands\n\n"
	}
	if interactive {
		helpText += ColorCyan + Bold + "**ENDRESULT**" + ColorReset + "\n"
	} else {
		helpText += "**ENDRESULT**\n"
	}

	return helpText
}
