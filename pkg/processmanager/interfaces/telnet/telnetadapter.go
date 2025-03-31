package telnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
	"github.com/freeflowuniverse/herolauncher/pkg/telnetserver"
)

// FormatError formats an error message
func FormatError(err error, interactive bool) string {
	if err == nil {
		return ""
	}

	if interactive {
		return fmt.Sprintf("%sError: %s%s\n", telnetserver.ColorRed, err.Error(), telnetserver.ColorReset)
	}
	return fmt.Sprintf("Error: %s\n", err.Error())
}

// FormatResult formats a command result
func FormatResult(result, jobID string, interactive bool) string {
	if jobID != "" {
		return fmt.Sprintf("jobid: %s\n%s", jobID, result)
	}
	return result
}

// FormatTable formats data as a table
func FormatTable(headers []string, rows [][]string, interactive bool) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	// Write headers
	if interactive {
		fmt.Fprintf(w, "%s", telnetserver.Bold)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	if interactive {
		fmt.Fprintf(w, "%s", telnetserver.ColorReset)
	}

	// Write rows
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
	return buf.String()
}

// TelnetAdapter represents an adapter between the process manager and telnet server
type TelnetAdapter struct {
	processManager interfaces.ProcessManagerInterface
	telnetServer   *telnetserver.TelnetServer
	logEnabled     bool
}

