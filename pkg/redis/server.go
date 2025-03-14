package redis

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
// It implements a Redis-compatible server using redcon.
type Server struct {
	mu   sync.RWMutex
	data map[string]*entry
}

// NewServer creates a new server instance and starts a cleanup goroutine.
// It also starts a Redis-compatible server on port 6379 in a separate goroutine.
func NewServer() *Server {
	s := &Server{
		data: make(map[string]*entry),
	}
	go s.cleanupExpiredKeys()
	go s.startRedisServer()
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

// Set stores a key with a value and an optional expiration duration.
func (s *Server) Set(key string, value interface{}, duration time.Duration) {
	s.set(key, value, duration)
}

// set is the internal implementation of Set
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

// Get retrieves the value for a key if it exists and is not expired.
func (s *Server) Get(key string) (interface{}, bool) {
	return s.get(key)
}

// get is the internal implementation of Get
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

// Del deletes a key and returns 1 if the key was present.
func (s *Server) Del(key string) int {
	return s.del(key)
}

// del is the internal implementation of Del
func (s *Server) del(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		return 1
	}
	return 0
}

// Keys returns all keys matching the given pattern.
// For simplicity, only "*" is fully supported.
func (s *Server) Keys(pattern string) []string {
	return s.keys(pattern)
}

// keys is the internal implementation of Keys
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

// GetHash retrieves the hash map stored at key.
func (s *Server) GetHash(key string) (map[string]string, bool) {
	return s.getHash(key)
}

// getHash is the internal implementation of GetHash
func (s *Server) getHash(key string) (map[string]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ent, exists := s.data[key]
	if !exists || (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		return nil, false
	}

	hash, ok := ent.value.(map[string]string)
	return hash, ok
}

// HSet sets a field in the hash stored at key. It returns 1 if the field is new.
func (s *Server) HSet(key, field, value string) int {
	return s.hset(key, field, value)
}

// hset is the internal implementation of HSet
func (s *Server) hset(key, field, value string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if exists && (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		// Key exists but has expired, delete it
		delete(s.data, key)
		exists = false
	}

	// Handle hash creation or update
	var hash map[string]string
	if exists {
		// Try to cast to map[string]string
		switch v := ent.value.(type) {
		case map[string]string:
			// Key exists and is a hash
			hash = v
		default:
			// Key exists but is not a hash, overwrite it
			hash = make(map[string]string)
			s.data[key] = &entry{value: hash, expiration: ent.expiration}
		}
	} else {
		// Key doesn't exist, create a new hash
		hash = make(map[string]string)
		s.data[key] = &entry{value: hash}
	}

	// Set the field in the hash
	_, fieldExists := hash[field]
	hash[field] = value

	// Return 1 if field was added, 0 if it was updated
	if fieldExists {
		return 0
	}
	return 1
}

// HGet retrieves the value of a field in the hash stored at key.
func (s *Server) HGet(key, field string) (string, bool) {
	return s.hget(key, field)
}

// hget is the internal implementation of HGet
func (s *Server) hget(key, field string) (string, bool) {
	hash, ok := s.getHash(key)
	if !ok {
		return "", false
	}
	val, exists := hash[field]
	return val, exists
}

// HDel deletes one or more fields from the hash stored at key.
// Returns the number of fields that were removed.
func (s *Server) HDel(key string, fields []string) int {
	return s.hdel(key, fields)
}

// hdel is the internal implementation of HDel
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

// HKeys returns all field names in the hash stored at key.
func (s *Server) HKeys(key string) []string {
	return s.hkeys(key)
}

// hkeys is the internal implementation of HKeys
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

// HLen returns the number of fields in the hash stored at key.
func (s *Server) HLen(key string) int {
	return s.hlen(key)
}

// hlen is the internal implementation of HLen
func (s *Server) hlen(key string) int {
	hash, ok := s.getHash(key)
	if !ok {
		return 0
	}
	return len(hash)
}

// Incr increments the integer value stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation.
func (s *Server) Incr(key string) (int64, error) {
	return s.incr(key)
}

// incr is the internal implementation of Incr
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

