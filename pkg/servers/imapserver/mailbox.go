package imapserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/freeflowuniverse/herolauncher/pkg/mail"
)

// Mailbox represents an IMAP mailbox (folder)
type Mailbox struct {
	backend  *Backend
	user     *User
	name     string
	messages []*Message
}

// Name returns the mailbox name
func (m *Mailbox) Name() string {
	return m.name
}

// Info returns information about the mailbox
func (m *Mailbox) Info() (*imap.MailboxInfo, error) {
	info := &imap.MailboxInfo{
		Attributes: []string{},
		Delimiter:  "/",
		Name:       m.name,
	}

	// Add attributes based on mailbox name
	if strings.EqualFold(m.name, "inbox") {
		// inbox is a special mailbox
		info.Attributes = append(info.Attributes, "\\Inbox")
	}

	// Handle standard mailboxes
	if strings.EqualFold(m.name, "sent") || strings.EqualFold(m.name, "sent items") {
		info.Attributes = append(info.Attributes, "\\Sent")
	} else if strings.EqualFold(m.name, "drafts") {
		info.Attributes = append(info.Attributes, "\\Drafts")
	} else if strings.EqualFold(m.name, "trash") {
		info.Attributes = append(info.Attributes, "\\Trash")
	} else if strings.EqualFold(m.name, "junk") || strings.EqualFold(m.name, "spam") {
		info.Attributes = append(info.Attributes, "\\Junk")
	}

	// Handle nested folders
	if strings.Contains(m.name, "/") {
		// This is a child folder, mark it as such
		info.Attributes = append(info.Attributes, "\\HasChildren")

		// Log for debugging
		if m.backend.debugMode {
			log.Printf("DEBUG: Mailbox %s is a nested folder", m.name)
		}
	}

	return info, nil
}

// Status returns the mailbox status
func (m *Mailbox) Status(items []imap.StatusItem) (*imap.MailboxStatus, error) {
	log.Printf("Getting status for mailbox %s", m.name)

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return nil, err
	}

	status := imap.NewMailboxStatus(m.name, items)
	status.Flags = []string{imap.SeenFlag, imap.AnsweredFlag, imap.FlaggedFlag, imap.DeletedFlag, imap.DraftFlag}
	status.PermanentFlags = []string{imap.SeenFlag, imap.AnsweredFlag, imap.FlaggedFlag, imap.DeletedFlag, imap.DraftFlag}

	// Filter messages to only include direct messages for this mailbox (not in subfolders)
	var directMessages []*Message
	lowerName := strings.ToLower(m.name)
	prefix := fmt.Sprintf("mail:in:%s:%s:", m.user.username, lowerName)

	for _, msg := range m.messages {
		// Check if this message belongs directly to this mailbox (not a subfolder)
		if msg.Key != "" && strings.HasPrefix(msg.Key, prefix) && !strings.Contains(msg.Key[len(prefix):], "/") {
			directMessages = append(directMessages, msg)
		}
	}

	if m.backend.debugMode {
		log.Printf("DEBUG: Mailbox %s has %d total messages, %d direct messages",
			m.name, len(m.messages), len(directMessages))
	}

	for _, item := range items {
		switch item {
		case imap.StatusMessages:
			// Only count direct messages, not messages in subfolders
			status.Messages = uint32(len(directMessages))
		case imap.StatusRecent:
			status.Recent = 0 // No recent messages for simplicity
		case imap.StatusUnseen:
			// Count unseen messages (only direct messages)
			unseen := 0
			for _, msg := range directMessages {
				if !contains(msg.Flags, imap.SeenFlag) {
					unseen++
				}
			}
			status.Unseen = uint32(unseen)
		case imap.StatusUidNext:
			// Use current timestamp as next UID for simplicity
			status.UidNext = uint32(time.Now().Unix())
		case imap.StatusUidValidity:
			// Use a fixed value for simplicity
			status.UidValidity = 1
		}
	}

	return status, nil
}

// SetSubscribed sets the mailbox subscription status
func (m *Mailbox) SetSubscribed(subscribed bool) error {
	// We accept all subscription changes but don't actually track them
	return nil
}

