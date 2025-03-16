package main

import (
	"flag"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func main() {
	// Parse command-line flags
	imapAddr := flag.String("imap-addr", "localhost:2143", "IMAP server address")
	username := flag.String("username", "jan", "IMAP username")
	password := flag.String("password", "password", "IMAP password")
	flag.Parse()

	// Connect to the IMAP server
	log.Printf("Connecting to IMAP server at %s...", *imapAddr)
	c, err := client.Dial(*imapAddr)
	if err != nil {
		log.Fatalf("Failed to connect to IMAP server: %v", err)
	}
	defer c.Logout()
	log.Printf("Connected to IMAP server")

	// Login
	if err := c.Login(*username, *password); err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	log.Printf("Logged in as %s", *username)

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Printf("Mailboxes:")
	for m := range mailboxes {
		log.Printf("* %s", m.Name)
	}

	if err := <-done; err != nil {
		log.Fatalf("Failed to list mailboxes: %v", err)
	}

	// Select inbox
	mbox, err := c.Select("inbox", false)
	if err != nil {
		log.Fatalf("Failed to select inbox: %v", err)
	}
	log.Printf("Inbox selected, %d messages", mbox.Messages)

	// Fetch messages
	if mbox.Messages > 0 {
		seqSet := new(imap.SeqSet)
		seqSet.AddRange(1, mbox.Messages)

		messages := make(chan *imap.Message, 10)
		done = make(chan error, 1)
		go func() {
			done <- c.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope}, messages)
		}()

		log.Printf("Messages in inbox:")
		for msg := range messages {
			log.Printf("* %d: %s", msg.SeqNum, msg.Envelope.Subject)
		}

		if err := <-done; err != nil {
			log.Fatalf("Failed to fetch messages: %v", err)
		}

		// Test moving a message to another folder
		if mbox.Messages > 0 {
			log.Printf("Testing move operation...")
			moveSeqSet := new(imap.SeqSet)
			moveSeqSet.AddNum(1) // Move the first message

			// Try to move the message to the "archive" folder
			err = c.Move(moveSeqSet, "archive")
			if err != nil {
				log.Printf("Move operation failed: %v", err)
			} else {
				log.Printf("Successfully moved message to archive folder")
			}

			// Select inbox again to see if the message was moved
			mbox, err = c.Select("inbox", false)
			if err != nil {
				log.Fatalf("Failed to select inbox: %v", err)
			}
			log.Printf("Inbox selected again, now has %d messages", mbox.Messages)

			// Select archive folder to see if the message was moved there
			mbox, err = c.Select("archive", false)
			if err != nil {
				log.Printf("Failed to select archive folder: %v", err)
			} else {
				log.Printf("Archive folder selected, has %d messages", mbox.Messages)
			}
		}

		// Test expunge operation
		log.Printf("Testing expunge operation...")
		
		// Select inbox again
		mbox, err = c.Select("inbox", false)
		if err != nil {
			log.Fatalf("Failed to select inbox: %v", err)
		}
		
		if mbox.Messages > 0 {
			// Mark a message for deletion
			delSeqSet := new(imap.SeqSet)
			delSeqSet.AddNum(1) // Mark the first message for deletion
			
			storeItems := imap.FormatFlagsOp(imap.AddFlags, true)
			err = c.Store(delSeqSet, storeItems, []interface{}{imap.DeletedFlag}, nil)
			if err != nil {
				log.Printf("Failed to mark message for deletion: %v", err)
			} else {
				log.Printf("Successfully marked message for deletion")
				
				// Expunge the mailbox
				err = c.Expunge(nil)
				if err != nil {
					log.Printf("Expunge operation failed: %v", err)
				} else {
					log.Printf("Successfully expunged mailbox")
					
					// Select inbox again to see if the message was deleted
					mbox, err = c.Select("inbox", false)
					if err != nil {
						log.Fatalf("Failed to select inbox: %v", err)
					}
					log.Printf("Inbox selected after expunge, now has %d messages", mbox.Messages)
				}
			}
		} else {
			log.Printf("No messages in inbox to test expunge operation")
		}
	} else {
		log.Printf("No messages in inbox")
	}

	// Logout
	if err := c.Logout(); err != nil {
		log.Printf("Failed to logout: %v", err)
	}
	log.Printf("Logged out")
}