// startRedisServer starts a Redis-compatible server on port 6378.
// expire sets an expiration time for a key
func (s *Server) expire(key string, duration time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return false
	}

	// Set expiration time
	item.expiration = time.Now().Add(duration)
	return true
}

// getTTL returns the time to live for a key in seconds
func (s *Server) getTTL(key string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists {
		// Key doesn't exist
		return -2
	}

	// If the key has no expiration
	if item.expiration.IsZero() {
		return -1
	}

	// If the key has expired
	if time.Now().After(item.expiration) {
		return -2
	}

	// Calculate remaining time in seconds
	ttl := int64(item.expiration.Sub(time.Now()).Seconds())
	return ttl
}

// scan returns a list of keys matching the pattern starting from cursor
func (s *Server) scan(cursor int, pattern string, count int) (int, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Get all keys
	allKeys := make([]string, 0, len(s.data))
	for k, item := range s.data {
		// Skip expired keys
		if !item.expiration.IsZero() && time.Now().After(item.expiration) {
			continue
		}
		
		// Check if key matches pattern
		if matched, _ := filepath.Match(pattern, k); matched {
			allKeys = append(allKeys, k)
		}
	}
	
	// Sort keys for consistent results
	sort.Strings(allKeys)
	
	// If cursor is beyond the end or there are no keys, return empty list
	if cursor >= len(allKeys) || len(allKeys) == 0 {
		return 0, []string{}
	}
	
	// Calculate end index
	end := cursor + count
	if end > len(allKeys) {
		end = len(allKeys)
	}
	
	// Get keys for this iteration
	keys := allKeys[cursor:end]
	
	// Calculate next cursor
	nextCursor := 0
	if end < len(allKeys) {
		nextCursor = end
	}
	
	return nextCursor, keys
}

// hscan iterates over fields in a hash that match a pattern
func (s *Server) hscan(key string, cursor int, pattern string, count int) (int, []string, []string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the hash
	hash, ok := s.getHash(key)
	if !ok {
		return 0, []string{}, []string{}
	}

	// Get all fields
	allFields := make([]string, 0, len(hash))
	for field := range hash {
		// Check if field matches pattern
		if matched, _ := filepath.Match(pattern, field); matched {
			allFields = append(allFields, field)
		}
	}

	// Sort fields for consistent results
	sort.Strings(allFields)

	// If cursor is beyond the end or there are no fields, return empty lists
	if cursor >= len(allFields) || len(allFields) == 0 {
		return 0, []string{}, []string{}
	}

	// Calculate end index
	end := cursor + count
	if end > len(allFields) {
		end = len(allFields)
	}

	// Get fields for this iteration
	fields := allFields[cursor:end]

	// Get corresponding values
	values := make([]string, len(fields))
	for i, field := range fields {
		values[i] = hash[field]
	}

	// Calculate next cursor
	nextCursor := 0
	if end < len(allFields) {
		nextCursor = end
	}

	return nextCursor, fields, values
}

// lpush adds one or more values to the head of a list
func (s *Server) lpush(key string, values []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if exists && (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		// Key exists but has expired, delete it
		delete(s.data, key)
		exists = false
	}

	var list []string
	if exists {
		// Try to cast to []string
		if l, ok := ent.value.([]string); ok {
			// Key exists and is a list
			list = l
		} else {
			// Key exists but is not a list, overwrite it
			list = []string{}
			s.data[key] = &entry{value: list, expiration: ent.expiration}
		}
	} else {
		// Key doesn't exist, create a new list
		list = []string{}
		s.data[key] = &entry{value: list}
	}

	// Add values to the head of the list
	newList := make([]string, len(values)+len(list))
	copy(newList, values)
	copy(newList[len(values):], list)

	// Update the list in the data store
	s.data[key].value = newList

	return len(newList)
}

// rpush adds one or more values to the tail of a list
func (s *Server) rpush(key string, values []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if exists && (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		// Key exists but has expired, delete it
		delete(s.data, key)
		exists = false
	}

	var list []string
	if exists {
		// Try to cast to []string
		if l, ok := ent.value.([]string); ok {
			// Key exists and is a list
			list = l
		} else {
			// Key exists but is not a list, overwrite it
			list = []string{}
			s.data[key] = &entry{value: list, expiration: ent.expiration}
		}
	} else {
		// Key doesn't exist, create a new list
		list = []string{}
		s.data[key] = &entry{value: list}
	}

	// Add values to the tail of the list
	newList := append(list, values...)

	// Update the list in the data store
	s.data[key].value = newList

	return len(newList)
}