// Check checks the mailbox for updates
func (m *Mailbox) Check() error {
	// Reload messages to get any updates
	return m.loadMessages()
}

// ListMessages returns a list of messages
func (m *Mailbox) ListMessages(uid bool, seqSet *imap.SeqSet, items []imap.FetchItem, ch chan<- *imap.Message) error {
	defer close(ch)

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

	// Sort messages by UID
	sort.Slice(m.messages, func(i, j int) bool {
		return m.messages[i].Uid < m.messages[j].Uid
	})

	for i, msg := range m.messages {
		seqNum := uint32(i + 1)

		// Check if this message should be included based on the sequence set
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}
		if !seqSet.Contains(id) {
			continue
		}

		// Convert to IMAP message
		imapMsg, err := msg.Fetch(seqNum, items)
		if err != nil {
			return err
		}

		ch <- imapMsg
	}

	return nil
}

// SearchMessages searches for messages matching the given criteria
func (m *Mailbox) SearchMessages(uid bool, criteria *imap.SearchCriteria) ([]uint32, error) {
	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return nil, err
	}

	var ids []uint32
	for i, msg := range m.messages {
		seqNum := uint32(i + 1)

		// Check if message matches criteria
		if msg.Match(seqNum, criteria) {
			if uid {
				ids = append(ids, msg.Uid)
			} else {
				ids = append(ids, seqNum)
			}
		}
	}

	return ids, nil
}

// CreateMessage adds a new message to the mailbox
func (m *Mailbox) CreateMessage(flags []string, date time.Time, body imap.Literal) error {
	// Ensure mailbox name is lowercase
	lowerName := strings.ToLower(m.name)
	log.Printf("Creating message in mailbox %s (lowercase: %s) for user %s", m.name, lowerName, m.user.username)

	// Read the message body
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(body); err != nil {
		return fmt.Errorf("failed to read message body: %w", err)
	}

	// Parse the message content
	// For simplicity, we'll extract basic headers and use the rest as the message body
	content := buf.String()
	headers, messageBody := parseEmailContent(content)

	// Create a new Email object
	email := &mail.Email{
		Message:     messageBody,
		Attachments: []mail.Attachment{}, // No attachments for now
	}

	// Set envelope fields
	email.SetFrom(headers["From"])
	email.SetTo(strings.Split(headers["To"], ","))
	email.SetSubject(headers["Subject"])

	// Generate a UID for the email
	uid, err := email.UID()
	if err != nil {
		return fmt.Errorf("failed to generate email UID: %w", err)
	}

	// Serialize the email to JSON
	emailJSON, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	// Store the email in Redis
	// Key format: mail:in:<username>:<mailbox>/<uid>
	// Always use lowercase mailbox name for consistency
	key := fmt.Sprintf("mail:in:%s:%s/%s", m.user.username, lowerName, uid)
	err = m.backend.redisClient.Set(m.backend.ctx, key, string(emailJSON), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store email in Redis: %w", err)
	}

	// If successful, reload the messages to include the new one
	return m.loadMessages()
}

