package imapserver

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/backend"
	"github.com/redis/go-redis/v9"
)

// Backend implements the go-imap backend interface
type Backend struct {
	redisClient *redis.Client
	ctx         context.Context
	debugMode   bool
}

// NewBackend creates a new IMAP backend
func NewBackend(redisClient *redis.Client, debugMode bool) *Backend {
	return &Backend{
		redisClient: redisClient,
		ctx:         context.Background(),
		debugMode:   debugMode,
	}
}

// Login authenticates a user. In this implementation, we accept any username/password
func (b *Backend) Login(_ *imap.ConnInfo, username, password string) (backend.User, error) {
	log.Printf("Login attempt for user: %s", username)
	// Accept any username/password combination
	return &User{
		backend:  b,
		username: username,
	}, nil
}

// User represents a user connected to the IMAP server
type User struct {
	backend  *Backend
	username string
}

// Username returns the user's username
func (u *User) Username() string {
	return u.username
}

// ListMailboxes returns a list of mailboxes available for this user
func (u *User) ListMailboxes(subscribed bool) ([]backend.Mailbox, error) {
	log.Printf("Listing mailboxes for user: %s", u.username)
	
	// Get all keys matching the pattern mail:in:username:*
	pattern := fmt.Sprintf("mail:in:%s:*", u.username)
	
	if u.backend.debugMode {
		log.Printf("DEBUG: Querying Redis with pattern: %s", pattern)
	}
	
	keys, err := u.backend.redisClient.Keys(u.backend.ctx, pattern).Result()
	if err != nil {
		log.Printf("ERROR: Failed to query Redis: %v", err)
		return nil, err
	}
	
	if u.backend.debugMode {
		log.Printf("DEBUG: Found %d keys for user %s", len(keys), u.username)
		for i, key := range keys {
			log.Printf("DEBUG: Key[%d]: %s", i, key)
		}
	} else {
		log.Printf("Found %d keys for user %s", len(keys), u.username)
	}

	// Extract unique folder names from keys
	folderMap := make(map[string]bool)
	for _, key := range keys {
		if u.backend.debugMode {
			log.Printf("DEBUG: Processing key: %s", key)
		}
		
		parts := strings.Split(key, ":")
		
		if len(parts) >= 4 {
			// The folder name is the 4th part (index 3)
			folderName := parts[3]
			
			if u.backend.debugMode {
				log.Printf("DEBUG: Original folder name from key: %s", folderName)
			}
			
			// Check if the folder name contains a UID (with slash)
			if strings.Contains(folderName, "/") {
				// Split by slash and take the folder part
				folderParts := strings.Split(folderName, "/")
				folderName = folderParts[0]
				if u.backend.debugMode {
					log.Printf("DEBUG: Folder name after removing UID part: %s", folderName)
				}
			}
			
			// Convert folder name to lowercase for consistency
			folderName = strings.ToLower(folderName)
			
			// Add the folder to our map
			log.Printf("Found mailbox: %s", folderName)
			folderMap[folderName] = true
			
			// Check for nested folders (e.g., sent/work)
			if strings.Contains(folderName, "/") {
				nestedParts := strings.Split(folderName, "/")
				for i := 1; i < len(nestedParts); i++ {
					nestedFolder := strings.Join(nestedParts[:i+1], "/")
					// Convert nested folder name to lowercase
					nestedFolder = strings.ToLower(nestedFolder)
					log.Printf("Found nested mailbox: %s", nestedFolder)
					folderMap[nestedFolder] = true
				}
			}
		}
	}

	// Create mailbox objects for each unique folder
	mailboxes := make([]backend.Mailbox, 0, len(folderMap))
	for folder := range folderMap {
		log.Printf("Creating mailbox object for: %s", folder)
		mailboxes = append(mailboxes, &Mailbox{
			backend:  u.backend,
			user:     u,
			name:     folder,
			messages: nil, // Messages will be loaded on demand
		})
	}
	
	if u.backend.debugMode {
		log.Printf("DEBUG: Created %d mailbox objects", len(mailboxes))
		for i, m := range mailboxes {
			log.Printf("DEBUG: Mailbox[%d]: %s", i, m.Name())
		}
	}

	// Only add standard mailboxes if there are no mailboxes found
	if len(folderMap) == 0 {
		if u.backend.debugMode {
			log.Printf("DEBUG: No mailboxes found in Redis, adding standard mailboxes")
		}
		
		// Use lowercase for standard mailboxes
		standardMailboxes := []string{"inbox", "sent", "drafts", "trash"}
		for _, stdBox := range standardMailboxes {
			log.Printf("Adding standard mailbox (no mailboxes found): %s", stdBox)
			mailboxes = append(mailboxes, &Mailbox{
				backend:  u.backend,
				user:     u,
				name:     stdBox,
				messages: nil,
			})
		}
	} else if u.backend.debugMode {
		log.Printf("DEBUG: Found %d mailboxes in Redis, not adding standard mailboxes", len(folderMap))
	}

	return mailboxes, nil
}