// lpop removes and returns the first element of a list
func (s *Server) lpop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if !exists || (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		// Key doesn't exist or has expired
		if exists {
			delete(s.data, key)
		}
		return "", false
	}

	// Try to cast to []string
	list, ok := ent.value.([]string)
	if !ok || len(list) == 0 {
		return "", false
	}

	// Get the first element
	val := list[0]

	// Remove the first element from the list
	if len(list) == 1 {
		delete(s.data, key)
	} else {
		s.data[key].value = list[1:]
	}

	return val, true
}

// rpop removes and returns the last element of a list
func (s *Server) rpop(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if !exists || (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		// Key doesn't exist or has expired
		if exists {
			delete(s.data, key)
		}
		return "", false
	}

	// Try to cast to []string
	list, ok := ent.value.([]string)
	if !ok || len(list) == 0 {
		return "", false
	}

	// Get the last element
	val := list[len(list)-1]

	// Remove the last element from the list
	if len(list) == 1 {
		delete(s.data, key)
	} else {
		s.data[key].value = list[:len(list)-1]
	}

	return val, true
}

// llen returns the length of a list
func (s *Server) llen(key string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if !exists || (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		return 0
	}

	// Try to cast to []string
	list, ok := ent.value.([]string)
	if !ok {
		return 0
	}

	return len(list)
}

// lrange returns a range of elements from a list
func (s *Server) lrange(key string, start, stop int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if key exists and is not expired
	ent, exists := s.data[key]
	if !exists || (!ent.expiration.IsZero() && time.Now().After(ent.expiration)) {
		return []string{}
	}

	// Try to cast to []string
	list, ok := ent.value.([]string)
	if !ok {
		return []string{}
	}

	// Handle negative indices
	listLen := len(list)
	if start < 0 {
		start = listLen + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = listLen + stop
	}

	// Ensure start and stop are within bounds
	if start >= listLen || start > stop {
		return []string{}
	}
	if stop >= listLen {
		stop = listLen - 1
	}

	// Return the range of elements
	return list[start : stop+1]
}

// getType returns the type of the value stored at key
func (s *Server) getType(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.data[key]
	if !exists || (!item.expiration.IsZero() && time.Now().After(item.expiration)) {
		// Key doesn't exist or has expired
		return "none"
	}

	switch v := item.value.(type) {
	case string:
		return "string"
	case map[string]string:
		return "hash"
	case map[string]interface{}:
		return "hash"
	case []string:
		return "list"
	default:
		// For debugging
		log.Printf("Unknown type for key %s: %T", key, v)
		return "none"
	}
}

// getInfo returns information about the server for the INFO command
func (s *Server) getInfo() string {
	s.mu.RLock()
	keyCount := len(s.data)
	s.mu.RUnlock()

	// Build the info string in Redis format
	info := "# Server\r\n"
	info += "redis_version:1.0.0\r\n"
	info += "redis_mode:standalone\r\n"
	info += "os:" + runtime.GOOS + "\r\n"
	info += "arch_bits:" + strconv.Itoa(32<<(^uint(0)>>63)) + "\r\n"
	info += "process_id:" + strconv.Itoa(os.Getpid()) + "\r\n"
	
	info += "\r\n# Clients\r\n"
	info += "connected_clients:1\r\n"
	
	info += "\r\n# Memory\r\n"
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info += "used_memory:" + strconv.FormatUint(m.Alloc, 10) + "\r\n"
	info += "used_memory_human:" + humanizeBytes(m.Alloc) + "\r\n"
	
	info += "\r\n# Stats\r\n"
	info += "keyspace_hits:0\r\n"
	info += "keyspace_misses:0\r\n"
	
	info += "\r\n# Keyspace\r\n"
	info += "db0:keys=" + strconv.Itoa(keyCount) + ",expires=0,avg_ttl=0\r\n"
	
	return info
}

