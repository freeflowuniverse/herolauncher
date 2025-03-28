// pkg/mycelium_client/client.go
package mycelium_client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// DefaultAPIPort is the default port on which the Mycelium HTTP API listens
const DefaultAPIPort = 8989

// Default timeout values
const (
	DefaultClientTimeout = 30 * time.Second
	DefaultReplyTimeout  = 60        // seconds
	DefaultReceiveWait   = 10        // seconds
)

// MyceliumClient represents a client for interacting with the Mycelium API
type MyceliumClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Mycelium client with the given base URL
// If baseURL is empty, it defaults to "http://localhost:8989"
func NewClient(baseURL string) *MyceliumClient {
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", DefaultAPIPort)
	}
	return &MyceliumClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: DefaultClientTimeout},
	}
}

// SetTimeout sets the HTTP client timeout
func (c *MyceliumClient) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

// Message Structures

// MessageDestination represents a destination for a message, either by IP or public key
type MessageDestination struct {
	IP string `json:"ip,omitempty"` // IPv6 address in the overlay network
	PK string `json:"pk,omitempty"` // Public key hex encoded
}

// PushMessage represents a message to be sent
type PushMessage struct {
	Dst     MessageDestination `json:"dst"`
	Topic   string             `json:"topic,omitempty"`
	Payload string             `json:"payload"` // Base64 encoded
}

// InboundMessage represents a received message
type InboundMessage struct {
	ID      string `json:"id"`
	SrcIP   string `json:"srcIp"`
	SrcPK   string `json:"srcPk"`
	DstIP   string `json:"dstIp"`
	DstPK   string `json:"dstPk"`
	Topic   string `json:"topic,omitempty"`
	Payload string `json:"payload"` // Base64 encoded
}

// MessageResponse represents the ID of a pushed message
type MessageResponse struct {
	ID string `json:"id"`
}

// NodeInfo represents general information about the Mycelium node
type NodeInfo struct {
	NodeSubnet string `json:"nodeSubnet"`
}

// PeerStats represents statistics about a peer
type PeerStats struct {
	Endpoint        Endpoint `json:"endpoint"`
	Type            string   `json:"type"`            // static, inbound, linkLocalDiscovery
	ConnectionState string   `json:"connectionState"` // alive, connecting, dead
	TxBytes         int64    `json:"txBytes,omitempty"`
	RxBytes         int64    `json:"rxBytes,omitempty"`
}

// Endpoint represents connection information for a peer
type Endpoint struct {
	Proto      string `json:"proto"`      // tcp, quic
	SocketAddr string `json:"socketAddr"` // IP:port
}

// Route represents a network route
type Route struct {
	Subnet  string      `json:"subnet"`
	NextHop string      `json:"nextHop"`
	Metric  interface{} `json:"metric"` // Can be int or string "infinite"
	Seqno   int         `json:"seqno"`
}

// Decode decodes the base64 payload of an inbound message
func (m *InboundMessage) Decode() ([]byte, error) {
	return base64.StdEncoding.DecodeString(m.Payload)
}

// GetNodeInfo retrieves general information about the Mycelium node
func (c *MyceliumClient) GetNodeInfo(ctx context.Context) (*NodeInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/v1/admin", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var info NodeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

// SendMessage sends a message to a specified destination
// If waitForReply is true, it will wait for a reply up to the specified timeout
func (c *MyceliumClient) SendMessage(ctx context.Context, dst MessageDestination, payload []byte, topic string, waitForReply bool, replyTimeout int) (*InboundMessage, string, error) {
	// Encode payload to base64
	encodedPayload := base64.StdEncoding.EncodeToString(payload)

	msg := PushMessage{
		Dst:     dst,
		Topic:   topic,
		Payload: encodedPayload,
	}

	reqBody, err := json.Marshal(msg)
	if err != nil {
		return nil, "", err
	}

	// Build URL with optional reply_timeout
	url := fmt.Sprintf("%s/api/v1/messages", c.BaseURL)
	if waitForReply && replyTimeout > 0 {
		url = fmt.Sprintf("%s?reply_timeout=%d", url, replyTimeout)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// If we got a reply (status 200)
	if resp.StatusCode == http.StatusOK && waitForReply {
		var reply InboundMessage
		if err := json.NewDecoder(resp.Body).Decode(&reply); err != nil {
			return nil, "", err
		}
		return &reply, "", nil
	}

	// If we just got a message ID (status 201)
	var result MessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", err
	}

	return nil, result.ID, nil
}

// ReplyToMessage sends a reply to a previously received message
func (c *MyceliumClient) ReplyToMessage(ctx context.Context, msgID string, payload []byte, topic string) error {
	encodedPayload := base64.StdEncoding.EncodeToString(payload)

	msg := PushMessage{
		Dst:     MessageDestination{}, // Not needed for replies
		Topic:   topic,
		Payload: encodedPayload,
	}

	reqBody, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/messages/reply/%s", c.BaseURL, msgID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ReceiveMessage waits for and receives a message, optionally filtering by topic
// If timeout is 0, it will return immediately if no message is available
func (c *MyceliumClient) ReceiveMessage(ctx context.Context, timeout int, topic string, peek bool) (*InboundMessage, error) {
	params := url.Values{}
	if timeout > 0 {
		params.Add("timeout", fmt.Sprintf("%d", timeout))
	}
	if topic != "" {
		params.Add("topic", topic)
	}
	if peek {
		params.Add("peek", "true")
	}

	url := fmt.Sprintf("%s/api/v1/messages?%s", c.BaseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// No message available
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var msg InboundMessage
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// GetMessageStatus checks the status of a previously sent message
func (c *MyceliumClient) GetMessageStatus(ctx context.Context, msgID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/messages/status/%s", c.BaseURL, msgID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return status, nil
}

// ListPeers retrieves a list of known peers
func (c *MyceliumClient) ListPeers(ctx context.Context) ([]PeerStats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/v1/admin/peers", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var peers []PeerStats
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return nil, err
	}

	return peers, nil
}

// AddPeer adds a new peer to the network
func (c *MyceliumClient) AddPeer(ctx context.Context, endpoint string) error {
	// The API expects a direct endpoint string, not a JSON object
	reqBody := []byte(endpoint)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/v1/admin/peers", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// RemovePeer removes a peer from the network
func (c *MyceliumClient) RemovePeer(ctx context.Context, endpoint string) error {
	url := fmt.Sprintf("%s/api/v1/admin/peers/%s", c.BaseURL, url.PathEscape(endpoint))
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListSelectedRoutes retrieves a list of selected routes
func (c *MyceliumClient) ListSelectedRoutes(ctx context.Context) ([]Route, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/v1/admin/routes/selected", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, err
	}

	return routes, nil
}

// ListFallbackRoutes retrieves a list of fallback routes
func (c *MyceliumClient) ListFallbackRoutes(ctx context.Context) ([]Route, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/api/v1/admin/routes/fallback", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, err
	}

	return routes, nil
}
