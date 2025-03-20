package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
)

func main() {
	fmt.Println("System Stats Test Program")
	fmt.Println("========================")

	// DISK INFORMATION
	fmt.Println("\n1. DISK INFORMATION")
	fmt.Println("------------------")
	
	// Get all disk stats
	diskStats, err := stats.GetDiskStats()
	if err != nil {
		fmt.Printf("Error getting disk stats: %v\n", err)
	} else {
		fmt.Printf("Found %d disks:\n", len(diskStats.Disks))
		for _, disk := range diskStats.Disks {
			fmt.Printf("  %s: %.1f GB total, %.1f GB free (%.1f%% used)\n", 
				disk.Path, disk.Total, disk.Free, disk.UsedPercent)
		}
	}

	// Get root disk info
	rootDisk, err := stats.GetRootDiskInfo()
	if err != nil {
		fmt.Printf("Error getting root disk info: %v\n", err)
	} else {
		fmt.Printf("\nRoot Disk: %.1f GB total, %.1f GB free (%.1f%% used)\n", 
			rootDisk.Total, rootDisk.Free, rootDisk.UsedPercent)
	}

	// Get formatted disk info
	fmt.Printf("Formatted Disk Info: %s\n", stats.GetFormattedDiskInfo())

	// SYSTEM INFORMATION
	fmt.Println("\n2. SYSTEM INFORMATION")
	fmt.Println("--------------------")
	
	// Get system info
	sysInfo, err := stats.GetSystemInfo()
	if err != nil {
		fmt.Printf("Error getting system info: %v\n", err)
	} else {
		fmt.Println("CPU Information:")
		fmt.Printf("  Cores: %d\n", sysInfo.CPU.Cores)
		fmt.Printf("  Model: %s\n", sysInfo.CPU.ModelName)
		fmt.Printf("  Usage: %.1f%%\n", sysInfo.CPU.UsagePercent)
		
		fmt.Println("\nMemory Information:")
		fmt.Printf("  Total: %.1f GB\n", sysInfo.Memory.Total)
		fmt.Printf("  Used: %.1f GB (%.1f%%)\n", sysInfo.Memory.Used, sysInfo.Memory.UsedPercent)
		fmt.Printf("  Free: %.1f GB\n", sysInfo.Memory.Free)
		
		fmt.Println("\nNetwork Information:")
		fmt.Printf("  Upload Speed: %s\n", sysInfo.Network.UploadSpeed)
		fmt.Printf("  Download Speed: %s\n", sysInfo.Network.DownloadSpeed)
	}

	// Get network speed directly
	fmt.Println("\nNetwork Speed Test:")
	netSpeed := stats.GetNetworkSpeedResult()
	fmt.Printf("  Upload: %s\n", netSpeed.UploadSpeed)
	fmt.Printf("  Download: %s\n", netSpeed.DownloadSpeed)

	// PROCESS INFORMATION
	fmt.Println("\n3. PROCESS INFORMATION")
	fmt.Println("---------------------")
	
	// Get process stats
	processStats, err := stats.GetProcessStats(5) // Get top 5 processes
	if err != nil {
		fmt.Printf("Error getting process stats: %v\n", err)
	} else {
		fmt.Printf("Total processes: %d (showing top %d)\n", 
			processStats.Total, len(processStats.Processes))
		
		fmt.Println("\nTop Processes by CPU Usage:")
		for i, proc := range processStats.Processes {
			fmt.Printf("  %d. PID %d: %s (CPU: %.1f%%, Memory: %.1f MB)\n", 
				i+1, proc.PID, proc.Name, proc.CPUPercent, proc.MemoryMB)
		}
	}

	// Get top processes directly
	fmt.Println("\nTop 3 Processes:")
	topProcs, err := stats.GetTopProcesses(3)
	if err != nil {
		fmt.Printf("Error getting top processes: %v\n", err)
	} else {
		for i, proc := range topProcs {
			fmt.Printf("  %d. %s (PID %d)\n", i+1, proc.Name, proc.PID)
		}
	}

	// COMBINED STATS
	fmt.Println("\n4. COMBINED STATS FUNCTIONS")
	fmt.Println("--------------------------")
	
	// Hardware stats
	fmt.Println("\nHardware Stats:")
	hardwareStats := stats.GetHardwareStats()
	for key, value := range hardwareStats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Hardware stats JSON
	fmt.Println("\nHardware Stats (JSON):")
	hardwareJSON := stats.GetHardwareStatsJSON()
	prettyJSON, _ := json.MarshalIndent(hardwareJSON, "", "  ")
	fmt.Println(string(prettyJSON))

	// Process stats JSON
	fmt.Println("\nProcess Stats (JSON):")
	processJSON := stats.GetProcessStatsJSON(3) // Top 3 processes
	prettyJSON, _ = json.MarshalIndent(processJSON, "", "  ")
	fmt.Println(string(prettyJSON))

	// Wait and measure network speed again
	fmt.Println("\nWaiting 2 seconds for another network speed measurement...")
	time.Sleep(2 * time.Second)
	
	// Get updated network speed
	updatedNetSpeed := stats.GetNetworkSpeedResult()
	fmt.Println("\nUpdated Network Speed:")
	fmt.Printf("  Upload: %s\n", updatedNetSpeed.UploadSpeed)
	fmt.Printf("  Download: %s\n", updatedNetSpeed.DownloadSpeed)
}
