package stats

import (
	"fmt"
	"math"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskInfo represents information about a disk
type DiskInfo struct {
	Path        string  `json:"path"`
	Total       float64 `json:"total_gb"`
	Free        float64 `json:"free_gb"`
	Used        float64 `json:"used_gb"`
	UsedPercent float64 `json:"used_percent"`
}

// DiskStats contains information about all disks
type DiskStats struct {
	Disks []DiskInfo `json:"disks"`
}

// GetDiskStats returns information about all disks
func GetDiskStats() (*DiskStats, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	stats := &DiskStats{
		Disks: make([]DiskInfo, 0, len(partitions)),
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		// Convert bytes to GB and round to 1 decimal place
		totalGB := math.Round(float64(usage.Total)/(1024*1024*1024)*10) / 10
		freeGB := math.Round(float64(usage.Free)/(1024*1024*1024)*10) / 10
		usedGB := math.Round(float64(usage.Used)/(1024*1024*1024)*10) / 10

		stats.Disks = append(stats.Disks, DiskInfo{
			Path:        partition.Mountpoint,
			Total:       totalGB,
			Free:        freeGB,
			Used:        usedGB,
			UsedPercent: math.Round(usage.UsedPercent*10) / 10,
		})
	}

	return stats, nil
}

// GetRootDiskInfo returns information about the root disk
func GetRootDiskInfo() (*DiskInfo, error) {
	usage, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get root disk usage: %w", err)
	}

	// Convert bytes to GB and round to 1 decimal place
	totalGB := math.Round(float64(usage.Total)/(1024*1024*1024)*10) / 10
	freeGB := math.Round(float64(usage.Free)/(1024*1024*1024)*10) / 10
	usedGB := math.Round(float64(usage.Used)/(1024*1024*1024)*10) / 10

	return &DiskInfo{
		Path:        "/",
		Total:       totalGB,
		Free:        freeGB,
		Used:        usedGB,
		UsedPercent: math.Round(usage.UsedPercent*10) / 10,
	}, nil
}

// GetFormattedDiskInfo returns a formatted string with disk information
func GetFormattedDiskInfo() string {
	diskInfo, err := GetRootDiskInfo()
	if err != nil {
		return "Unknown"
	}
	
	return fmt.Sprintf("%.0fGB (%.0fGB free)", diskInfo.Total, diskInfo.Free)
}