// NewTelnetAdapter creates a new telnet adapter
func NewTelnetAdapter(processManager interfaces.ProcessManagerInterface) *TelnetAdapter {
	log.Println("Creating new telnet adapter for process manager")
	adapter := &TelnetAdapter{
		processManager: processManager,
		logEnabled:     true,
	}

	// Create telnet server with auth and command handlers
	server := telnetserver.NewTelnetServer(
		// Auth handler
		func(secret string) bool {
			return secret == processManager.GetSecret()
		},
		// Command handler
		adapter.handleCommand,
		// Debug mode
		false,
		// Process manager - nil since we're using the interface
		nil,
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
func (ta *TelnetAdapter) handleCommand(session *telnetserver.Session, command string) error {
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
		return FormatError(fmt.Errorf("empty command"), interactive)
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
		
		// Parse the command parameters
		name := ""
		command := ""
		logEnabled := false
		deadline := 0
		cron := ""
		jobID := ""
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			} else if strings.HasPrefix(part, "command:") {
				command = strings.Trim(strings.TrimPrefix(part, "command:"), "'\"")
			} else if strings.HasPrefix(part, "log_enabled:") {
				logStr := strings.Trim(strings.TrimPrefix(part, "log_enabled:"), "'\"")
				logEnabled = strings.ToLower(logStr) == "true"
			} else if strings.HasPrefix(part, "deadline:") {
				deadlineStr := strings.Trim(strings.TrimPrefix(part, "deadline:"), "'\"")
				var err error
				deadline, err = strconv.Atoi(deadlineStr)
				if err != nil {
					deadline = 0
				}
			} else if strings.HasPrefix(part, "cron:") {
				cron = strings.Trim(strings.TrimPrefix(part, "cron:"), "'\"")
			} else if strings.HasPrefix(part, "job_id:") {
				jobID = strings.Trim(strings.TrimPrefix(part, "job_id:"), "'\"")
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		if command == "" {
			return FormatError(fmt.Errorf("missing required parameter: command"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessStart(name, command, logEnabled, deadline, cron, jobID)

	case strings.HasPrefix(cmd, "!!process.stop"):
		if ta.logEnabled {
			log.Println("Handling process.stop command")
		}
		
		// Parse the command parameters
		name := ""
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessStop(name)

	case strings.HasPrefix(cmd, "!!process.restart"):
		if ta.logEnabled {
			log.Println("Handling process.restart command")
		}
		
		// Parse the command parameters
		name := ""
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessRestart(name)

	case strings.HasPrefix(cmd, "!!process.delete"):
		if ta.logEnabled {
			log.Println("Handling process.delete command")
		}
		
		// Parse the command parameters
		name := ""
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessDelete(name)

	case strings.HasPrefix(cmd, "!!process.status"):
		if ta.logEnabled {
			log.Println("Handling process.status command")
		}
		
		// Parse the command parameters
		name := ""
		format := "text"
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			} else if strings.HasPrefix(part, "format:") {
				format = strings.Trim(strings.TrimPrefix(part, "format:"), "'\"")
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessStatus(name, format)

	case strings.HasPrefix(cmd, "!!process.list"):
		if ta.logEnabled {
			log.Println("Handling process.list command")
		}
		
		// Handle the command
		actionResult = ta.handleProcessList()

	case strings.HasPrefix(cmd, "!!process.log"):
		if ta.logEnabled {
			log.Println("Handling process.log command")
		}
		
		// Parse the command parameters
		name := ""
		lines := 10 // Default to 10 lines
		
		// Extract parameters from the command string
		for _, part := range parts[1:] {
			if strings.HasPrefix(part, "name:") {
				name = strings.Trim(strings.TrimPrefix(part, "name:"), "'\"")
			} else if strings.HasPrefix(part, "lines:") {
				linesStr := strings.Trim(strings.TrimPrefix(part, "lines:"), "'\"")
				var err error
				lines, err = strconv.Atoi(linesStr)
				if err != nil {
					lines = 10
				}
			}
		}
		
		// Validate required parameters
		if name == "" {
			return FormatError(fmt.Errorf("missing required parameter: name"), interactive)
		}
		
		// Handle the command
		actionResult = ta.handleProcessLog(name, lines, interactive)

	case cmd == "!!help" || cmd == "!!?" || cmd == "!!process.help":
		if ta.logEnabled {
			log.Println("Handling help command")
		}
		
		// Generate help text
		actionResult = ta.generateHelpText(interactive)

	default:
		if ta.logEnabled {
			log.Printf("Unknown command: '%s'", cmd)
		}
		return FormatError(fmt.Errorf("unknown command: %s", cmd), interactive)
	}

	// Format the result
	result.WriteString(FormatResult(actionResult, jobID, interactive))
	return result.String()
}

// handleProcessStart handles the process.start command
func (ta *TelnetAdapter) handleProcessStart(name, command string, logEnabled bool, deadline int, cron, jobID string) string {
	if ta.logEnabled {
		log.Printf("Starting process: name=%s, command=%s, logEnabled=%v, deadline=%d, cron=%s, jobID=%s", name, command, logEnabled, deadline, cron, jobID)
	}

	err := ta.processManager.StartProcess(name, command, logEnabled, deadline, cron, jobID)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error starting process: %v", err)
		}
		return fmt.Sprintf("Error starting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' started successfully", name)
}

// handleProcessList handles the process.list command
func (ta *TelnetAdapter) handleProcessList() string {
	if ta.logEnabled {
		log.Println("Listing processes")
	}

	processes := ta.processManager.ListProcesses()
	if len(processes) == 0 {
		return "No processes found"
	}

	// Format as a table
	headers := []string{"Name", "Status", "PID", "CPU%", "Memory (MB)", "Command"}
	rows := make([][]string, 0, len(processes))

	for _, p := range processes {
		status := string(p.Status)
		if p.Error != "" {
			status = fmt.Sprintf("%s (Error: %s)", status, p.Error)
		}

		rows = append(rows, []string{
			p.Name,
			status,
			fmt.Sprintf("%d", p.PID),
			fmt.Sprintf("%.2f", p.CPUPercent),
			fmt.Sprintf("%.2f", p.MemoryMB),
			p.Command,
		})
	}

	// Format as JSON if requested
	return FormatTable(headers, rows, true)
}

// handleProcessDelete handles the process.delete command
func (ta *TelnetAdapter) handleProcessDelete(name string) string {
	if ta.logEnabled {
		log.Printf("Deleting process: %s", name)
	}

	err := ta.processManager.DeleteProcess(name)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error deleting process: %v", err)
		}
		return fmt.Sprintf("Error deleting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' deleted successfully", name)
}

// handleProcessStatus handles the process.status command
func (ta *TelnetAdapter) handleProcessStatus(name string, format string) string {
	if ta.logEnabled {
		log.Printf("Getting status for process: %s (format: %s)", name, format)
	}

	process, err := ta.processManager.GetProcessStatus(name)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error getting process status: %v", err)
		}
		return fmt.Sprintf("Error getting process status: %v", err)
	}

	// Format as JSON if requested
	if format == "json" {
		jsonData, err := json.MarshalIndent(process, "", "  ")
		if err != nil {
			if ta.logEnabled {
				log.Printf("Error marshaling process status to JSON: %v", err)
			}
			return fmt.Sprintf("Error marshaling process status to JSON: %v", err)
		}
		return string(jsonData)
	}

	// Format as text
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Name: %s\n", process.Name))
	result.WriteString(fmt.Sprintf("Command: %s\n", process.Command))
	result.WriteString(fmt.Sprintf("Status: %s\n", process.Status))
	result.WriteString(fmt.Sprintf("PID: %d\n", process.PID))
	result.WriteString(fmt.Sprintf("CPU: %.2f%%\n", process.CPUPercent))
	result.WriteString(fmt.Sprintf("Memory: %.2f MB\n", process.MemoryMB))
	result.WriteString(fmt.Sprintf("Start Time: %s\n", process.StartTime.Format("2006-01-02 15:04:05")))
	if process.Error != "" {
		result.WriteString(fmt.Sprintf("Error: %s\n", process.Error))
	}

	return result.String()
}

