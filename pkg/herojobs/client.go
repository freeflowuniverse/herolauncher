package herojobs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// Client represents a client for the HeroJobs service
type Client struct {
	socketPath string
	conn       net.Conn
}

// NewClient creates a new HeroJobs client
func NewClient(socketPath string) (*Client, error) {
	return &Client{
		socketPath: socketPath,
	}, nil
}

// Connect connects to the HeroJobs server
func (c *Client) Connect() error {
	// Connect to the server
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	c.conn = conn
	return nil
}

// Close closes the connection to the HeroJobs server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendCommand sends a command to the server and returns the response
func (c *Client) sendCommand(cmd string, data interface{}) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("not connected to server")
	}

	// Send command
	if _, err := fmt.Fprintln(c.conn, cmd); err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	// Send data as JSON if provided
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}

		if _, err := fmt.Fprintln(c.conn, string(jsonData)); err != nil {
			return "", fmt.Errorf("failed to send data: %w", err)
		}
	}

	// Send empty line to mark end of data
	if _, err := fmt.Fprintln(c.conn, ""); err != nil {
		return "", fmt.Errorf("failed to send end marker: %w", err)
	}

	// Read response
	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(response), nil
}

// SubmitJob submits a job to the server
func (c *Client) SubmitJob(job *Job) (*Job, error) {
	// Send PUT command with job data
	response, err := c.sendCommand(CmdPut, job)
	if err != nil {
		return nil, err
	}

	// Parse response as job
	respJob, err := NewJobFromJSON(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return respJob, nil
}

// GetJob gets a job from the server
func (c *Client) GetJob(jobID string) (*Job, error) {
	// Send GET command with job ID
	data := map[string]string{
		"jobid": jobID,
	}

	response, err := c.sendCommand(CmdGet, data)
	if err != nil {
		return nil, err
	}

	// Parse response as job
	job, err := NewJobFromJSON(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return job, nil
}

// DeleteJob deletes a job from the server
func (c *Client) DeleteJob(jobID string) error {
	// Send DELETE command with job ID
	data := map[string]string{
		"jobid": jobID,
	}

	response, err := c.sendCommand(CmdDelete, data)
	if err != nil {
		return err
	}

	// Parse response
	var respData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &respData); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	status, ok := respData["status"].(string)
	if !ok || status != "success" {
		errMsg, _ := respData["error"].(string)
		return fmt.Errorf("failed to delete job: %s", errMsg)
	}

	return nil
}

// ListJobs lists jobs from the server
func (c *Client) ListJobs(circleID, topic string) ([]string, error) {
	// Send LIST command with circle ID and topic
	data := map[string]string{
		"circleid": circleID,
		"topic":    topic,
	}

	response, err := c.sendCommand(CmdList, data)
	if err != nil {
		return nil, err
	}

	// Parse response
	var respData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &respData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	status, ok := respData["status"].(string)
	if !ok || status != "success" {
		errMsg, _ := respData["error"].(string)
		return nil, fmt.Errorf("failed to list jobs: %s", errMsg)
	}

	// Get job IDs
	jobsInterface, ok := respData["jobs"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	// Convert to string slice
	jobIDs := make([]string, len(jobsInterface))
	for i, job := range jobsInterface {
		jobIDs[i], _ = job.(string)
	}

	return jobIDs, nil
}

// QueueSize gets the size of a queue
func (c *Client) QueueSize(circleID, topic string) (int64, error) {
	// Send QUEUESIZE command with circle ID and topic
	data := map[string]string{
		"circleid": circleID,
		"topic":    topic,
	}

	response, err := c.sendCommand(CmdQueueSize, data)
	if err != nil {
		return 0, err
	}

	// Parse response
	var respData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &respData); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	status, ok := respData["status"].(string)
	if !ok || status != "success" {
		errMsg, _ := respData["error"].(string)
		return 0, fmt.Errorf("failed to get queue size: %s", errMsg)
	}

	// Get size
	size, ok := respData["size"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid size in response")
	}

	return int64(size), nil
}

// QueueEmpty empties a queue
func (c *Client) QueueEmpty(circleID, topic string) error {
	// Send QUEUEEMPTY command with circle ID and topic
	data := map[string]string{
		"circleid": circleID,
		"topic":    topic,
	}

	response, err := c.sendCommand(CmdQueueEmpty, data)
	if err != nil {
		return err
	}

	// Parse response
	var respData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &respData); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	status, ok := respData["status"].(string)
	if !ok || status != "success" {
		errMsg, _ := respData["error"].(string)
		return fmt.Errorf("failed to empty queue: %s", errMsg)
	}

	return nil
}

// QueueGet gets a job from a queue without removing it
func (c *Client) QueueGet(circleID, topic string) (*Job, error) {
	// Send QUEUEGET command with circle ID and topic
	data := map[string]string{
		"circleid": circleID,
		"topic":    topic,
	}

	response, err := c.sendCommand(CmdQueueGet, data)
	if err != nil {
		return nil, err
	}

	// Parse response as job
	job, err := NewJobFromJSON(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return job, nil
}

// QueueFetch gets and removes a job from a queue
func (c *Client) QueueFetch(circleID, topic string) (*Job, error) {
	// Send QUEUEFETCH command with circle ID and topic
	data := map[string]string{
		"circleid": circleID,
		"topic":    topic,
	}

	response, err := c.sendCommand(CmdQueueFetch, data)
	if err != nil {
		return nil, err
	}

	// Parse response as job
	job, err := NewJobFromJSON(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return job, nil
}

// CreateJob creates a new job with the given parameters
func (c *Client) CreateJob(circleID, topic, sessionKey, heroScript, rhaiScript string) (*Job, error) {
	// Create job
	job := NewJob()
	job.CircleID = circleID
	job.Topic = topic
	job.SessionKey = sessionKey
	job.HeroScript = heroScript
	job.RhaiScript = rhaiScript

	// Submit job
	return c.SubmitJob(job)
}
