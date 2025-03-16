package processmanager

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// TelnetAdapter represents an adapter between the process manager and telnet server
type TelnetAdapter struct {
	processManager *ProcessManager
	telnetServer   *telnet.Server
	logEnabled     bool
}

// NewTelnetAdapter creates a new telnet adapter
func NewTelnetAdapter(processManager *ProcessManager) *TelnetAdapter {
	log.Println("Creating new telnet adapter for process manager")
	adapter := &TelnetAdapter{
		processManager: processManager,
		logEnabled:     true,
	}

	// Create telnet server with auth and command handlers
	server := telnet.NewServer(
		// Auth handler
		func(secret string) bool {
			return secret == processManager.GetSecret()
		},
		// Command handler
		adapter.handleCommand,
		// Debug mode
		false,
	)

	adapter.telnetServer = server
	return adapter
}

// Start starts the telnet server on the specified socket path
func (ta *TelnetAdapter) Start(socketPath string) error {
	log.Printf("Starting telnet server on socket: %s", socketPath)
	err := ta.telnetServer.Start(socketPath)
	if err != nil {
		log.Printf("Error starting telnet server: %v", err)
		return err
	}
	log.Println("Telnet server started successfully")
	return nil
}

// Stop stops the telnet server
func (ta *TelnetAdapter) Stop() error {
	log.Println("Stopping telnet server")
	err := ta.telnetServer.Stop()
	if err != nil {
		log.Printf("Error stopping telnet server: %v", err)
		return err
	}
	log.Println("Telnet server stopped successfully")
	return nil
}

// handleCommand handles commands from clients
func (ta *TelnetAdapter) handleCommand(session *telnet.Session, command string) error {
	// Handle empty command
	if command == "" {
		return nil
	}

	// Log the received command
	if ta.logEnabled {
		log.Printf("Received command: '%s'", command)
	}

	// Trim any leading/trailing whitespace
	command = strings.TrimSpace(command)

	// Process command
	if strings.HasPrefix(command, "!!") || strings.HasPrefix(command, "!!process.") {
		if ta.logEnabled {
			log.Printf("Executing heroscript command: '%s'", command)
		}
		result := ta.executeHeroscript(command, session.IsInteractive())
		session.Write(result)
		return nil
	}

	// Unknown command
	if ta.logEnabled {
		log.Printf("Unknown command received: '%s'", command)
	}
	session.PrintlnYellow(fmt.Sprintf("Unknown command: %s", command))
	session.PrintlnYellow("Use '?' or 'help' to see available commands")
	return nil
}

// executeHeroscript executes a command and returns the result
func (ta *TelnetAdapter) executeHeroscript(script string, interactive bool) string {
	// For now, we'll just handle the commands directly without a playbook parser
	// In a real implementation, you would parse the script properly

	// Trim any leading/trailing whitespace
	script = strings.TrimSpace(script)

	// Log the script being executed
	if ta.logEnabled {
		log.Printf("Executing heroscript: '%s'", script)
	}

	// Extract command parts
	parts := strings.Fields(script)
	if len(parts) == 0 {
		if ta.logEnabled {
			log.Println("Error: empty command")
		}
		return telnet.FormatError(fmt.Errorf("empty command"), interactive)
	}

	// Extract job ID if present
	jobID := ""

	// Process the command
	var result strings.Builder
	var actionResult string

	// Extract command name
	cmd := parts[0]

	// Process based on command name
	switch {
	case strings.HasPrefix(cmd, "!!process.start"):
		if ta.logEnabled {
			log.Println("Handling process.start command")
		}
		actionResult = "Process start command received\n"
	case strings.HasPrefix(cmd, "!!process.list"):
		if ta.logEnabled {
			log.Println("Handling process.list command")
		}
		actionResult = ta.handleProcessList()
	case strings.HasPrefix(cmd, "!!process.delete"):
		if ta.logEnabled {
			log.Println("Handling process.delete command")
		}
		actionResult = "Process delete command received\n"
	case strings.HasPrefix(cmd, "!!process.status"):
		if ta.logEnabled {
			log.Println("Handling process.status command")
		}
		actionResult = "Process status command received\n"
	case strings.HasPrefix(cmd, "!!process.restart"):
		if ta.logEnabled {
			log.Println("Handling process.restart command")
		}
		actionResult = "Process restart command received\n"
	case strings.HasPrefix(cmd, "!!process.stop"):
		if ta.logEnabled {
			log.Println("Handling process.stop command")
		}
		actionResult = "Process stop command received\n"
	case strings.HasPrefix(cmd, "!!process.log"):
		if ta.logEnabled {
			log.Println("Handling process.log command")
		}
		actionResult = "Process log command received\n"
	case cmd == "!!help" || cmd == "?" || cmd == "h":
		if ta.logEnabled {
			log.Println("Handling help command")
		}
		actionResult = ta.generateHelpText(interactive)
	default:
		if ta.logEnabled {
			log.Printf("Unknown command received: '%s'", cmd)
		}
		actionResult = fmt.Sprintf("Unknown command: %s\n", cmd)
	}

	result.WriteString(actionResult)

	formattedResult := telnet.FormatResult(result.String(), jobID, interactive)

	if ta.logEnabled {
		log.Printf("Command result: %s", strings.ReplaceAll(formattedResult, "\n", " "))
	}

	return formattedResult
}

