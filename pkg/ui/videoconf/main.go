package videoconf

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/yuin/goldmark"
)

// Config holds the configuration for the video conferencing UI
type Config struct {
	Port             int
	AlternativePorts []int
	TemplatesPath    string
	StaticPath       string
	SrtURL           string
}

// VideoConf represents the video conferencing UI server
type VideoConf struct {
	app           *fiber.App
	config        Config
	apiKey        string
	apiSecret     string
	livekitURL    string
	recordingsDir string
}

// ConnectionDetails represents the connection details for a LiveKit room
type ConnectionDetails struct {
	ServerURL        string `json:"serverUrl"`
	RoomName         string `json:"roomName"`
	ParticipantToken string `json:"participantToken"`
	ParticipantName  string `json:"participantName"`
}

// RecordingRequest represents a request to start recording a room
type RecordingRequest struct {
	RoomName string `json:"roomName"`
	Identity string `json:"identity"` // Identity of the participant to record
}

// RecordingResponse represents the response from a recording request
type RecordingResponse struct {
	EgressID string `json:"egressId"`
	Status   string `json:"status"`
}

// DefaultConfig returns the default configuration for the video conferencing UI
func DefaultConfig() Config {
	return Config{
		Port:             8088,
		AlternativePorts: []int{8089, 8090, 8091, 8092},
		TemplatesPath:    "./web/templates",
		StaticPath:       "./web/static",
		SrtURL:           "srt://localhost:22222",
	}
}

