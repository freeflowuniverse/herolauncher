package stats

import (
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// ProcessInfo contains information about a single process
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryMB    float64 `json:"memory_mb"`
	CreateTime  string  `json:"create_time"`
	IsCurrent   bool    `json:"is_current"`
	CommandLine string  `json:"command_line,omitempty"`
}

// ProcessStats contains information about all processes
type ProcessStats struct {
	Processes []ProcessInfo `json:"processes"`
	Total     int           `json:"total"`
	Filtered  int           `json:"filtered"`
}

// GetProcessStats returns information about all processes
func GetProcessStats(limit int) (*ProcessStats, error) {
	// Get process information
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// Process data to return
	stats := &ProcessStats{
		Processes: make([]ProcessInfo, 0, len(processes)),
		Total:     len(processes),
	}

	// Current process ID
	currentPid := int32(os.Getpid())

	// Get stats for each process
	for _, p := range processes {
		// Skip processes we can't access
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Get memory info
		memInfo, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		// Get CPU percent
		cpuPercent, err := p.CPUPercent()
		if err != nil {
			continue
		}

		// Skip processes with very minimal CPU and memory usage to reduce data size
		if cpuPercent < 0.01 && float64(memInfo.RSS)/(1024*1024) < 1 {
			continue
		}

		// Get process status
		status := "unknown"
		statusSlice, err := p.Status()
		if err == nil && len(statusSlice) > 0 {
			status = statusSlice[0]
		}

		// Get process creation time
		createTime, err := p.CreateTime()
		if err != nil {
			createTime = 0
		}

		// Format creation time as string
		createTimeStr := "N/A"
		if createTime > 0 {
			createTimeStr = time.Unix(createTime/1000, 0).Format("2006-01-02 15:04:05")
		}

		// Calculate memory in MB and round to 1 decimal place
		memoryMB := math.Round(float64(memInfo.RSS)/(1024*1024)*10) / 10

		// Round CPU percent to 1 decimal place
		cpuPercent = math.Round(cpuPercent*10) / 10

		// Check if this is the current process
		isCurrent := p.Pid == currentPid

		// Get command line (may be empty if no permission)
		cmdline := ""
		cmdlineSlice, err := p.Cmdline()
		if err == nil {
			cmdline = cmdlineSlice
		}

		// Add process info
		stats.Processes = append(stats.Processes, ProcessInfo{
			PID:         p.Pid,
			Name:        name,
			Status:      status,
			CPUPercent:  cpuPercent,
			MemoryMB:    memoryMB,
			CreateTime:  createTimeStr,
			IsCurrent:   isCurrent,
			CommandLine: cmdline,
		})
	}

	stats.Filtered = len(stats.Processes)

	// Sort processes by CPU usage (descending)
	sort.Slice(stats.Processes, func(i, j int) bool {
		return stats.Processes[i].CPUPercent > stats.Processes[j].CPUPercent
	})

	// Limit to top N processes if requested
	if limit > 0 && len(stats.Processes) > limit {
		stats.Processes = stats.Processes[:limit]
	}

	return stats, nil
}

// GetTopProcesses returns the top N processes by CPU usage
func GetTopProcesses(n int) ([]ProcessInfo, error) {
	stats, err := GetProcessStats(n)
	if err != nil {
		return nil, err
	}
	
	return stats.Processes, nil
}

// GetProcessStatsJSON returns process statistics in a format suitable for JSON responses
func GetProcessStatsJSON(limit int) map[string]interface{} {
	// Get process stats
	processStats, err := GetProcessStats(limit)
	if err != nil {
		return map[string]interface{}{
			"processes": []interface{}{},
			"total":     0,
			"filtered":  0,
		}
	}
	
	// Convert to JSON-friendly format
	return map[string]interface{}{
		"processes": processStats.Processes,
		"total":     processStats.Total,
		"filtered":  processStats.Filtered,
	}
}
