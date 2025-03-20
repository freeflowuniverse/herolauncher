package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/system/stats"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

// TestResult stores the results of a single test run
type TestResult struct {
	StartTime       time.Time
	EndTime         time.Time
	SystemInfoTime  time.Duration
	DiskStatsTime   time.Duration
	ProcessTime     time.Duration
	NetworkTime     time.Duration
	HardwareTime    time.Duration
	TotalTime       time.Duration
	UserCPU         float64
	SystemCPU       float64
	TotalCPU        float64
	OverallCPU      float64
	MemoryUsageMB   float32
	NumGoroutines   int
}

func main() {
	// Parse command line flags
	intervalPtr := flag.Int("interval", 5, "Interval between tests in seconds")
	sleepPtr := flag.Int("sleep", 0, "Sleep time between operations in milliseconds")
	cpuProfilePtr := flag.String("cpuprofile", "", "Write cpu profile to file")
	flag.Parse()

	// If CPU profiling is enabled, set it up
	if *cpuProfilePtr != "" {
		f, err := os.Create(*cpuProfilePtr)
		if err != nil {
			fmt.Printf("Error creating CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Printf("Error starting CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	fmt.Println("StatsManager Performance Test")
	fmt.Println("============================")
	fmt.Printf("This test measures the performance of retrieving stats from Redis cache\n")
	fmt.Printf("It will run every %d seconds and print performance metrics\n", *intervalPtr)
	fmt.Printf("Sleep between operations: %d ms\n", *sleepPtr)
	fmt.Println("Press Ctrl+C to exit and view summary statistics")
	fmt.Println()

	// Create a new stats manager with Redis connection
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

	// Initialize the cache with initial values
	fmt.Println("Initializing cache with initial values...")
	_, _ = manager.GetSystemInfo()
	_, _ = manager.GetDiskStats()
	_, _ = manager.GetProcessStats(10)
	_ = manager.GetNetworkSpeedResult()
	_ = manager.GetHardwareStatsJSON()
	fmt.Println("Cache initialized. Starting performance test...")
	fmt.Println()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Create a ticker for running tests at the specified interval
	ticker := time.NewTicker(time.Duration(*intervalPtr) * time.Second)
	defer ticker.Stop()
	
	// Store the sleep duration between operations
	sleepDuration := time.Duration(*sleepPtr) * time.Millisecond

	// Get the current process for CPU and memory measurements
	currentProcess, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		fmt.Printf("Error getting current process: %v\n", err)
		os.Exit(1)
	}

	// Store test results
	var results []TestResult
	
	// Print header
	fmt.Printf("%-20s %-20s %-12s %-12s %-12s %-12s %-12s %-12s %-12s %-12s %-12s %-12s %-12s\n",
		"Start Time", "End Time", "System(ms)", "Disk(ms)", "Process(ms)", "Network(ms)", "Hardware(ms)", "Total(ms)", "UserCPU(%)", "SysCPU(%)", "TotalCPU(%)", "Memory(MB)", "Goroutines")
	fmt.Println(strings.Repeat("-", 180))

	// Run the test until interrupted
	for {
		select {
		case <-ticker.C:
			// Run a test and record the results
			result := runTest(manager, currentProcess, sleepDuration)
			results = append(results, result)
			
			// Print the result
			fmt.Printf("%-20s %-20s %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12.2f %-12d\n",
				result.StartTime.Format("15:04:05.000000"),
				result.EndTime.Format("15:04:05.000000"),
				float64(result.SystemInfoTime.Microseconds())/1000,
				float64(result.DiskStatsTime.Microseconds())/1000,
				float64(result.ProcessTime.Microseconds())/1000,
				float64(result.NetworkTime.Microseconds())/1000,
				float64(result.HardwareTime.Microseconds())/1000,
				float64(result.TotalTime.Microseconds())/1000,
				result.UserCPU,
				result.SystemCPU,
				result.TotalCPU,
				result.MemoryUsageMB,
				result.NumGoroutines)
			
		case <-sigChan:
			// Calculate and print summary statistics
			fmt.Println("\nTest Summary:")
			fmt.Println(strings.Repeat("-", 50))
			
			var totalSystemTime, totalDiskTime, totalProcessTime, totalNetworkTime, totalHardwareTime, totalTime time.Duration
			var totalUserCPU, totalSystemCPU, totalCombinedCPU, totalOverallCPU float64
			var totalMemory float32
			
			for _, r := range results {
				totalSystemTime += r.SystemInfoTime
				totalDiskTime += r.DiskStatsTime
				totalProcessTime += r.ProcessTime
				totalNetworkTime += r.NetworkTime
				totalHardwareTime += r.HardwareTime
				totalTime += r.TotalTime
				totalUserCPU += r.UserCPU
				totalSystemCPU += r.SystemCPU
				totalCombinedCPU += r.TotalCPU
				totalOverallCPU += r.OverallCPU
				totalMemory += r.MemoryUsageMB
			}
			
			count := float64(len(results))
			if count > 0 {
				fmt.Printf("Average System Info Time:  %.2f ms\n", float64(totalSystemTime.Microseconds())/(count*1000))
				fmt.Printf("Average Disk Stats Time:   %.2f ms\n", float64(totalDiskTime.Microseconds())/(count*1000))
				fmt.Printf("Average Process Time:      %.2f ms\n", float64(totalProcessTime.Microseconds())/(count*1000))
				fmt.Printf("Average Network Time:      %.2f ms\n", float64(totalNetworkTime.Microseconds())/(count*1000))
				fmt.Printf("Average Hardware Time:     %.2f ms\n", float64(totalHardwareTime.Microseconds())/(count*1000))
				fmt.Printf("Average Total Time:        %.2f ms\n", float64(totalTime.Microseconds())/(count*1000))
				fmt.Printf("Average User CPU:          %.2f%%\n", totalUserCPU/count)
				fmt.Printf("Average System CPU:        %.2f%%\n", totalSystemCPU/count)
				fmt.Printf("Average Process CPU:       %.2f%%\n", totalCombinedCPU/count)
				fmt.Printf("Average Overall CPU:       %.2f%%\n", totalOverallCPU/count)
				fmt.Printf("Average Memory Usage:      %.2f MB\n", float64(totalMemory)/count)
			}
			
			fmt.Println("\nTest completed. Exiting...")
			return
		}
	}
}

// runTest runs a single test iteration and returns the results
func runTest(manager *stats.StatsManager, proc *process.Process, sleepBetweenOps time.Duration) TestResult {
	// Get initial CPU times for the process
	initialTimes, _ := proc.Times()
	
	// Get initial overall CPU usage
	_, _ = cpu.Percent(0, false) // Discard initial reading, we'll only use the final reading
	
	result := TestResult{
		StartTime: time.Now(),
	}
	
	// Measure total time
	totalStart := time.Now()
	
	// Measure system info time
	start := time.Now()
	_, _ = manager.GetSystemInfo()
	result.SystemInfoTime = time.Since(start)
	
	// Sleep between operations if configured
	if sleepBetweenOps > 0 {
		time.Sleep(sleepBetweenOps)
	}
	
	// Measure disk stats time
	start = time.Now()
	_, _ = manager.GetDiskStats()
	result.DiskStatsTime = time.Since(start)
	
	// Sleep between operations if configured
	if sleepBetweenOps > 0 {
		time.Sleep(sleepBetweenOps)
	}
	
	// Measure process stats time
	start = time.Now()
	_, _ = manager.GetProcessStats(10)
	result.ProcessTime = time.Since(start)
	
	// Sleep between operations if configured
	if sleepBetweenOps > 0 {
		time.Sleep(sleepBetweenOps)
	}
	
	// Measure network speed time
	start = time.Now()
	_ = manager.GetNetworkSpeedResult()
	result.NetworkTime = time.Since(start)
	
	// Sleep between operations if configured
	if sleepBetweenOps > 0 {
		time.Sleep(sleepBetweenOps)
	}
	
	// Measure hardware stats time
	start = time.Now()
	_ = manager.GetHardwareStatsJSON()
	result.HardwareTime = time.Since(start)
	
	// Record total time
	result.TotalTime = time.Since(totalStart)
	result.EndTime = time.Now()
	
	// Get final CPU times for the process
	finalTimes, _ := proc.Times()
	
	// Calculate CPU usage for this specific operation
	if initialTimes != nil && finalTimes != nil {
		result.UserCPU = (finalTimes.User - initialTimes.User) * 100
		result.SystemCPU = (finalTimes.System - initialTimes.System) * 100
		result.TotalCPU = result.UserCPU + result.SystemCPU
	}
	
	// Get overall CPU usage
	finalOverallCPU, _ := cpu.Percent(0, false)
	if len(finalOverallCPU) > 0 {
		result.OverallCPU = finalOverallCPU[0]
	}
	
	// Measure memory usage
	memInfo, _ := proc.MemoryInfo()
	if memInfo != nil {
		result.MemoryUsageMB = float32(memInfo.RSS) / (1024 * 1024)
	}
	
	// Record number of goroutines
	result.NumGoroutines = runtime.NumGoroutine()
	
	return result
}
