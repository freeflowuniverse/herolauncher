# Redis Server Package

A lightweight, in-memory Redis-compatible server implementation in Go. This package provides a Redis-like server that can be embedded in your Go applications.

## Features

- Supports both TCP and Unix socket connections
- In-memory data storage with key expiration
- Implements common Redis commands
- Thread-safe operations
- Automatic cleanup of expired keys

## Supported Commands

The server implements the following Redis commands:

- Basic: `PING`, `SET`, `GET`, `DEL`, `KEYS`, `EXISTS`, `TYPE`, `TTL`, `INFO`, `INCR`
- Hash operations: `HSET`, `HGET`, `HDEL`, `HKEYS`, `HLEN`
- List operations: `LPUSH`, `RPUSH`, `LPOP`, `RPOP`, `LLEN`, `LRANGE`
- Cursor-based iteration: `SCAN`, `HSCAN`

## Usage

### Basic Usage

```go
import "github.com/freeflowuniverse/herolauncher/pkg/redisserver"

// Create a new server with default configuration
server := redisserver.NewServer(redisserver.ServerConfig{
    TCPPort: "6379",                  // TCP port to listen on
    UnixSocketPath: "/tmp/redis.sock" // Unix socket path (optional)
})

// The server starts automatically and runs in background goroutines
```

### Connecting to the Server

You can connect to the server using any Redis client. For example, using the `go-redis` package:

```go
import "github.com/redis/go-redis/v9"

// Connect via TCP
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Or connect via Unix socket
unixClient := redis.NewClient(&redis.Options{
    Network: "unix",
    Addr:    "/tmp/redis.sock",
})
```
