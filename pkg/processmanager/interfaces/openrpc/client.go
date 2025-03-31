package openrpc

import (
	"encoding/json"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager/client"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
)

// Client provides a client for interacting with process manager operations via RPC
type Client struct {
	client.BaseClient
	secret string
}

// NewClient creates a new client for the process manager API
func NewClient(socketPath, secret string) *Client {
	return &Client{
		BaseClient: *client.NewClient(socketPath, secret),
		secret:     secret,
	}
}

// Secret returns the client's secret
func (c *Client) Secret() string {
	return c.secret
}

// StartProcess starts a new process with the given name and command
func (c *Client) StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) (interfaces.ProcessStartResult, error) {
	params := map[string]interface{}{
		"name":        name,
		"command":     command,
		"log_enabled": logEnabled,
		"deadline":    deadline,
		"cron":        cron,
		"job_id":      jobID,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ProcessStartResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.start", paramsJSON, "")
	if err != nil {
		return interfaces.ProcessStartResult{}, fmt.Errorf("failed to start process: %w", err)
	}

	// Convert result to ProcessStartResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ProcessStartResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var startResult interfaces.ProcessStartResult
	if err := json.Unmarshal(resultJSON, &startResult); err != nil {
		return interfaces.ProcessStartResult{}, fmt.Errorf("failed to unmarshal process start result: %w", err)
	}

	return startResult, nil
}

// StopProcess stops a running process
func (c *Client) StopProcess(name string) (interfaces.ProcessStopResult, error) {
	params := map[string]string{
		"name": name,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ProcessStopResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.stop", paramsJSON, "")
	if err != nil {
		return interfaces.ProcessStopResult{}, fmt.Errorf("failed to stop process: %w", err)
	}

	// Convert result to ProcessStopResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ProcessStopResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var stopResult interfaces.ProcessStopResult
	if err := json.Unmarshal(resultJSON, &stopResult); err != nil {
		return interfaces.ProcessStopResult{}, fmt.Errorf("failed to unmarshal process stop result: %w", err)
	}

	return stopResult, nil
}

// RestartProcess restarts a process
func (c *Client) RestartProcess(name string) (interfaces.ProcessRestartResult, error) {
	params := map[string]string{
		"name": name,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ProcessRestartResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.restart", paramsJSON, "")
	if err != nil {
		return interfaces.ProcessRestartResult{}, fmt.Errorf("failed to restart process: %w", err)
	}

	// Convert result to ProcessRestartResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ProcessRestartResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var restartResult interfaces.ProcessRestartResult
	if err := json.Unmarshal(resultJSON, &restartResult); err != nil {
		return interfaces.ProcessRestartResult{}, fmt.Errorf("failed to unmarshal process restart result: %w", err)
	}

	return restartResult, nil
}

// DeleteProcess deletes a process from the manager
func (c *Client) DeleteProcess(name string) (interfaces.ProcessDeleteResult, error) {
	params := map[string]string{
		"name": name,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ProcessDeleteResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.delete", paramsJSON, "")
	if err != nil {
		return interfaces.ProcessDeleteResult{}, fmt.Errorf("failed to delete process: %w", err)
	}

	// Convert result to ProcessDeleteResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ProcessDeleteResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var deleteResult interfaces.ProcessDeleteResult
	if err := json.Unmarshal(resultJSON, &deleteResult); err != nil {
		return interfaces.ProcessDeleteResult{}, fmt.Errorf("failed to unmarshal process delete result: %w", err)
	}

	return deleteResult, nil
}

// GetProcessStatus gets the status of a process
func (c *Client) GetProcessStatus(name string, format string) (interface{}, error) {
	params := map[string]string{
		"name":   name,
		"format": format,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.status", paramsJSON, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	if format == "text" {
		// For text format, return the raw result
		return result, nil
	}

	// For JSON format, convert to ProcessStatus
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var statusResult interfaces.ProcessStatus
	if err := json.Unmarshal(resultJSON, &statusResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process status result: %w", err)
	}

	return statusResult, nil
}

// ListProcesses lists all processes
func (c *Client) ListProcesses(format string) (interface{}, error) {
	params := map[string]string{
		"format": format,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.list", paramsJSON, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %w", err)
	}

	if format == "text" {
		// For text format, return the raw result
		return result, nil
	}

	// For JSON format, convert to []ProcessStatus
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	var listResult []interfaces.ProcessStatus
	if err := json.Unmarshal(resultJSON, &listResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process list result: %w", err)
	}

	return listResult, nil
}

// GetProcessLogs gets logs for a process
func (c *Client) GetProcessLogs(name string, lines int) (interfaces.ProcessLogResult, error) {
	params := map[string]interface{}{
		"name":  name,
		"lines": lines,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return interfaces.ProcessLogResult{}, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	result, err := c.Request("process.log", paramsJSON, "")
	if err != nil {
		return interfaces.ProcessLogResult{}, fmt.Errorf("failed to get process logs: %w", err)
	}

	// Convert result to ProcessLogResult
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return interfaces.ProcessLogResult{}, fmt.Errorf("failed to marshal result: %w", err)
	}

	var logResult interfaces.ProcessLogResult
	if err := json.Unmarshal(resultJSON, &logResult); err != nil {
		return interfaces.ProcessLogResult{}, fmt.Errorf("failed to unmarshal process log result: %w", err)
	}

	return logResult, nil
}
