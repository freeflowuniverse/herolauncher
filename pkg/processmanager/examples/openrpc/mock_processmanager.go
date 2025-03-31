package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
)

// MockProcessManager implements the interfaces.ProcessManagerInterface for testing purposes
type MockProcessManager struct {
	processes map[string]*processmanager.ProcessInfo
	logs      map[string][]string
	mutex     sync.RWMutex
	secret    string
}

// Ensure MockProcessManager implements interfaces.ProcessManagerInterface
var _ interfaces.ProcessManagerInterface = (*MockProcessManager)(nil)

// NewMockProcessManager creates a new mock process manager
func NewMockProcessManager() *MockProcessManager {
	return &MockProcessManager{
		processes: make(map[string]*processmanager.ProcessInfo),
		logs:      make(map[string][]string),
		secret:    "mock-secret",
	}
}

// StartProcess starts a new process
func (m *MockProcessManager) StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.processes[name]; exists {
		return fmt.Errorf("process %s already exists", name)
	}

	process := &processmanager.ProcessInfo{
		Name:       name,
		Command:    command,
		PID:        12345, // Mock PID
		Status:     processmanager.ProcessStatusRunning,
		CPUPercent: 0.5,
		MemoryMB:   10.0,
		StartTime:  time.Now(),
		LogEnabled: logEnabled,
		Cron:       cron,
		JobID:      jobID,
		Deadline:   deadline,
	}

	m.processes[name] = process
	m.logs[name] = []string{
		fmt.Sprintf("[%s] Process started: %s", time.Now().Format(time.RFC3339), command),
	}

	return nil
}

// StopProcess stops a running process
func (m *MockProcessManager) StopProcess(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	process, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("process %s does not exist", name)
	}

	if process.Status != processmanager.ProcessStatusRunning {
		return fmt.Errorf("process %s is not running", name)
	}

	process.Status = processmanager.ProcessStatusStopped

	m.logs[name] = append(m.logs[name], fmt.Sprintf("[%s] Process stopped", time.Now().Format(time.RFC3339)))

	return nil
}

// RestartProcess restarts a process
func (m *MockProcessManager) RestartProcess(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	process, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("process %s does not exist", name)
	}

	process.Status = processmanager.ProcessStatusRunning
	process.StartTime = time.Now()

	m.logs[name] = append(m.logs[name], fmt.Sprintf("[%s] Process restarted", time.Now().Format(time.RFC3339)))

	return nil
}

// DeleteProcess deletes a process
func (m *MockProcessManager) DeleteProcess(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.processes[name]; !exists {
		return fmt.Errorf("process %s does not exist", name)
	}

	delete(m.processes, name)
	delete(m.logs, name)

	return nil
}

// GetProcessStatus gets the status of a process
func (m *MockProcessManager) GetProcessStatus(name string) (*processmanager.ProcessInfo, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	process, exists := m.processes[name]
	if !exists {
		return nil, fmt.Errorf("process %s does not exist", name)
	}

	// Return a copy to avoid race conditions
	return &processmanager.ProcessInfo{
		Name:       process.Name,
		Command:    process.Command,
		PID:        process.PID,
		Status:     process.Status,
		CPUPercent: process.CPUPercent,
		MemoryMB:   process.MemoryMB,
		StartTime:  process.StartTime,
		LogEnabled: process.LogEnabled,
		Cron:       process.Cron,
		JobID:      process.JobID,
		Deadline:   process.Deadline,
	}, nil
}

// ListProcesses lists all processes
func (m *MockProcessManager) ListProcesses() []*processmanager.ProcessInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	processes := make([]*processmanager.ProcessInfo, 0, len(m.processes))
	for _, process := range m.processes {
		// Create a copy to avoid race conditions
		processes = append(processes, &processmanager.ProcessInfo{
			Name:       process.Name,
			Command:    process.Command,
			PID:        process.PID,
			Status:     process.Status,
			CPUPercent: process.CPUPercent,
			MemoryMB:   process.MemoryMB,
			StartTime:  process.StartTime,
			LogEnabled: process.LogEnabled,
			Cron:       process.Cron,
			JobID:      process.JobID,
			Deadline:   process.Deadline,
		})
	}

	return processes
}

// GetProcessLogs gets the logs for a process
func (m *MockProcessManager) GetProcessLogs(name string, maxLines int) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	logs, exists := m.logs[name]
	if !exists {
		return "", fmt.Errorf("logs for process %s do not exist", name)
	}

	if maxLines <= 0 || maxLines > len(logs) {
		return strings.Join(logs, "\n"), nil
	}

	return strings.Join(logs[len(logs)-maxLines:], "\n"), nil
}

// SetLogsBasePath sets the base path for logs (mock implementation does nothing)
func (m *MockProcessManager) SetLogsBasePath(path string) {
	// No-op for mock
}

// GetSecret returns the authentication secret
func (m *MockProcessManager) GetSecret() string {
	return m.secret
}
