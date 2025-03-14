package imapserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sort"
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

	for _, item := range items {
		switch item {
		case imap.StatusMessages:
			status.Messages = uint32(len(m.messages))
		case imap.StatusRecent:
			status.Recent = 0 // No recent messages for simplicity
		case imap.StatusUnseen:
			// Count unseen messages
			unseen := 0
			for _, msg := range m.messages {
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
		From:        headers["From"],
		To:          strings.Split(headers["To"], ","),
		Subject:     headers["Subject"],
		Message:     messageBody,
		Attachments: []mail.Attachment{}, // No attachments for now
	}
	
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
	// Make sure messages are loaded
	if err := m.loadMessages(); err != nil {
		return err
	}

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

		// Update flags
		switch operation {
		case imap.SetFlags:
			msg.Flags = flags
		case imap.AddFlags:
			msg.Flags = addFlags(msg.Flags, flags)
		case imap.RemoveFlags:
			msg.Flags = removeFlags(msg.Flags, flags)
		}
	}

	return nil
}

// CopyMessages copies the specified messages to another mailbox
func (m *Mailbox) CopyMessages(uid bool, seqSet *imap.SeqSet, destName string) error {
	// This is not implemented as we're using Redis as a read-only source
	return fmt.Errorf("copying messages is not supported")
}

// Expunge permanently removes messages marked for deletion
func (m *Mailbox) Expunge() error {
	// This is not implemented as we're using Redis as a read-only source
	return fmt.Errorf("expunging messages is not supported")
}

// loadMessages loads messages from Redis for this mailbox
func (m *Mailbox) loadMessages() error {
	log.Printf("Loading messages for mailbox %s, user %s", m.name, m.user.username)
	
	// Always use lowercase mailbox name for consistency
	lowerName := strings.ToLower(m.name)
	pattern := fmt.Sprintf("mail:in:%s:%s*", m.user.username, lowerName)
	
	if m.backend.debugMode {
		log.Printf("DEBUG: Using Redis pattern with lowercase name: %s", pattern)
	} else {
		log.Printf("Using Redis pattern with lowercase name: %s", pattern)
	}
	
	// Use SCAN instead of KEYS for better performance and reliability
	var cursor uint64
	var keys []string
	var allKeys []string
	var scanErr error
	
	for {
		keys, cursor, scanErr = m.backend.redisClient.Scan(m.backend.ctx, cursor, pattern, 10).Result()
		if scanErr != nil {
			log.Printf("ERROR: Failed to scan keys from Redis: %v", scanErr)
			return scanErr
		}
		
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	
	if m.backend.debugMode {
		log.Printf("DEBUG: Found %d keys with pattern %s using SCAN", len(allKeys), pattern)
		for i, key := range allKeys {
			log.Printf("DEBUG: Key[%d]: %s", i, key)
		}
	} else {
		log.Printf("Found %d keys with pattern %s", len(allKeys), pattern)
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
		
		// Use SCAN for the generic pattern too
		var genericCursor uint64
		var genericKeys []string
		var allUserKeys []string
		
		for {
			genericKeys, genericCursor, scanErr = m.backend.redisClient.Scan(m.backend.ctx, genericCursor, allKeysPattern, 10).Result()
			if scanErr != nil {
				log.Printf("ERROR: Failed to scan all keys from Redis: %v", scanErr)
				return scanErr
			}
			
			allUserKeys = append(allUserKeys, genericKeys...)
			if genericCursor == 0 {
				break
			}
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
	// Extract the epoch part for simplicity
	// In a real implementation, you might want to handle this differently
	if len(uidStr) > 10 {
		uidStr = uidStr[:10] // Take first 10 chars which should be the epoch
	}
	
	// Parse as time and convert to uint32
	timestamp, err := time.Parse(time.RFC3339, uidStr)
	if err == nil {
		return uint32(timestamp.Unix())
	}
	
	// If parsing fails, use a simple hash of the string
	var hash uint32
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
