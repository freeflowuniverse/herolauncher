// Package processmanager provides functionality for managing and monitoring processes.
package processmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/freeflowuniverse/herolauncher/pkg/logger"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessStatus represents the status of a process
type ProcessStatus string

const (
	// ProcessStatusRunning indicates the process is running
	ProcessStatusRunning ProcessStatus = "running"
	// ProcessStatusStopped indicates the process is stopped
	ProcessStatusStopped ProcessStatus = "stopped"
	// ProcessStatusFailed indicates the process failed to start or crashed
	ProcessStatusFailed ProcessStatus = "failed"
	// ProcessStatusCompleted indicates the process completed successfully
	ProcessStatusCompleted ProcessStatus = "completed"
)

// ProcessInfo represents information about a managed process
type ProcessInfo struct {
	Name       string        `json:"name"`
	Command    string        `json:"command"`
	PID        int32         `json:"pid"`
	Status     ProcessStatus `json:"status"`
	CPUPercent float64       `json:"cpu_percent"`
	MemoryMB   float64       `json:"memory_mb"`
	StartTime  time.Time     `json:"start_time"`
	LogEnabled bool          `json:"log_enabled"`
	Cron       string        `json:"cron,omitempty"`
	JobID      string        `json:"job_id,omitempty"`
	Deadline   int           `json:"deadline,omitempty"`
	Error      string        `json:"error,omitempty"`

	cmd        *exec.Cmd
	ctx        context.Context
	cancel     context.CancelFunc
	procLogger *logger.Logger // Logger instance for this process
	mutex      sync.Mutex
}

// ProcessManager manages multiple processes
type ProcessManager struct {
	processes    map[string]*ProcessInfo
	mutex        sync.RWMutex
	logsBasePath string // Base path for all process logs
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	// Default logs path
	logsPath := filepath.Join(os.TempDir(), "herolauncher", "process_logs")

	return &ProcessManager{
		processes:    make(map[string]*ProcessInfo),
		logsBasePath: logsPath,
	}
}

// SetLogsBasePath sets the base directory path for process logs
func (pm *ProcessManager) SetLogsBasePath(path string) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.logsBasePath = path
}

