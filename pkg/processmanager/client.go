package processmanager

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// Client represents a client for the process manager telnet server
type Client struct {
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	secret     string
}

// NewClient creates a new process manager client
func NewClient(socketPath, secret string) *Client {
	return &Client{
		socketPath: socketPath,
		secret:     secret,
	}
}

// Connect connects to the process manager telnet server
func (c *Client) Connect() error {
	// Connect to the Unix domain socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to socket: %v", err)
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)

	// Read welcome message
	welcome, err := c.reader.ReadString('\n')
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read welcome message: %v", err)
	}

	// Authenticate
	if !strings.Contains(welcome, "not authenticated") {
		c.conn.Close()
		return fmt.Errorf("unexpected welcome message: %s", welcome)
	}

	// Send secret
	_, err = c.conn.Write([]byte(c.secret + "\n"))
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to send secret: %v", err)
	}

	// Read authentication response
	authResponse, err := c.reader.ReadString('\n')
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to read authentication response: %v", err)
	}

	if !strings.Contains(authResponse, "authenticated") {
		c.conn.Close()
		return fmt.Errorf("authentication failed: %s", authResponse)
	}

	return nil
}

// Close closes the connection to the process manager
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendCommand sends a command to the process manager and returns the result
func (c *Client) SendCommand(command string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// Send command with a trailing newline to execute it
	_, err := c.conn.Write([]byte(command + "\n\n"))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}

	// Read response
	var result strings.Builder
	inResult := false
	resultComplete := false
	
	// Set a timeout for reading the response
	err = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return "", fmt.Errorf("failed to set read deadline: %v", err)
	}
	
	// Read until we get the complete result or timeout
	for !resultComplete {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			// If we've already started receiving a result, return what we have
			if inResult {
				break
			}
			return result.String(), fmt.Errorf("failed to read response: %v", err)
		}

		if strings.HasPrefix(line, "**RESULT**") {
			inResult = true
			result.WriteString(line)
			continue
		}

		if strings.HasPrefix(line, "**ENDRESULT**") {
			result.WriteString(line)
			resultComplete = true
			break
		}

		if inResult {
			result.WriteString(line)
		}
	}
	
	// Reset the read deadline
	err = c.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return result.String(), fmt.Errorf("failed to reset read deadline: %v", err)
	}

	return result.String(), nil
}

// StartProcess starts a new process
func (c *Client) StartProcess(name, command string, logEnabled bool, deadline int, cron, jobID string) (string, error) {
	heroscript := fmt.Sprintf("!!process.start name:'%s' command:'%s' log:%t", name, command, logEnabled)
	
	if deadline > 0 {
		heroscript += fmt.Sprintf(" deadline:%d", deadline)
	}
	
	if cron != "" {
		heroscript += fmt.Sprintf(" cron:'%s'", cron)
	}
	
	if jobID != "" {
		heroscript += fmt.Sprintf(" jobid:'%s'", jobID)
	}
	
	return c.SendCommand(heroscript)
}

// ListProcesses lists all processes
func (c *Client) ListProcesses(format string) (string, error) {
	heroscript := "!!process.list"
	
	if format != "" {
		heroscript += fmt.Sprintf(" format:'%s'", format)
	}
	
	return c.SendCommand(heroscript)
}

// DeleteProcess deletes a process
func (c *Client) DeleteProcess(name string) (string, error) {
	heroscript := fmt.Sprintf("!!process.delete name:'%s'", name)
	return c.SendCommand(heroscript)
}

// GetProcessStatus gets the status of a process
func (c *Client) GetProcessStatus(name, format string) (string, error) {
	heroscript := fmt.Sprintf("!!process.status name:'%s'", name)
	
	if format != "" {
		heroscript += fmt.Sprintf(" format:'%s'", format)
	}
	
	return c.SendCommand(heroscript)
}

// RestartProcess restarts a process
func (c *Client) RestartProcess(name string) (string, error) {
	heroscript := fmt.Sprintf("!!process.restart name:'%s'", name)
	return c.SendCommand(heroscript)
}

// StopProcess stops a process
func (c *Client) StopProcess(name string) (string, error) {
	heroscript := fmt.Sprintf("!!process.stop name:'%s'", name)
	return c.SendCommand(heroscript)
}
