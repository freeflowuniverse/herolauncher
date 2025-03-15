package processmanager

import (
	"strings"
	"sync"
)

// RingBuffer implements a fixed-size buffer that overwrites oldest data when full
type RingBuffer struct {
	buffer    []byte
	size      int
	capacity  int
	mutex     sync.Mutex
	lineCount int
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buffer:   make([]byte, 0, capacity),
		capacity: capacity,
	}
}

// Write writes data to the ring buffer, overwriting oldest data if necessary
func (rb *RingBuffer) Write(data []byte) (n int, err error) {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	// Count newlines in the data
	for _, b := range data {
		if b == '\n' {
			rb.lineCount++
		}
	}

	// If data is larger than capacity, just keep the last capacity bytes
	if len(data) >= rb.capacity {
		copy(rb.buffer, data[len(data)-rb.capacity:])
		rb.size = rb.capacity
		return len(data), nil
	}

	// If buffer will exceed capacity, shift data
	if rb.size+len(data) > rb.capacity {
		// Calculate how many bytes we need to remove
		removeBytes := rb.size + len(data) - rb.capacity
		
		// Shift buffer to remove oldest data
		copy(rb.buffer, rb.buffer[removeBytes:rb.size])
		rb.size -= removeBytes
	}

	// Append new data
	if rb.size+len(data) <= cap(rb.buffer) {
		rb.buffer = rb.buffer[:rb.size+len(data)]
	} else {
		newBuf := make([]byte, rb.size+len(data))
		copy(newBuf, rb.buffer[:rb.size])
		rb.buffer = newBuf
	}
	
	copy(rb.buffer[rb.size:], data)
	rb.size += len(data)

	return len(data), nil
}

// GetLastLines returns the last n lines from the buffer
func (rb *RingBuffer) GetLastLines(n int) string {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	if rb.size == 0 {
		return ""
	}

	// Convert buffer to string
	content := string(rb.buffer[:rb.size])
	
	// Split into lines
	lines := strings.Split(content, "\n")
	
	// If the last element is empty (due to trailing newline), remove it
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	
	// Return last n lines or all lines if fewer than n
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}
	
	return strings.Join(lines[start:], "\n")
}

// GetContent returns the entire content of the buffer
func (rb *RingBuffer) GetContent() string {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	if rb.size == 0 {
		return ""
	}

	return string(rb.buffer[:rb.size])
}

// Clear empties the buffer
func (rb *RingBuffer) Clear() {
	rb.mutex.Lock()
	defer rb.mutex.Unlock()

	rb.buffer = make([]byte, 0, rb.capacity)
	rb.size = 0
	rb.lineCount = 0
}
