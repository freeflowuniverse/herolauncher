package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func main() {
	// Parse command line flags
	imapAddr := flag.String("imap-addr", "localhost:1143", "IMAP server address")
	username := flag.String("username", "jan", "IMAP username (any value works)")
	password := flag.String("password", "testpass", "IMAP password (any value works)")
	flag.Parse()

	log.Println("Connecting to IMAP server at", *imapAddr)

	// Connect to the IMAP server
	c, err := client.Dial(*imapAddr)
	if err != nil {
		log.Fatal("Failed to connect to IMAP server:", err)
	}
	log.Println("Connected to IMAP server")

	// Don't forget to logout
	defer func() {
		if c != nil {
			c.Logout()
			log.Println("Logged out from IMAP server")
		}
	}()

	// Login (the server accepts any credentials)
	if err := c.Login(*username, *password); err != nil {
		log.Fatal("Failed to login:", err)
	}
	log.Println("Logged in as", *username)

	// List mailboxes
	listMailboxes(c)

	// Get the list of available mailboxes
	mailboxNames := listMailboxesAndGetNames(c)
	
	// Try to select each mailbox and find one with messages
	var mbox *imap.MailboxStatus
	var selectedMailbox string
	
	log.Println("Searching for mailboxes with messages...")
	for _, mailboxName := range mailboxNames {
		log.Println("Trying to select mailbox:", mailboxName)
		mbox, err = c.Select(mailboxName, false)
		if err == nil {
			log.Printf("Selected mailbox: %s with %d messages\n", mbox.Name, mbox.Messages)
			selectedMailbox = mailboxName
			
			// If this mailbox has messages, use it
			if mbox.Messages > 0 {
				log.Printf("Found mailbox with messages: %s (%d messages)\n", mbox.Name, mbox.Messages)
				break
			}
		} else {
			log.Printf("Failed to select mailbox %s: %v\n", mailboxName, err)
		}
	}
	
	// If no mailbox with messages was found, use the last successfully selected mailbox
	if mbox == nil || (mbox.Messages == 0 && selectedMailbox == "") {
		log.Fatal("Failed to select any mailbox with messages")
	}
	
	// If the selected mailbox has no messages, inform the user
	if mbox.Messages == 0 {
		log.Printf("No messages found in any mailbox, using mailbox: %s\n", selectedMailbox)
	}

	// If there are still no messages, exit
	if mbox.Messages == 0 {
		log.Println("No messages found in any mailbox")
		return
	}

	// Get the last 5 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 5 {
		from = mbox.Messages - 4
	}
	log.Printf("Fetching messages %d to %d\n", from, to)

	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchFlags, imap.FetchUid}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, items, messages)
	}()

	// Process fetched messages
	var lastUID uint32
	for msg := range messages {
		log.Printf("* Message %d (UID %d):\n", msg.SeqNum, msg.Uid)
		log.Printf("  Subject: %s\n", msg.Envelope.Subject)
		log.Printf("  From: %v\n", formatAddresses(msg.Envelope.From))
		log.Printf("  To: %v\n", formatAddresses(msg.Envelope.To))
		log.Printf("  Flags: %v\n", msg.Flags)

		// Process message body
		r := msg.GetBody(section)
		if r == nil {
			log.Println("  No message body")
			continue
		}

		// Parse message body
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Println("  Error parsing message:", err)
			continue
		}

		// Print message body
		header := mr.Header
		if date, err := header.Date(); err == nil {
			log.Printf("  Date: %s\n", date.Format("2006-01-02 15:04:05"))
		}

		// Just try to read the first part for a preview
		p, err := mr.NextPart()
		if err == nil {
			// Read the body
			body, _ := io.ReadAll(p.Body)
			if len(body) > 200 {
				body = body[:200]
			}
			log.Printf("  Body preview: %s...\n", string(body))
		}

		// Save the last UID for deletion test
		lastUID = msg.Uid
	}

	if err := <-done; err != nil {
		log.Fatal("Error fetching messages:", err)
	}

	// Test deleting a message if we have a UID
	if lastUID > 0 {
		log.Printf("Testing message deletion for UID %d\n", lastUID)

		// Mark the message as deleted
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(lastUID)

		// First, mark the message with the \Deleted flag
		flagsToAdd := []interface{}{imap.DeletedFlag}
		err = c.UidStore(seqSet, imap.AddFlags, flagsToAdd, nil)
		if err != nil {
			log.Println("Error marking message as deleted:", err)
		} else {
			log.Println("Marked message as deleted")

			// Then, expunge the mailbox to remove messages marked as deleted
			err = c.Expunge(nil)
			if err != nil {
				log.Println("Error expunging mailbox:", err)
			} else {
				log.Println("Expunged mailbox, deleted messages removed")
			}
		}

		// Verify the message is gone by trying to fetch it
		messages = make(chan *imap.Message, 1)
		done = make(chan error, 1)
		go func() {
			done <- c.UidFetch(seqSet, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		// Check if any messages were returned
		messageCount := 0
		for range messages {
			messageCount++
		}

		if err := <-done; err != nil {
			log.Println("Error fetching deleted message:", err)
		}

		if messageCount == 0 {
			log.Println("Message successfully deleted")
		} else {
			log.Println("Message still exists after deletion attempt")
		}
	}

	// Test setting a flag on a message
	if mbox.Messages > 0 {
		// Get the first message
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(1)

		// Mark the message as seen
		flagsToAdd := []interface{}{imap.SeenFlag}
		err = c.Store(seqSet, imap.AddFlags, flagsToAdd, nil)
		if err != nil {
			log.Println("Error marking message as seen:", err)
		} else {
			log.Println("Marked message as seen")
		}
	}

	log.Println("IMAP client operations completed successfully")
}

// listMailboxes lists all available mailboxes
func listMailboxes(c *client.Client) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Available mailboxes:")
	for m := range mailboxes {
		log.Printf("- %s\n", m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal("Error listing mailboxes:", err)
	}
}

// listMailboxesAndGetNames lists all available mailboxes and returns their names as a slice
func listMailboxesAndGetNames(c *client.Client) []string {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var mailboxNames []string
	log.Println("Available mailboxes:")
	for m := range mailboxes {
		log.Printf("- %s\n", m.Name)
		mailboxNames = append(mailboxNames, m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal("Error listing mailboxes:", err)
	}
	
	return mailboxNames
}

// formatAddresses formats a list of addresses
func formatAddresses(addrs []*imap.Address) string {
	if len(addrs) == 0 {
		return ""
	}

	var formatted []string
	for _, addr := range addrs {
		if addr.PersonalName != "" {
			formatted = append(formatted, fmt.Sprintf("%s <%s@%s>", addr.PersonalName, addr.MailboxName, addr.HostName))
		} else {
			formatted = append(formatted, fmt.Sprintf("%s@%s", addr.MailboxName, addr.HostName))
		}
	}

	return strings.Join(formatted, ", ")
}
