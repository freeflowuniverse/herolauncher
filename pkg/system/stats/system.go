package stats

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemInfo contains information about the system's CPU and memory
type SystemInfo struct {
	CPU     CPUInfo     `json:"cpu"`
	Memory  MemoryInfo  `json:"memory"`
	Network NetworkInfo `json:"network"`
}

// CPUInfo contains information about the CPU
type CPUInfo struct {
	Cores       int     `json:"cores"`
	ModelName   string  `json:"model_name"`
	UsagePercent float64 `json:"usage_percent"`
}

// MemoryInfo contains information about the system memory
type MemoryInfo struct {
	Total       float64 `json:"total_gb"`
	Used        float64 `json:"used_gb"`
	Free        float64 `json:"free_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkInfo contains information about network usage
type NetworkInfo struct {
	UploadSpeed   string `json:"upload_speed"`
	DownloadSpeed string `json:"download_speed"`
	BytesSent     uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
}

// NetworkSpeedResult contains the upload and download speeds
type NetworkSpeedResult struct {
	UploadSpeed   string `json:"upload_speed"`
	DownloadSpeed string `json:"download_speed"`
}

// UptimeProvider defines an interface for getting system uptime
type UptimeProvider interface {
	GetUptime() string
}

// GetSystemInfo returns information about the system's CPU and memory
func GetSystemInfo() (*SystemInfo, error) {
	// Get CPU info
	cpuInfo := CPUInfo{
		Cores:     runtime.NumCPU(),
		ModelName: "Unknown",
	}
	
	// Try to get detailed CPU info
	info, err := cpu.Info()
	if err == nil && len(info) > 0 {
		cpuInfo.ModelName = info[0].ModelName
	}
	
	// Get CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		cpuInfo.UsagePercent = math.Round(cpuPercent[0]*10) / 10
	}
	
	// Get memory info
	memInfo := MemoryInfo{}
	virtualMem, err := mem.VirtualMemory()
	if err == nil {
		memInfo.Total = float64(virtualMem.Total) / (1024 * 1024 * 1024) // Convert to GB
		memInfo.Used = float64(virtualMem.Used) / (1024 * 1024 * 1024)
		memInfo.Free = float64(virtualMem.Free) / (1024 * 1024 * 1024)
		memInfo.UsedPercent = math.Round(virtualMem.UsedPercent*10) / 10
	}
	
	// Get network speed
	uploadSpeed, downloadSpeed := GetNetworkSpeed()
	
	// Create and return the system info
	return &SystemInfo{
		CPU:    cpuInfo,
		Memory: memInfo,
		Network: NetworkInfo{
			UploadSpeed:   uploadSpeed,
			DownloadSpeed: downloadSpeed,
		},
	}, nil
}

// GetNetworkSpeed returns the current network speed in Mbps
func GetNetworkSpeed() (string, string) {
	networkUpSpeed := "Unknown"
	networkDownSpeed := "Unknown"

	// Get initial network counters
	countersStart, err := net.IOCounters(false)
	if err != nil || len(countersStart) == 0 {
		return networkUpSpeed, networkDownSpeed
	}

	// Wait a short time to measure the difference
	time.Sleep(500 * time.Millisecond)

	// Get updated network counters
	countersEnd, err := net.IOCounters(false)
	if err != nil || len(countersEnd) == 0 {
		return networkUpSpeed, networkDownSpeed
	}

	// Calculate the difference in bytes
	bytesSent := countersEnd[0].BytesSent - countersStart[0].BytesSent
	bytesRecv := countersEnd[0].BytesRecv - countersStart[0].BytesRecv

	// Convert to Mbps (megabits per second)
	// 500ms = 0.5s, so multiply by 2 to get per second
	// Then convert bytes to bits (*8) and to megabits (/1024/1024)
	mbpsSent := float64(bytesSent) * 2 * 8 / 1024 / 1024
	mbpsRecv := float64(bytesRecv) * 2 * 8 / 1024 / 1024

	// Format the speeds with appropriate units
	if mbpsSent < 1 {
		networkUpSpeed = fmt.Sprintf("%.1f Kbps", mbpsSent*1024)
	} else {
		networkUpSpeed = fmt.Sprintf("%.1f Mbps", mbpsSent)
	}

	if mbpsRecv < 1 {
		networkDownSpeed = fmt.Sprintf("%.1f Kbps", mbpsRecv*1024)
	} else {
		networkDownSpeed = fmt.Sprintf("%.1f Mbps", mbpsRecv)
	}

	return networkUpSpeed, networkDownSpeed
}

// GetNetworkSpeedResult returns the network speed as a struct
func GetNetworkSpeedResult() NetworkSpeedResult {
	uploadSpeed, downloadSpeed := GetNetworkSpeed()
	return NetworkSpeedResult{
		UploadSpeed:   uploadSpeed,
		DownloadSpeed: downloadSpeed,
	}
}

// GetFormattedCPUInfo returns a formatted string with CPU information
func GetFormattedCPUInfo() string {
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return "Unknown"
	}
	
	return fmt.Sprintf("%d cores (%s)", sysInfo.CPU.Cores, sysInfo.CPU.ModelName)
}

// GetFormattedMemoryInfo returns a formatted string with memory information
func GetFormattedMemoryInfo() string {
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return "Unknown"
	}
	
	return fmt.Sprintf("%.1fGB (%.1fGB used)", sysInfo.Memory.Total, sysInfo.Memory.Used)
}

// GetFormattedNetworkInfo returns a formatted string with network information
func GetFormattedNetworkInfo() string {
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return "Unknown"
	}
	
	return fmt.Sprintf("Up: %s\nDown: %s", sysInfo.Network.UploadSpeed, sysInfo.Network.DownloadSpeed)
}

// GetHardwareStats returns a map with hardware statistics
func GetHardwareStats() map[string]interface{} {
	// Create the hardware stats map
	hardwareStats := map[string]interface{}{
		"cpu":     GetFormattedCPUInfo(),
		"memory":  GetFormattedMemoryInfo(),
		"disk":    GetFormattedDiskInfo(),
		"network": GetFormattedNetworkInfo(),
	}
	
	return hardwareStats
}

// GetHardwareStatsJSON returns hardware statistics in a format suitable for JSON responses
func GetHardwareStatsJSON() map[string]interface{} {
	// Get system information
	sysInfo, err := GetSystemInfo()
	if err != nil {
		return map[string]interface{}{
			"error": "Failed to get system info: " + err.Error(),
		}
	}
	
	// Get disk information
	diskInfo, err := GetRootDiskInfo()
	if err != nil {
		return map[string]interface{}{
			"error": "Failed to get disk info: " + err.Error(),
		}
	}
	
	// Create the hardware stats map
	hardwareStats := map[string]interface{}{
		"cpu": map[string]interface{}{
			"cores":        sysInfo.CPU.Cores,
			"model":        sysInfo.CPU.ModelName,
			"usage_percent": sysInfo.CPU.UsagePercent,
		},
		"memory": map[string]interface{}{
			"total_gb":     sysInfo.Memory.Total,
			"used_gb":      sysInfo.Memory.Used,
			"free_gb":      sysInfo.Memory.Free,
			"used_percent": sysInfo.Memory.UsedPercent,
		},
		"disk": map[string]interface{}{
			"total_gb":     diskInfo.Total,
			"free_gb":      diskInfo.Free,
			"used_gb":      diskInfo.Used,
			"used_percent": diskInfo.UsedPercent,
		},
		"network": map[string]interface{}{
			"upload_speed":   sysInfo.Network.UploadSpeed,
			"download_speed": sysInfo.Network.DownloadSpeed,
		},
	}
	
	return hardwareStats
}