// GetMailbox returns a mailbox by name
func (u *User) GetMailbox(name string) (backend.Mailbox, error) {
	// Convert mailbox name to lowercase for consistency
	lowerName := strings.ToLower(name)
	log.Printf("Getting mailbox %s (lowercase: %s) for user: %s", name, lowerName, u.username)
	
	// Create a new mailbox object with lowercase name
	mailbox := &Mailbox{
		backend:  u.backend,
		user:     u,
		name:     lowerName,
		messages: nil, // Messages will be loaded on demand
	}

	// Load messages for this mailbox
	if err := mailbox.loadMessages(); err != nil {
		return nil, err
	}

	return mailbox, nil
}

// CreateMailbox creates a new mailbox
func (u *User) CreateMailbox(name string) error {
	// Convert mailbox name to lowercase for consistency
	lowerName := strings.ToLower(name)
	log.Printf("Creating mailbox %s (lowercase: %s) for user: %s", name, lowerName, u.username)
	// Since we're using Redis as a key-value store, we don't need to explicitly create mailboxes
	// They will be created implicitly when messages are added
	return nil
}

// DeleteMailbox deletes a mailbox
func (u *User) DeleteMailbox(name string) error {
	// Convert mailbox name to lowercase for consistency
	lowerName := strings.ToLower(name)
	log.Printf("Deleting mailbox %s (lowercase: %s) for user: %s", name, lowerName, u.username)
	
	// Get all keys matching the pattern mail:in:username:mailbox:*
	pattern := fmt.Sprintf("mail:in:%s:%s:*", u.username, lowerName)
	keys, err := u.backend.redisClient.Keys(u.backend.ctx, pattern).Result()
	if err != nil {
		return err
	}

	// Delete all keys in this mailbox
	if len(keys) > 0 {
		return u.backend.redisClient.Del(u.backend.ctx, keys...).Err()
	}

	return nil
}

// RenameMailbox renames a mailbox
func (u *User) RenameMailbox(existingName, newName string) error {
	// Convert mailbox names to lowercase for consistency
	lowerExistingName := strings.ToLower(existingName)
	lowerNewName := strings.ToLower(newName)
	log.Printf("Renaming mailbox %s (lowercase: %s) to %s (lowercase: %s) for user: %s", 
		existingName, lowerExistingName, newName, lowerNewName, u.username)
	
	// Get all keys matching the pattern mail:in:username:oldmailbox:*
	pattern := fmt.Sprintf("mail:in:%s:%s:*", u.username, lowerExistingName)
	keys, err := u.backend.redisClient.Keys(u.backend.ctx, pattern).Result()
	if err != nil {
		return err
	}

	// For each key, create a new key with the new mailbox name and copy the value
	for _, key := range keys {
		// Get the value
		value, err := u.backend.redisClient.Get(u.backend.ctx, key).Result()
		if err != nil {
			return err
		}

		// Create a new key with the new mailbox name
		parts := strings.Split(key, ":")
		if len(parts) >= 5 {
			newKey := fmt.Sprintf("mail:in:%s:%s:%s", u.username, lowerNewName, parts[4])
			
			// Set the new key with the same value
			if err := u.backend.redisClient.Set(u.backend.ctx, newKey, value, 0).Err(); err != nil {
				return err
			}
			
			// Delete the old key
			if err := u.backend.redisClient.Del(u.backend.ctx, key).Err(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Logout is called when a user logs out
func (u *User) Logout() error {
	log.Printf("User logged out: %s", u.username)
	return nil
}
