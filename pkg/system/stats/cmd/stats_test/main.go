package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
)

func main() {
	fmt.Println("System Stats Test Program")
	fmt.Println("========================")

	// Create a new stats manager with Redis connection
	config := &stats.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		Debug:         false,
		QueueSize:     100,
		DefaultTimeout: 5 * time.Second,
		ExpirationTimes: map[string]time.Duration{
			"system":   30 * time.Second,  // System info expires after 30 seconds
			"disk":     60 * time.Second,  // Disk info expires after 1 minute
			"process":  15 * time.Second,  // Process info expires after 15 seconds
			"network":  20 * time.Second,  // Network info expires after 20 seconds
			"hardware": 60 * time.Second,  // Hardware stats expire after 1 minute
		},
	}
	
	manager, err := stats.NewStatsManager(config)
	if err != nil {
		fmt.Printf("Error creating stats manager: %v\n", err)
		os.Exit(1)
	}
	defer manager.Close()

	// DISK INFORMATION
	fmt.Println("\n1. DISK INFORMATION")
	fmt.Println("------------------")
	
	// Get all disk stats using the manager
	diskStats, err := manager.GetDiskStats()
	if err != nil {
		fmt.Printf("Error getting disk stats: %v\n", err)
	} else {
		fmt.Printf("Found %d disks:\n", len(diskStats.Disks))
		for _, disk := range diskStats.Disks {
			fmt.Printf("  %s: %.1f GB total, %.1f GB free (%.1f%% used)\n", 
				disk.Path, disk.Total, disk.Free, disk.UsedPercent)
		}
	}

	// Get root disk info using the manager
	rootDisk, err := manager.GetRootDiskInfo()
	if err != nil {
		fmt.Printf("Error getting root disk info: %v\n", err)
	} else {
		fmt.Printf("\nRoot Disk: %.1f GB total, %.1f GB free (%.1f%% used)\n", 
			rootDisk.Total, rootDisk.Free, rootDisk.UsedPercent)
	}

	// Get formatted disk info
	fmt.Printf("Formatted Disk Info: %s\n", manager.GetFormattedDiskInfo())

	// SYSTEM INFORMATION
	fmt.Println("\n2. SYSTEM INFORMATION")
	fmt.Println("--------------------")
	
	// Get system info using the manager
	sysInfo, err := manager.GetSystemInfo()
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

	// Get network speed using the manager
	fmt.Println("\nNetwork Speed Test:")
	netSpeed := manager.GetNetworkSpeedResult()
	fmt.Printf("  Upload: %s\n", netSpeed.UploadSpeed)
	fmt.Printf("  Download: %s\n", netSpeed.DownloadSpeed)

	// PROCESS INFORMATION
	fmt.Println("\n3. PROCESS INFORMATION")
	fmt.Println("---------------------")
	
	// Get process stats using the manager
	processStats, err := manager.GetProcessStats(5) // Get top 5 processes
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

	// Get top processes using the manager
	fmt.Println("\nTop 3 Processes:")
	topProcs, err := manager.GetTopProcesses(3)
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
	
	// Hardware stats using the manager
	fmt.Println("\nHardware Stats:")
	hardwareStats := manager.GetHardwareStats()
	for key, value := range hardwareStats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	// Hardware stats JSON using the manager
	fmt.Println("\nHardware Stats (JSON):")
	hardwareJSON := manager.GetHardwareStatsJSON()
	prettyJSON, _ := json.MarshalIndent(hardwareJSON, "", "  ")
	fmt.Println(string(prettyJSON))

	// Process stats JSON using the manager
	fmt.Println("\nProcess Stats (JSON):")
	processJSON := manager.GetProcessStatsJSON(3) // Top 3 processes
	prettyJSON, _ = json.MarshalIndent(processJSON, "", "  ")
	fmt.Println(string(prettyJSON))

	// Wait and measure network speed again
	fmt.Println("\nWaiting 2 seconds for another network speed measurement...")
	time.Sleep(2 * time.Second)
	
	// Get updated network speed using the manager
	updatedNetSpeed := manager.GetNetworkSpeedResult()
	fmt.Println("\nUpdated Network Speed:")
	fmt.Printf("  Upload: %s\n", updatedNetSpeed.UploadSpeed)
	fmt.Printf("  Download: %s\n", updatedNetSpeed.DownloadSpeed)
	
	// CACHE MANAGEMENT
	fmt.Println("\n5. CACHE MANAGEMENT")
	fmt.Println("------------------")
	
	// Force update of system stats
	fmt.Println("\nForcing update of system stats...")
	err = manager.ForceUpdate("system")
	if err != nil {
		fmt.Printf("Error forcing update: %v\n", err)
	} else {
		fmt.Println("System stats updated successfully")
	}
	
	// Get updated system info
	updatedSysInfo, err := manager.GetSystemInfo()
	if err != nil {
		fmt.Printf("Error getting updated system info: %v\n", err)
	} else {
		fmt.Println("\nUpdated CPU Usage: " + fmt.Sprintf("%.1f%%", updatedSysInfo.CPU.UsagePercent))
	}
	
	// Clear cache for disk stats
	fmt.Println("\nClearing cache for disk stats...")
	err = manager.ClearCache("disk")
	if err != nil {
		fmt.Printf("Error clearing cache: %v\n", err)
	} else {
		fmt.Println("Disk stats cache cleared successfully")
	}
	
	// Toggle debug mode
	fmt.Println("\nToggling debug mode (direct fetching without cache)...")
	manager.Debug = !manager.Debug
	fmt.Printf("Debug mode is now: %v\n", manager.Debug)
}