// UpdateMessagesFlags updates flags for the specified messages
func (m *Mailbox) UpdateMessagesFlags(uid bool, seqSet *imap.SeqSet, operation imap.FlagsOp, flags []string) error {
	log.Printf("Updating flags for messages in mailbox %s, operation: %v, flags: %v", m.name, operation, flags)

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

	// Keep track of modified messages
	var modifiedMessages []*Message

	for i, msg := range m.messages {
		seqNum := uint32(i + 1)

		// Check if this message should be included
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}
		if !seqSet.Contains(id) {
			continue
		}

		// Store original flags to check if they changed
		originalFlags := make([]string, len(msg.Flags))
		copy(originalFlags, msg.Flags)

		// Update flags
		switch operation {
		case imap.SetFlags:
			msg.Flags = flags
		case imap.AddFlags:
			msg.Flags = addFlags(msg.Flags, flags)
		case imap.RemoveFlags:
			msg.Flags = removeFlags(msg.Flags, flags)
		}

		// Check if flags have changed
		if !equalFlags(originalFlags, msg.Flags) {
			modifiedMessages = append(modifiedMessages, msg)
			if m.backend.debugMode {
				log.Printf("DEBUG: Message UID %d flags changed from %v to %v", msg.Uid, originalFlags, msg.Flags)
			}
		}
	}

	// Persist flag changes to Redis
	if len(modifiedMessages) > 0 {
		log.Printf("Persisting flag changes for %d messages to Redis", len(modifiedMessages))

		for _, msg := range modifiedMessages {
			// We need to update the message in Redis
			if msg.Key == "" {
				log.Printf("WARNING: Cannot update message in Redis, key is empty for UID %d", msg.Uid)
				continue
			}

			// Get the current message data from Redis
			emailJSON, err := m.backend.redisClient.Get(m.backend.ctx, msg.Key).Result()
			if err != nil {
				log.Printf("ERROR: Failed to get message data from Redis for key %s: %v", msg.Key, err)
				continue
			}

			// Parse the email JSON
			var email mail.Email
			err = json.Unmarshal([]byte(emailJSON), &email)
			if err != nil {
				log.Printf("ERROR: Failed to unmarshal email JSON for key %s: %v", msg.Key, err)
				continue
			}

			// Update the flags in the email object
			email.Flags = msg.Flags

			// Marshal the updated email back to JSON
			updatedJSON, err := json.Marshal(email)
			if err != nil {
				log.Printf("ERROR: Failed to marshal updated email to JSON for key %s: %v", msg.Key, err)
				continue
			}

			// Save the updated email back to Redis
			_, err = m.backend.redisClient.Set(m.backend.ctx, msg.Key, string(updatedJSON), 0).Result()
			if err != nil {
				log.Printf("ERROR: Failed to save updated email to Redis for key %s: %v", msg.Key, err)
				continue
			}

			if m.backend.debugMode {
				log.Printf("DEBUG: Successfully updated flags in Redis for message UID %d, key %s", msg.Uid, msg.Key)
			}
		}
	}

	return nil
}

// equalFlags checks if two flag slices contain the same flags (order doesn't matter)
func equalFlags(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	// Create maps to count occurrences of each flag
	countA := make(map[string]int)
	countB := make(map[string]int)

	for _, flag := range a {
		countA[flag]++
	}

	for _, flag := range b {
		countB[flag]++
	}

	// Compare the maps
	for flag, count := range countA {
		if countB[flag] != count {
			return false
		}
	}

	return true
}

// CopyMessages copies the specified messages to another mailbox
func (m *Mailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, destName string) error {
	log.Printf("Copying messages to mailbox %s", destName)

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

	// Find the destination mailbox
	destMailbox, err := m.user.GetMailbox(destName)
	if err != nil {
		return fmt.Errorf("destination mailbox not found: %w", err)
	}

	// Sort messages by UID
	sort.Slice(m.messages, func(i, j int) bool {
		return m.messages[i].Uid < m.messages[j].Uid
	})

	for i, msg := range m.messages {
		seqNum := uint32(i + 1)

		// Check if this message should be included based on the sequence set
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}
		if !seqSet.Contains(id) {
			continue
		}

		// Create a new message in the destination mailbox
		// We need to convert the Email back to a literal for CreateMessage
		emailContent := formatEmailContent(msg.Email)
		literal := bytes.NewBufferString(emailContent)

		// Copy the message to the destination mailbox
		err := destMailbox.CreateMessage(msg.Flags, time.Now(), literal)
		if err != nil {
			return fmt.Errorf("failed to copy message: %w", err)
		}
	}

	return nil
}

