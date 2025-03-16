package mail

import (
	"fmt"
	"strings"
)

// CalculateSize calculates the total size of the email in bytes
func (e *Email) CalculateSize() uint32 {
	size := uint32(len(e.Message))

	// Add size of attachments
	for _, att := range e.Attachments {
		size += uint32(len(att.Data))
	}

	// Add estimated size of envelope data if available
	if e.Envelope != nil {
		size += uint32(len(e.Envelope.Subject))
		size += uint32(len(e.Envelope.MessageId))
		size += uint32(len(e.Envelope.InReplyTo))

		// Add size of address fields
		for _, addr := range e.Envelope.From {
			size += uint32(len(addr))
		}
		for _, addr := range e.Envelope.To {
			size += uint32(len(addr))
		}
		for _, addr := range e.Envelope.Cc {
			size += uint32(len(addr))
		}
		for _, addr := range e.Envelope.Bcc {
			size += uint32(len(addr))
		}
	}

	return size
}

// GetBodyStructure generates and returns a description of the MIME structure of the email
// This can be used by IMAP clients to understand the structure of the message
// should return a string in the format of an IMAP body structure
//
//	type BodyStructure struct {
//		MIMEType    string          `json:"mime_type,omitempty"`    // MIME Type (e.g., text/plain, multipart/mixed)
//		Encoding    string          `json:"encoding,omitempty"`    // Encoding method (e.g., quoted-printable, base64)
//		Size        uint32          `json:"size,omitempty"`        // Size of the message body
//		Parts       []*BodyStructure `json:"parts,omitempty"`      // Parts for multipart messages
//		Disposition string          `json:"disposition,omitempty"` // Content disposition (inline, attachment)
//	}
func (e *Email) GetBodyStructure() string {
	// If there are no attachments, return a simple text structure
	if len(e.Attachments) == 0 {
		return "(\"text\" \"plain\" (\"charset\" \"utf-8\") NIL NIL \"7bit\" " +
			fmt.Sprintf("%d %d", len(e.Message), countLines(e.Message)) + " NIL NIL NIL)"
	}

	// For emails with attachments, create a multipart/mixed structure
	result := "(\"multipart\" \"mixed\" NIL NIL NIL \"7bit\" NIL NIL ("

	// Add the text part
	result += "(\"text\" \"plain\" (\"charset\" \"utf-8\") NIL NIL \"7bit\" " +
		fmt.Sprintf("%d %d", len(e.Message), countLines(e.Message)) + " NIL NIL NIL)"

	// Add each attachment
	for _, att := range e.Attachments {
		// Default to application/octet-stream if content type is empty
		contentType := att.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Split content type into type and subtype
		parts := strings.SplitN(contentType, "/", 2)
		if len(parts) != 2 {
			parts = []string{"application", "octet-stream"}
		}

		// Add the attachment part
		result += fmt.Sprintf(" (\"application\" \"%s\" (\"name\" \"%s\") NIL NIL \"base64\" %d NIL (\"attachment\" (\"filename\" \"%s\")) NIL)",
			parts[1], att.Filename, len(att.Data), att.Filename)
	}

	// Close the structure
	result += ")"

	return result
}

// countLines counts the number of lines in a string
func countLines(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

// Helper methods to access fields from the Envelope

// From returns the From address from the Envelope
func (e *Email) From() string {
	if e.Envelope == nil || len(e.Envelope.From) == 0 {
		return ""
	}
	return e.Envelope.From[0]
}

// To returns the To addresses from the Envelope
func (e *Email) To() []string {
	if e.Envelope == nil {
		return nil
	}
	return e.Envelope.To
}

// Cc returns the Cc addresses from the Envelope
func (e *Email) Cc() []string {
	if e.Envelope == nil {
		return nil
	}
	return e.Envelope.Cc
}

// Bcc returns the Bcc addresses from the Envelope
func (e *Email) Bcc() []string {
	if e.Envelope == nil {
		return nil
	}
	return e.Envelope.Bcc
}

// Subject returns the Subject from the Envelope
func (e *Email) Subject() string {
	if e.Envelope == nil {
		return ""
	}
	return e.Envelope.Subject
}

// Date returns the Date from the Envelope
func (e *Email) Date() int64 {
	if e.Envelope == nil {
		return 0
	}
	return e.Envelope.Date
}

// EnsureEnvelope ensures that the email has an envelope, creating one if needed
func (e *Email) EnsureEnvelope() {
	if e.Envelope == nil {
		e.Envelope = &Envelope{}
	}
}

// SetFrom sets the From address in the Envelope
func (e *Email) SetFrom(from string) {
	e.EnsureEnvelope()
	e.Envelope.From = []string{from}
}

// SetTo sets the To addresses in the Envelope
func (e *Email) SetTo(to []string) {
	e.EnsureEnvelope()
	e.Envelope.To = to
}

// SetCc sets the Cc addresses in the Envelope
func (e *Email) SetCc(cc []string) {
	e.EnsureEnvelope()
	e.Envelope.Cc = cc
}

// SetBcc sets the Bcc addresses in the Envelope
func (e *Email) SetBcc(bcc []string) {
	e.EnsureEnvelope()
	e.Envelope.Bcc = bcc
}

// SetSubject sets the Subject in the Envelope
func (e *Email) SetSubject(subject string) {
	e.EnsureEnvelope()
	e.Envelope.Subject = subject
}

// SetDate sets the Date in the Envelope
func (e *Email) SetDate(date int64) {
	e.EnsureEnvelope()
	e.Envelope.Date = date
}
