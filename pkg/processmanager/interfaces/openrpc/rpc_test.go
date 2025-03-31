package openrpc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessManagerRPC(t *testing.T) {
	// Create a temporary directory for the socket
	tempDir, err := os.MkdirTemp("", "processmanager-rpc-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a socket path
	socketPath := filepath.Join(tempDir, "process-manager.sock")

	// Create a test secret
	secret := "test-secret"

	// Create a process manager
	pm := processmanager.NewProcessManager(secret)
	pm.SetLogsBasePath(filepath.Join(tempDir, "logs"))

	// Create and start the server
	server, err := NewServer(pm, socketPath)
	require.NoError(t, err)

	// Start the server in a goroutine
	go func() {
		err := server.Start()
		if err != nil {
			t.Logf("Error starting server: %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a client
	client := NewClient(socketPath, secret)

	// Test process start
	t.Run("StartProcess", func(t *testing.T) {
		result, err := client.StartProcess("test-process", "echo 'Hello, World!'", true, 0, "", "")
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Message)
		assert.NotZero(t, result.PID)
	})

	// Test process status
	t.Run("GetProcessStatus", func(t *testing.T) {
		status, err := client.GetProcessStatus("test-process", "json")
		require.NoError(t, err)
		
		processStatus, ok := status.(interfaces.ProcessStatus)
		require.True(t, ok)
		assert.Equal(t, "test-process", processStatus.Name)
		assert.Equal(t, "echo 'Hello, World!'", processStatus.Command)
	})

	// Test process list
	t.Run("ListProcesses", func(t *testing.T) {
		processList, err := client.ListProcesses("json")
		require.NoError(t, err)
		
		processes, ok := processList.([]interfaces.ProcessStatus)
		require.True(t, ok)
		assert.NotEmpty(t, processes)
		
		// Find our test process
		found := false
		for _, proc := range processes {
			if proc.Name == "test-process" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	// Test process logs
	t.Run("GetProcessLogs", func(t *testing.T) {
		// Wait a bit for logs to be generated
		time.Sleep(100 * time.Millisecond)
		
		logs, err := client.GetProcessLogs("test-process", 10)
		require.NoError(t, err)
		assert.True(t, logs.Success)
	})

	// Test process restart
	t.Run("RestartProcess", func(t *testing.T) {
		result, err := client.RestartProcess("test-process")
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Message)
	})

	// Test process stop
	t.Run("StopProcess", func(t *testing.T) {
		result, err := client.StopProcess("test-process")
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Message)
	})

	// Test process delete
	t.Run("DeleteProcess", func(t *testing.T) {
		result, err := client.DeleteProcess("test-process")
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotEmpty(t, result.Message)
	})

	// Stop the server
	err = server.Stop()
	require.NoError(t, err)
}

// TestProcessManagerRPCWithMock tests the RPC interface with a mock process manager
func TestProcessManagerRPCWithMock(t *testing.T) {
	// This test would use a mock implementation of the ProcessManagerInterface
	// to test the RPC layer without actually starting real processes
	t.Skip("Mock implementation test to be added")
}