// Expunge permanently removes messages marked for deletion
func (m *Mailbox) Expunge() error {
	log.Printf("Expunging deleted messages from mailbox %s", m.name)

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

	// Find messages marked for deletion
	var messagesToDelete []*Message
	for _, msg := range m.messages {
		if contains(msg.Flags, imap.DeletedFlag) {
			messagesToDelete = append(messagesToDelete, msg)
		}
	}

	if len(messagesToDelete) == 0 {
		log.Printf("No messages marked for deletion in mailbox %s", m.name)
		return nil
	}

	log.Printf("Found %d messages to delete", len(messagesToDelete))

	// Delete each message from Redis
	for _, msg := range messagesToDelete {
		// Get the original Redis key for this message
		// We need to find the exact key that was used to store this message
		// since it could be in a nested folder
		lowerName := strings.ToLower(m.name)

		// Create patterns to match both direct messages and nested folders
		patterns := []string{
			fmt.Sprintf("mail:in:%s:%s:%d", m.user.username, lowerName, msg.Uid),
			fmt.Sprintf("mail:in:%s:%s/*:%d", m.user.username, lowerName, msg.Uid),
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Looking for message with UID %d using patterns: %v", msg.Uid, patterns)
		}

		// Check each pattern for matching keys
		deleted := false
		for _, pattern := range patterns {
			keys, err := m.backend.redisClient.Keys(m.backend.ctx, pattern).Result()
			if err != nil {
				log.Printf("ERROR: Failed to get keys from Redis with pattern %s: %v", pattern, err)
				continue
			}

			if m.backend.debugMode {
				log.Printf("DEBUG: Found %d keys matching pattern %s", len(keys), pattern)
			}

			// Delete all matching keys
			for _, key := range keys {
				if m.backend.debugMode {
					log.Printf("DEBUG: Deleting message with key %s", key)
				}

				_, err := m.backend.redisClient.Del(m.backend.ctx, key).Result()
				if err != nil {
					log.Printf("Error deleting message with key %s: %v", key, err)
					// Continue with other deletions even if one fails
				} else {
					log.Printf("Successfully deleted message with key %s", key)
					deleted = true
				}
			}
		}

		if !deleted && m.backend.debugMode {
			log.Printf("WARNING: Could not find any Redis keys for message with UID %d", msg.Uid)
		}
	}

	// Reload messages to reflect the changes
	return m.loadMessages()
}

// MoveMessages moves the specified messages to another mailbox
func (m *Mailbox) MoveMessages(uid bool, seqSet *imap.SeqSet, destName string) error {
	log.Printf("Moving messages from %s to %s", m.name, destName)

	// First, copy the messages to the destination mailbox
	err := m.CopyMessages(uid, seqSet, destName)
	if err != nil {
		return fmt.Errorf("failed to copy messages during move operation: %w", err)
	}

	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

	// Find messages that match the sequence set
	var messagesToDelete []*Message
	for i, msg := range m.messages {
		seqNum := uint32(i + 1)

		// Check if this message should be included based on the sequence set
		var id uint32
		if uid {
			id = msg.Uid
		} else {
			id = seqNum
		}

		if seqSet.Contains(id) {
			// Mark message for deletion
			msg.Flags = addFlags(msg.Flags, []string{imap.DeletedFlag})
			messagesToDelete = append(messagesToDelete, msg)

			// Log the message being marked for deletion
			if m.backend.debugMode {
				log.Printf("DEBUG: Marking message with UID %d for deletion during move operation", msg.Uid)
			}
		}
	}

	if len(messagesToDelete) == 0 {
		log.Printf("No messages found to move from %s to %s", m.name, destName)
		return nil
	}

	log.Printf("Marked %d messages for deletion after copying to %s", len(messagesToDelete), destName)

	// Directly delete the messages from Redis without relying on Expunge
	// This ensures that the messages are actually deleted from Redis
	for _, msg := range messagesToDelete {
		// Get the original Redis key for this message
		lowerName := strings.ToLower(m.name)

		// Create patterns to match both direct messages and nested folders
		patterns := []string{
			fmt.Sprintf("mail:in:%s:%s:%d", m.user.username, lowerName, msg.Uid),
			fmt.Sprintf("mail:in:%s:%s/*:%d", m.user.username, lowerName, msg.Uid),
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Looking for message with UID %d to delete during move using patterns: %v", msg.Uid, patterns)
		}

		// Check each pattern for matching keys
		deleted := false
		for _, pattern := range patterns {
			keys, err := m.backend.redisClient.Keys(m.backend.ctx, pattern).Result()
			if err != nil {
				log.Printf("ERROR: Failed to get keys from Redis with pattern %s: %v", pattern, err)
				continue
			}

			if m.backend.debugMode {
				log.Printf("DEBUG: Found %d keys matching pattern %s during move operation", len(keys), pattern)
			}

			// Delete all matching keys
			for _, key := range keys {
				if m.backend.debugMode {
					log.Printf("DEBUG: Deleting message with key %s during move operation", key)
				}

				_, err := m.backend.redisClient.Del(m.backend.ctx, key).Result()
				if err != nil {
					log.Printf("Error deleting message with key %s during move: %v", key, err)
				} else {
					log.Printf("Successfully deleted message with key %s during move", key)
					deleted = true
				}
			}
		}

		if !deleted && m.backend.debugMode {
			log.Printf("WARNING: Could not find any Redis keys for message with UID %d during move", msg.Uid)
		}
	}

	// Reload messages to reflect the changes
	if err := m.loadMessages(); err != nil {
		return fmt.Errorf("failed to reload messages after move operation: %w", err)
	}

	log.Printf("Successfully moved %d messages from %s to %s", len(messagesToDelete), m.name, destName)
	return nil
}

