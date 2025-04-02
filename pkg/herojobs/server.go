package herojobs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Command types
const (
	CmdPut         = "PUT"
	CmdGet         = "GET"
	CmdDelete      = "DELETE"
	CmdList        = "LIST"
	CmdQueueSize   = "QUEUESIZE"
	CmdQueueEmpty  = "QUEUEEMPTY"
	CmdQueueGet    = "QUEUEGET"
	CmdQueueFetch  = "QUEUEFETCH"
)

// Server represents a Unix domain socket server for the HeroJobs service
type Server struct {
	socketPath  string
	redisClient *RedisClient
	listener    net.Listener
	clients     map[net.Conn]bool
	clientsMu   sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	running     bool
}

// NewServer creates a new HeroJobs server
func NewServer(socketPath string, redisAddr string, isRedisUnixSocket bool) (*Server, error) {
	// Create Redis client
	redisClient, err := NewRedisClient(redisAddr, isRedisUnixSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		socketPath:  socketPath,
		redisClient: redisClient,
		clients:     make(map[net.Conn]bool),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	// Check if socket file already exists
	if _, err := os.Stat(s.socketPath); err == nil {
		// Socket file exists, remove it
		if err := os.Remove(s.socketPath); err != nil {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}
	}

	// Create Unix domain socket
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	s.listener = listener
	s.running = true

	// Set socket permissions
	if err := os.Chmod(s.socketPath, 0666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Accept connections in a goroutine
	s.wg.Add(1)
	go s.acceptConnections()

	// Setup signal handling
	s.setupSignalHandling()

	fmt.Printf("HeroJobs server started on socket: %s\n", s.socketPath)
	return nil
}

// Stop stops the server
func (s *Server) Stop() error {
	if !s.running {
		return nil
	}

	s.running = false

	// Signal all goroutines to stop
	s.cancel()

	// Close the listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	// Close all client connections
	s.clientsMu.Lock()
	for conn := range s.clients {
		conn.Close()
		delete(s.clients, conn)
	}
	s.clientsMu.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	// Close Redis client
	if err := s.redisClient.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	// Remove socket file
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	fmt.Println("HeroJobs server stopped")
	return nil
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		// Use a separate goroutine to accept connections so we can check for context cancellation
		connCh := make(chan net.Conn)
		errCh := make(chan error)

		go func() {
			conn, err := s.listener.Accept()
			if err != nil {
				errCh <- err
				return
			}
			connCh <- conn
		}()

		select {
		case <-s.ctx.Done():
			// Context was canceled, exit the loop
			return
		case conn := <-connCh:
			// Handle the connection in a goroutine
			s.wg.Add(1)
			go s.handleConnection(conn)
		case err := <-errCh:
			if s.running {
				fmt.Printf("Failed to accept connection: %v\n", err)
			} else {
				// If we're not running, this is expected during shutdown
				return
			}
		}
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()

	// Add client to the map
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	// Ensure client is removed when connection closes
	defer func() {
		conn.Close()
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
	}()

	// Create a reader for the connection
	reader := bufio.NewReader(conn)
	
	// Process client requests
	for {
		// Check if context is done
		select {
		case <-s.ctx.Done():
			return
		default:
			// Continue processing
		}

		// Read command (first line)
		cmdLine, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading command: %v\n", err)
			}
			return
		}

		// Process command
		cmd := strings.TrimSpace(cmdLine)
		if cmd == "" {
			continue
		}

		// Convert to uppercase for consistency
		cmd = strings.ToUpper(cmd)

		// Read JSON data if needed
		var jsonData string
		var jsonBuffer strings.Builder
		
		// Commands that require JSON data
		if cmd == CmdPut || cmd == CmdGet || cmd == CmdDelete || 
		   cmd == CmdList || cmd == CmdQueueSize || cmd == CmdQueueEmpty || 
		   cmd == CmdQueueGet || cmd == CmdQueueFetch {
			// Read until we get an empty line (marks end of JSON)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Error reading JSON data: %v\n", err)
					}
					return
				}
				
				if strings.TrimSpace(line) == "" {
					// Empty line marks end of JSON
					break
				}
				
				jsonBuffer.WriteString(line)
			}
			
			jsonData = jsonBuffer.String()
		}

		// Process the command
		response, err := s.processCommand(cmd, jsonData)
		if err != nil {
			errResponse := map[string]string{
				"status": "error",
				"error":  err.Error(),
			}
			errJSON, _ := json.Marshal(errResponse)
			conn.Write(errJSON)
			conn.Write([]byte("\n"))
		} else {
			conn.Write([]byte(response))
			conn.Write([]byte("\n"))
		}
	}
}