// handleProcessRestart handles the process.restart command
func (ta *TelnetAdapter) handleProcessRestart(name string) string {
	if ta.logEnabled {
		log.Printf("Restarting process: %s", name)
	}

	err := ta.processManager.RestartProcess(name)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error restarting process: %v", err)
		}
		return fmt.Sprintf("Error restarting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' restarted successfully", name)
}

// handleProcessStop handles the process.stop command
func (ta *TelnetAdapter) handleProcessStop(name string) string {
	if ta.logEnabled {
		log.Printf("Stopping process: %s", name)
	}

	err := ta.processManager.StopProcess(name)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error stopping process: %v", err)
		}
		return fmt.Sprintf("Error stopping process: %v", err)
	}

	return fmt.Sprintf("Process '%s' stopped successfully", name)
}

// handleProcessLog handles the process.log command
func (ta *TelnetAdapter) handleProcessLog(name string, lines int, interactive bool) string {
	if ta.logEnabled {
		log.Printf("Getting logs for process: %s (lines: %d)", name, lines)
	}

	logs, err := ta.processManager.GetProcessLogs(name, lines)
	if err != nil {
		if ta.logEnabled {
			log.Printf("Error getting process logs: %v", err)
		}
		return fmt.Sprintf("Error getting process logs: %v", err)
	}

	if logs == "" {
		return "No logs available"
	}

	// Format logs
	if interactive {
		var result strings.Builder
		result.WriteString(fmt.Sprintf("%sLogs for process '%s':%s\n", telnetserver.Bold, name, telnetserver.ColorReset))
		result.WriteString(logs)
		return result.String()
	}

	return fmt.Sprintf("Logs for process '%s':\n%s", name, logs)
}

// generateHelpText generates help text for available commands
func (ta *TelnetAdapter) generateHelpText(interactive bool) string {
	var result strings.Builder

	if interactive {
		result.WriteString(fmt.Sprintf("%sAvailable Commands:%s\n", telnetserver.Bold, telnetserver.ColorReset))
	} else {
		result.WriteString("Available Commands:\n")
	}

	result.WriteString("!!process.start name:<name> command:<command> [log_enabled:<true|false>] [deadline:<seconds>] [cron:<cron_expr>] [job_id:<job_id>]\n")
	result.WriteString("  - Starts a new process with the given name and command\n\n")

	result.WriteString("!!process.stop name:<name>\n")
	result.WriteString("  - Stops a running process\n\n")

	result.WriteString("!!process.restart name:<name>\n")
	result.WriteString("  - Restarts a process\n\n")

	result.WriteString("!!process.delete name:<name>\n")
	result.WriteString("  - Deletes a process\n\n")

	result.WriteString("!!process.status name:<name> [format:<text|json>]\n")
	result.WriteString("  - Gets the status of a process\n\n")

	result.WriteString("!!process.list\n")
	result.WriteString("  - Lists all processes\n\n")

	result.WriteString("!!process.log name:<name> [lines:<num_lines>]\n")
	result.WriteString("  - Gets the logs for a process\n\n")

	result.WriteString("!!help or !!?\n")
	result.WriteString("  - Shows this help text\n")

	return result.String()
}