// loadMessages loads messages from Redis for this mailbox
func (m *Mailbox) loadMessages() error {
	log.Printf("Loading messages for mailbox %s, user %s", m.name, m.user.username)

	// Always use lowercase mailbox name for consistency
	lowerName := strings.ToLower(m.name)

	// Create patterns to match both direct messages and nested folders
	// For example, for inbox, we want to match both:
	// - mail:in:username:inbox:*  (direct messages in this mailbox)
	// - mail:in:username:inbox/*  (messages in subfolders)

	// First pattern is for direct messages in this mailbox (not in subfolders)
	directPattern := fmt.Sprintf("mail:in:%s:%s:*", m.user.username, lowerName)

	// Second pattern is for messages in subfolders
	subfolderPattern := fmt.Sprintf("mail:in:%s:%s/*", m.user.username, lowerName)

	// We'll collect all keys from both patterns
	var allKeys []string
	var directKeys []string
	var subfolderKeys []string

	if m.backend.debugMode {
		log.Printf("DEBUG: Using direct pattern: %s", directPattern)
		log.Printf("DEBUG: Using subfolder pattern: %s", subfolderPattern)
	} else {
		log.Printf("Using patterns: direct=%s, subfolder=%s", directPattern, subfolderPattern)
	}

	// Get direct messages
	keys, err := m.backend.redisClient.Keys(m.backend.ctx, directPattern).Result()
	if err != nil {
		log.Printf("ERROR: Failed to get keys from Redis with pattern %s: %v", directPattern, err)
		return err
	}

	if m.backend.debugMode {
		log.Printf("DEBUG: Found %d direct message keys with pattern %s", len(keys), directPattern)
	}
	directKeys = keys

	// Get subfolder messages
	keys, err = m.backend.redisClient.Keys(m.backend.ctx, subfolderPattern).Result()
	if err != nil {
		log.Printf("ERROR: Failed to get keys from Redis with pattern %s: %v", subfolderPattern, err)
		return err
	}

	if m.backend.debugMode {
		log.Printf("DEBUG: Found %d subfolder message keys with pattern %s", len(keys), subfolderPattern)
	}
	subfolderKeys = keys

	// Combine all keys
	allKeys = append(directKeys, subfolderKeys...)

	if m.backend.debugMode {
		log.Printf("DEBUG: Found %d total keys (%d direct, %d subfolder)", len(allKeys), len(directKeys), len(subfolderKeys))
		for i, key := range allKeys {
			log.Printf("DEBUG: Key[%d]: %s", i, key)
		}
	} else {
		log.Printf("Found %d total keys (%d direct, %d subfolder)", len(allKeys), len(directKeys), len(subfolderKeys))
	}

	// If still no keys found, try with a more generic pattern
	if len(allKeys) == 0 {
		// List all keys for this user to debug
		allKeysPattern := fmt.Sprintf("mail:in:%s:*", m.user.username)

		if m.backend.debugMode {
			log.Printf("DEBUG: No keys found, listing all keys for user with pattern: %s", allKeysPattern)
		} else {
			log.Printf("No keys found, listing all keys for user with pattern: %s", allKeysPattern)
		}

		// Use KEYS command for the generic pattern
		allUserKeys, err := m.backend.redisClient.Keys(m.backend.ctx, allKeysPattern).Result()
		if err != nil {
			log.Printf("ERROR: Failed to get all keys from Redis: %v", err)
			return err
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Found %d total keys for user %s using SCAN", len(allUserKeys), m.user.username)
			for i, k := range allUserKeys {
				log.Printf("DEBUG: AvailableKey[%d]: %s", i, k)
			}
		} else {
			log.Printf("Found %d total keys for user %s", len(allUserKeys), m.user.username)
			for _, k := range allUserKeys {
				log.Printf("Available key: %s", k)
			}
		}
	}

	if m.backend.debugMode {
		log.Printf("DEBUG: Processing %d keys matching pattern for mailbox %s", len(allKeys), m.name)
	} else {
		log.Printf("Found %d keys matching pattern for mailbox %s", len(allKeys), m.name)
	}

	// Clear existing messages
	m.messages = nil

	// Load each message
	for i, key := range allKeys {
		if m.backend.debugMode {
			log.Printf("DEBUG: Processing key[%d]: %s", i, key)
		} else {
			log.Printf("Processing key: %s", key)
		}

		// Get the email JSON from Redis
		if m.backend.debugMode {
			log.Printf("DEBUG: Fetching email data from Redis for key: %s", key)
		}

		emailJSON, err := m.backend.redisClient.Get(m.backend.ctx, key).Result()
		if err != nil {
			log.Printf("ERROR: Failed to get email from Redis: %v", err)
			continue
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Retrieved email data (%d bytes)", len(emailJSON))
		}

		// Parse the email JSON
		if m.backend.debugMode {
			log.Printf("DEBUG: Unmarshaling email JSON data")
		}

		var email mail.Email
		if err := json.Unmarshal([]byte(emailJSON), &email); err != nil {
			log.Printf("ERROR: Failed to unmarshal email JSON: %v", err)

			if m.backend.debugMode {
				// Show a snippet of the JSON for debugging
				jsonPreview := emailJSON
				if len(jsonPreview) > 100 {
					jsonPreview = jsonPreview[:100] + "..."
				}
				log.Printf("DEBUG: Invalid JSON data: %s", jsonPreview)
			}

			continue
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Successfully parsed email: From=%s, To=%v, Subject=%s",
				email.From, email.To, email.Subject)
		}

		// Extract UID from the key
		// The key format appears to be mail:in:jan:mailbox:UID or mail:in:jan:mailbox/UID
		if m.backend.debugMode {
			log.Printf("DEBUG: Extracting UID from key: %s", key)
		}

		uidStr := ""
		parts := strings.Split(key, ":")

		if m.backend.debugMode {
			log.Printf("DEBUG: Key parts: %v", parts)
		}

		if len(parts) >= 4 {
			// Check if the last part contains a slash (for nested folders)
			lastPart := parts[len(parts)-1]

			if m.backend.debugMode {
				log.Printf("DEBUG: Last part of key: %s", lastPart)
			}

			if strings.Contains(lastPart, "/") {
				// Format is likely mail:in:username:mailbox/UID
				slashParts := strings.Split(lastPart, "/")

				if m.backend.debugMode {
					log.Printf("DEBUG: Slash parts: %v", slashParts)
				}

				if len(slashParts) > 1 {
					uidStr = slashParts[len(slashParts)-1]
					if m.backend.debugMode {
						log.Printf("DEBUG: Extracted UID from slash part: %s", uidStr)
					}
				}
			} else {
				// Format is likely mail:in:username:mailbox:UID
				uidStr = lastPart
				if m.backend.debugMode {
					log.Printf("DEBUG: Using last part as UID: %s", uidStr)
				}
			}
		}

		if uidStr == "" {
			log.Printf("ERROR: Could not extract UID from key: %s", key)
			continue
		}

		if m.backend.debugMode {
			log.Printf("DEBUG: Successfully extracted UID: %s", uidStr)
		} else {
			log.Printf("Extracted UID: %s", uidStr)
		}

		// Create a new message
		if m.backend.debugMode {
			log.Printf("DEBUG: Creating message object with UID string: %s", uidStr)
		}

		parsedUID := parseUID(uidStr)

		if m.backend.debugMode {
			log.Printf("DEBUG: Parsed UID string to numeric value: %d", parsedUID)
		}

		msg := &Message{
			Email: &email,
			Uid:   parsedUID,
			Flags: []string{}, // No flags by default
			Key:   key,        // Store the Redis key
		}

		m.messages = append(m.messages, msg)

		if m.backend.debugMode {
			log.Printf("DEBUG: Added message with UID: %d, Subject: %s", msg.Uid, email.Subject)
		} else {
			log.Printf("Added message with UID: %d", msg.Uid)
		}
	}

	return nil
}