// processCommand processes a command and its associated JSON data
func (s *Server) processCommand(cmd, jsonData string) (string, error) {
	switch cmd {
	case CmdPut:
		// Parse JSON as job
		job, err := NewJobFromJSON(jsonData)
		if err != nil {
			return "", fmt.Errorf("invalid job JSON: %w", err)
		}
		
		// Store job in Redis
		if err := s.redisClient.EnqueueJob(job); err != nil {
			return "", fmt.Errorf("failed to enqueue job: %w", err)
		}
		
		// Return the job as JSON
		return job.ToJSON()
		
	case CmdGet:
		// Parse job ID from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		jobID, ok := data["jobid"]
		if !ok {
			return "", fmt.Errorf("missing jobid field")
		}
		
		// Get job
		job, err := s.redisClient.GetJob(jobID)
		if err != nil {
			return "", fmt.Errorf("failed to get job: %w", err)
		}
		
		// Return job as JSON
		return job.ToJSON()
		
	case CmdDelete:
		// Parse job ID from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		jobID, ok := data["jobid"]
		if !ok {
			return "", fmt.Errorf("missing jobid field")
		}
		
		// Delete job
		if err := s.redisClient.DeleteJob(jobID); err != nil {
			return "", fmt.Errorf("failed to delete job: %w", err)
		}
		
		// Return success
		response := map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Job %s deleted", jobID),
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}
		
		return string(responseJSON), nil
		
	case CmdList:
		// Parse circle and topic from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		circleID := data["circleid"]
		topic := data["topic"]
		
		// List jobs
		jobIDs, err := s.redisClient.ListJobs(circleID, topic)
		if err != nil {
			return "", fmt.Errorf("failed to list jobs: %w", err)
		}
		
		// Return job IDs as JSON
		response := map[string]interface{}{
			"status": "success",
			"jobs":   jobIDs,
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}
		
		return string(responseJSON), nil
		
	case CmdQueueSize:
		// Parse circle and topic from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		circleID, ok := data["circleid"]
		if !ok {
			return "", fmt.Errorf("missing circleid field")
		}
		
		topic, ok := data["topic"]
		if !ok {
			topic = "default"
		}
		
		// Get queue size
		size, err := s.redisClient.QueueSize(circleID, topic)
		if err != nil {
			return "", fmt.Errorf("failed to get queue size: %w", err)
		}
		
		// Return queue size as JSON
		response := map[string]interface{}{
			"status": "success",
			"size":   size,
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}
		
		return string(responseJSON), nil
		
	case CmdQueueEmpty:
		// Parse circle and topic from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		circleID, ok := data["circleid"]
		if !ok {
			return "", fmt.Errorf("missing circleid field")
		}
		
		topic, ok := data["topic"]
		if !ok {
			topic = "default"
		}
		
		// Empty queue
		if err := s.redisClient.QueueEmpty(circleID, topic); err != nil {
			return "", fmt.Errorf("failed to empty queue: %w", err)
		}
		
		// Return success
		response := map[string]string{
			"status":  "success",
			"message": fmt.Sprintf("Queue for circle %s and topic %s emptied", circleID, topic),
		}
		responseJSON, err := json.Marshal(response)
		if err != nil {
			return "", fmt.Errorf("failed to marshal response: %w", err)
		}
		
		return string(responseJSON), nil
		
	case CmdQueueGet:
		// Parse circle and topic from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		circleID, ok := data["circleid"]
		if !ok {
			return "", fmt.Errorf("missing circleid field")
		}
		
		topic, ok := data["topic"]
		if !ok {
			topic = "default"
		}
		
		// Get job from queue
		job, err := s.redisClient.QueueGet(circleID, topic)
		if err != nil {
			return "", fmt.Errorf("failed to get job from queue: %w", err)
		}
		
		// Return job as JSON
		return job.ToJSON()
		
	case CmdQueueFetch:
		// Parse circle and topic from JSON
		var data map[string]string
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			return "", fmt.Errorf("invalid JSON format: %w", err)
		}
		
		circleID, ok := data["circleid"]
		if !ok {
			return "", fmt.Errorf("missing circleid field")
		}
		
		topic, ok := data["topic"]
		if !ok {
			topic = "default"
		}
		
		// Fetch job from queue
		job, err := s.redisClient.QueueFetch(circleID, topic)
		if err != nil {
			return "", fmt.Errorf("failed to fetch job from queue: %w", err)
		}
		
		// Return job as JSON
		return job.ToJSON()
		
	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}

// setupSignalHandling sets up signal handling for graceful shutdown
func (s *Server) setupSignalHandling() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("Received shutdown signal")
		s.Stop()
	}()
}
