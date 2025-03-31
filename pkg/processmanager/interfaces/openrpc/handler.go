package openrpc

import (
	"encoding/json"
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager/interfaces"
)

// Handler implements the OpenRPC handlers for process manager operations
type Handler struct {
	processManager interfaces.ProcessManagerInterface
	secret         string
}

// NewHandler creates a new RPC handler for process manager operations
func NewHandler(processManager interfaces.ProcessManagerInterface, secret string) *Handler {
	return &Handler{
		processManager: processManager,
		secret:         secret,
	}
}

// GetHandlers returns a map of RPC handlers for the OpenRPC manager
func (h *Handler) GetHandlers() map[string]openrpcmanager.RPCHandler {
	return map[string]openrpcmanager.RPCHandler{
		"process.start":   h.handleProcessStart,
		"process.stop":    h.handleProcessStop,
		"process.restart": h.handleProcessRestart,
		"process.delete":  h.handleProcessDelete,
		"process.status":  h.handleProcessStatus,
		"process.list":    h.handleProcessList,
		"process.log":     h.handleProcessLog,
	}
}

// handleProcessStart handles the process.start method
func (h *Handler) handleProcessStart(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name       string `json:"name"`
		Command    string `json:"command"`
		LogEnabled bool   `json:"log_enabled"`
		Deadline   int    `json:"deadline"`
		Cron       string `json:"cron"`
		JobID      string `json:"job_id"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	err := h.processManager.StartProcess(
		request.Name,
		request.Command,
		request.LogEnabled,
		request.Deadline,
		request.Cron,
		request.JobID,
	)

	result := interfaces.ProcessStartResult{
		Success: err == nil,
		Message: "",
	}

	if err != nil {
		result.Message = err.Error()
		return result, nil
	}

	// Get the process info to return the PID
	procInfo, err := h.processManager.GetProcessStatus(request.Name)
	if err != nil {
		result.Message = fmt.Sprintf("Process started but failed to get status: %v", err)
		return result, nil
	}

	result.PID = procInfo.PID
	result.Message = fmt.Sprintf("Process '%s' started successfully with PID %d", request.Name, procInfo.PID)
	return result, nil
}

// handleProcessStop handles the process.stop method
func (h *Handler) handleProcessStop(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	err := h.processManager.StopProcess(request.Name)

	result := interfaces.ProcessStopResult{
		Success: err == nil,
		Message: "",
	}

	if err != nil {
		result.Message = err.Error()
	} else {
		result.Message = fmt.Sprintf("Process '%s' stopped successfully", request.Name)
	}

	return result, nil
}

// handleProcessRestart handles the process.restart method
func (h *Handler) handleProcessRestart(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	err := h.processManager.RestartProcess(request.Name)

	result := interfaces.ProcessRestartResult{
		Success: err == nil,
		Message: "",
	}

	if err != nil {
		result.Message = err.Error()
		return result, nil
	}

	// Get the process info to return the PID
	procInfo, err := h.processManager.GetProcessStatus(request.Name)
	if err != nil {
		result.Message = fmt.Sprintf("Process restarted but failed to get status: %v", err)
		return result, nil
	}

	result.PID = procInfo.PID
	result.Message = fmt.Sprintf("Process '%s' restarted successfully with PID %d", request.Name, procInfo.PID)
	return result, nil
}

// handleProcessDelete handles the process.delete method
func (h *Handler) handleProcessDelete(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	err := h.processManager.DeleteProcess(request.Name)

	result := interfaces.ProcessDeleteResult{
		Success: err == nil,
		Message: "",
	}

	if err != nil {
		result.Message = err.Error()
	} else {
		result.Message = fmt.Sprintf("Process '%s' deleted successfully", request.Name)
	}

	return result, nil
}

// handleProcessStatus handles the process.status method
func (h *Handler) handleProcessStatus(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name   string `json:"name"`
		Format string `json:"format"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	procInfo, err := h.processManager.GetProcessStatus(request.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	if request.Format == "text" {
		// Format as text using the processmanager's FormatProcessInfo function
		textResult, err := processmanager.FormatProcessInfo(procInfo, "text")
		if err != nil {
			return nil, fmt.Errorf("failed to format process info: %w", err)
		}
		return map[string]interface{}{
			"text": textResult,
		}, nil
	}

	// Default to JSON format
	return convertProcessInfoToStatus(procInfo), nil
}

// handleProcessList handles the process.list method
func (h *Handler) handleProcessList(params json.RawMessage) (interface{}, error) {
	var request struct {
		Format string `json:"format"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	processes := h.processManager.ListProcesses()

	if request.Format == "text" {
		// Format as text using the processmanager's FormatProcessList function
		textResult, err := processmanager.FormatProcessList(processes, "text")
		if err != nil {
			return nil, fmt.Errorf("failed to format process list: %w", err)
		}
		return map[string]interface{}{
			"text": textResult,
		}, nil
	}

	// Default to JSON format
	result := make([]interfaces.ProcessStatus, 0, len(processes))
	for _, proc := range processes {
		result = append(result, convertProcessInfoToStatus(proc))
	}
	return result, nil
}

// handleProcessLog handles the process.log method
func (h *Handler) handleProcessLog(params json.RawMessage) (interface{}, error) {
	var request struct {
		Name  string `json:"name"`
		Lines int    `json:"lines"`
	}
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	logs, err := h.processManager.GetProcessLogs(request.Name, request.Lines)

	result := interfaces.ProcessLogResult{
		Success: err == nil,
		Message: "",
		Logs:    logs,
	}

	if err != nil {
		result.Message = err.Error()
		result.Logs = ""
	} else {
		result.Message = fmt.Sprintf("Retrieved %d lines of logs for process '%s'", request.Lines, request.Name)
	}

	return result, nil
}

// convertProcessInfoToStatus converts a ProcessInfo to a ProcessStatus
func convertProcessInfoToStatus(info *processmanager.ProcessInfo) interfaces.ProcessStatus {
	return interfaces.ProcessStatus{
		Name:       info.Name,
		Command:    info.Command,
		PID:        info.PID,
		Status:     string(info.Status),
		CPUPercent: info.CPUPercent,
		MemoryMB:   info.MemoryMB,
		StartTime:  info.StartTime,
		LogEnabled: info.LogEnabled,
		Cron:       info.Cron,
		JobID:      info.JobID,
		Deadline:   info.Deadline,
		Error:      info.Error,
	}
}
