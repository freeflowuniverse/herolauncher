# Mycelium Client

A Go client for the Mycelium overlay network. This package allows you to connect to a Mycelium node via its HTTP API and perform operations like sending/receiving messages and managing peers.

## Features

- Send and receive messages through the Mycelium network
- List, add, and remove peers
- View network routes
- Query node information
- Reply to received messages
- Check message status

## Usage

### Basic Client Usage

```go
// Create a new client with default configuration (localhost:8989)
client := mycelium_client.NewClient("")

// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Get node info
info, err := client.GetNodeInfo(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Node subnet: %s\n", info.NodeSubnet)

// List peers
peers, err := client.ListPeers(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Found %d peers\n", len(peers))

// Send a message
dest := mycelium_client.MessageDestination{
    PK: "publicKeyHexString", // or IP: "myceliumIPv6Address"
}
payload := []byte("Hello from mycelium client!")
waitForReply := false
replyTimeout := 0 // not used when waitForReply is false
_, msgID, err := client.SendMessage(ctx, dest, payload, "example.topic", waitForReply, replyTimeout)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Message sent with ID: %s\n", msgID)

// Receive a message with 10 second timeout
msg, err := client.ReceiveMessage(ctx, 10, "", false)
if err != nil {
    log.Fatal(err)
}
if msg != nil {
    payload, _ := msg.Decode()
    fmt.Printf("Received message: %s\n", string(payload))
}
```

### Command Line Tool

The package includes a command-line tool for interacting with a Mycelium node:

```
Usage: mycelium-client [flags] COMMAND [args...]

Flags:
  -api string
        Mycelium API URL (default "http://localhost:8989")
  -json
        Output in JSON format
  -timeout int
        Client timeout in seconds (default 30)

Commands:
  info                        Get node information
  peers                       List connected peers
  add-peer ENDPOINT           Add a new peer
  del-peer ENDPOINT           Remove a peer
  send [--pk=PK|--ip=IP] [--topic=TOPIC] [--wait] [--reply-timeout=N] MESSAGE
                              Send a message to a destination
  receive [--topic=TOPIC] [--timeout=N]
                              Receive a message
  reply ID [--topic=TOPIC] MESSAGE
                              Reply to a message
  status ID                   Get status of a sent message
  routes [selected|fallback]  List routes (default: selected)
```

## Building the Command Line Tool

```bash
cd pkg/mycelium_client/cmd
go build -o mycelium-client
```

## Examples

See the `examples` directory for full usage examples.

## Notes

- This client requires a running Mycelium node accessible via HTTP API.
- The default API endpoint is http://localhost:8989.
- Messages are automatically encoded/decoded from base64 when working with the API.
