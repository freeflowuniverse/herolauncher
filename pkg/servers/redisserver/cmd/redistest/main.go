package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/servers/redisserver"
	"github.com/redis/go-redis/v9"
)

// TestResult represents the result of a single test
type TestResult struct {
	Name        string
	Passed      bool
	Description string
	Error       error
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Name   string
	Tests  []TestResult
	Passed int
	Failed int
}

func (ts *TestSuite) AddResult(name string, passed bool, description string, err error) {
	result := TestResult{
		Name:        name,
		Passed:      passed,
		Description: description,
		Error:       err,
	}
	ts.Tests = append(ts.Tests, result)
	if passed {
		ts.Passed++
	} else {
		ts.Failed++
	}
}

func (ts *TestSuite) PrintResults() {
	fmt.Printf("\n=== Test Suite: %s ===\n", ts.Name)
	for _, test := range ts.Tests {
		status := "✅ PASS"
		if !test.Passed {
			status = "❌ FAIL"
		}
		fmt.Printf("%s: %s - %s\n", status, test.Name, test.Description)
		if test.Error != nil {
			fmt.Printf("   Error: %v\n", test.Error)
		}
	}
	fmt.Printf("\nSummary: %d passed, %d failed\n", ts.Passed, ts.Failed)
}

func main() {
	// Parse command line flags
	tcpPort := flag.String("tcp-port", "7777", "Redis server TCP port")
	unixSocket := flag.String("unix-socket", "/tmp/redis-test.sock", "Redis server Unix domain socket path")
	debug := flag.Bool("debug", false, "Enable debug output")
	dbNum := flag.Int("db", 0, "Redis database number")
	flag.Parse()

	// Start Redis server in a goroutine
	log.Printf("Starting Redis server on TCP port %s and Unix socket %s", *tcpPort, *unixSocket)

	// Create a wait group to ensure the server is started before testing
	var wg sync.WaitGroup
	wg.Add(1)

	// Remove the Unix socket file if it already exists
	if *unixSocket != "" {
		if _, err := os.Stat(*unixSocket); err == nil {
			log.Printf("Removing existing Unix socket file: %s", *unixSocket)
			if err := os.Remove(*unixSocket); err != nil {
				log.Printf("Warning: Failed to remove existing Unix socket file: %v", err)
			}
		}
	}

	// Start the Redis server in a goroutine
	go func() {
		// Create a new server instance
		_ = redisserver.NewServer(redisserver.ServerConfig{TCPPort: *tcpPort, UnixSocketPath: *unixSocket})

		// Signal that the server is ready
		wg.Done()

		// Keep the server running
		select {}
	}()

	// Wait for the server to start
	wg.Wait()

	// Give the server a moment to initialize, especially for Unix socket
	time.Sleep(1 * time.Second)

	// Test TCP connection
	log.Println("Testing TCP connection")
	tcpAddr := fmt.Sprintf("localhost:%s", *tcpPort)
	runTests(tcpAddr, *debug, *dbNum)

	// Test Unix socket connection if supported
	log.Println("\nTesting Unix socket connection")
	runTests(*unixSocket, *debug, *dbNum)
}