// StartProcess starts a new process with the given name and command
func (pm *ProcessManager) StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Check if process already exists
	if _, exists := pm.processes[name]; exists {
		return fmt.Errorf("process with name '%s' already exists", name)
	}

	// Create process info
	ctx, cancel := context.WithCancel(context.Background())
	procInfo := &ProcessInfo{
		Name:       name,
		Command:    command,
		Status:     ProcessStatusStopped,
		LogEnabled: logEnabled,
		Cron:       cron,
		JobID:      jobID,
		Deadline:   deadline,
		StartTime:  time.Now(),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Set up logging if enabled
	if logEnabled {
		// Create a process-specific log directory
		processLogDir := filepath.Join(pm.logsBasePath, name)

		// Initialize the logger for this process
		loggerInstance, err := logger.New(processLogDir)
		if err != nil {
			return fmt.Errorf("failed to create logger: %v", err)
		}
		procInfo.procLogger = loggerInstance
	}

	// Start the process
	// Determine if the command starts with a path (has slashes) or contains shell operators
	var cmd *exec.Cmd
	hasShellOperators := strings.ContainsAny(command, ";&|<>$()`")

	// Split the command into parts but preserve quoted sections
	commandParts := parseCommand(command)

	if !hasShellOperators && len(commandParts) > 0 && strings.Contains(commandParts[0], "/") {
		// Command has an absolute or relative path and no shell operators, handle it directly
		execPath := commandParts[0]
		execPath = filepath.Clean(execPath) // Clean the path

		// Check if the executable exists and is executable
		if _, err := os.Stat(execPath); err == nil {
			if fileInfo, err := os.Stat(execPath); err == nil {
				if fileInfo.Mode()&0111 != 0 { // Check if executable
					// Use direct execution with the absolute path
					var args []string
					if len(commandParts) > 1 {
						args = commandParts[1:]
					}

					cmd = exec.CommandContext(ctx, execPath, args...)
					goto setupOutput // Skip the shell execution
				}
			}
		}
	}

	// If we get here, use shell execution
	cmd = exec.CommandContext(ctx, "sh", "-c", command)

setupOutput:

	// Set up output redirection
	if logEnabled && procInfo.procLogger != nil {
		// Create stdout writer that logs to the process logger
		stdoutWriter := &logWriter{
			procLogger: procInfo.procLogger,
			category:   "stdout",
			logType:    logger.LogTypeStdout,
		}

		// Create stderr writer that logs to the process logger
		stderrWriter := &logWriter{
			procLogger: procInfo.procLogger,
			category:   "stderr",
			logType:    logger.LogTypeError,
		}

		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	} else {
		// Discard output if logging is disabled
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	procInfo.cmd = cmd
	err := cmd.Start()
	if err != nil {
		// Set logger to nil to allow garbage collection
		if logEnabled && procInfo.procLogger != nil {
			procInfo.procLogger = nil
		}
		return fmt.Errorf("failed to start process: %v", err)
	}

	procInfo.PID = int32(cmd.Process.Pid)
	procInfo.Status = ProcessStatusRunning

	// Store the process
	pm.processes[name] = procInfo

	// Set up deadline if specified
	if deadline > 0 {
		go func() {
			select {
			case <-time.After(time.Duration(deadline) * time.Second):
				pm.StopProcess(name)
			case <-ctx.Done():
				// Process was stopped or completed
			}
		}()
	}

	// Monitor the process in a goroutine
	go pm.monitorProcess(name)

	return nil
}

// monitorProcess monitors a process's status and resources
func (pm *ProcessManager) monitorProcess(name string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.mutex.RLock()
			procInfo, exists := pm.processes[name]
			pm.mutex.RUnlock()

			if !exists || procInfo.Status != ProcessStatusRunning {
				return
			}

			// Update process info
			procInfo.mutex.Lock()

			// Check if process is still running
			if procInfo.cmd.ProcessState != nil && procInfo.cmd.ProcessState.Exited() {
				if procInfo.cmd.ProcessState.Success() {
					procInfo.Status = ProcessStatusCompleted
				} else {
					procInfo.Status = ProcessStatusFailed
					procInfo.Error = fmt.Sprintf("process exited with code %d", procInfo.cmd.ProcessState.ExitCode())
				}
				procInfo.mutex.Unlock()
				return
			}

			// Update CPU and memory usage
			if proc, err := process.NewProcess(procInfo.PID); err == nil {
				if cpuPercent, err := proc.CPUPercent(); err == nil {
					procInfo.CPUPercent = cpuPercent
				}
				if memInfo, err := proc.MemoryInfo(); err == nil && memInfo != nil {
					procInfo.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
				}
			}

			procInfo.mutex.Unlock()
		}
	}
}

// StopProcess stops a running process
func (pm *ProcessManager) StopProcess(name string) error {
	pm.mutex.Lock()
	procInfo, exists := pm.processes[name]
	pm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("process '%s' not found", name)
	}

	procInfo.mutex.Lock()
	defer procInfo.mutex.Unlock()

	// Check if already stopped
	if procInfo.Status != ProcessStatusRunning {
		return fmt.Errorf("process '%s' is not running", name)
	}

	// Try to flush any remaining logs
	if stdout, ok := procInfo.cmd.Stdout.(*logWriter); ok {
		stdout.flush()
	}
	if stderr, ok := procInfo.cmd.Stderr.(*logWriter); ok {
		stderr.flush()
	}

	// Cancel the context to stop the process
	procInfo.cancel()

	// Try graceful termination first
	var err error
	if procInfo.cmd != nil && procInfo.cmd.Process != nil {
		// Attempt to terminate the process
		err = procInfo.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			// If graceful termination fails, force kill
			err = procInfo.cmd.Process.Kill()
		}
	}

	procInfo.Status = ProcessStatusStopped

	// We keep the procLogger available so logs can still be viewed after stopping

	return nil
}

