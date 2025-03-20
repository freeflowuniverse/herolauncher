package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// StatsManager is a factory for managing system statistics with caching
type StatsManager struct {
	// Redis client for caching
	redisClient *redis.Client

	// Expiration times for different types of stats in seconds
	Expiration map[string]time.Duration

	// Debug mode - if true, requests are direct without caching
	Debug bool

	// Queue for requesting stats updates
	updateQueue chan string

	// Mutex for thread-safe operations
	mu sync.Mutex

	// Context for controlling the background goroutine
	ctx    context.Context
	cancel context.CancelFunc
	
	// Default timeout for waiting for stats
	defaultTimeout time.Duration
	
	// Logger for StatsManager operations
	logger *log.Logger
}

// NewStatsManager creates a new StatsManager with Redis connection
func NewStatsManager(config *Config) (*StatsManager, error) {
	// Use default config if nil is provided
	if config == nil {
		config = DefaultConfig()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create context with cancel for the background goroutine
	ctx, cancel := context.WithCancel(ctx)

	// Create logger
	logger := log.New(os.Stdout, "[StatsManager] ", log.LstdFlags)

	// Create the manager
	manager := &StatsManager{
		redisClient:    client,
		Expiration:     config.ExpirationTimes,
		Debug:          config.Debug,
		updateQueue:    make(chan string, config.QueueSize),
		ctx:            ctx,
		cancel:         cancel,
		defaultTimeout: config.DefaultTimeout,
		logger:         logger,
	}

	// Start the background goroutine for updates
	go manager.updateWorker()

	// Initialize cache with first fetch
	manager.initializeCache()

	return manager, nil
}

// NewStatsManagerWithDefaults creates a new StatsManager with default settings
func NewStatsManagerWithDefaults() (*StatsManager, error) {
	return NewStatsManager(DefaultConfig())
}

// Close closes the StatsManager and its connections
func (sm *StatsManager) Close() error {
	// Stop the background goroutine
	sm.cancel()

	// Close Redis connection
	return sm.redisClient.Close()
}

// updateWorker is a background goroutine that processes update requests
func (sm *StatsManager) updateWorker() {
	sm.logger.Println("Starting stats update worker goroutine")
	for {
		select {
		case <-sm.ctx.Done():
			// Context cancelled, exit the goroutine
			sm.logger.Println("Stopping stats update worker goroutine")
			return
		case statsType := <-sm.updateQueue:
			// Process the update request
			sm.logger.Printf("Processing update request for %s stats", statsType)
			sm.fetchAndCacheStats(statsType)
		}
	}
}

// fetchAndCacheStats fetches stats and caches them in Redis
func (sm *StatsManager) fetchAndCacheStats(statsType string) {
	var data interface{}
	var err error
	
	sm.logger.Printf("Fetching %s stats", statsType)
	startTime := time.Now()
	
	// Fetch the requested stats
	switch statsType {
	case "system":
		data, err = GetSystemInfo()
	case "disk":
		data, err = GetDiskStats()
	case "process":
		data, err = GetProcessStats(0) // Get all processes
	case "root_disk":
		data, err = GetRootDiskInfo()
	case "network":
		data = GetNetworkSpeedResult()
	case "hardware":
		data = GetHardwareStatsJSON()
	default:
		sm.logger.Printf("Unknown stats type: %s", statsType)
		return // Unknown stats type
	}
	
	if err != nil {
		// Log error but continue
		sm.logger.Printf("Error fetching %s stats: %v", statsType, err)
		return
	}
	
	// Marshal to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		sm.logger.Printf("Error marshaling %s stats: %v", statsType, err)
		return
	}
	
	// Cache in Redis
	key := fmt.Sprintf("stats:%s", statsType)
	err = sm.redisClient.Set(sm.ctx, key, jsonData, sm.Expiration[statsType]).Err()
	if err != nil {
		sm.logger.Printf("Error caching %s stats: %v", statsType, err)
		return
	}
	
	// Set last update time
	lastUpdateKey := fmt.Sprintf("stats:%s:last_update", statsType)
	sm.redisClient.Set(sm.ctx, lastUpdateKey, time.Now().Unix(), 0)
	
	sm.logger.Printf("Successfully cached %s stats in %v", statsType, time.Since(startTime))
}

