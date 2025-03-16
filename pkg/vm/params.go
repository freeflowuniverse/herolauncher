package vm

import (
	"strconv"
	"strings"
)

// Params is a simple helper for working with parameter maps
type Params map[string]string

// Get returns the value for a key
func (p Params) Get(key string) string {
	if val, ok := p[key]; ok {
		return val
	}
	return ""
}

// GetInt returns the value for a key as an integer
func (p Params) GetInt(key string) int {
	val := p.Get(key)
	if val == "" {
		return 0
	}
	
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	
	return i
}

// GetBool returns the value for a key as a boolean
func (p Params) GetBool(key string) bool {
	val := strings.ToLower(p.Get(key))
	return val == "true" || val == "1" || val == "yes" || val == "y"
}