// RestartProcess restarts a process
func (pm *ProcessManager) RestartProcess(name string) error {
	pm.mutex.Lock()
	procInfo, exists := pm.processes[name]
	if !exists {
		pm.mutex.Unlock()
		return fmt.Errorf("process '%s' not found", name)
	}

	// Save the process configuration
	command := procInfo.Command
	logEnabled := procInfo.LogEnabled
	deadline := procInfo.Deadline
	cron := procInfo.Cron
	jobID := procInfo.JobID
	pm.mutex.Unlock()

	// Stop the process
	err := pm.StopProcess(name)
	if err != nil && err.Error() != fmt.Sprintf("process '%s' is not running", name) {
		return fmt.Errorf("failed to stop process: %v", err)
	}

	// Delete the process
	pm.DeleteProcess(name)

	// Start the process again
	return pm.StartProcess(name, command, logEnabled, deadline, cron, jobID)
}

// DeleteProcess removes a process from the manager
func (pm *ProcessManager) DeleteProcess(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	procInfo, exists := pm.processes[name]
	if !exists {
		return fmt.Errorf("process '%s' not found", name)
	}

	// Lock the process info to ensure thread safety
	procInfo.mutex.Lock()
	defer procInfo.mutex.Unlock()

	// Stop the process if it's running
	if procInfo.Status == ProcessStatusRunning {
		procInfo.cancel()
		_ = procInfo.cmd.Process.Kill()
	}

	// Delete the log directory for this process if it exists
	if pm.logsBasePath != "" {
		processLogDir := filepath.Join(pm.logsBasePath, name)
		if _, err := os.Stat(processLogDir); err == nil {
			// Directory exists, delete it and its contents
			os.RemoveAll(processLogDir)
		}
	}

	// Always set logger to nil to allow garbage collection
	// This ensures that when a service with the same name is started again,
	// a new logger instance will be created
	procInfo.procLogger = nil

	// Remove the process from the map
	delete(pm.processes, name)

	return nil
}

// GetProcessStatus returns the status of a process
func (pm *ProcessManager) GetProcessStatus(name string) (*ProcessInfo, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	procInfo, exists := pm.processes[name]
	if !exists {
		return nil, fmt.Errorf("process '%s' not found", name)
	}

	// Make a copy to avoid race conditions
	procInfo.mutex.Lock()
	infoCopy := &ProcessInfo{
		Name:       procInfo.Name,
		Command:    procInfo.Command,
		PID:        procInfo.PID,
		Status:     procInfo.Status,
		CPUPercent: procInfo.CPUPercent,
		MemoryMB:   procInfo.MemoryMB,
		StartTime:  procInfo.StartTime,
		LogEnabled: procInfo.LogEnabled,
		Cron:       procInfo.Cron,
		JobID:      procInfo.JobID,
		Deadline:   procInfo.Deadline,
		Error:      procInfo.Error,
	}
	procInfo.mutex.Unlock()

	return infoCopy, nil
}

// ListProcesses returns a list of all processes
func (pm *ProcessManager) ListProcesses() []*ProcessInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	processes := make([]*ProcessInfo, 0, len(pm.processes))
	for _, procInfo := range pm.processes {
		procInfo.mutex.Lock()
		infoCopy := &ProcessInfo{
			Name:       procInfo.Name,
			Command:    procInfo.Command,
			PID:        procInfo.PID,
			Status:     procInfo.Status,
			CPUPercent: procInfo.CPUPercent,
			MemoryMB:   procInfo.MemoryMB,
			StartTime:  procInfo.StartTime,
			LogEnabled: procInfo.LogEnabled,
			Cron:       procInfo.Cron,
			JobID:      procInfo.JobID,
			Deadline:   procInfo.Deadline,
			Error:      procInfo.Error,
		}
		procInfo.mutex.Unlock()
		processes = append(processes, infoCopy)
	}

	return processes
}

// GetProcessLogs returns the logs for a specific process
func (pm *ProcessManager) GetProcessLogs(name string, lines int) (string, error) {
	pm.mutex.RLock()
	procInfo, exists := pm.processes[name]
	pm.mutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("process '%s' not found", name)
	}

	// Set default line count for logs
	if lines <= 0 {
		// Default to a high number to essentially show all logs
		lines = 10000
	}

	// Check if logger exists
	if !procInfo.LogEnabled || procInfo.procLogger == nil {
		return "", fmt.Errorf("logging is not enabled for process '%s'", name)
	}

	// Search for the most recent logs
	results, err := procInfo.procLogger.Search(logger.SearchArgs{
		MaxItems: lines,
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve logs: %v", err)
	}

	// Format the results
	var logBuffer strings.Builder
	for _, item := range results {
		timestamp := item.Timestamp.Format("2006-01-02 15:04:05")
		prefix := " "
		if item.LogType == logger.LogTypeError {
			prefix = "E"
		}
		logBuffer.WriteString(fmt.Sprintf("%s %s %s - %s\n",
			timestamp, prefix, item.Category, item.Message))
	}

	return logBuffer.String(), nil
}

