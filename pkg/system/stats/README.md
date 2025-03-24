# System Stats Package

The `stats` package provides a comprehensive solution for collecting, caching, and retrieving system statistics in Go applications. It uses Redis for caching to minimize system resource usage when frequently accessing system metrics.

## Overview

This package offers a thread-safe, configurable system for monitoring:
- CPU usage and information
- Memory utilization
- Disk space and usage
- Process statistics
- Network speed and throughput
- Hardware information

The `StatsManager` provides a central interface for accessing all system statistics with built-in caching, background updates, and configurable expiration times.

## Key Components

### StatsManager

The core component that manages all statistics collection with Redis-based caching:

- Thread-safe operations with mutex protection
- Background worker for asynchronous updates
- Configurable expiration times for different stat types
- Debug mode for direct fetching without caching
- Automatic cache initialization and updates

### Statistics Types

The package collects various system statistics:

1. **System Information** (`system`)
   - CPU details (cores, model, usage percentage)
   - Memory information (total, used, free, usage percentage)
   - Network speed (upload and download)

2. **Disk Statistics** (`disk`)
   - Information for all mounted partitions
   - Total, used, and free space
   - Usage percentages

3. **Process Statistics** (`process`)
   - List of running processes
   - CPU and memory usage per process
   - Process status and creation time
   - Command line information

4. **Network Speed** (`network`)
   - Current upload and download speeds
   - Formatted with appropriate units (Mbps/Kbps)

5. **Hardware Statistics** (`hardware`)
   - Combined system, disk, and network information
   - Formatted for easy display or JSON responses

## Configuration

The `Config` struct allows customization of:

- Redis connection settings (address, password, database)
- Expiration times for different types of statistics
- Debug mode toggle
- Default timeout for waiting for stats
- Maximum queue size for update requests

Default configuration is provided through the `DefaultConfig()` function.

## Usage Examples

### Basic Usage

```go
// Create a stats manager with default settings
manager, err := stats.NewStatsManagerWithDefaults()
if err != nil {
    log.Fatalf("Error creating stats manager: %v", err)
}
defer manager.Close()

// Get system information
sysInfo, err := manager.GetSystemInfo()
if err != nil {
    log.Printf("Error getting system info: %v", err)
} else {
    fmt.Printf("CPU Cores: %d\n", sysInfo.CPU.Cores)
    fmt.Printf("Memory Used: %.1f GB (%.1f%%)\n", 
        sysInfo.Memory.Used, sysInfo.Memory.UsedPercent)
}

// Get disk statistics
diskStats, err := manager.GetDiskStats()
if err != nil {
    log.Printf("Error getting disk stats: %v", err)
} else {
    for _, disk := range diskStats.Disks {
        fmt.Printf("%s: %.1f GB total, %.1f GB free (%.1f%% used)\n", 
            disk.Path, disk.Total, disk.Free, disk.UsedPercent)
    }
}

// Get top processes by CPU usage
processes, err := manager.GetTopProcesses(5)
if err != nil {
    log.Printf("Error getting top processes: %v", err)
} else {
    for i, proc := range processes {
        fmt.Printf("%d. %s (PID: %d, CPU: %.1f%%, Memory: %.1f MB)\n", 
            i+1, proc.Name, proc.PID, proc.CPUPercent, proc.MemoryMB)
    }
}
```

### Custom Configuration

```go
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
        "process":  30 * time.Second,  // Process info expires after 30 seconds
        "network":  30 * time.Second,  // Network info expires after 30 seconds
        "hardware": 120 * time.Second, // Hardware stats expire after 2 minutes
    },
}

manager, err := stats.NewStatsManager(config)
if err != nil {
    log.Fatalf("Error creating stats manager: %v", err)
}
defer manager.Close()
```

### Force Updates and Cache Management

```go
// Force an immediate update of system stats
err := manager.ForceUpdate("system")
if err != nil {
    log.Printf("Error forcing update: %v", err)
}

// Clear the cache for a specific stats type
err = manager.ClearCache("disk")
if err != nil {
    log.Printf("Error clearing cache: %v", err)
}

// Clear the entire cache
err = manager.ClearCache("")
if err != nil {
    log.Printf("Error clearing entire cache: %v", err)
}
```

### Debug Mode

```go
// Enable debug mode to bypass caching
manager.Debug = true

// Get stats directly without using cache
sysInfo, err := manager.GetSystemInfo()

// Disable debug mode to resume caching
manager.Debug = false
```

