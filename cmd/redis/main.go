package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/redis"
	"github.com/tidwall/redcon"
)

func main() {
	server := redis.NewServer()
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
				server.Set(key, value, duration)
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'get' command")
					return
				}
				key := string(cmd.Args[1])
				v, ok := server.Get(key)
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
					count += server.Del(key)
				}
				conn.WriteInt(count)
			case "keys":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'keys' command")
					return
				}
				pattern := string(cmd.Args[1])
				keys := server.Keys(pattern)
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
				added := server.HSet(key, field, value)
				conn.WriteInt(added)
			case "hget":
				// Usage: HGET key field
				if len(cmd.Args) < 3 {
					conn.WriteError("ERR wrong number of arguments for 'hget' command")
					return
				}
				key := string(cmd.Args[1])
				field := string(cmd.Args[2])
				v, ok := server.HGet(key, field)
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
				removed := server.HDel(key, fields)
				conn.WriteInt(removed)
			case "hkeys":
				// Usage: HKEYS key
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'hkeys' command")
					return
				}
				key := string(cmd.Args[1])
				fields := server.HKeys(key)
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
				length := server.HLen(key)
				conn.WriteInt(length)
			case "incr":
				if len(cmd.Args) < 2 {
					conn.WriteError("ERR wrong number of arguments for 'incr' command")
					return
				}
				key := string(cmd.Args[1])
				newVal, err := server.Incr(key)
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
