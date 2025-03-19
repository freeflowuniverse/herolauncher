package api

import "time"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// Executor Models

// ExecuteCommandRequest represents a request to execute a command
type ExecuteCommandRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// ExecuteCommandResponse represents the response from executing a command
type ExecuteCommandResponse struct {
	JobID string `json:"job_id"`
}

// JobResponse represents a job response
type JobResponse struct {
	ID        string    `json:"id"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
	Output    string    `json:"output"`
	Error     string    `json:"error"`
}

// Package Manager Models

// InstallPackageRequest represents a request to install a package
type InstallPackageRequest struct {
	PackageName string `json:"package_name"`
}

// InstallPackageResponse represents the response from installing a package
type InstallPackageResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

// UninstallPackageRequest represents a request to uninstall a package
type UninstallPackageRequest struct {
	PackageName string `json:"package_name"`
}

// UninstallPackageResponse represents the response from uninstalling a package
type UninstallPackageResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

// ListPackagesResponse represents the response from listing packages
type ListPackagesResponse struct {
	Packages []string `json:"packages"`
}

// SearchPackagesResponse represents the response from searching packages
type SearchPackagesResponse struct {
	Packages []string `json:"packages"`
}



// Redis Models

// SetKeyRequest represents a request to set a key
type SetKeyRequest struct {
	Key               string `json:"key"`
	Value             string `json:"value"`
	ExpirationSeconds int    `json:"expiration_seconds"`
}

// SetKeyResponse represents the response from setting a key
type SetKeyResponse struct {
	Success bool `json:"success"`
}

// GetKeyResponse represents the response from getting a key
type GetKeyResponse struct {
	Value string `json:"value"`
}

// DeleteKeyResponse represents the response from deleting a key
type DeleteKeyResponse struct {
	Count int `json:"count"`
}

// GetKeysResponse represents the response from getting keys
type GetKeysResponse struct {
	Keys []string `json:"keys"`
}

// HSetKeyRequest represents a request to set a hash field
type HSetKeyRequest struct {
	Key   string `json:"key"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// HSetKeyResponse represents the response from setting a hash field
type HSetKeyResponse struct {
	Added bool `json:"added"`
}

// HGetKeyResponse represents the response from getting a hash field
type HGetKeyResponse struct {
	Value string `json:"value"`
}

// HDelKeyRequest represents a request to delete hash fields
type HDelKeyRequest struct {
	Key    string   `json:"key"`
	Fields []string `json:"fields"`
}

// HDelKeyResponse represents the response from deleting hash fields
type HDelKeyResponse struct {
	Count int `json:"count"`
}

// HKeysResponse represents the response from getting hash keys
type HKeysResponse struct {
	Fields []string `json:"fields"`
}

// HLenResponse represents the response from getting hash length
type HLenResponse struct {
	Length int `json:"length"`
}

// IncrKeyResponse represents the response from incrementing a key
type IncrKeyResponse struct {
	Value int64 `json:"value"`
}