// parseUID converts a string UID to uint32
func parseUID(uidStr string) uint32 {
	// Try to parse as integer first
	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err == nil {
		return uint32(uid)
	}

	// If the string contains non-numeric characters, extract the numeric part
	var numericPart string
	for _, c := range uidStr {
		if c >= '0' && c <= '9' {
			numericPart += string(c)
		}
	}

	// If we extracted a numeric part, try to parse it
	if numericPart != "" {
		uid, err := strconv.ParseUint(numericPart, 10, 32)
		if err == nil {
			return uint32(uid)
		}
	}

	// If all else fails, use a hash of the string
	var hash uint32 = 1
	for _, c := range uidStr {
		hash = hash*31 + uint32(c)
	}
	return hash
}

// Helper function to check if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to add flags to a slice
func addFlags(slice []string, flags []string) []string {
	result := make([]string, len(slice))
	copy(result, slice)

	for _, flag := range flags {
		if !contains(result, flag) {
			result = append(result, flag)
		}
	}
	return result
}

// Helper function to remove flags from a slice
func removeFlags(slice []string, flags []string) []string {
	result := []string{}

	for _, s := range slice {
		if !contains(flags, s) {
			result = append(result, s)
		}
	}
	return result
}

// formatEmailContent converts an Email struct back to a string representation
func formatEmailContent(email *mail.Email) string {
	// Create a simple email format
	var buf strings.Builder

	// Add headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", email.From()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To(), ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject()))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")

	// Add blank line to separate headers from body
	buf.WriteString("\r\n")

	// Add message body
	buf.WriteString(email.Message)

	// Add attachments as base64 encoded content if needed
	// (not implemented in this simple version)

	return buf.String()
}

