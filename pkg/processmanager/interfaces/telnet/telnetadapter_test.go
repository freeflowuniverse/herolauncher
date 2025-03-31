package telnet

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProcessManager is a mock implementation of ProcessManagerInterface for testing
type MockProcessManager struct {
	mock.Mock
	secret string
}

// Verify that MockProcessManager implements the ProcessManagerInterface
var _ interfaces.ProcessManagerInterface = (*MockProcessManager)(nil)

// GetSecret returns the authentication secret
func (m *MockProcessManager) GetSecret() string {
	return m.secret
}

// StartProcess mocks starting a process
func (m *MockProcessManager) StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) error {
	args := m.Called(name, command, logEnabled, deadline, cron, jobID)
	return args.Error(0)
}

// StopProcess mocks stopping a process
func (m *MockProcessManager) StopProcess(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// RestartProcess mocks restarting a process
func (m *MockProcessManager) RestartProcess(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// DeleteProcess mocks deleting a process
func (m *MockProcessManager) DeleteProcess(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// GetProcessStatus mocks getting a process status
func (m *MockProcessManager) GetProcessStatus(name string) (*processmanager.ProcessInfo, error) {
	args := m.Called(name)
	return args.Get(0).(*processmanager.ProcessInfo), args.Error(1)
}

// ListProcesses mocks listing processes
func (m *MockProcessManager) ListProcesses() []*processmanager.ProcessInfo {
	args := m.Called()
	return args.Get(0).([]*processmanager.ProcessInfo)
}

// GetProcessLogs mocks getting process logs
func (m *MockProcessManager) GetProcessLogs(name string, lines int) (string, error) {
	args := m.Called(name, lines)
	return args.String(0), args.Error(1)
}

func TestTelnetAdapter_StartStop(t *testing.T) {
	// Create a temporary socket path
	tempDir, err := os.MkdirTemp("", "telnet_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	socketPath := filepath.Join(tempDir, "test.sock")

	// Create mock process manager
	mockPM := &MockProcessManager{
		secret: "test-secret",
	}

	// Create telnet adapter
	adapter := NewTelnetAdapter(mockPM)
	assert.NotNil(t, adapter)

	// Start the telnet server
	err = adapter.Start(socketPath)
	assert.NoError(t, err)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check that the socket file exists
	_, err = os.Stat(socketPath)
	assert.NoError(t, err)

	// Stop the telnet server
	err = adapter.Stop()
	assert.NoError(t, err)

	// Give it a moment to stop
	time.Sleep(100 * time.Millisecond)
}

func TestTelnetAdapter_HandleCommands(t *testing.T) {
	// Create mock process manager
	mockPM := &MockProcessManager{
		secret: "test-secret",
	}

	// Setup expectations
	mockPM.On("StartProcess", "test-process", "echo hello", true, 0, "", "").
		Return(nil)
	mockPM.On("StopProcess", "test-process").
		Return(nil)
	mockPM.On("RestartProcess", "test-process").
		Return(nil)
	mockPM.On("DeleteProcess", "test-process").
		Return(nil)
	mockPM.On("GetProcessStatus", "test-process").
		Return(&processmanager.ProcessInfo{
			Name:       "test-process",
			Command:    "echo hello",
			PID:        1234,
			Status:     processmanager.ProcessStatusRunning,
			CPUPercent: 1.5,
			MemoryMB:   10.5,
			StartTime:  time.Now(),
			LogEnabled: true,
		}, nil)
	mockPM.On("ListProcesses").
		Return([]*processmanager.ProcessInfo{
			{
				Name:       "test-process",
				Command:    "echo hello",
				PID:        1234,
				Status:     processmanager.ProcessStatusRunning,
				CPUPercent: 1.5,
				MemoryMB:   10.5,
				StartTime:  time.Now(),
				LogEnabled: true,
			},
		})
	mockPM.On("GetProcessLogs", "test-process", 10).
		Return("Log line 1\nLog line 2\n", nil)

	// Create telnet adapter
	adapter := NewTelnetAdapter(mockPM)
	assert.NotNil(t, adapter)

	// Test individual handler functions directly instead of through executeHeroscript
	// Test handleProcessStart
	result := adapter.handleProcessStart("test-process", "echo hello", true, 0, "", "")
	assert.Contains(t, result, "Process 'test-process' started successfully")

	// Test handleProcessStop
	result = adapter.handleProcessStop("test-process")
	assert.Contains(t, result, "Process 'test-process' stopped successfully")

	// Test handleProcessRestart
	result = adapter.handleProcessRestart("test-process")
	assert.Contains(t, result, "Process 'test-process' restarted successfully")

	// Test handleProcessDelete
	result = adapter.handleProcessDelete("test-process")
	assert.Contains(t, result, "Process 'test-process' deleted successfully")

	// Test handleProcessStatus
	result = adapter.handleProcessStatus("test-process", "text")
	assert.Contains(t, result, "Name: test-process")
	assert.Contains(t, result, "Status: running")

	// Test handleProcessList
	result = adapter.handleProcessList()
	assert.Contains(t, result, "test-process")

	// Test handleProcessLog
	result = adapter.handleProcessLog("test-process", 10, false)
	assert.Contains(t, result, "Log line 1")
	assert.Contains(t, result, "Log line 2")

	// Test generateHelpText
	result = adapter.generateHelpText(false)
	assert.Contains(t, result, "Available Commands")

	// Verify all expectations were met
	mockPM.AssertExpectations(t)
}

func TestTelnetAdapter_ExecuteHeroscript_InvalidCommand(t *testing.T) {
	// Create mock process manager
	mockPM := &MockProcessManager{
		secret: "test-secret",
	}

	// Create telnet adapter
	adapter := NewTelnetAdapter(mockPM)
	assert.NotNil(t, adapter)

	// Test invalid command
	result := adapter.executeHeroscript("!!invalid", false)
	assert.Contains(t, result, "Error: unknown command")
}

func TestFormatError(t *testing.T) {
	// Test with nil error
	result := FormatError(nil, false)
	assert.Equal(t, "", result)

	// Test with error in non-interactive mode
	result = FormatError(fmt.Errorf("test error"), false)
	assert.Equal(t, "Error: test error\n", result)

	// Test with error in interactive mode
	result = FormatError(fmt.Errorf("test error"), true)
	assert.Contains(t, result, "Error: test error")
}

func TestFormatResult(t *testing.T) {
	// Test without job ID
	result := FormatResult("test result", "", false)
	assert.Equal(t, "test result", result)

	// Test with job ID
	result = FormatResult("test result", "job123", false)
	assert.Equal(t, "jobid: job123\ntest result", result)
}

func TestFormatTable(t *testing.T) {
	headers := []string{"Header1", "Header2"}
	rows := [][]string{
		{"Value1", "Value2"},
		{"Value3", "Value4"},
	}

	// Test in non-interactive mode
	result := FormatTable(headers, rows, false)
	assert.Contains(t, result, "Header1")
	assert.Contains(t, result, "Header2")
	assert.Contains(t, result, "Value1")
	assert.Contains(t, result, "Value2")
	assert.Contains(t, result, "Value3")
	assert.Contains(t, result, "Value4")

	// Test in interactive mode
	result = FormatTable(headers, rows, true)
	assert.Contains(t, result, "Header1")
	assert.Contains(t, result, "Header2")
	assert.Contains(t, result, "Value1")
	assert.Contains(t, result, "Value2")
	assert.Contains(t, result, "Value3")
	assert.Contains(t, result, "Value4")
}
