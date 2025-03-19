package videoconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/livekit/protocol/auth"
)

// LiveKitClient is a minimal client for LiveKit operations
type LiveKitClient struct {
	url       string
	apiKey    string
	apiSecret string
}

// Room represents a LiveKit room
type Room struct {
	Name            string    `json:"name"`
	Sid             string    `json:"sid"`
	EmptyTimeout    int       `json:"emptyTimeout"`
	MaxParticipants int       `json:"maxParticipants"`
	CreationTime    time.Time `json:"creationTime"`
	TurnPassword    string    `json:"turnPassword"`
	Enabled         bool      `json:"enabled"`
}

// RoomResponse represents the response from the LiveKit API for room operations
type RoomResponse struct {
	Room Room `json:"room"`
}

// RoomsResponse represents the response from the LiveKit API for listing rooms
type RoomsResponse struct {
	Rooms []Room `json:"rooms"`
}

// NewLiveKitClient creates a new minimal LiveKit client
func NewLiveKitClient() (*LiveKitClient, error) {
	// Get LiveKit configuration from environment variables
	url := os.Getenv("LIVEKIT_URL")
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")

	// Check if required environment variables are set
	if url == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing required environment variables: LIVEKIT_URL, LIVEKIT_API_KEY, LIVEKIT_API_SECRET")
	}

	return &LiveKitClient{
		url:       url,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}, nil
}

// makeRequest makes an authenticated request to the LiveKit API
func (c *LiveKitClient) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// Create the request
	req, err := http.NewRequest(method, c.url+path, reqBody)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create the access token for authentication
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)
	at.SetValidFor(5 * time.Minute) // Short-lived token for API requests
	token, err := at.ToJWT()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("LiveKit API error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

// CreateRoom creates a new LiveKit room
func (c *LiveKitClient) CreateRoom(name string) (*Room, error) {
	body := map[string]interface{}{
		"name":            name,
		"emptyTimeout":    60 * 60, // 1 hour in seconds
		"maxParticipants": 20,
	}

	respData, err := c.makeRequest("POST", "/twirp/livekit.RoomService/CreateRoom", body)
	if err != nil {
		return nil, err
	}

	var resp RoomResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}

	return &resp.Room, nil
}

// ListRooms lists all active LiveKit rooms
func (c *LiveKitClient) ListRooms() ([]Room, error) {
	respData, err := c.makeRequest("POST", "/twirp/livekit.RoomService/ListRooms", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	var resp RoomsResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}

	return resp.Rooms, nil
}

// GenerateToken generates a token for a participant to join a room
func (c *LiveKitClient) GenerateToken(roomName, participantName string, canPublish, canSubscribe bool) (string, error) {
	// Create a new token
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)

	// Create the video grant with the appropriate permissions
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}

	// Set publish and subscribe permissions using the provided methods
	grant.SetCanPublish(canPublish)
	grant.SetCanSubscribe(canSubscribe)

	at.AddGrant(grant)
	at.SetIdentity(participantName)
	at.SetValidFor(24 * time.Hour) // Token valid for 24 hours

	return at.ToJWT()
}
