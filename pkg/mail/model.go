package mail

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

type Email struct {
	From        string       `json:"from"`
	To          []string     `json:"to"`
	Subject     string       `json:"subject"`
	Message     string       `json:"message"`
	Attachments []Attachment `json:"attachments"`
}

// UID returns the Blake2b-192 hash of the email in JSON format
func (e *Email) UID() (string, error) {
	// Marshal the email to JSON
	emailJSON, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("failed to marshal email: %w", err)
	}

	// Create a Blake2b-192 hash (24 bytes) from the email JSON
	hash, err := blake2b.New(24, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create Blake2b hash: %w", err)
	}

	// Write the data to the hash
	_, err = hash.Write(emailJSON)
	if err != nil {
		return "", fmt.Errorf("failed to write to hash: %w", err)
	}

	// Get the hash sum and convert to hex string
	hashSum := hash.Sum(nil)
	hashHex := hex.EncodeToString(hashSum)

	return hashHex, nil
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // Base64 encoded binary data
}
