package redisserver

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/redcon"
)

func (s *Server) startRedisServer(addr string, networkType string) {
	networkDesc := "TCP"
	netType := "tcp"
	
	if networkType == "unix" {
		networkDesc = "Unix socket"
		netType = "unix"
	}
	
	log.Printf("Starting Redis-like server on %s (%s)", addr, networkDesc)
	
	// Use ListenAndServeNetwork to support both TCP and Unix sockets
	err := redcon.ListenAndServeNetwork(netType, addr,
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
			case "exists":
				// Usage: EXISTS key [key ...]
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'exists' command")
					return
				}
				keys := make([]string, 0, len(cmd.Args)-1)
				for i := 1; i < len(cmd.Args); i++ {
					keys = append(keys, string(cmd.Args[i]))
				}
				count := s.exists(keys)
				conn.WriteInt(count)
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