// humanizeBytes converts bytes to a human-readable string
func humanizeBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	
	var value float64
	var unit string
	
	switch {
	case bytes >= GB:
		value = float64(bytes) / GB
		unit = "GB"
	case bytes >= MB:
		value = float64(bytes) / MB
		unit = "MB"
	case bytes >= KB:
		value = float64(bytes) / KB
		unit = "KB"
	default:
		return strconv.FormatUint(bytes, 10) + "B"
	}
	
	return fmt.Sprintf("%.2f%s", value, unit)
}

func (s *Server) startRedisServer() {
	log.Println("Starting Redis-like server on :6378")
	err := redcon.ListenAndServe(":6378",
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
				s.set(key, value, duration)
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'get' command")
					return
				}
				key := string(cmd.Args[1])
				v, ok := s.get(key)
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
					count += s.del(key)
				}
				conn.WriteInt(count)
			case "keys":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'keys' command")
					return
				}
				pattern := string(cmd.Args[1])
				keys := s.keys(pattern)
				conn.WriteArray(len(keys))
				for _, k := range keys {
					conn.WriteBulkString(k)
				}
			case "hset":
				// Usage: HSET key field value
				if len(cmd.Args) < 4 {
					conn.WriteError("ERR wrong number of arguments for 'hset' command")
					return
				}
				key := string(cmd.Args[1])
				field := string(cmd.Args[2])
				value := string(cmd.Args[3])
				added := s.hset(key, field, value)
				conn.WriteInt(added)
			case "hget":
				// Usage: HGET key field
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'hget' command")
					return
				}
				key := string(cmd.Args[1])
				field := string(cmd.Args[2])
				v, ok := s.hget(key, field)
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
				removed := s.hdel(key, fields)
				conn.WriteInt(removed)
			case "hkeys":
				// Usage: HKEYS key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'hkeys' command")
					return
				}
				key := string(cmd.Args[1])
				fields := s.hkeys(key)
				conn.WriteArray(len(fields))
				for _, field := range fields {
					conn.WriteBulkString(field)
				}
			case "hlen":
				// Usage: HLEN key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'hlen' command")
					return
				}
				key := string(cmd.Args[1])
				length := s.hlen(key)
				conn.WriteInt(length)
			case "incr":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'incr' command")
					return
				}
				key := string(cmd.Args[1])
				newVal, err := s.incr(key)
				if err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}
				conn.WriteInt64(newVal)
			case "info":
				// Return basic information about the server
				info := s.getInfo()
				conn.WriteBulkString(info)
			case "type":
				// Usage: TYPE key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'type' command")
					return
				}
				key := string(cmd.Args[1])
				keyType := s.getType(key)
				conn.WriteBulkString(keyType)
			case "ttl":
				// Usage: TTL key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'ttl' command")
					return
				}
				key := string(cmd.Args[1])
				ttl := s.getTTL(key)
				conn.WriteInt64(ttl)
			case "expire":
				// Usage: EXPIRE key seconds
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'expire' command")
					return
				}
				key := string(cmd.Args[1])
				seconds, err := strconv.ParseInt(string(cmd.Args[2]), 10, 64)
				if err != nil {
					conn.WriteError("ERR value is not an integer or out of range")
					return
				}
				success := s.expire(key, time.Duration(seconds)*time.Second)
				if success {
					conn.WriteInt(1)
				} else {
					conn.WriteInt(0)
				}
			case "scan":
				// Usage: SCAN cursor [MATCH pattern] [COUNT count]
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'scan' command")
					return
				}
				
				cursor := string(cmd.Args[1])
				cursorInt, err := strconv.Atoi(cursor)
				if err != nil {
					conn.WriteError("ERR invalid cursor")
					return
				}
				
				// Default values
				pattern := "*"
				count := 10
				
				// Parse optional arguments
				for i := 2; i < len(cmd.Args); i++ {
					arg := strings.ToLower(string(cmd.Args[i]))
					if arg == "match" && i+1 < len(cmd.Args) {
						pattern = string(cmd.Args[i+1])
						i++
					} else if arg == "count" && i+1 < len(cmd.Args) {
						count, err = strconv.Atoi(string(cmd.Args[i+1]))
						if err != nil {
							conn.WriteError("ERR value is not an integer or out of range")
							return
						}
						i++
					}
				}
				
				// Get matching keys
				nextCursor, keys := s.scan(cursorInt, pattern, count)
				
				// Write response
				conn.WriteArray(2)
				conn.WriteBulkString(strconv.Itoa(nextCursor))
				conn.WriteArray(len(keys))
				for _, key := range keys {
					conn.WriteBulkString(key)
				}
			case "hscan":
				// Usage: HSCAN key cursor [MATCH pattern] [COUNT count]
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'hscan' command")
					return
				}
				
				key := string(cmd.Args[1])
				cursor := string(cmd.Args[2])
				cursorInt, err := strconv.Atoi(cursor)
				if err != nil {
					conn.WriteError("ERR invalid cursor")
					return
				}
				
				// Default values
				pattern := "*"
				count := 10
				
				// Parse optional arguments
				for i := 3; i < len(cmd.Args); i++ {
					arg := strings.ToLower(string(cmd.Args[i]))
					if arg == "match" && i+1 < len(cmd.Args) {
						pattern = string(cmd.Args[i+1])
						i++
					} else if arg == "count" && i+1 < len(cmd.Args) {
						count, err = strconv.Atoi(string(cmd.Args[i+1]))
						if err != nil {
							conn.WriteError("ERR value is not an integer or out of range")
							return
						}
						i++
					}
				}
				
				// Get matching fields and values
				nextCursor, fields, values := s.hscan(key, cursorInt, pattern, count)
				
				// Write response
				conn.WriteArray(2)
				conn.WriteBulkString(strconv.Itoa(nextCursor))
				
				// Write field-value pairs
				conn.WriteArray(len(fields) * 2) // Each field has a corresponding value
				for i := 0; i < len(fields); i++ {
					conn.WriteBulkString(fields[i])
					conn.WriteBulkString(values[i])
				}
				case "lpush":
				// Usage: LPUSH key value [value ...]
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'lpush' command")
					return
				}
				key := string(cmd.Args[1])
				values := make([]string, len(cmd.Args)-2)
				for i := 2; i < len(cmd.Args); i++ {
					values[i-2] = string(cmd.Args[i])
				}
				length := s.lpush(key, values)
				conn.WriteInt(length)
			
			case "rpush":
				// Usage: RPUSH key value [value ...]
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'rpush' command")
					return
				}
				key := string(cmd.Args[1])
				values := make([]string, len(cmd.Args)-2)
				for i := 2; i < len(cmd.Args); i++ {
					values[i-2] = string(cmd.Args[i])
				}
				length := s.rpush(key, values)
				conn.WriteInt(length)
			
			case "lpop":
				// Usage: LPOP key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'lpop' command")
					return
				}
				key := string(cmd.Args[1])
				val, ok := s.lpop(key)
				if !ok {
					conn.WriteNull()
					return
				}
				conn.WriteBulkString(val)
			
			case "rpop":
				// Usage: RPOP key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'rpop' command")
					return
				}
				key := string(cmd.Args[1])
				val, ok := s.rpop(key)
				if !ok {
					conn.WriteNull()
					return
				}
				conn.WriteBulkString(val)
			
			case "llen":
				// Usage: LLEN key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'llen' command")
					return
				}
				key := string(cmd.Args[1])
				length := s.llen(key)
				conn.WriteInt(length)
			
			case "lrange":
				// Usage: LRANGE key start stop
				if len(cmd.Args) < 4 {
					conn.WriteError("ERR wrong number of arguments for 'lrange' command")
					return
				}
				key := string(cmd.Args[1])
				start, err := strconv.Atoi(string(cmd.Args[2]))
				if err != nil {
					conn.WriteError("ERR value is not an integer or out of range")
					return
				}
				stop, err := strconv.Atoi(string(cmd.Args[3]))
				if err != nil {
					conn.WriteError("ERR value is not an integer or out of range")
					return
				}
				values := s.lrange(key, start, stop)
				conn.WriteArray(len(values))
				for _, val := range values {
					conn.WriteBulkString(val)
				}
			
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
		log.Printf("Error starting Redis server: %v", err)
	}
}
