package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/freeflowuniverse/herolauncher/pkg/openrpcmanager"
)

func main() {
	// Parse command line arguments
	socketPath := flag.String("socket", "/tmp/openrpc.sock", "Path to the Unix socket")
	secret := flag.String("secret", "test-secret", "Secret for authenticated methods")
	flag.Parse()

	// Create a simple OpenRPC schema
	schema := openrpcmanager.OpenRPCSchema{
		OpenRPC: "1.2.6",
		Info: openrpcmanager.InfoObject{
			Title:   "Hero Launcher API",
			Version: "1.0.0",
		},
		Methods: []openrpcmanager.MethodObject{
			{
				Name: "echo",
				Params: []openrpcmanager.ContentDescriptorObject{
					{
						Name:   "message",
						Schema: openrpcmanager.SchemaObject{"type": "object"},
					},
				},
				Result: &openrpcmanager.ContentDescriptorObject{
					Name:   "result",
					Schema: openrpcmanager.SchemaObject{"type": "object"},
				},
			},
			{
				Name:   "ping",
				Params: []openrpcmanager.ContentDescriptorObject{},
				Result: &openrpcmanager.ContentDescriptorObject{
					Name:   "result",
					Schema: openrpcmanager.SchemaObject{"type": "string"},
				},
			},
			{
				Name:   "secure.info",
				Params: []openrpcmanager.ContentDescriptorObject{},
				Result: &openrpcmanager.ContentDescriptorObject{
					Name:   "result",
					Schema: openrpcmanager.SchemaObject{"type": "object"},
				},
			},
		},
	}

	// Create handlers
	handlers := map[string]openrpcmanager.RPCHandler{
		"echo": func(params json.RawMessage) (interface{}, error) {
			var data interface{}
			if err := json.Unmarshal(params, &data); err != nil {
				return nil, err
			}
			return data, nil
		},
		"ping": func(params json.RawMessage) (interface{}, error) {
			return "pong", nil
		},
		"secure.info": func(params json.RawMessage) (interface{}, error) {
			return map[string]interface{}{
				"server":  "Hero Launcher",
				"version": "1.0.0",
				"status":  "running",
			}, nil
		},
	}

	// Create OpenRPC manager
	manager, err := openrpcmanager.NewOpenRPCManager(schema, handlers, *secret)
	if err != nil {
		log.Fatalf("Failed to create OpenRPC manager: %v", err)
	}

	// Create Unix server
	server, err := openrpcmanager.NewUnixServer(manager, *socketPath)
	if err != nil {
		log.Fatalf("Failed to create Unix server: %v", err)
	}

	// Start the server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start Unix server: %v", err)
	}
	defer server.Stop()

	fmt.Printf("OpenRPC server started on Unix socket: %s\n", *socketPath)
	fmt.Println("Available methods:")
	for _, method := range manager.ListMethods() {
		fmt.Printf("  - %s\n", method)
	}
	fmt.Println("\nPress Ctrl+C to stop the server")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
}
