package interfaces

import (
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

// ProcessManagerInterface defines the interface for process management operations
type ProcessManagerInterface interface {
	// StartProcess starts a new process with the given name and command
	StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) error

	// StopProcess stops a running process
	StopProcess(name string) error

	// RestartProcess restarts a process
	RestartProcess(name string) error

	// DeleteProcess removes a process from the manager
	DeleteProcess(name string) error

	// GetProcessStatus returns the status of a process
	GetProcessStatus(name string) (*processmanager.ProcessInfo, error)

	// ListProcesses returns a list of all processes
	ListProcesses() []*processmanager.ProcessInfo

	// GetProcessLogs returns the logs for a specific process
	GetProcessLogs(name string, lines int) (string, error)

	// GetSecret returns the authentication secret
	GetSecret() string
}

// ProcessStartResult represents the result of starting a process
type ProcessStartResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PID     int32  `json:"pid"`
}

// ProcessStopResult represents the result of stopping a process
type ProcessStopResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProcessRestartResult represents the result of restarting a process
type ProcessRestartResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PID     int32  `json:"pid"`
}

// ProcessDeleteResult represents the result of deleting a process
type ProcessDeleteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProcessLogResult represents the result of getting process logs
type ProcessLogResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Logs    string `json:"logs"`
}

// ProcessStatus represents detailed information about a process
type ProcessStatus struct {
	Name       string    `json:"name"`
	Command    string    `json:"command"`
	PID        int32     `json:"pid"`
	Status     string    `json:"status"`
	CPUPercent float64   `json:"cpu_percent"`
	MemoryMB   float64   `json:"memory_mb"`
	StartTime  time.Time `json:"start_time"`
	LogEnabled bool      `json:"log_enabled"`
	Cron       string    `json:"cron,omitempty"`
	JobID      string    `json:"job_id,omitempty"`
	Deadline   int       `json:"deadline,omitempty"`
	Error      string    `json:"error,omitempty"`
}