// initializeCache initializes the cache with initial values
func (sm *StatsManager) initializeCache() {
	sm.logger.Println("Initializing stats cache")
	
	// Queue initial fetches for all stats types
	statsTypes := []string{"system", "disk", "process", "root_disk", "network", "hardware"}
	for _, statsType := range statsTypes {
		sm.logger.Printf("Queueing initial fetch for %s stats", statsType)
		sm.updateQueue <- statsType
	}
}

// getFromCache gets stats from cache or triggers an update if expired
func (sm *StatsManager) getFromCache(statsType string, result interface{}) error {
	// In debug mode, fetch directly without caching
	if sm.Debug {
		sm.logger.Printf("Debug mode enabled, fetching %s stats directly", statsType)
		return sm.fetchDirect(statsType, result)
	}
	
	key := fmt.Sprintf("stats:%s", statsType)
	sm.logger.Printf("Getting %s stats from cache", statsType)
	
	// Get from Redis
	jsonData, err := sm.redisClient.Get(sm.ctx, key).Bytes()
	if err == redis.Nil {
		// Not in cache, fetch directly and wait
		sm.logger.Printf("%s stats not found in cache, fetching directly", statsType)
		return sm.fetchDirectAndCache(statsType, result)
	} else if err != nil {
		sm.logger.Printf("Redis error when getting %s stats: %v", statsType, err)
		return fmt.Errorf("redis error: %w", err)
	}
	
	// Unmarshal the data
	if err := json.Unmarshal(jsonData, result); err != nil {
		sm.logger.Printf("Error unmarshaling %s stats: %v", statsType, err)
		return fmt.Errorf("error unmarshaling data: %w", err)
	}
	
	// Check if data is expired
	lastUpdateKey := fmt.Sprintf("stats:%s:last_update", statsType)
	lastUpdateStr, err := sm.redisClient.Get(sm.ctx, lastUpdateKey).Result()
	if err == nil {
		var lastUpdate int64
		fmt.Sscanf(lastUpdateStr, "%d", &lastUpdate)
		
		// If expired, queue an update for next time
		expiration := sm.Expiration[statsType]
		updateTime := time.Unix(lastUpdate, 0)
		age := time.Since(updateTime)
		
		sm.logger.Printf("%s stats age: %v (expiration: %v)", statsType, age, expiration)
		
		if age > expiration {
			sm.logger.Printf("%s stats expired, queueing update for next request", statsType)
			// Queue update for next request
			select {
			case sm.updateQueue <- statsType:
				// Successfully queued
				sm.logger.Printf("Successfully queued %s stats update", statsType)
			default:
				// Queue is full, skip update
				sm.logger.Printf("Update queue full, skipping %s stats update", statsType)
			}
		}
	}
	
	return nil
}

// fetchDirect fetches stats directly without caching
func (sm *StatsManager) fetchDirect(statsType string, result interface{}) error {
	var data interface{}
	var err error

	// Fetch the requested stats
	switch statsType {
	case "system":
		data, err = GetSystemInfo()
	case "disk":
		data, err = GetDiskStats()
	case "process":
		data, err = GetProcessStats(0) // Get all processes
	case "root_disk":
		data, err = GetRootDiskInfo()
	case "network":
		data = GetNetworkSpeedResult()
	case "hardware":
		data = GetHardwareStatsJSON()
	default:
		return fmt.Errorf("unknown stats type: %s", statsType)
	}

	if err != nil {
		return err
	}

	// Convert data to the expected type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, result)
}

