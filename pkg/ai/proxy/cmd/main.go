package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/ai/proxy"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	// Start the server in a goroutine
	go runServerMode()

	// Wait a moment for the server to start
	time.Sleep(2 * time.Second)

	// Test the proxy with a client
	testProxyWithClient()

	// Keep the main function running
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")
}

// testProxyWithClient tests the proxy using the OpenAI Go client
func testProxyWithClient() {
	log.Println("Testing proxy with OpenAI Go client...")

	// Create a client that points to our proxy
	// Note: The server is using "/ai" as the prefix for all routes
	client := openai.NewClient(
		option.WithAPIKey("test-key"),                  // This is our test key, not a real OpenAI key
		option.WithBaseURL("http://localhost:8080/ai"), // Use the /ai prefix to match the server routes
	)

	// Create a completion request
	chatCompletion, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		},
		Model: "gpt-3.5-turbo", // Use a model that our proxy supports
	})

	if err != nil {
		log.Fatalf("Error creating completion: %v", err)
	}

	// Print the response
	log.Printf("Completion response: %s", chatCompletion.Choices[0].Message.Content)
	log.Println("Proxy test completed successfully!")
}

// runServerMode starts the proxy server with example configurations
func runServerMode() {
	// Get the OpenAI API key from environment variable
	openaiKey := os.Getenv("OPENAIKEY")
	if openaiKey == "" {
		log.Println("ERROR: OPENAIKEY environment variable is not set")
		os.Exit(1)
	}

	// Create a proxy configuration
	config := proxy.ProxyConfig{
		Port:             8080,                     // Use a non-privileged port for testing
		OpenAIBaseURL:    "https://api.openai.com", // Default OpenAI API URL
		DefaultOpenAIKey: openaiKey,                // Fallback API key if user doesn't have one
	}

	// Create a new factory with the configuration
	factory := proxy.NewFactory(config)

	// Add some example user configurations with the test key
	factory.AddUserConfig("test-key", proxy.UserConfig{
		Budget:      10000,           // 10,000 tokens
		ModelGroups: []string{"all"}, // Allow access to all models
		OpenAIKey:   "",              // Empty means use the default key
	})

	// Print debug info
	log.Printf("Added user config for 'test-key'")

	// Create a new server with the factory
	server := proxy.NewServer(factory)

	// Start the server
	fmt.Printf("OpenAI Proxy Server listening on port %d\n", config.Port)
	if err := server.Start(); err != nil {
		log.Printf("Error starting server: %v", err)
	}
}
