package processmanagerhandler

import (
	"fmt"

	"github.com/freeflowuniverse/herolauncher/pkg/handlerfactory/core"
	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

// ProcessManagerHandler handles process manager-related actions
type ProcessManagerHandler struct {
	core.BaseHandler
	pm *processmanager.ProcessManager
}

// NewProcessManagerHandler creates a new process manager handler
func NewProcessManagerHandler() *ProcessManagerHandler {
	return &ProcessManagerHandler{
		BaseHandler: core.BaseHandler{
			ActorName: "process",
		},
		pm: processmanager.NewProcessManager(""), // Empty string as secret was removed from ProcessManager
	}
}

// Start handles the process.start action
func (h *ProcessManagerHandler) Start(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	command := params.Get("command")
	if command == "" {
		return "Error: Command is required"
	}

	logEnabled := params.GetBoolDefault("log", true)
	deadline := params.GetIntDefault("deadline", 0)
	cron := params.Get("cron")
	jobID := params.Get("job_id")

	err = h.pm.StartProcess(name, command, logEnabled, deadline, cron, jobID)
	if err != nil {
		return fmt.Sprintf("Error starting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' started successfully", name)
}

// Stop handles the process.stop action
func (h *ProcessManagerHandler) Stop(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	err = h.pm.StopProcess(name)
	if err != nil {
		return fmt.Sprintf("Error stopping process: %v", err)
	}

	return fmt.Sprintf("Process '%s' stopped successfully", name)
}

// Restart handles the process.restart action
func (h *ProcessManagerHandler) Restart(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	err = h.pm.RestartProcess(name)
	if err != nil {
		return fmt.Sprintf("Error restarting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' restarted successfully", name)
}

// Delete handles the process.delete action
func (h *ProcessManagerHandler) Delete(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	err = h.pm.DeleteProcess(name)
	if err != nil {
		return fmt.Sprintf("Error deleting process: %v", err)
	}

	return fmt.Sprintf("Process '%s' deleted successfully", name)
}

// List handles the process.list action
func (h *ProcessManagerHandler) List(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	processes := h.pm.ListProcesses()
	if len(processes) == 0 {
		return "No processes found"
	}

	format := params.Get("format")
	if format == "" {
		format = "json"
	}

	output, err := processmanager.FormatProcessList(processes, format)
	if err != nil {
		return fmt.Sprintf("Error formatting process list: %v", err)
	}

	return output
}

// Status handles the process.status action
func (h *ProcessManagerHandler) Status(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	procInfo, err := h.pm.GetProcessStatus(name)
	if err != nil {
		return fmt.Sprintf("Error getting process status: %v", err)
	}

	format := params.Get("format")
	if format == "" {
		format = "json"
	}

	output, err := processmanager.FormatProcessInfo(procInfo, format)
	if err != nil {
		return fmt.Sprintf("Error formatting process status: %v", err)
	}

	return output
}

// Logs handles the process.logs action
func (h *ProcessManagerHandler) Logs(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	name := params.Get("name")
	if name == "" {
		return "Error: Process name is required"
	}

	lines := params.GetIntDefault("lines", 100)

	logs, err := h.pm.GetProcessLogs(name, lines)
	if err != nil {
		return fmt.Sprintf("Error getting process logs: %v", err)
	}

	return logs
}

// SetLogsPath handles the process.set_logs_path action
func (h *ProcessManagerHandler) SetLogsPath(script string) string {
	params, err := h.ParseParams(script)
	if err != nil {
		return fmt.Sprintf("Error parsing parameters: %v", err)
	}

	path := params.Get("path")
	if path == "" {
		return "Error: Path is required"
	}

	h.pm.SetLogsBasePath(path)
	return fmt.Sprintf("Process logs path set to '%s'", path)
}

// Help handles the process.help action
func (h *ProcessManagerHandler) Help(script string) string {
	return `Process Manager Handler Commands:
  process.start name:<name> command:<command> [log:true|false] [deadline:<seconds>] [cron:<cron_expr>] [job_id:<id>]
  process.stop name:<name>
  process.restart name:<name>
  process.delete name:<name>
  process.list [format:json|table|text]
  process.status name:<name> [format:json|table|text]
  process.logs name:<name> [lines:<count>]
  process.set_logs_path path:<path>
  process.help`
}
