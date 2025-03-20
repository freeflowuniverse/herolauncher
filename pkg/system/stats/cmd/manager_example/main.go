package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
)

func main() {
	fmt.Println("Stats Manager Example")
	fmt.Println("====================")

	// Create a new stats manager with Redis connection
	// Create a custom configuration
	config := &stats.Config{
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       0,
		Debug:         false,
		QueueSize:     100,
		DefaultTimeout: 5 * time.Second,
		ExpirationTimes: map[string]time.Duration{
			"system":   60 * time.Second,  // System info expires after 60 seconds
			"disk":     300 * time.Second, // Disk info expires after 5 minutes
			"process":  60 * time.Second,  // Process info expires after 1 minute
			"network":  30 * time.Second,  // Network info expires after 30 seconds
			"hardware": 120 * time.Second, // Hardware stats expire after 2 minutes
		},
	}
	
	manager, err := stats.NewStatsManager(config)
	if err != nil {
		fmt.Printf("Error creating stats manager: %v\n", err)
		os.Exit(1)
	}
	defer manager.Close()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		manager.Close()
		os.Exit(0)
	}()

	// Example 1: Get system info
	fmt.Println("\n1. SYSTEM INFORMATION")
	fmt.Println("--------------------")
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

	// Example 2: Get disk stats
	fmt.Println("\n2. DISK INFORMATION")
	fmt.Println("------------------")
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

	// Example 3: Get process stats
	fmt.Println("\n3. PROCESS INFORMATION")
	fmt.Println("---------------------")
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

	// Example 4: Demonstrate caching by getting the same data multiple times
	fmt.Println("\n4. CACHING DEMONSTRATION")
	fmt.Println("----------------------")
	fmt.Println("Getting network speed multiple times (should use cache):")
	
	for i := 0; i < 3; i++ {
		netSpeed := manager.GetNetworkSpeedResult()
		fmt.Printf("  Request %d: Upload: %s, Download: %s\n", 
			i+1, netSpeed.UploadSpeed, netSpeed.DownloadSpeed)
		time.Sleep(500 * time.Millisecond)
	}

	// Example 5: Get hardware stats JSON
	fmt.Println("\n5. HARDWARE STATS JSON")
	fmt.Println("--------------------")
	hardwareJSON := manager.GetHardwareStatsJSON()
	prettyJSON, _ := json.MarshalIndent(hardwareJSON, "", "  ")
	fmt.Println(string(prettyJSON))

	// Example 6: Debug mode demonstration
	fmt.Println("\n6. DEBUG MODE DEMONSTRATION")
	fmt.Println("--------------------------")
	fmt.Println("Enabling debug mode (direct fetching without cache)...")
	manager.Debug = true
	
	fmt.Println("Getting system info in debug mode:")
	debugSysInfo, err := manager.GetSystemInfo()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("  CPU Usage: %.1f%%\n", debugSysInfo.CPU.UsagePercent)
		fmt.Printf("  Memory Used: %.1f GB (%.1f%%)\n", 
			debugSysInfo.Memory.Used, debugSysInfo.Memory.UsedPercent)
	}
	
	// Reset debug mode
	manager.Debug = false

	// Example 7: Modify expiration times
	fmt.Println("\n7. CUSTOM EXPIRATION TIMES")
	fmt.Println("------------------------")
	fmt.Println("Current expiration times:")
	for statsType, duration := range manager.Expiration {
		fmt.Printf("  %s: %v\n", statsType, duration)
	}
	
	fmt.Println("\nChanging system stats expiration to 10 seconds...")
	manager.Expiration["system"] = 10 * time.Second
	
	fmt.Println("Updated expiration times:")
	for statsType, duration := range manager.Expiration {
		fmt.Printf("  %s: %v\n", statsType, duration)
	}

	fmt.Println("\nDemo complete. Press Ctrl+C to exit.")
	
	// Keep the program running
	select {}
}
