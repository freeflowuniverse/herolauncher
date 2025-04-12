package redisserver

import (
	"fmt"
	"strconv"
)

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