func runTests(addr string, debug bool, dbNum int) {
	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Network:      getNetworkType(addr),
		Addr:         addr,
		DB:           dbNum,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	defer redisClient.Close()

	ctx := context.Background()

	// Check connection
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Connected to Redis: %s", pong)

	// Run test suites
	connectionTestSuite := &TestSuite{Name: "Connection Tests"}
	stringTestSuite := &TestSuite{Name: "String Operations"}
	hashTestSuite := &TestSuite{Name: "Hash Operations"}
	listTestSuite := &TestSuite{Name: "List Operations"}
	patternTestSuite := &TestSuite{Name: "Pattern Matching"}
	miscTestSuite := &TestSuite{Name: "Miscellaneous Operations"}

	// Run connection tests
	runConnectionTests(ctx, redisClient, connectionTestSuite)

	// Run string operation tests
	runStringTests(ctx, redisClient, stringTestSuite)

	// Run hash operation tests
	runHashTests(ctx, redisClient, hashTestSuite)

	// Run list operation tests
	runListTests(ctx, redisClient, listTestSuite)

	// Run pattern matching tests
	runPatternTests(ctx, redisClient, patternTestSuite)

	// Run miscellaneous tests
	runMiscTests(ctx, redisClient, miscTestSuite)

	// Print test results
	connectionTestSuite.PrintResults()
	stringTestSuite.PrintResults()
	hashTestSuite.PrintResults()
	listTestSuite.PrintResults()
	patternTestSuite.PrintResults()
	miscTestSuite.PrintResults()

	// Print overall summary
	totalTests := connectionTestSuite.Passed + connectionTestSuite.Failed +
		stringTestSuite.Passed + stringTestSuite.Failed +
		hashTestSuite.Passed + hashTestSuite.Failed +
		listTestSuite.Passed + listTestSuite.Failed +
		patternTestSuite.Passed + patternTestSuite.Failed +
		miscTestSuite.Passed + miscTestSuite.Failed

	totalPassed := connectionTestSuite.Passed + stringTestSuite.Passed +
		hashTestSuite.Passed + listTestSuite.Passed +
		patternTestSuite.Passed + miscTestSuite.Passed

	totalFailed := connectionTestSuite.Failed + stringTestSuite.Failed +
		hashTestSuite.Failed + listTestSuite.Failed +
		patternTestSuite.Failed + miscTestSuite.Failed

	fmt.Printf("\n=== Overall Test Summary ===\n")
	fmt.Printf("Total Tests: %d\n", totalTests)
	fmt.Printf("Passed: %d\n", totalPassed)
	fmt.Printf("Failed: %d\n", totalFailed)
	fmt.Printf("Success Rate: %.2f%%\n", float64(totalPassed)/float64(totalTests)*100)

	// Debug output
	if debug {
		log.Println("\nListing keys in Redis using SCAN:")
		var cursor uint64
		var allKeys []string

		for {
			keys, nextCursor, err := redisClient.Scan(ctx, cursor, "*", 10).Result()
			if err != nil {
				log.Printf("Error scanning keys: %v", err)
				break
			}

			allKeys = append(allKeys, keys...)
			if nextCursor == 0 {
				break
			}
			cursor = nextCursor
		}

		log.Printf("Found %d total keys using SCAN", len(allKeys))
		for i, k := range allKeys {
			if i < 20 { // Limit output to first 20 keys
				log.Printf("Key[%d]: %s", i, k)
			}
		}
		if len(allKeys) > 20 {
			log.Printf("... and %d more keys", len(allKeys)-20)
		}
	}
}

func runConnectionTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Test basic connection
	_, err := client.Ping(ctx).Result()
	suite.AddResult("Ping", err == nil, "Basic connectivity test", err)

	// Test INFO command
	info, err := client.Info(ctx).Result()
	suite.AddResult("Info", err == nil && len(info) > 0, "Server information retrieval", err)
}

func runStringTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Test SET and GET
	err := client.Set(ctx, "test:string:key1", "value1", 0).Err()
	suite.AddResult("Set", err == nil, "Set a string value", err)

	val, err := client.Get(ctx, "test:string:key1").Result()
	suite.AddResult("Get", err == nil && val == "value1", "Get a string value", err)

	// Test SET with expiration
	err = client.Set(ctx, "test:string:expiring", "expiring-value", 2*time.Second).Err()
	suite.AddResult("SetWithExpiration", err == nil, "Set a string value with expiration", err)

	// Test TTL
	ttl, err := client.TTL(ctx, "test:string:expiring").Result()
	suite.AddResult("TTL", err == nil && ttl > 0, fmt.Sprintf("Check TTL of expiring key (TTL: %v)", ttl), err)

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Test expired key
	_, err = client.Get(ctx, "test:string:expiring").Result()
	suite.AddResult("ExpiredKey", err == redis.Nil, "Verify expired key is removed", err)

	// Test DEL
	client.Set(ctx, "test:string:to-delete", "delete-me", 0)
	affected, err := client.Del(ctx, "test:string:to-delete").Result()
	suite.AddResult("Del", err == nil && affected == 1, "Delete a key", err)

	// Test EXISTS
	client.Set(ctx, "test:string:exists", "i-exist", 0)
	exists, err := client.Exists(ctx, "test:string:exists").Result()
	suite.AddResult("Exists", err == nil && exists == 1, "Check if a key exists", err)

	// Test INCR
	client.Del(ctx, "test:counter")
	incr, err := client.Incr(ctx, "test:counter").Result()
	suite.AddResult("Incr", err == nil && incr == 1, "Increment a counter", err)

	incr, err = client.Incr(ctx, "test:counter").Result()
	suite.AddResult("IncrAgain", err == nil && incr == 2, "Increment a counter again", err)

	// Test TYPE
	client.Set(ctx, "test:string:type", "string-value", 0)
	keyType, err := client.Type(ctx, "test:string:type").Result()
	suite.AddResult("Type", err == nil && keyType == "string", "Get type of a key", err)
}

func runHashTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Test HSET and HGET
	err := client.HSet(ctx, "test:hash:user1", "name", "John").Err()
	suite.AddResult("HSet", err == nil, "Set a hash field", err)

	val, err := client.HGet(ctx, "test:hash:user1", "name").Result()
	suite.AddResult("HGet", err == nil && val == "John", "Get a hash field", err)

	// Test HSET multiple fields
	err = client.HSet(ctx, "test:hash:user1", "age", "30", "city", "New York").Err()
	suite.AddResult("HSetMultiple", err == nil, "Set multiple hash fields", err)

	// Test HGETALL
	fields, err := client.HGetAll(ctx, "test:hash:user1").Result()
	suite.AddResult("HGetAll", err == nil && len(fields) == 3,
		fmt.Sprintf("Get all hash fields (found %d fields)", len(fields)), err)

	// Test HKEYS
	keys, err := client.HKeys(ctx, "test:hash:user1").Result()
	suite.AddResult("HKeys", err == nil && len(keys) == 3,
		fmt.Sprintf("Get all hash keys (found %d keys)", len(keys)), err)

	// Test HLEN
	length, err := client.HLen(ctx, "test:hash:user1").Result()
	suite.AddResult("HLen", err == nil && length == 3,
		fmt.Sprintf("Get hash length (length: %d)", length), err)

	// Test HDEL
	deleted, err := client.HDel(ctx, "test:hash:user1", "city").Result()
	suite.AddResult("HDel", err == nil && deleted == 1, "Delete a hash field", err)

	// Test TYPE for hash
	keyType, err := client.Type(ctx, "test:hash:user1").Result()
	suite.AddResult("HashType", err == nil && keyType == "hash", "Get type of a hash key", err)
}

func runListTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Clear any existing list
	client.Del(ctx, "test:list:items")

	// Test LPUSH
	count, err := client.LPush(ctx, "test:list:items", "item1", "item2").Result()
	suite.AddResult("LPush", err == nil && count == 2,
		fmt.Sprintf("Push items to the head of a list (count: %d)", count), err)

	// Test RPUSH
	count, err = client.RPush(ctx, "test:list:items", "item3", "item4").Result()
	suite.AddResult("RPush", err == nil && count == 4,
		fmt.Sprintf("Push items to the tail of a list (count: %d)", count), err)

	// Test LLEN
	length, err := client.LLen(ctx, "test:list:items").Result()
	suite.AddResult("LLen", err == nil && length == 4,
		fmt.Sprintf("Get list length (length: %d)", length), err)

	// Test LRANGE
	items, err := client.LRange(ctx, "test:list:items", 0, -1).Result()
	suite.AddResult("LRange", err == nil && len(items) == 4,
		fmt.Sprintf("Get range of list items (found %d items)", len(items)), err)

	// Test LPOP
	item, err := client.LPop(ctx, "test:list:items").Result()
	suite.AddResult("LPop", err == nil && item == "item2",
		fmt.Sprintf("Pop item from the head of a list (item: %s)", item), err)

	// Test RPOP
	item, err = client.RPop(ctx, "test:list:items").Result()
	suite.AddResult("RPop", err == nil && item == "item4",
		fmt.Sprintf("Pop item from the tail of a list (item: %s)", item), err)

	// Test TYPE for list
	keyType, err := client.Type(ctx, "test:list:items").Result()
	suite.AddResult("ListType", err == nil && keyType == "list", "Get type of a list key", err)
}

func runPatternTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Create test keys
	client.FlushDB(ctx)
	client.Set(ctx, "user:1:name", "Alice", 0)
	client.Set(ctx, "user:1:email", "alice@example.com", 0)
	client.Set(ctx, "user:2:name", "Bob", 0)
	client.Set(ctx, "user:2:email", "bob@example.com", 0)
	client.Set(ctx, "product:1", "Laptop", 0)
	client.Set(ctx, "product:2", "Phone", 0)

	// Test KEYS with pattern
	keys, err := client.Keys(ctx, "user:*").Result()
	suite.AddResult("KeysPattern", err == nil && len(keys) == 4,
		fmt.Sprintf("Get keys matching pattern 'user:*' (found %d keys)", len(keys)), err)

	keys, err = client.Keys(ctx, "user:1:*").Result()
	suite.AddResult("KeysSpecificPattern", err == nil && len(keys) == 2,
		fmt.Sprintf("Get keys matching pattern 'user:1:*' (found %d keys)", len(keys)), err)

	// Test SCAN with pattern
	var cursor uint64
	var allKeys []string

	for {
		keys, nextCursor, err := client.Scan(ctx, cursor, "product:*", 10).Result()
		if err != nil {
			suite.AddResult("ScanPattern", false, "Scan keys matching pattern 'product:*'", err)
			break
		}

		allKeys = append(allKeys, keys...)
		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	suite.AddResult("ScanPattern", len(allKeys) == 2,
		fmt.Sprintf("Scan keys matching pattern 'product:*' (found %d keys)", len(allKeys)), nil)

	// Test wildcard patterns
	for i := 1; i <= 5; i++ {
		client.Set(ctx, fmt.Sprintf("test:wildcard:%d", i), strconv.Itoa(i), 0)
	}

	keys, err = client.Keys(ctx, "test:wildcard:?").Result()
	suite.AddResult("SingleCharWildcard", err == nil && len(keys) == 5,
		fmt.Sprintf("Get keys matching single character wildcard (found %d keys)", len(keys)), err)
}

func runMiscTests(ctx context.Context, client *redis.Client, suite *TestSuite) {
	// Test FLUSHDB
	err := client.FlushDB(ctx).Err()
	suite.AddResult("FlushDB", err == nil, "Flush the current database", err)

	// Test key count after flush
	keys, err := client.Keys(ctx, "*").Result()
	suite.AddResult("KeysAfterFlush", err == nil && len(keys) == 0,
		fmt.Sprintf("Verify no keys after flush (found %d keys)", len(keys)), err)

	// Test multiple operations in sequence
	client.Set(ctx, "test:multi:1", "value1", 0)
	client.Set(ctx, "test:multi:2", "value2", 0)

	val1, err1 := client.Get(ctx, "test:multi:1").Result()
	val2, err2 := client.Get(ctx, "test:multi:2").Result()

	suite.AddResult("MultipleOperations", err1 == nil && err2 == nil && val1 == "value1" && val2 == "value2",
		"Perform multiple operations in sequence", nil)

	// Test non-existent key
	_, err = client.Get(ctx, "test:nonexistent").Result()
	suite.AddResult("NonExistentKey", err == redis.Nil, "Get a non-existent key", nil)
}

// getNetworkType determines if the address is a TCP or Unix socket
func getNetworkType(addr string) string {
	if addr[0] == '/' {
		return "unix"
	}
	return "tcp"
}
