package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
)

// RPCRequest represents an outgoing RPC request
type RPCRequest struct {
	Method     string          `json:"method"`
	Params     json.RawMessage `json:"params"`
	ID         interface{}     `json:"id,omitempty"`
	Secret     string          `json:"secret,omitempty"`
	JSONRPC    string          `json:"jsonrpc"`
}

// RPCResponse represents an incoming RPC response
type RPCResponse struct {
	Result  interface{}     `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
}

// RPCError represents an RPC error
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// Parse command line arguments
	socketPath := flag.String("socket", "/tmp/openrpc.sock", "Path to the Unix socket")
	method := flag.String("method", "rpc.discover", "RPC method to call")
	params := flag.String("params", "{}", "JSON parameters for the method")
	secret := flag.String("secret", "", "Secret for authenticated methods")
	flag.Parse()

	// Connect to the Unix socket
	conn, err := net.Dial("unix", *socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to Unix socket: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Create the request
	var paramsJSON json.RawMessage
	if err := json.Unmarshal([]byte(*params), &paramsJSON); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid JSON parameters: %v\n", err)
		os.Exit(1)
	}

	request := RPCRequest{
		Method:  *method,
		Params:  paramsJSON,
		ID:      1,
		Secret:  *secret,
		JSONRPC: "2.0",
	}

	// Marshal the request
	requestData, err := json.Marshal(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal request: %v\n", err)
		os.Exit(1)
	}

	// Send the request
	_, err = conn.Write(requestData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send request: %v\n", err)
		os.Exit(1)
	}

	// Read the response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response: %v\n", err)
		os.Exit(1)
	}

	// Parse the response
	var response RPCResponse
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal response: %v\n", err)
		os.Exit(1)
	}

	// Check for errors
	if response.Error != nil {
		fmt.Fprintf(os.Stderr, "Error: %s (code: %d)\n", response.Error.Message, response.Error.Code)
		if response.Error.Data != nil {
			fmt.Fprintf(os.Stderr, "Error data: %v\n", response.Error.Data)
		}
		os.Exit(1)
	}

	// Print the result
	resultJSON, err := json.MarshalIndent(response.Result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(resultJSON))
}