// New creates a new video conferencing UI server
func New(config Config) *VideoConf {
	// Initialize template engine with reload enabled for development
	engine := pug.New(config.TemplatesPath, ".pug")
	engine.Reload(true) // Enable reloading for development

	// Add template functions
	engine.AddFunc("title", func(s string) string {
		return strings.Title(s)
	})

	// Add function to render unescaped HTML
	engine.AddFunc("unescaped", func(s string) template.HTML {
		return template.HTML(s)
	})

	// Add function to format timestamps
	engine.AddFunc("formatTime", func(timestamp int64) string {
		// Convert Unix timestamp (seconds) to time.Time
		t := time.Unix(timestamp, 0)
		// Format the time in a human-readable format
		return t.Format("Jan 02, 2006 15:04:05")
	})

	// Add function to render markdown as HTML using goldmark
	engine.AddFunc("markdown", func(content string) template.HTML {
		var buf strings.Builder
		md := goldmark.New()
		if err := md.Convert([]byte(content), &buf); err != nil {
			log.Printf("Error converting markdown to HTML: %v", err)
			return template.HTML(content)
		}
		return template.HTML(buf.String())
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views: engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Static files
	app.Static("/", config.StaticPath)
	app.Static("/css", config.StaticPath+"/css")
	app.Static("/js", config.StaticPath+"/js")
	app.Static("/images", config.StaticPath+"/images")

	// Get LiveKit configuration from environment variables
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	livekitURL := os.Getenv("LIVEKIT_URL")

	// Check if required environment variables are set
	if apiKey == "" || apiSecret == "" || livekitURL == "" {
		log.Printf("Warning: LiveKit environment variables not set (LIVEKIT_API_KEY, LIVEKIT_API_SECRET, LIVEKIT_URL)")
		log.Printf("Video conferencing functionality will be limited")
	}

	// Set up recordings directory
	recordingsDir := os.Getenv("RECORDINGS_DIR")
	if recordingsDir == "" {
		// Default to a directory within static path if not specified
		recordingsDir = filepath.Join(config.StaticPath, "recordings")
	}

	// Create recordings directory if it doesn't exist
	if err := os.MkdirAll(recordingsDir, 0755); err != nil {
		log.Printf("Failed to create recordings directory: %v\n", err)
	}

	return &VideoConf{
		app:           app,
		config:        config,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		livekitURL:    livekitURL,
		recordingsDir: recordingsDir,
	}
}

// SetupRoutes configures the routes for the video conferencing UI
func (vc *VideoConf) SetupRoutes() {
	// Test route for debugging template issues
	vc.app.Get("/test", func(c *fiber.Ctx) error {
		return c.Render("test", fiber.Map{})
	})

	// Home page
	vc.app.Get("/", func(c *fiber.Ctx) error {
		// Get rooms from LiveKit
		var rooms []*livekit.Room
		var connectionError bool

		roomClient := lksdk.NewRoomServiceClient(vc.livekitURL, vc.apiKey, vc.apiSecret)
		res, err := roomClient.ListRooms(context.Background(), &livekit.ListRoomsRequest{})

		if err != nil {
			log.Printf("Error listing rooms: %v", err)
			connectionError = true
			// Return empty list instead of nil to avoid nil pointer dereference
			rooms = []*livekit.Room{}
		} else {
			rooms = res.Rooms
		}

		// Read the markdown content
		mdContent, err := ioutil.ReadFile(vc.config.StaticPath + "/../content/ow_meet_header.md")
		if err != nil {
			log.Printf("Error reading markdown file: %v", err)
			mdContent = []byte("# OurWorld Meet\n\nSecure, private video conferencing solution.")
		}

		return c.Render("home", fiber.Map{
			"title":           "OurWorld Meet",
			"rooms":           rooms,
			"mdContent":       string(mdContent),
			"connectionError": connectionError,
		})
	})

	// Room page
	vc.app.Get("/rooms/:roomId", func(c *fiber.Ctx) error {
		roomId := c.Params("roomId")

		// With our minimal client, we don't check if the room exists
		// Just render the room page directly
		return c.Render("room", fiber.Map{
			"roomName": roomId,
		})
	})

	// API endpoints

	// Create a new room
	vc.app.Post("/api/room", func(c *fiber.Ctx) error {
		// Log the incoming request for debugging
		log.Printf("Received request to create a new room")

		// Parse request body
		type CreateRoomRequest struct {
			Name string `json:"name"`
			// EmptyTimeout    int    `json:"emptyTimeout,omitempty"`
			MaxParticipants int `json:"maxParticipants,omitempty"`
		}

		// Get LiveKit configuration from VideoConf struct
		hostURL := vc.livekitURL
		if hostURL == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "LiveKit URL not set",
			})
		}

		// Convert WebSocket URL to HTTP URL if needed
		if strings.HasPrefix(hostURL, "wss://") {
			hostURL = "https://" + strings.TrimPrefix(hostURL, "wss://")
		} else if strings.HasPrefix(hostURL, "ws://") {
			hostURL = "http://" + strings.TrimPrefix(hostURL, "ws://")
		}

		roomClient := lksdk.NewRoomServiceClient(hostURL, vc.apiKey, vc.apiSecret)

		var req CreateRoomRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body",
			})
		}
		log.Printf("Decoded request: %+v", req)

		// Validate request
		if req.Name == "" {
			// Generate a room ID if not provided
			req.Name = generateRoomId()
		}

		// create a new room
		room, err := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
			Name: req.Name,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create room: " + err.Error(),
			})
		}

		log.Printf("Created room: %+v", room)

		// Redirect to the new room page
		return c.Redirect(fmt.Sprintf("/rooms/%s", room.Name), fiber.StatusSeeOther)
	})

	// Get room info
	vc.app.Get("/api/room/:roomId", func(c *fiber.Ctx) error {
		roomId := c.Params("roomId")

		// Here you would typically fetch room information from a database or service
		// For now, we'll just return the room template
		return c.Render("room", fiber.Map{
			"title":  "Conference Room",
			"roomId": roomId,
		})
	})

	// Create a new room
	vc.app.Post("/api/rooms", func(c *fiber.Ctx) error {
		// Parse form data
		roomName := c.FormValue("roomName", "")

		if roomName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Room name is required",
			})
		}

		// Generate a room ID if not provided
		roomId := roomName
		if strings.TrimSpace(roomId) == "" {
			roomId = generateRoomId()
		}

		roomClient := lksdk.NewRoomServiceClient(vc.livekitURL, vc.apiKey, vc.apiSecret)
		// create a new room
		room, _ := roomClient.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
			Name: roomId,
		})

		// Redirect to the new room
		return c.Redirect("/room/" + room.Sid)
	})

	// GET endpoint to handle connection details
	vc.app.Get("/api/connection-details", func(c *fiber.Ctx) error {
		// Extract query parameters
		roomName := c.Query("roomName")
		if roomName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required query parameter: roomName",
			})
		}

		participantName := c.Query("participantName")
		if participantName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required query parameter: participantName",
			})
		}

		metadata := c.Query("metadata", "")
		region := c.Query("region", "")

		// Determine the LiveKit server URL based on region
		livekitServerURL := vc.livekitURL
		if region != "" {
			url, err := vc.getLiveKitURL(region)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Invalid region: %v", err),
				})
			}
			livekitServerURL = url
		}

		// Generate participant token
		participantToken, err := vc.createParticipantToken(participantName, roomName, metadata)
		if err != nil {
			log.Printf("Error generating token for room %s, participant %s: %v", roomName, participantName, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create participant token: %v", err),
			})
		}

		// Create connection details response
		connectionDetails := ConnectionDetails{
			ServerURL:        livekitServerURL,
			RoomName:         roomName,
			ParticipantToken: participantToken,
			ParticipantName:  participantName,
		}

		// Return JSON response
		return c.JSON(connectionDetails)
	})

	// Start recording a room
	vc.app.Post("/api/recording/start", func(c *fiber.Ctx) error {
		var req RecordingRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request body: " + err.Error(),
			})
		}

		// Validate request
		if req.RoomName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Room name is required",
			})
		}
		if req.Identity == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Participant identity is required",
			})
		}

		// Create LiveKit room client to get participant tracks
		roomClient := lksdk.NewRoomServiceClient(vc.livekitURL, vc.apiKey, vc.apiSecret)

		// Get participants in the room to find tracks
		ctx := context.Background()
		listReq := &livekit.ListParticipantsRequest{
			Room: req.RoomName,
		}
		response, err := roomClient.ListParticipants(ctx, listReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get room participants: " + err.Error(),
			})
		}

		// Find the participant to record
		var targetParticipant *livekit.ParticipantInfo
		for _, p := range response.Participants {
			if p.Identity == req.Identity {
				targetParticipant = p
				break
			}
		}

		if targetParticipant == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Participant not found in room",
			})
		}

		// Generate a unique filename prefix based on room name, participant and timestamp
		timestamp := time.Now().Format("20060102-150405")
		filenamePrefix := fmt.Sprintf("%s-%s-%s", req.RoomName, req.Identity, timestamp)

		// Create LiveKit egress client
		egressClient := lksdk.NewEgressClient(vc.livekitURL, vc.apiKey, vc.apiSecret)

		// Generate a unique session ID for this recording
		sessionID := fmt.Sprintf("%s_%s", filenamePrefix, GenerateRandomString(4))

		// Handle IPv6 addresses properly by enclosing them in square brackets if needed
		log.Printf("Using SRT URL: %s", vc.config.SrtURL)
		streamOutput := &livekit.StreamOutput{
			Protocol: livekit.StreamProtocol_SRT,
			Urls:     []string{vc.config.SrtURL},
		}

		// Log detailed information about the stream output
		log.Printf("Stream output configuration: Protocol=%v, URLs=%v",
			streamOutput.Protocol, streamOutput.Urls)

		// Create the ParticipantEgressRequest
		participantEgressReq := &livekit.ParticipantEgressRequest{
			RoomName: req.RoomName,
			Identity: req.Identity,
			StreamOutputs: []*livekit.StreamOutput{
				streamOutput,
			},
		}

		// Log detailed information about the egress request
		log.Printf("Participant egress request details: RoomName=%s, Identity=%s, StreamOutputs=%+v",
			participantEgressReq.RoomName, participantEgressReq.Identity, participantEgressReq.StreamOutputs)

		// Log network interface information to help diagnose connectivity issues
		log.Printf("Network interfaces for diagnostic purposes:")
		interfaces, err := net.Interfaces()
		if err == nil {
			for _, iface := range interfaces {
				addrs, err := iface.Addrs()
				if err == nil {
					for _, addr := range addrs {
						log.Printf("Interface: %s, Address: %s", iface.Name, addr.String())
					}
				}
			}
		} else {
			log.Printf("Error getting network interfaces: %v", err)
		}

		// Start the participant egress
		log.Printf("Starting participant egress with request: %+v", participantEgressReq)
		egressInfo, err := egressClient.StartParticipantEgress(ctx, participantEgressReq)
		if err != nil {
			log.Printf("Error: Failed to start participant recording: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to start recording: " + err.Error(),
			})
		}

		log.Printf("Egress started successfully with ID: %s, status: %s", egressInfo.EgressId, egressInfo.Status.String())

		// Log more detailed information about the egress response
		log.Printf("Egress response details: EgressId=%s, Status=%s, StartedAt=%v, Error=%v",
			egressInfo.EgressId, egressInfo.Status.String(),
			time.Unix(egressInfo.StartedAt, 0), egressInfo.Error)

		// Log additional information about egress outputs
		log.Printf("Checking for stream result details in egress response")
		if streamResults := egressInfo.GetStreamResults(); len(streamResults) > 0 {
			for i, result := range streamResults {
				log.Printf("Stream result %d: URL=%s, Status=%s", i, result.GetUrl(), result.GetStatus().String())
			}
		} else {
			log.Printf("WARNING: No stream results available in egress response")
		}

		// Return response with the egress ID and session ID for later reference
		egressID := egressInfo.EgressId
		status := egressInfo.Status.String()

		// Store the session ID and egress ID mapping for later reference
		// This could be used to retrieve the recording file path later
		log.Printf("Started recording with egressID: %s, sessionID: %s\n", egressID, sessionID)

		return c.JSON(RecordingResponse{
			EgressID: egressID,
			Status:   status,
		})
	})

	// Stop recording a room
	vc.app.Post("/api/recording/stop", func(c *fiber.Ctx) error {
		egressID := c.FormValue("egressId")
		if egressID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Egress ID is required",
			})
		}

		// Create LiveKit client
		egressClient := lksdk.NewEgressClient(vc.livekitURL, vc.apiKey, vc.apiSecret)

		// First check the egress status
		log.Printf("Getting egress info for ID: %s", egressID)

		egressInfo, err := egressClient.ListEgress(context.Background(), &livekit.ListEgressRequest{
			EgressId: egressID,
		})
		if err != nil {
			log.Printf("Error getting egress info: %v", err)
			log.Printf("Error type: %T", err)
			log.Printf("Error details: %+v", err)
			// Continue with stop attempt even if we can't get info
		} else {
			log.Printf("Successfully retrieved egress info with %d items", len(egressInfo.Items))
			// Check if we have any items in the list
			if len(egressInfo.Items) == 0 {
				log.Printf("No egress found with ID: %s", egressID)
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Recording not found",
				})
			}

			// Get the first item (should be the only one)
			egress := egressInfo.Items[0]

			// Log the current status and details
			log.Printf("Egress status before stopping: %s", egress.Status.String())
			// Format times properly
			var endedAtTime, updatedAtTime string
			if egress.EndedAt > 0 {
				endedAtTime = time.Unix(egress.EndedAt, 0).String()
			} else {
				endedAtTime = "not set"
			}
			if egress.UpdatedAt > 0 {
				updatedAtTime = time.Unix(egress.UpdatedAt, 0).String()
			} else {
				updatedAtTime = "not set"
			}

			log.Printf("Egress details: Room=%s, StartedAt=%v, EndedAt=%v, UpdatedAt=%v, Error=%s",
				egress.RoomName,
				time.Unix(egress.StartedAt, 0),
				endedAtTime,
				updatedAtTime,
				egress.Error)

			// Log additional details about the egress
			log.Printf("Egress additional details: SourceType=%s, ErrorCode=%d, Details=%s",
				egress.SourceType.String(), egress.ErrorCode, egress.Details)

			// Check for stream results
			if streamResults := egress.GetStreamResults(); len(streamResults) > 0 {
				for i, result := range streamResults {
					log.Printf("Stream result %d: URL=%s, Status=%s",
						i, result.GetUrl(), result.GetStatus().String())
				}
			} else {
				log.Printf("WARNING: No stream results available in egress response")
			}

			// If egress failed, log more details
			if egress.Status == livekit.EgressStatus_EGRESS_FAILED {
				log.Printf("Egress failure reason: %s", egress.Error)
				log.Printf("Egress failure code: %d", egress.ErrorCode)
				log.Printf("Egress failure details: %s", egress.Details)
			}

			// If already in a terminal state, return success without trying to stop
			if egress.Status == livekit.EgressStatus_EGRESS_COMPLETE ||
				egress.Status == livekit.EgressStatus_EGRESS_FAILED ||
				egress.Status == livekit.EgressStatus_EGRESS_ABORTED {
				log.Printf("Egress already in terminal state: %s, not attempting to stop", egress.Status.String())
				return c.JSON(fiber.Map{
					"status":  "already_stopped",
					"message": fmt.Sprintf("Recording was already in %s state", egress.Status.String()),
					"error":   egress.Error,
				})
			}
		}

		// Try to stop egress
		log.Printf("Attempting to stop egress with ID: %s", egressID)

		stopResult, err := egressClient.StopEgress(context.Background(), &livekit.StopEgressRequest{
			EgressId: egressID,
		})
		if err != nil {
			// Log detailed error information
			log.Printf("Error stopping egress: %v", err)
			log.Printf("Error type: %T", err)
			log.Printf("Error details: %+v", err)

			// Check if the error is because the egress is already in a terminal state
			if strings.Contains(err.Error(), "EGRESS_FAILED") ||
				strings.Contains(err.Error(), "EGRESS_COMPLETE") ||
				strings.Contains(err.Error(), "EGRESS_ABORTED") {
				log.Printf("Egress already in terminal state according to error: %v", err)
				return c.JSON(fiber.Map{
					"status":  "already_stopped",
					"message": "Recording was already stopped or failed",
					"error":   err.Error(),
				})
			}

			log.Printf("Error stopping egress: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to stop recording: " + err.Error(),
			})
		}

		log.Printf("Successfully stopped egress, result status: %s", stopResult.Status.String())
		// Format EndedAt time properly
		var endedAtTime string
		if stopResult.EndedAt > 0 {
			endedAtTime = time.Unix(stopResult.EndedAt, 0).String()
		} else {
			endedAtTime = "not set"
		}

		log.Printf("Detailed stop result: EgressId=%s, Status=%s, StartedAt=%v, EndedAt=%v, Error=%s",
			stopResult.EgressId, stopResult.Status.String(),
			time.Unix(stopResult.StartedAt, 0),
			endedAtTime,
			stopResult.Error)

		return c.JSON(fiber.Map{
			"status": "stopped",
		})
	})
}

