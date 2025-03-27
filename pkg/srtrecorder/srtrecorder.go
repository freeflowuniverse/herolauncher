package srtrecorder

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/haivision/srtgo"
)

// SRTSession represents an active SRT recording session
type SRTSession struct {
	ID         string
	OutputPath string
	Socket     *srtgo.SrtSocket
	Mutex      sync.Mutex
}

// SRTRecorder represents an SRT server for handling and recording streams
type SRTRecorder struct {
	port         int
	sessions     map[string]*SRTSession
	sessionMutex sync.Mutex
	outputDir    string
	running      bool
	mutex        sync.Mutex
	stopChan     chan struct{}
}

// NewSRTRecorder creates a new SRT recorder
func NewSRTRecorder(port int, outputDir string) *SRTRecorder {
	return &SRTRecorder{
		port:      port,
		sessions:  make(map[string]*SRTSession),
		outputDir: outputDir,
		stopChan:  make(chan struct{}),
	}
}

// IsRunning returns true if the SRT recorder is currently running
func (s *SRTRecorder) IsRunning() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.running
}

// GetPort returns the current port of the SRT recorder
func (s *SRTRecorder) GetPort() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.port
}

// GetSessionCount returns the number of active SRT sessions
func (s *SRTRecorder) GetSessionCount() int {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	return len(s.sessions)
}

// Start starts the SRT recorder
func (s *SRTRecorder) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return nil
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	s.running = true
	s.stopChan = make(chan struct{})

	// Start accepting connections in a goroutine
	go s.acceptConnections()

	log.Printf("SRT recorder listening on port %d", s.port)
	return nil
}

// Stop stops the SRT recorder
func (s *SRTRecorder) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return nil
	}

	// Signal the acceptConnections goroutine to stop
	close(s.stopChan)

	// Close all active sessions
	s.sessionMutex.Lock()
	for _, session := range s.sessions {
		if session.Socket != nil {
			session.Socket.Close()
		}
	}
	s.sessions = make(map[string]*SRTSession)
	s.sessionMutex.Unlock()

	s.running = false
	log.Printf("SRT recorder stopped")
	return nil
}

// acceptConnections accepts incoming SRT connections
func (s *SRTRecorder) acceptConnections() {
	options := make(map[string]string)
	options["transtype"] = "live"
	options["latency"] = "200"

	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Continue with accepting connections
		}

		// Check if we're still running
		s.mutex.Lock()
		if !s.running {
			s.mutex.Unlock()
			return
		}
		port := s.port
		s.mutex.Unlock()

		// Create a new SRT socket
		sck := srtgo.NewSrtSocket("0.0.0.0", uint16(port), options)
		if sck == nil {
			log.Printf("SRT Error: Failed to create socket on port %d", port)
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// Set up the socket
		err := sck.Listen(1)
		if err != nil {
			log.Printf("SRT Error: Failed to listen: %v", err)
			sck.Close()
			time.Sleep(5 * time.Second) // Wait before retrying
			continue
		}

		// Accept a connection
		conn, _, err := sck.Accept()
		if err != nil {
			log.Printf("SRT Error: Failed to accept connection: %v", err)
			sck.Close()
			continue
		}

		// Generate a session ID
		sessionID := fmt.Sprintf("srt_%d", time.Now().UnixNano())
		outputPath := filepath.Join(s.outputDir, sessionID+".ts")

		// Create a new session
		session := &SRTSession{
			ID:         sessionID,
			OutputPath: outputPath,
			Socket:     conn,
		}

		// Store the session
		s.sessionMutex.Lock()
		s.sessions[sessionID] = session
		s.sessionMutex.Unlock()

		// Handle the connection in a goroutine
		go s.handleConnection(sessionID, conn, outputPath)
	}
}

// handleConnection handles an incoming SRT connection and records the stream
func (s *SRTRecorder) handleConnection(sessionID string, conn *srtgo.SrtSocket, outputPath string) {
	defer func() {
		conn.Close()
		
		// Remove the session
		s.sessionMutex.Lock()
		delete(s.sessions, sessionID)
		s.sessionMutex.Unlock()
		
		log.Printf("SRT: Connection closed for session %s", sessionID)
	}()
	
	log.Printf("SRT: Starting to record stream to %s", outputPath)
	
	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		log.Printf("SRT Error: Failed to create output file: %v", err)
		return
	}
	defer file.Close()
	
	// Read from the SRT connection and write to the file
	buf := make([]byte, 1316) // SRT typically uses 1316 bytes as MTU
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("SRT Error: Failed to read from connection: %v", err)
			return
		}
		
		if n == 0 {
			log.Printf("SRT: Connection closed by peer")
			return
		}
		
		// Write the data to the file
		_, err = file.Write(buf[:n])
		if err != nil {
			log.Printf("SRT Error: Failed to write to file: %v", err)
			return
		}
	}
}

// GetSessionFilePath returns the file path for a given session ID
func (s *SRTRecorder) GetSessionFilePath(sessionID string) (string, bool) {
	s.sessionMutex.Lock()
	defer s.sessionMutex.Unlock()
	
	session, exists := s.sessions[sessionID]
	if !exists {
		return "", false
	}
	
	return session.OutputPath, true
}
