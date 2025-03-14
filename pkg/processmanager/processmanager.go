// Package processmanager provides functionality for managing and monitoring processes.
package processmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

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
	logFile    *os.File
	mutex      sync.Mutex
}

// ProcessManager manages multiple processes
type ProcessManager struct {
	processes map[string]*ProcessInfo
	mutex     sync.RWMutex
	secret    string
}

// NewProcessManager creates a new process manager
func NewProcessManager(secret string) *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*ProcessInfo),
		secret:    secret,
	}
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
		logFile, err := os.OpenFile(fmt.Sprintf("%s.log", name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create log file: %v", err)
		}
		procInfo.logFile = logFile
	}

	// Start the process
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if logEnabled && procInfo.logFile != nil {
		cmd.Stdout = procInfo.logFile
		cmd.Stderr = procInfo.logFile
	}
	
	procInfo.cmd = cmd
	err := cmd.Start()
	if err != nil {
		if logEnabled && procInfo.logFile != nil {
			procInfo.logFile.Close()
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
	defer pm.mutex.Unlock()

	procInfo, exists := pm.processes[name]
	if !exists {
		return fmt.Errorf("process '%s' not found", name)
	}

	if procInfo.Status != ProcessStatusRunning {
		return fmt.Errorf("process '%s' is not running", name)
	}

	procInfo.mutex.Lock()
	defer procInfo.mutex.Unlock()

	// Cancel the context to stop the process
	procInfo.cancel()
	
	// Wait for the process to exit
	err := procInfo.cmd.Process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}

	procInfo.Status = ProcessStatusStopped

	// Close log file if it exists
	if procInfo.logFile != nil {
		procInfo.logFile.Close()
		procInfo.logFile = nil
	}

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

	// Stop the process if it's running
	if procInfo.Status == ProcessStatusRunning {
		procInfo.mutex.Lock()
		procInfo.cancel()
		_ = procInfo.cmd.Process.Kill()
		
		// Close log file if it exists
		if procInfo.logFile != nil {
			procInfo.logFile.Close()
		}
		procInfo.mutex.Unlock()
	}

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

// GetSecret returns the authentication secret
func (pm *ProcessManager) GetSecret() string {
	return pm.secret
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