// fetchDirectAndCache fetches stats directly and caches them
func (sm *StatsManager) fetchDirectAndCache(statsType string, result interface{}) error {
	var data interface{}
	var err error

	// Fetch the requested stats
	switch statsType {
	case "system":
		data, err = GetSystemInfo()
	case "disk":
		data, err = GetDiskStats()
	case "process":
		data, err = GetProcessStats(0) // Get all processes
	case "root_disk":
		data, err = GetRootDiskInfo()
	case "network":
		data = GetNetworkSpeedResult()
	case "hardware":
		data = GetHardwareStatsJSON()
	default:
		return fmt.Errorf("unknown stats type: %s", statsType)
	}

	if err != nil {
		return err
	}

	// Convert data to the expected type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Cache in Redis
	key := fmt.Sprintf("stats:%s", statsType)
	err = sm.redisClient.Set(sm.ctx, key, jsonData, sm.Expiration[statsType]).Err()
	if err != nil {
		return err
	}

	// Set last update time
	lastUpdateKey := fmt.Sprintf("stats:%s:last_update", statsType)
	sm.redisClient.Set(sm.ctx, lastUpdateKey, time.Now().Unix(), 0)

	return json.Unmarshal(jsonData, result)
}

// waitForCachedData waits for data to be available in cache with timeout
func (sm *StatsManager) waitForCachedData(statsType string, timeout time.Duration) bool {
	key := fmt.Sprintf("stats:%s", statsType)
	startTime := time.Now()

	sm.logger.Printf("Waiting for %s stats to be available in cache (timeout: %v)", statsType, timeout)

	for {
		// Check if data exists
		exists, err := sm.redisClient.Exists(sm.ctx, key).Result()
		if err == nil && exists > 0 {
			sm.logger.Printf("%s stats found in cache after %v", statsType, time.Since(startTime))
			return true
		}

		// Check timeout
		if time.Since(startTime) > timeout {
			sm.logger.Printf("Timeout waiting for %s stats in cache", statsType)
			return false
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}
}

// ClearCache clears all cached stats or a specific stats type
func (sm *StatsManager) ClearCache(statsType string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if statsType == "" {
		// Clear all stats
		sm.logger.Println("Clearing all cached stats")
		statsTypes := []string{"system", "disk", "process", "root_disk", "network", "hardware"}
		for _, t := range statsTypes {
			key := fmt.Sprintf("stats:%s", t)
			lastUpdateKey := fmt.Sprintf("stats:%s:last_update", t)

			sm.redisClient.Del(sm.ctx, key)
			sm.redisClient.Del(sm.ctx, lastUpdateKey)
		}
	} else {
		// Clear specific stats type
		sm.logger.Printf("Clearing cached %s stats", statsType)
		key := fmt.Sprintf("stats:%s", statsType)
		lastUpdateKey := fmt.Sprintf("stats:%s:last_update", statsType)

		sm.redisClient.Del(sm.ctx, key)
		sm.redisClient.Del(sm.ctx, lastUpdateKey)
	}

	return nil
}

// ForceUpdate forces an immediate update of stats
func (sm *StatsManager) ForceUpdate(statsType string) error {
	sm.logger.Printf("Forcing immediate update of %s stats", statsType)

	// Clear the cache for this stats type
	err := sm.ClearCache(statsType)
	if err != nil {
		return err
	}

	// Fetch and cache directly
	switch statsType {
	case "system", "disk", "process", "root_disk", "network", "hardware":
		sm.fetchAndCacheStats(statsType)
		return nil
	default:
		return fmt.Errorf("unknown stats type: %s", statsType)
	}
}

// GetSystemInfo gets system information with caching
func (sm *StatsManager) GetSystemInfo() (*SystemInfo, error) {
	var result SystemInfo

	// Try to get from cache
	err := sm.getFromCache("system", &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDiskStats gets disk statistics with caching
func (sm *StatsManager) GetDiskStats() (*DiskStats, error) {
	var result DiskStats

	// Try to get from cache
	err := sm.getFromCache("disk", &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetRootDiskInfo gets root disk information with caching
func (sm *StatsManager) GetRootDiskInfo() (*DiskInfo, error) {
	var result DiskInfo

	// Try to get from cache
	err := sm.getFromCache("root_disk", &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProcessStats gets process statistics with caching
func (sm *StatsManager) GetProcessStats(limit int) (*ProcessStats, error) {
	var result ProcessStats

	// Try to get from cache
	err := sm.getFromCache("process", &result)
	if err != nil {
		return nil, err
	}

	// Apply limit if needed
	if limit > 0 && len(result.Processes) > limit {
		result.Processes = result.Processes[:limit]
	}

	return &result, nil
}

// GetTopProcesses gets top processes by CPU usage with caching
func (sm *StatsManager) GetTopProcesses(n int) ([]ProcessInfo, error) {
	stats, err := sm.GetProcessStats(n)
	if err != nil {
		return nil, err
	}

	return stats.Processes, nil
}

// GetNetworkSpeedResult gets network speed with caching
func (sm *StatsManager) GetNetworkSpeedResult() NetworkSpeedResult {
	var result NetworkSpeedResult

	// Try to get from cache
	err := sm.getFromCache("network", &result)
	if err != nil {
		// Fallback to direct fetch on error
		uploadSpeed, downloadSpeed := GetNetworkSpeed()
		return NetworkSpeedResult{
			UploadSpeed:   uploadSpeed,
			DownloadSpeed: downloadSpeed,
		}
	}

	return result
}

// GetHardwareStats gets hardware statistics with caching
func (sm *StatsManager) GetHardwareStats() map[string]interface{} {
	var result map[string]interface{}

	// Try to get from cache
	err := sm.getFromCache("hardware", &result)
	if err != nil {
		// Fallback to direct fetch on error
		return GetHardwareStats()
	}

	return result
}

// GetHardwareStatsJSON gets hardware statistics in JSON format with caching
func (sm *StatsManager) GetHardwareStatsJSON() map[string]interface{} {
	var result map[string]interface{}

	// Try to get from cache
	err := sm.getFromCache("hardware", &result)
	if err != nil {
		// Fallback to direct fetch on error
		return GetHardwareStatsJSON()
	}

	return result
}

// GetFormattedCPUInfo gets formatted CPU info with caching
func (sm *StatsManager) GetFormattedCPUInfo() string {
	sysInfo, err := sm.GetSystemInfo()
	if err != nil {
		return "Unknown"
	}

	return fmt.Sprintf("%d cores (%s)", sysInfo.CPU.Cores, sysInfo.CPU.ModelName)
}

// GetFormattedMemoryInfo gets formatted memory info with caching
func (sm *StatsManager) GetFormattedMemoryInfo() string {
	sysInfo, err := sm.GetSystemInfo()
	if err != nil {
		return "Unknown"
	}

	return fmt.Sprintf("%.1fGB (%.1fGB used)", sysInfo.Memory.Total, sysInfo.Memory.Used)
}

// GetFormattedDiskInfo gets formatted disk info with caching
func (sm *StatsManager) GetFormattedDiskInfo() string {
	diskInfo, err := sm.GetRootDiskInfo()
	if err != nil {
		return "Unknown"
	}

	return fmt.Sprintf("%.0fGB (%.0fGB free)", diskInfo.Total, diskInfo.Free)
}

// GetFormattedNetworkInfo gets formatted network info with caching
func (sm *StatsManager) GetFormattedNetworkInfo() string {
	netSpeed := sm.GetNetworkSpeedResult()

	return fmt.Sprintf("Up: %s\nDown: %s", netSpeed.UploadSpeed, netSpeed.DownloadSpeed)
}

// GetProcessStatsJSON gets process statistics in JSON format with caching
func (sm *StatsManager) GetProcessStatsJSON(limit int) map[string]interface{} {
	// Get process stats
	processStats, err := sm.GetProcessStats(limit)
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