// Start starts the video conferencing UI server
func (vc *VideoConf) Start() error {
	log.Printf("Starting video conferencing UI server on port %d", vc.config.Port)
	return vc.app.Listen(fmt.Sprintf(":%d", vc.config.Port))
}

// GetApp returns the underlying Fiber app
func (vc *VideoConf) GetApp() *fiber.App {
	return vc.app
}

// generateRoomId generates a random room ID
func generateRoomId() string {
	// Initialize random source with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random 6-character string
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 6)
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}

	return string(result)
}

// getLiveKitURL returns the LiveKit server URL based on the region
func (vc *VideoConf) getLiveKitURL(region string) (string, error) {
	// Get LiveKit URL from VideoConf struct
	baseURL := vc.livekitURL
	if baseURL == "" {
		return "", fmt.Errorf("LiveKit URL not set")
	}

	// Convert WebSocket URL to HTTP URL if needed
	if strings.HasPrefix(baseURL, "wss://") {
		baseURL = "https://" + strings.TrimPrefix(baseURL, "wss://")
	} else if strings.HasPrefix(baseURL, "ws://") {
		baseURL = "http://" + strings.TrimPrefix(baseURL, "ws://")
	}

	// In the future, we could implement region-specific URLs if needed
	// For now, we just return the base URL regardless of region
	return baseURL, nil
}

