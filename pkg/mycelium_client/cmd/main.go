// pkg/mycelium_client/cmd/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/mycelium_client"
)

type config struct {
	baseURL      string
	command      string
	peerEndpoint string
	message      string
	destination  string
	topic        string
	timeout      int
	wait         bool
	replyTimeout int
	messageID    string
	outputJSON   bool
}

// Commands
const (
	cmdInfo    = "info"
	cmdPeers   = "peers"
	cmdAddPeer = "add-peer"
	cmdDelPeer = "del-peer"
	cmdSend    = "send"
	cmdReceive = "receive"
	cmdReply   = "reply"
	cmdStatus  = "status"
	cmdRoutes  = "routes"
)

func main() {
	// Create config with default values
	cfg := config{
		baseURL:      fmt.Sprintf("http://localhost:%d", mycelium_client.DefaultAPIPort),
		timeout:      30,
		replyTimeout: mycelium_client.DefaultReplyTimeout,
	}

	// Parse command line flags
	flag.StringVar(&cfg.baseURL, "api", cfg.baseURL, "Mycelium API URL")
	flag.IntVar(&cfg.timeout, "timeout", cfg.timeout, "Client timeout in seconds")
	flag.BoolVar(&cfg.outputJSON, "json", false, "Output in JSON format")
	flag.Parse()

	// Get the command
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cfg.command = args[0]
	args = args[1:]

	// Create client
	client := mycelium_client.NewClient(cfg.baseURL)
	client.SetTimeout(time.Duration(cfg.timeout) * time.Second)

	// Create context with cancellation for graceful shutdowns
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Execute command
	var err error
	switch cfg.command {
	case cmdInfo:
		err = showNodeInfo(ctx, client)

	case cmdPeers:
		err = listPeers(ctx, client, cfg.outputJSON)

	case cmdAddPeer:
		if len(args) < 1 {
			fmt.Println("Missing peer endpoint argument")
			printUsage()
			os.Exit(1)
		}
		cfg.peerEndpoint = args[0]
		err = addPeer(ctx, client, cfg.peerEndpoint)

	case cmdDelPeer:
		if len(args) < 1 {
			fmt.Println("Missing peer endpoint argument")
			printUsage()
			os.Exit(1)
		}
		cfg.peerEndpoint = args[0]
		err = removePeer(ctx, client, cfg.peerEndpoint)

	case cmdSend:
		parseMessageArgs(&cfg, args)
		err = sendMessage(ctx, client, cfg)

	case cmdReceive:
		parseReceiveArgs(&cfg, args)
		err = receiveMessage(ctx, client, cfg)

	case cmdReply:
		parseReplyArgs(&cfg, args)
		err = replyToMessage(ctx, client, cfg)

	case cmdStatus:
		if len(args) < 1 {
			fmt.Println("Missing message ID argument")
			printUsage()
			os.Exit(1)
		}
		cfg.messageID = args[0]
		err = getMessageStatus(ctx, client, cfg.messageID)

	case cmdRoutes:
		var routeType string
		if len(args) > 0 {
			routeType = args[0]
		}
		err = listRoutes(ctx, client, routeType, cfg.outputJSON)

	default:
		fmt.Printf("Unknown command: %s\n", cfg.command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: mycelium-client [flags] COMMAND [args...]")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
	fmt.Println("\nCommands:")
	fmt.Println("  info                        Get node information")
	fmt.Println("  peers                       List connected peers")
	fmt.Println("  add-peer ENDPOINT           Add a new peer")
	fmt.Println("  del-peer ENDPOINT           Remove a peer")
	fmt.Println("  send [--pk=PK|--ip=IP] [--topic=TOPIC] [--wait] [--reply-timeout=N] MESSAGE")
	fmt.Println("                              Send a message to a destination")
	fmt.Println("  receive [--topic=TOPIC] [--timeout=N]")
	fmt.Println("                              Receive a message")
	fmt.Println("  reply ID [--topic=TOPIC] MESSAGE")
	fmt.Println("                              Reply to a message")
	fmt.Println("  status ID                   Get status of a sent message")
	fmt.Println("  routes [selected|fallback]  List routes (default: selected)")
}

func parseMessageArgs(cfg *config, args []string) {
	// Create a temporary flag set
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	fs.StringVar(&cfg.destination, "pk", "", "Destination public key (hex encoded)")
	fs.StringVar(&cfg.destination, "ip", "", "Destination IP address")
	fs.StringVar(&cfg.topic, "topic", "", "Message topic")
	fs.BoolVar(&cfg.wait, "wait", false, "Wait for reply")
	fs.IntVar(&cfg.replyTimeout, "reply-timeout", cfg.replyTimeout, "Reply timeout in seconds")

	// Parse args
	fs.Parse(args)

	// Remaining args are the message
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		fmt.Println("Missing message content")
		printUsage()
		os.Exit(1)
	}
	cfg.message = strings.Join(remainingArgs, " ")
}

func parseReceiveArgs(cfg *config, args []string) {
	// Create a temporary flag set
	fs := flag.NewFlagSet("receive", flag.ExitOnError)
	fs.StringVar(&cfg.topic, "topic", "", "Message topic filter")
	fs.IntVar(&cfg.timeout, "timeout", 10, "Receive timeout in seconds")

	// Parse args
	fs.Parse(args)
}

func parseReplyArgs(cfg *config, args []string) {
	if len(args) < 1 {
		fmt.Println("Missing message ID argument")
		printUsage()
		os.Exit(1)
	}

	cfg.messageID = args[0]
	args = args[1:]

	// Create a temporary flag set
	fs := flag.NewFlagSet("reply", flag.ExitOnError)
	fs.StringVar(&cfg.topic, "topic", "", "Message topic")

	// Parse args
	fs.Parse(args)

	// Remaining args are the message
	remainingArgs := fs.Args()
	if len(remainingArgs) == 0 {
		fmt.Println("Missing reply message content")
		printUsage()
		os.Exit(1)
	}
	cfg.message = strings.Join(remainingArgs, " ")
}

func showNodeInfo(ctx context.Context, client *mycelium_client.MyceliumClient) error {
	info, err := client.GetNodeInfo(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Node Information:")
	fmt.Printf("  Subnet: %s\n", info.NodeSubnet)
	return nil
}

func listPeers(ctx context.Context, client *mycelium_client.MyceliumClient, jsonOutput bool) error {
	peers, err := client.ListPeers(ctx)
	if err != nil {
		return err
	}

	if jsonOutput {
		// TODO: Output JSON
		fmt.Printf("Found %d peers\n", len(peers))
	} else {
		fmt.Printf("Connected Peers (%d):\n", len(peers))
		if len(peers) == 0 {
			fmt.Println("  No peers connected")
			return nil
		}

		for i, peer := range peers {
			fmt.Printf("  %d. %s://%s\n", i+1, peer.Endpoint.Proto, peer.Endpoint.SocketAddr)
			fmt.Printf("     Type: %s, State: %s\n", peer.Type, peer.ConnectionState)
			if peer.TxBytes > 0 || peer.RxBytes > 0 {
				fmt.Printf("     TX: %d bytes, RX: %d bytes\n", peer.TxBytes, peer.RxBytes)
			}
		}
	}
	return nil
}

func addPeer(ctx context.Context, client *mycelium_client.MyceliumClient, endpoint string) error {
	if err := client.AddPeer(ctx, endpoint); err != nil {
		return err
	}
	fmt.Printf("Peer added: %s\n", endpoint)
	return nil
}

func removePeer(ctx context.Context, client *mycelium_client.MyceliumClient, endpoint string) error {
	if err := client.RemovePeer(ctx, endpoint); err != nil {
		return err
	}
	fmt.Printf("Peer removed: %s\n", endpoint)
	return nil
}

func sendMessage(ctx context.Context, client *mycelium_client.MyceliumClient, cfg config) error {
	var dst mycelium_client.MessageDestination

	if cfg.destination == "" {
		return fmt.Errorf("destination is required (--pk or --ip)")
	}

	// Determine destination type
	if strings.HasPrefix(cfg.destination, "--pk=") {
		dst.PK = strings.TrimPrefix(cfg.destination, "--pk=")
	} else if strings.HasPrefix(cfg.destination, "--ip=") {
		dst.IP = strings.TrimPrefix(cfg.destination, "--ip=")
	} else {
		// Try to guess format
		if strings.Contains(cfg.destination, ":") {
			dst.IP = cfg.destination
		} else {
			dst.PK = cfg.destination
		}
	}

	// Send message
	payload := []byte(cfg.message)
	reply, id, err := client.SendMessage(ctx, dst, payload, cfg.topic, cfg.wait, cfg.replyTimeout)
	if err != nil {
		return err
	}

	if reply != nil {
		fmt.Println("Received reply:")
		printMessage(reply)
	} else {
		fmt.Printf("Message sent successfully. ID: %s\n", id)
	}

	return nil
}

func receiveMessage(ctx context.Context, client *mycelium_client.MyceliumClient, cfg config) error {
	fmt.Printf("Waiting for message (timeout: %d seconds)...\n", cfg.timeout)
	msg, err := client.ReceiveMessage(ctx, cfg.timeout, cfg.topic, false)
	if err != nil {
		return err
	}

	if msg == nil {
		fmt.Println("No message received within timeout")
		return nil
	}

	fmt.Println("Message received:")
	printMessage(msg)
	return nil
}

func replyToMessage(ctx context.Context, client *mycelium_client.MyceliumClient, cfg config) error {
	if err := client.ReplyToMessage(ctx, cfg.messageID, []byte(cfg.message), cfg.topic); err != nil {
		return err
	}

	fmt.Printf("Reply sent to message ID: %s\n", cfg.messageID)
	return nil
}

func getMessageStatus(ctx context.Context, client *mycelium_client.MyceliumClient, messageID string) error {
	status, err := client.GetMessageStatus(ctx, messageID)
	if err != nil {
		return err
	}

	fmt.Printf("Message Status (ID: %s):\n", messageID)
	for k, v := range status {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return nil
}

func listRoutes(ctx context.Context, client *mycelium_client.MyceliumClient, routeType string, jsonOutput bool) error {
	var routes []mycelium_client.Route
	var err error

	// Default to selected routes
	if routeType == "" || routeType == "selected" {
		routes, err = client.ListSelectedRoutes(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Selected Routes (%d):\n", len(routes))
	} else if routeType == "fallback" {
		routes, err = client.ListFallbackRoutes(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("Fallback Routes (%d):\n", len(routes))
	} else {
		return fmt.Errorf("unknown route type: %s (use 'selected' or 'fallback')", routeType)
	}

	if jsonOutput {
		// TODO: Output JSON
		fmt.Printf("Found %d routes\n", len(routes))
	} else {
		if len(routes) == 0 {
			fmt.Println("  No routes found")
			return nil
		}

		for i, route := range routes {
			fmt.Printf("  %d. Subnet: %s\n", i+1, route.Subnet)
			fmt.Printf("     Next Hop: %s\n", route.NextHop)
			fmt.Printf("     Metric: %v, Sequence: %d\n", route.Metric, route.Seqno)
		}
	}
	return nil
}

func printMessage(msg *mycelium_client.InboundMessage) {
	payload, err := msg.Decode()
	fmt.Printf("  ID: %s\n", msg.ID)
	fmt.Printf("  From: %s (IP: %s)\n", msg.SrcPK, msg.SrcIP)
	fmt.Printf("  To: %s (IP: %s)\n", msg.DstPK, msg.DstIP)
	if msg.Topic != "" {
		fmt.Printf("  Topic: %s\n", msg.Topic)
	}
	if err != nil {
		fmt.Printf("  Payload (base64): %s\n", msg.Payload)
		fmt.Printf("  Error decoding payload: %v\n", err)
	} else {
		fmt.Printf("  Payload: %s\n", string(payload))
	}
}