// parseEmailContent extracts headers and body from an email message
func parseEmailContent(content string) (map[string]string, string) {
	headers := make(map[string]string)

	// Split the content into header and body parts
	parts := strings.SplitN(content, "\r\n\r\n", 2)
	if len(parts) < 2 {
		// Try with just \n\n if \r\n\r\n doesn't work
		parts = strings.SplitN(content, "\n\n", 2)
	}

	headerPart := ""
	bodyPart := ""

	if len(parts) >= 2 {
		headerPart = parts[0]
		bodyPart = parts[1]
	} else {
		// If no clear separation, treat everything as body
		bodyPart = content
	}

	// Parse headers
	headerLines := strings.Split(headerPart, "\n")
	currentHeader := ""
	currentValue := ""

	for _, line := range headerLines {
		// Remove carriage returns
		line = strings.TrimRight(line, "\r")

		// Check if this is a continuation of a previous header
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// Continuation of previous header
			currentValue += " " + strings.TrimSpace(line)
			continue
		}

		// If we have a current header, save it before processing the new one
		if currentHeader != "" {
			headers[currentHeader] = currentValue
			currentHeader = ""
			currentValue = ""
		}

		// Check if this is a new header
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			currentHeader = parts[0]
			currentValue = strings.TrimSpace(parts[1])
		}
	}

	// Save the last header if there is one
	if currentHeader != "" {
		headers[currentHeader] = currentValue
	}

	// Ensure we have the basic headers
	if _, ok := headers["From"]; !ok {
		headers["From"] = "unknown@example.com"
	}
	if _, ok := headers["To"]; !ok {
		headers["To"] = "unknown@example.com"
	}
	if _, ok := headers["Subject"]; !ok {
		headers["Subject"] = "No Subject"
	}

	return headers, bodyPart
}
