package imapserver

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/freeflowuniverse/herolauncher/pkg/mail"
)

// Message represents an email message
type Message struct {
	Email *mail.Email
	Uid   uint32
	Flags []string
	Key   string // Redis key where this message is stored
}

// Fetch converts a Message to an imap.Message
func (m *Message) Fetch(seqNum uint32, items []imap.FetchItem) (*imap.Message, error) {
	msg := imap.NewMessage(seqNum, items)
	msg.Uid = m.Uid
	msg.Flags = m.Flags

	for _, item := range items {
		switch item {
		case imap.FetchEnvelope:
			msg.Envelope = m.createEnvelope()
		case imap.FetchBody, imap.FetchBodyStructure:
			msg.BodyStructure = m.createBodyStructure(item == imap.FetchBodyStructure)
		case imap.FetchFlags:
			// Flags already set above
		case imap.FetchInternalDate:
			// Use InternalDate from the Email struct if available, otherwise use current time
			if m.Email.InternalDate > 0 {
				msg.InternalDate = time.Unix(m.Email.InternalDate, 0)
			} else {
				msg.InternalDate = time.Now()
			}
		case imap.FetchRFC822Size:
			// Use Size from the Email struct if available, otherwise calculate it
			if m.Email.Size > 0 {
				msg.Size = m.Email.Size
			} else {
				// Calculate size using the CalculateSize method
				msg.Size = m.Email.CalculateSize()
			}
		default:
			// Handle section fetch (BODY[...], BODY.PEEK[...])
			if section, err := imap.ParseBodySectionName(item); err == nil {
				msg.Body[section] = m.getBodySection(section)
			}
		}
	}

	return msg, nil
}