// FormatProcessInfo formats process information based on the specified format
func FormatProcessInfo(procInfo *ProcessInfo, format string) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(procInfo, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal process info: %v", err)
		}
		return string(data), nil
	default:
		// Default to a simple text format
		return fmt.Sprintf("Name: %s\nStatus: %s\nPID: %d\nCPU: %.2f%%\nMemory: %.2f MB\nStarted: %s\n",
			procInfo.Name, procInfo.Status, procInfo.PID, procInfo.CPUPercent,
			procInfo.MemoryMB, procInfo.StartTime.Format(time.RFC3339)), nil
	}
}

// FormatProcessList formats a list of processes based on the specified format
func FormatProcessList(processes []*ProcessInfo, format string) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(processes, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal process list: %v", err)
		}
		return string(data), nil
	default:
		// Default to a simple text format
		result := ""
		for _, proc := range processes {
			result += fmt.Sprintf("Name: %s, Status: %s, PID: %d, CPU: %.2f%%, Memory: %.2f MB\n",
				proc.Name, proc.Status, proc.PID, proc.CPUPercent, proc.MemoryMB)
		}
		return result, nil
	}
}

// logWriter is a writer implementation that sends output to a logger
// It's used to capture stdout/stderr and convert them to structured logs
type logWriter struct {
	procLogger *logger.Logger
	category   string
	logType    logger.LogType
	buffer     strings.Builder
}

// Write implements the io.Writer interface
func (lw *logWriter) Write(p []byte) (int, error) {
	// Add the data to our buffer
	lw.buffer.Write(p)

	// Check if we have a complete line ending with newline
	bufStr := lw.buffer.String()

	// Process each complete line
	for {
		// Find the next newline character
		idx := strings.IndexByte(bufStr, '\n')
		if idx == -1 {
			break // No more complete lines
		}

		// Extract the line (without the newline)
		line := strings.TrimSpace(bufStr[:idx])

		// Log the line (only if non-empty)
		if line != "" {
			err := lw.procLogger.Log(logger.LogItemArgs{
				Category: lw.category,
				Message:  line,
				LogType:  lw.logType,
			})

			if err != nil {
				// Just continue on error, don't want to break the process output
				fmt.Fprintf(os.Stderr, "Error logging process output: %v\n", err)
			}
		}

		// Move to the next part of the buffer
		bufStr = bufStr[idx+1:]
	}

	// Keep any partial line in the buffer
	lw.buffer.Reset()
	lw.buffer.WriteString(bufStr)

	// Always report success to avoid breaking the process
	return len(p), nil
}

// flush should be called when the process exits to log any remaining content
func (lw *logWriter) flush() {
	if lw.buffer.Len() > 0 {
		// Log any remaining content that didn't end with a newline
		line := strings.TrimSpace(lw.buffer.String())
		if line != "" {
			err := lw.procLogger.Log(logger.LogItemArgs{
				Category: lw.category,
				Message:  line,
				LogType:  lw.logType,
			})

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error flushing process output: %v\n", err)
			}
		}
		lw.buffer.Reset()
	}
}

// parseCommand parses a command string into parts, respecting quotes
func parseCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	quoteChar := ' ' // placeholder

	for _, r := range cmd {
		switch {
		case r == '\'' || r == '"':
			if inQuote {
				if r == quoteChar { // closing quote
					inQuote = false
				} else { // different quote char inside quotes
					current.WriteRune(r)
				}
			} else { // opening quote
				inQuote = true
				quoteChar = r
			}
		case unicode.IsSpace(r):
			if inQuote { // space inside quote
				current.WriteRune(r)
			} else if current.Len() > 0 { // end of arg
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	// Add the last part if not empty
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