// handleProcessStart handles the process.start command
func (ta *TelnetAdapter) handleProcessStart(name, command string, logEnabled bool, deadline int, cron, jobID string) string {
	if name == "" {
		return "Error: name parameter is required\n"
	}

	if command == "" {
		return "Error: command parameter is required\n"
	}

	err := ta.processManager.StartProcess(name, command, logEnabled, deadline, cron, jobID)
	if err != nil {
		return fmt.Sprintf("Error starting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' started successfully\n", name)
}

// handleProcessList handles the process.list command
func (ta *TelnetAdapter) handleProcessList() string {
	format := "" // Default format
	processes := ta.processManager.ListProcesses()

	if format == "json" {
		jsonData, err := json.MarshalIndent(processes, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error marshaling processes to JSON: %v\n", err)
		}
		return string(jsonData) + "\n"
	}

	// Format as table
	if len(processes) == 0 {
		return "No processes found\n"
	}

	headers := []string{"Name", "Status", "PID", "Command", "Cron"}
	rows := make([][]string, 0, len(processes))

	for _, proc := range processes {
		status := "Stopped"
		pid := "-"
		if proc.Status == "running" {
			status = "Running"
			pid = fmt.Sprintf("%d", proc.PID)
		}

		cronStr := proc.Cron
		if cronStr == "" {
			cronStr = "-"
		}

		rows = append(rows, []string{
			proc.Name,
			status,
			pid,
			proc.Command,
			cronStr,
		})
	}

	return telnet.FormatTable(headers, rows, false)
}

// handleProcessDelete handles the process.delete command
func (ta *TelnetAdapter) handleProcessDelete(name string) string {
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ta.processManager.DeleteProcess(name)
	if err != nil {
		return fmt.Sprintf("Error deleting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' deleted successfully\n", name)
}

// handleProcessStatus handles the process.status command
func (ta *TelnetAdapter) handleProcessStatus(name string, format string) string {
	if name == "" {
		return "Error: name parameter is required\n"
	}

	process, err := ta.processManager.GetProcessStatus(name)
	if err != nil {
		return fmt.Sprintf("Error getting process: %v\n", err)
	}

	if format == "json" {
		jsonData, err := json.MarshalIndent(process, "", "  ")
		if err != nil {
			return fmt.Sprintf("Error marshaling process to JSON: %v\n", err)
		}
		return string(jsonData) + "\n"
	}

	// Format as text
	status := string(process.Status)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Name: %s\n", process.Name))
	result.WriteString(fmt.Sprintf("Status: %s\n", status))
	result.WriteString(fmt.Sprintf("Command: %s\n", process.Command))

	if process.Status == "running" {
		result.WriteString(fmt.Sprintf("PID: %d\n", process.PID))
	}

	if process.Cron != "" {
		result.WriteString(fmt.Sprintf("Cron: %s\n", process.Cron))
	}

	return result.String()
}

// handleProcessRestart handles the process.restart command
func (ta *TelnetAdapter) handleProcessRestart(name string) string {
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ta.processManager.RestartProcess(name)
	if err != nil {
		return fmt.Sprintf("Error restarting process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' restarted successfully\n", name)
}

// handleProcessStop handles the process.stop command
func (ta *TelnetAdapter) handleProcessStop(name string) string {
	if name == "" {
		return "Error: name parameter is required\n"
	}

	err := ta.processManager.StopProcess(name)
	if err != nil {
		return fmt.Sprintf("Error stopping process: %v\n", err)
	}

	return fmt.Sprintf("Process '%s' stopped successfully\n", name)
}

// handleProcessLog handles the process.log command
func (ta *TelnetAdapter) handleProcessLog(name string, lines int, interactive bool) string {
	// Get process name
	if name == "" {
		return "Error: name parameter is required\n"
	}

	// If lines is not specified, default to 20
	if lines <= 0 {
		lines = 20
	}

	// Get logs
	logs, err := ta.processManager.GetProcessLogs(name, lines)
	if err != nil {
		return fmt.Sprintf("Error getting logs: %v\n", err)
	}

	// Format the output
	var output strings.Builder
	if interactive {
		output.WriteString(fmt.Sprintf("%sLast %d lines of logs for process '%s':%s\n",
			telnet.Bold, lines, name, telnet.ColorReset))
		output.WriteString(fmt.Sprintf("%s%s\n", telnet.ColorGreen, logs))
	} else {
		output.WriteString(fmt.Sprintf("Last %d lines of logs for process '%s':\n", lines, name))
		output.WriteString(logs)
		output.WriteString("\n")
	}

	return output.String()
}

// generateHelpText generates help text for available commands
func (ta *TelnetAdapter) generateHelpText(interactive bool) string {
	var helpText strings.Builder

	// Process commands
	helpText.WriteString("Process management commands:\n")
	helpText.WriteString("  !!process.start name:'<name>' command:'<command>' [log:true|false] [deadline:<seconds>] [cron:'<schedule>'] [jobid:'<id>']\n")
	helpText.WriteString("  !!process.list [format:'json']\n")
	helpText.WriteString("  !!process.delete name:'<name>'\n")
	helpText.WriteString("  !!process.status name:'<name>' [format:'json']\n")
	helpText.WriteString("  !!process.restart name:'<name>'\n")
	helpText.WriteString("  !!process.stop name:'<name>'\n")
	helpText.WriteString("  !!process.log name:'<name>' [lines:20]\n\n")

	// Special commands
	helpText.WriteString("Special commands:\n")
	helpText.WriteString("  !!help, ? or h - Show this help text\n")
	helpText.WriteString("  !!interactive or !!i - Toggle interactive mode with colors\n")
	helpText.WriteString("  !!exit or !!quit or q or Ctrl+C - Close the connection\n")
	helpText.WriteString("  <empty line> - Execute previous command or pending command\n")
	helpText.WriteString("  Up arrow - Navigate to previous commands\n\n")

	return helpText.String()
}
