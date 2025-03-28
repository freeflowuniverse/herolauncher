// pkg/mycelium_client/examples/basic_usage.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/mycelium_client"
)

func main() {
	// Create a new client with default configuration (localhost:8989)
	client := mycelium_client.NewClient("")

	// Set a custom timeout if needed
	client.SetTimeout(60 * time.Second)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example 1: Get node info
	fmt.Println("Getting node info...")
	info, err := client.GetNodeInfo(ctx)
	if err != nil {
		log.Printf("Failed to get node info: %v", err)
	} else {
		fmt.Printf("Node subnet: %s\n", info.NodeSubnet)
	}

	// Example 2: List peers
	fmt.Println("\nListing peers...")
	peers, err := client.ListPeers(ctx)
	if err != nil {
		log.Printf("Failed to list peers: %v", err)
	} else {
		fmt.Printf("Found %d peers:\n", len(peers))
		for i, peer := range peers {
			fmt.Printf("  %d. %s://%s (%s)\n",
				i+1,
				peer.Endpoint.Proto,
				peer.Endpoint.SocketAddr,
				peer.ConnectionState)
		}
	}

	// Example 3: Send a message (if there are peers)
	if len(os.Args) > 1 && os.Args[1] == "send" {
		fmt.Println("\nSending a message...")

		// In a real application, you would get this from the peer
		// This is just a placeholder public key
		dest := mycelium_client.MessageDestination{
			PK: "bb39b4a3a4efd70f3e05e37887677e02efbda14681d0acd3882bc0f754792c32",
		}

		payload := []byte("Hello from mycelium client!")
		topic := "exampletopic"

		// Send without waiting for reply
		_, msgID, err := client.SendMessage(ctx, dest, payload, topic, false, 0)
		if err != nil {
			log.Printf("Failed to send message: %v", err)
		} else {
			fmt.Printf("Message sent with ID: %s\n", msgID)
		}
	}

	// Example 4: Receive a message (with a short timeout)
	if len(os.Args) > 1 && os.Args[1] == "receive" {
		fmt.Println("\nWaiting for a message (5 seconds)...")
		receiveCtx, receiveCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer receiveCancel()

		msg, err := client.ReceiveMessage(receiveCtx, 5, "", false)
		if err != nil {
			log.Printf("Error receiving message: %v", err)
		} else if msg == nil {
			fmt.Println("No message received within timeout")
		} else {
			payload, err := msg.Decode()
			if err != nil {
				log.Printf("Failed to decode message payload: %v", err)
			} else {
				fmt.Printf("Received message (ID: %s):\n", msg.ID)
				fmt.Printf("  From: %s\n", msg.SrcPK)
				fmt.Printf("  Topic: %s\n", msg.Topic)
				fmt.Printf("  Payload: %s\n", string(payload))
			}
		}
	}
}
