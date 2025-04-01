package pages

import (
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/processmanager"
)

// ProcessDisplayInfo represents information about a process for display purposes
type ProcessDisplayInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	StartTime string `json:"start_time"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
}

// ConvertToDisplayInfo converts a ProcessInfo from the processmanager package to ProcessDisplayInfo
func ConvertToDisplayInfo(info *processmanager.ProcessInfo) ProcessDisplayInfo {
	// Calculate uptime from start time
	uptime := formatUptime(time.Since(info.StartTime))
	
	return ProcessDisplayInfo{
		ID:        fmt.Sprintf("%d", info.PID),
		Name:      info.Name,
		Status:    string(info.Status),
		Uptime:    uptime,
		StartTime: info.StartTime.Format("2006-01-02 15:04:05"),
		CPU:       fmt.Sprintf("%.2f%%", info.CPUPercent),
		Memory:    fmt.Sprintf("%.2f MB", info.MemoryMB),
	}
}

// formatUptime formats a duration as a human-readable uptime string
func formatUptime(duration time.Duration) string {
	totalSeconds := int(duration.Seconds())
	days := totalSeconds / (24 * 3600)
	hours := (totalSeconds % (24 * 3600)) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d minutes, %d seconds", minutes, seconds)
	} else {
		return fmt.Sprintf("%d seconds", seconds)
	}
}