// Match checks if the message matches the search criteria
func (m *Message) Match(seqNum uint32, criteria *imap.SearchCriteria) bool {
	// Check sequence number criteria
	if criteria.SeqNum != nil && !criteria.SeqNum.Contains(seqNum) {
		return false
	}

	// Check UID criteria
	if criteria.Uid != nil && !criteria.Uid.Contains(m.Uid) {
		return false
	}

	// Check flags criteria
	for _, f := range criteria.WithFlags {
		if !contains(m.Flags, f) {
			return false
		}
	}
	for _, f := range criteria.WithoutFlags {
		if contains(m.Flags, f) {
			return false
		}
	}

	// Check header criteria
	if criteria.Header != nil {
		for field, values := range criteria.Header {
			switch strings.ToLower(field) {
			case "from":
				if !matchAny(m.Email.From(), values) {
					return false
				}
			case "to":
				if !matchAnyInSlice(m.Email.To(), values) {
					return false
				}
			case "subject":
				if !matchAny(m.Email.Subject(), values) {
					return false
				}
			}
		}
	}

	// Check body criteria
	if len(criteria.Body) > 0 {
		matched := false
		for _, value := range criteria.Body {
			if strings.Contains(strings.ToLower(m.Email.Message), strings.ToLower(value)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check text criteria (header + body)
	if len(criteria.Text) > 0 {
		allText := strings.ToLower(fmt.Sprintf("%s %s %s %s",
			m.Email.From,
			strings.Join(m.Email.To, " "),
			m.Email.Subject,
			m.Email.Message,
		))
		
		matched := false
		for _, value := range criteria.Text {
			if strings.Contains(allText, strings.ToLower(value)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If we made it here, the message matches all criteria
	return true
}

// createEnvelope creates an IMAP envelope from the email
func (m *Message) createEnvelope() *imap.Envelope {
	// If the email already has an envelope defined in the mail.Email struct, use that
	if m.Email.Envelope != nil {
		// Convert mail.Envelope to imap.Envelope
		env := &imap.Envelope{
			Date:    time.Unix(m.Email.Envelope.Date, 0),
			Subject: m.Email.Envelope.Subject,
			From:    make([]*imap.Address, 0, len(m.Email.Envelope.From)),
			Sender:  make([]*imap.Address, 0, len(m.Email.Envelope.Sender)),
			To:      make([]*imap.Address, 0, len(m.Email.Envelope.To)),
			Cc:      make([]*imap.Address, 0, len(m.Email.Envelope.Cc)),
			Bcc:     make([]*imap.Address, 0, len(m.Email.Envelope.Bcc)),
		}
		
		// Convert addresses
		for _, from := range m.Email.Envelope.From {
			env.From = append(env.From, parseAddress(from))
		}
		
		for _, sender := range m.Email.Envelope.Sender {
			env.Sender = append(env.Sender, parseAddress(sender))
		}
		
		for _, to := range m.Email.Envelope.To {
			env.To = append(env.To, parseAddress(to))
		}
		
		for _, cc := range m.Email.Envelope.Cc {
			env.Cc = append(env.Cc, parseAddress(cc))
		}
		
		for _, bcc := range m.Email.Envelope.Bcc {
			env.Bcc = append(env.Bcc, parseAddress(bcc))
		}
		
		// Set message IDs
		env.MessageId = m.Email.Envelope.MessageId
		env.InReplyTo = m.Email.Envelope.InReplyTo
		
		return env
	}
	
	// Otherwise, create a new envelope from the basic email fields using accessor methods
	env := &imap.Envelope{
		Date:    time.Now(), // Use current time for simplicity
		Subject: m.Email.Subject(),
		From:    []*imap.Address{parseAddress(m.Email.From())},
		Sender:  []*imap.Address{parseAddress(m.Email.From())},
		To:      make([]*imap.Address, 0, len(m.Email.To())),
	}

	// Add recipients
	for _, to := range m.Email.To() {
		env.To = append(env.To, parseAddress(to))
	}
	
	// Add CC recipients if available
	cc := m.Email.Cc()
	if len(cc) > 0 {
		env.Cc = make([]*imap.Address, 0, len(cc))
		for _, ccAddr := range cc {
			env.Cc = append(env.Cc, parseAddress(ccAddr))
		}
	}
	
	// Add BCC recipients if available
	bcc := m.Email.Bcc()
	if len(bcc) > 0 {
		env.Bcc = make([]*imap.Address, 0, len(bcc))
		for _, bccAddr := range bcc {
			env.Bcc = append(env.Bcc, parseAddress(bccAddr))
		}
	}

	return env
}

// createBodyStructure creates an IMAP body structure
func (m *Message) createBodyStructure(extended bool) *imap.BodyStructure {
	// Generate a body structure using the GetBodyStructure method
	// and parse it into an imap.BodyStructure
	// For now, we'll continue with the existing implementation
	
	// Create a simple text body structure
	bs := &imap.BodyStructure{
		MIMEType:    "text",
		MIMESubType: "plain",
		Params:      map[string]string{"charset": "utf-8"},
		Size:        uint32(len(m.Email.Message)),
		Extended:    extended,
	}

	// If there are attachments, create a multipart structure
	if len(m.Email.Attachments) > 0 {
		bs = &imap.BodyStructure{
			MIMEType:    "multipart",
			MIMESubType: "mixed",
			Params:      map[string]string{},
			Parts: []*imap.BodyStructure{
				// Text part
				{
					MIMEType:    "text",
					MIMESubType: "plain",
					Params:      map[string]string{"charset": "utf-8"},
					Size:        uint32(len(m.Email.Message)),
					Extended:    extended,
				},
			},
			Extended: extended,
		}

		// Add attachment parts
		for _, att := range m.Email.Attachments {
			attPart := &imap.BodyStructure{
				MIMEType:    strings.Split(att.ContentType, "/")[0],
				MIMESubType: strings.Split(att.ContentType, "/")[1],
				Params: map[string]string{
					"name": att.Filename,
				},
				Disposition:      "attachment",
				DispositionParams: map[string]string{"filename": att.Filename},
				Size:             uint32(len(att.Data)),
				Extended:         extended,
			}
			bs.Parts = append(bs.Parts, attPart)
		}
	}

	return bs
}

// getBodySection returns a specific section of the message body
func (m *Message) getBodySection(section *imap.BodySectionName) imap.Literal {
	// Create a buffer to store the message
	var buf bytes.Buffer

	// If no section specified or section is HEADER, return the header
	if section.Specifier == imap.HeaderSpecifier || section.Specifier == imap.EntireSpecifier {
		header := createHeader(m.Email)
		if section.Specifier == imap.HeaderSpecifier {
			buf.Write(header)
			return bytes.NewReader(buf.Bytes())
		}
		
		// For entire message, write header followed by body
		buf.Write(header)
		buf.WriteString("\r\n")
	}

	// If section is TEXT or ENTIRE, include the message body
	if section.Specifier == imap.TextSpecifier || section.Specifier == imap.EntireSpecifier {
		buf.WriteString(m.Email.Message)
		
		// Add attachments if any
		if len(m.Email.Attachments) > 0 {
			for _, att := range m.Email.Attachments {
				buf.WriteString(fmt.Sprintf("\r\n\r\n--Attachment: %s--\r\n", att.Filename))
			}
		}
	}

	return bytes.NewReader(buf.Bytes())
}

// createHeader creates an RFC822 header from the email
func createHeader(email *mail.Email) []byte {
	var buf bytes.Buffer
	
	// Add From header
	buf.WriteString(fmt.Sprintf("From: %s\r\n", email.From()))
	
	// Add To header
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To(), ", ")))
	
	// Add Subject header
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject()))
	
	// Add Date header
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	
	// Add Message-ID header
	buf.WriteString(fmt.Sprintf("Message-ID: <%d@example.com>\r\n", time.Now().UnixNano()))
	
	// Add MIME headers
	if len(email.Attachments) > 0 {
		buf.WriteString("MIME-Version: 1.0\r\n")
		buf.WriteString("Content-Type: multipart/mixed; boundary=attachment-boundary\r\n")
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	}
	
	return buf.Bytes()
}

// parseAddress parses an email address into an IMAP address
func parseAddress(addr string) *imap.Address {
	// Simple parsing for "name <email>" format
	name := ""
	email := addr
	
	if strings.Contains(addr, "<") && strings.Contains(addr, ">") {
		parts := strings.Split(addr, "<")
		name = strings.TrimSpace(parts[0])
		email = strings.TrimSuffix(parts[1], ">")
	}
	
	// Split email into mailbox and host parts
	mailbox := email
	host := ""
	if strings.Contains(email, "@") {
		parts := strings.Split(email, "@")
		mailbox = parts[0]
		host = parts[1]
	}
	
	return &imap.Address{
		PersonalName: name,
		MailboxName:  mailbox,
		HostName:     host,
	}
}

// matchAny checks if a string contains any of the given values
func matchAny(s string, values []string) bool {
	s = strings.ToLower(s)
	for _, value := range values {
		if strings.Contains(s, strings.ToLower(value)) {
			return true
		}
	}
	return false
}

// matchAnyInSlice checks if any string in the slice contains any of the given values
func matchAnyInSlice(slice []string, values []string) bool {
	for _, s := range slice {
		if matchAny(s, values) {
			return true
		}
	}
	return false
}
