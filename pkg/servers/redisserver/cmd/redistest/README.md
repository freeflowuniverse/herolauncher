# Redis Server Test Tool

This tool provides comprehensive testing for the Redis server implementation in the `pkg/redisserver` package.

## Features

- Tests both TCP and Unix socket connections
- Organized test suites for different Redis functionality:
  - Connection tests
  - String operations
  - Hash operations
  - List operations
  - Pattern matching
  - Miscellaneous operations
- Detailed test results with pass/fail status and error information
- Summary statistics for overall test coverage

## Usage

```bash
go run main.go [options]
```

### Options

- `-tcp-port string`: Redis server TCP port (default "7777")
- `-unix-socket string`: Redis server Unix domain socket path (default "/tmp/redis-test.sock")
- `-debug`: Enable debug output (default false)
- `-db int`: Redis database number (default 0)

## Example

```bash
# Run with default settings
go run main.go

# Run with custom port and socket
go run main.go -tcp-port 6379 -unix-socket /tmp/custom-redis.sock

# Run with debug output
go run main.go -debug
```

## Test Coverage

The tool tests the following Redis functionality:

1. **Connection Tests**
   - PING
   - INFO

2. **String Operations**
   - SET/GET
   - SET with expiration
   - TTL
   - DEL
   - EXISTS
   - INCR
   - TYPE

3. **Hash Operations**
   - HSET/HGET
   - HKEYS
   - HLEN
   - HDEL
   - Type checking

4. **List Operations**
   - LPUSH/RPUSH
   - LLEN
   - LRANGE
   - LPOP/RPOP
   - Type checking

5. **Pattern Matching**
   - KEYS with various patterns
   - SCAN with patterns
   - Wildcard pattern matching

6. **Miscellaneous Operations**
   - FLUSHDB
   - Multiple sequential operations
   - Non-existent key handling
