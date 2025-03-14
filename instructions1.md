
create a golang project

there will be multiple 

- modules
    - one is for installers
    - one is for a fiber web server with a web ui, swagger UI and opeapi rest interface (v3.1.0 swagger)
    - a generic redis server

- on the fiber webserver create multiple endpoints nicely structures as separate directories underneith the module
    - executor (for executing commands, results in jobs)
    - package manager (on basis of apt, brew, scoop)
    - create an openapi interface for each of those v3.1.0
    - integrate in generic way the goswagger interface so people can use the rest interface from web

- create a main server which connects to all the modules


### code for the redis server

```go
package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/redcon"
)

// entry represents a stored value. For strings, value is stored as a string.
// For hashes, value is stored as a map[string]string.
type entry struct {
	value      interface{}
	expiration time.Time // zero means no expiration
}

// Server holds the in-memory datastore and provides thread-safe access.
type Server struct {
	mu   sync.RWMutex
	data map[string]*entry
}

// NewServer creates a new server instance and starts a cleanup goroutine.
func NewServer() *Server {
	s := &Server{
		data: make(map[string]*entry),
	}
	go s.cleanupExpiredKeys()
	return s
}

// cleanupExpiredKeys periodically removes expired keys.
func (s *Server) cleanupExpiredKeys() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for k, ent := range s.data {
			if !ent.expiration.IsZero() && now.After(ent.expiration) {
				delete(s.data, k)
			}
		}
		s.mu.Unlock()
	}
}

// set stores a key with a value and an optional expiration duration.
func (s *Server) set(key string, value interface{}, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var exp time.Time
	if duration > 0 {
		exp = time.Now().Add(duration)
	}
	s.data[key] = &entry{
		value:      value,
		expiration: exp,
	}
}

// get retrieves the value for a key if it exists and is not expired.
func (s *Server) get(key string) (interface{}, bool) {
	s.mu.RLock()
	ent, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
		// Key has expired; remove it.
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return nil, false
	}
	return ent.value, true
}

// del deletes a key and returns 1 if the key was present.
func (s *Server) del(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		return 1
	}
	return 0
}

// keys returns all keys matching the given pattern.
// For simplicity, only "*" is fully supported.
func (s *Server) keys(pattern string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []string
	// Simple pattern matching: if pattern is "*", return all nonexpired keys.
	if pattern == "*" {
		for k, ent := range s.data {
			if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
				continue
			}
			result = append(result, k)
		}
	} else {
		// For any other pattern, do a simple substring match.
		for k, ent := range s.data {
			if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
				continue
			}
			if strings.Contains(k, pattern) {
				result = append(result, k)
			}
		}
	}
	return result
}

// getHash retrieves the hash map stored at key.
func (s *Server) getHash(key string) (map[string]string, bool) {
	v, ok := s.get(key)
	if !ok {
		return nil, false
	}
	hash, ok := v.(map[string]string)
	return hash, ok
}

// hset sets a field in the hash stored at key. It returns 1 if the field is new.
func (s *Server) hset(key, field, value string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	var hash map[string]string
	ent, exists := s.data[key]
	if exists {
		if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
			// expired; recreate a new hash.
			hash = make(map[string]string)
			s.data[key] = &entry{value: hash}
		} else {
			var ok bool
			hash, ok = ent.value.(map[string]string)
			if !ok {
				// Overwrite if the key holds a non-hash value.
				hash = make(map[string]string)
				s.data[key] = &entry{value: hash}
			}
		}
	} else {
		hash = make(map[string]string)
		s.data[key] = &entry{value: hash}
	}
	_, fieldExists := hash[field]
	hash[field] = value
	if fieldExists {
		return 0
	}
	return 1
}

// hget retrieves the value of a field in the hash stored at key.
func (s *Server) hget(key, field string) (string, bool) {
	hash, ok := s.getHash(key)
	if !ok {
		return "", false
	}
	val, exists := hash[field]
	return val, exists
}

// hdel deletes one or more fields from the hash stored at key.
// Returns the number of fields that were removed.
func (s *Server) hdel(key string, fields []string) int {
	hash, ok := s.getHash(key)
	if !ok {
		return 0
	}
	count := 0
	for _, field := range fields {
		if _, exists := hash[field]; exists {
			delete(hash, field)
			count++
		}
	}
	return count
}

// hkeys returns all field names in the hash stored at key.
func (s *Server) hkeys(key string) []string {
	hash, ok := s.getHash(key)
	if !ok {
		return nil
	}
	var keys []string
	for field := range hash {
		keys = append(keys, field)
	}
	return keys
}

// hlen returns the number of fields in the hash stored at key.
func (s *Server) hlen(key string) int {
	hash, ok := s.getHash(key)
	if !ok {
		return 0
	}
	return len(hash)
}

// incr increments the integer value stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
func (s *Server) incr(key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var current int64
	ent, exists := s.data[key]
	if exists {
		if !ent.expiration.IsZero() && time.Now().After(ent.expiration) {
			current = 0
		} else {
			switch v := ent.value.(type) {
			case string:
				var err error
				current, err = strconv.ParseInt(v, 10, 64)
				if err != nil {
					return 0, err
				}
			case int:
				current = int64(v)
			case int64:
				current = v
			default:
				return 0, fmt.Errorf("value is not an integer")
			}
		}
	}
	current++
	// Store the new value as a string.
	s.data[key] = &entry{
		value: strconv.FormatInt(current, 10),
	}
	return current, nil
}

func main() {
	server := NewServer()
	log.Println("Starting Redis-like server on :6379")
	err := redcon.ListenAndServe(":6379",
		func(conn redcon.Conn, cmd redcon.Command) {
			// Every command is expected to have at least one argument (the command name).
			if len(cmd.Args) == 0 {
				conn.WriteError("ERR empty command")
				return
			}
			command := strings.ToLower(string(cmd.Args[0]))
			switch command {
			case "ping":
				conn.WriteString("PONG")
			case "set":
				// Usage: SET key value [EX seconds]
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'set' command")
					return
				}
				key := string(cmd.Args[1])
				value := string(cmd.Args[2])
				duration := time.Duration(0)
				// Check for an expiration option (only EX is supported here).
				if len(cmd.Args) > 3 {
					if strings.ToLower(string(cmd.Args[3])) == "ex" && len(cmd.Args) > 4 {
						seconds, err := strconv.Atoi(string(cmd.Args[4]))
						if err != nil {
							conn.WriteError("ERR invalid expire time")
							return
						}
						duration = time.Duration(seconds) * time.Second
					}
				}
				server.set(key, value, duration)
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'get' command")
					return
				}
				key := string(cmd.Args[1])
				v, ok := server.get(key)
				if !ok {
					conn.WriteNull()
					return
				}
				// Only string type is returned by GET.
				switch val := v.(type) {
				case string:
					conn.WriteBulkString(val)
				default:
					conn.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
				}
			case "del":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'del' command")
					return
				}
				count := 0
				for i := 1; i < len(cmd.Args); i++ {
					key := string(cmd.Args[i])
					count += server.del(key)
				}
				conn.WriteInt(count)
			case "keys":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'keys' command")
					return
				}
				pattern := string(cmd.Args[1])
				keys := server.keys(pattern)
				res := make([][]byte, len(keys))
				for i, k := range keys {
					res[i] = []byte(k)
				}
				conn.WriteArray(res)
			case "hset":
				// Usage: HSET key field value
				if len(cmd.Args) < 4 {
					conn.WriteError("ERR wrong number of arguments for 'hset' command")
					return
				}
				key := string(cmd.Args[1])
				field := string(cmd.Args[2])
				value := string(cmd.Args[3])
				added := server.hset(key, field, value)
				conn.WriteInt(added)
			case "hget":
				// Usage: HGET key field
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'hget' command")
					return
				}
				key := string(cmd.Args[1])
				field := string(cmd.Args[2])
				v, ok := server.hget(key, field)
				if !ok {
					conn.WriteNull()
					return
				}
				conn.WriteBulkString(v)
			case "hdel":
				// Usage: HDEL key field [field ...]
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'hdel' command")
					return
				}
				key := string(cmd.Args[1])
				fields := make([]string, 0, len(cmd.Args)-2)
				for i := 2; i < len(cmd.Args); i++ {
					fields = append(fields, string(cmd.Args[i]))
				}
				removed := server.hdel(key, fields)
				conn.WriteInt(removed)
			case "hkeys":
				// Usage: HKEYS key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'hkeys' command")
					return
				}
				key := string(cmd.Args[1])
				fields := server.hkeys(key)
				res := make([][]byte, len(fields))
				for i, field := range fields {
					res[i] = []byte(field)
				}
				conn.WriteArray(res)
			case "hlen":
				// Usage: HLEN key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'hlen' command")
					return
				}
				key := string(cmd.Args[1])
				length := server.hlen(key)
				conn.WriteInt(length)
			case "incr":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'incr' command")
					return
				}
				key := string(cmd.Args[1])
				newVal, err := server.incr(key)
				if err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}
				conn.WriteInt64(newVal)
			default:
				conn.WriteError("ERR unknown command '" + command + "'")
			}
		},
		// Accept connection: always allow.
		func(conn redcon.Conn) bool { return true },
		// On connection close.
		func(conn redcon.Conn, err error) {},
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

test above code, test with a redis client it works