// createParticipantToken generates a token for a participant to join a room
func (vc *VideoConf) createParticipantToken(participantName string, roomName string, metadata string) (string, error) {
	// Generate a random string for the participant's identity
	randomStr := GenerateRandomString(4)
	identity := fmt.Sprintf("%s__%s", participantName, randomStr)

	// Debug: Log the API key and secret (truncated for security)
	log.Printf("Using API Key: %s... (truncated)", vc.apiKey[:min(len(vc.apiKey), 5)])
	log.Printf("Using API Secret: %s... (truncated)", vc.apiSecret[:min(len(vc.apiSecret), 5)])

	// Create a new access token using the LiveKit SDK - following docs exactly
	at := auth.NewAccessToken(vc.apiKey, vc.apiSecret)

	// Create a video grant
	grant := &auth.VideoGrant{
		Room:     roomName,
		RoomJoin: true,
	}

	// Set permissions using the proper setter methods
	grant.SetCanPublish(true)
	grant.SetCanPublishData(true)
	grant.SetCanSubscribe(true)

	// Add grant and set identity in a chain as shown in the documentation
	at.AddGrant(grant).
		SetIdentity(identity).
		SetName(participantName).
		SetValidFor(300 * time.Second) // Token expiration: 5 minutes

	// Set metadata if provided
	if metadata != "" {
		at.SetMetadata(metadata)
	}

	// Generate the JWT
	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %v", err)
	}

	// Debug: Log that we generated a token
	log.Printf("Generated token for room %s, participant %s", roomName, participantName)

	return token, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) string {
	// Initialize random source with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define character set
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)

	// Generate random string
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}

	return string(result)
}
