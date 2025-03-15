package mail

type Email struct {
	// Content fields
	Message     string       `json:"message"`     // The email body content
	Attachments []Attachment `json:"attachments"` // Any file attachments

	// IMAP specific fields
	Flags        []string  `json:"flags,omitempty"`         // IMAP flags like \Seen, \Deleted, etc.
	InternalDate int64     `json:"internal_date,omitempty"` // Unix timestamp when the email was received
	Size         uint32    `json:"size,omitempty"`          // Size of the message in bytes
	Envelope     *Envelope `json:"envelope,omitempty"`      // IMAP envelope information (contains From, To, Subject, etc.)
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Data        string `json:"data"` // Base64 encoded binary data
}

// Envelope represents an IMAP envelope structure
type Envelope struct {
	Date      int64    `json:"date,omitempty"`        // Date the message was sent
	Subject   string   `json:"subject,omitempty"`     // Subject of the message
	From      []string `json:"from,omitempty"`        // Sender addresses
	Sender    []string `json:"sender,omitempty"`      // Actual sender addresses
	ReplyTo   []string `json:"reply_to,omitempty"`    // Reply-To addresses
	To        []string `json:"to,omitempty"`          // Recipient addresses
	Cc        []string `json:"cc,omitempty"`          // CC addresses
	Bcc       []string `json:"bcc,omitempty"`         // BCC addresses
	InReplyTo string   `json:"in_reply_to,omitempty"` // Message-ID of the message being replied to
	MessageId string   `json:"message_id,omitempty"`  // Message-ID
}
