package models

import (
	"sync"
)

// Store represents a database interface for all models
// This is kept for backward compatibility
type Store struct {
	mu sync.RWMutex
}

// NewStore creates a new store with connection to the database
func NewStore() *Store {
	return &Store{}
}
