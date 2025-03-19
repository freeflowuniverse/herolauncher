package videoconf

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/pug/v2"
	"github.com/livekit/protocol/auth"
	livekit "github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

// Config holds the configuration for the video conferencing UI
type Config struct {
	Port          int
	TemplatesPath string
	StaticPath    string
}

// VideoConf represents the video conferencing UI server
type VideoConf struct {
	app        *fiber.App
	config     Config
	livekit    *LiveKitClient
	apiKey     string
	apiSecret  string
	livekitURL string
}

// ConnectionDetails represents the connection details for a LiveKit room
type ConnectionDetails struct {
	ServerURL        string `json:"serverUrl"`
	RoomName         string `json:"roomName"`
	ParticipantToken string `json:"participantToken"`
	ParticipantName  string `json:"participantName"`
}

// DefaultConfig returns the default configuration for the video conferencing UI
func DefaultConfig() Config {
	return Config{
		Port:          8088,
		TemplatesPath: "./web/templates",
		StaticPath:    "./web/static",
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

	// Initialize LiveKit client
	livekit, err := NewLiveKitClient()
	if err != nil {
		log.Printf("Warning: LiveKit client initialization failed: %v", err)
		log.Printf("Video conferencing functionality will be limited")
	}

	return &VideoConf{
		app:        app,
		config:     config,
		livekit:    livekit,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		livekitURL: livekitURL,
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
		var roomsList []fiber.Map

		// If LiveKit client is available, fetch real rooms
		if vc.livekit != nil {
			rooms, err := vc.livekit.ListRooms()
			if err != nil {
				log.Printf("Error listing rooms: %v", err)
			} else {
				for _, room := range rooms {
					// We don't have ListParticipants in our minimal client
					// Just show 0 participants for now
					roomsList = append(roomsList, fiber.Map{
						"ID":               room.Name,
						"Name":             room.Name,
						"ParticipantCount": 0,
						"CreatedAt":        room.CreationTime.Format("2006-01-02 15:04"),
					})
				}
			}
		}

		return c.Render("home", fiber.Map{
			"title": "Video Conference",
			"rooms": roomsList,
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

		apiKey := vc.apiKey
		apiSecret := vc.apiSecret

		if apiKey == "" || apiSecret == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "LiveKit API key or secret not set",
			})
		}

		roomClient := lksdk.NewRoomServiceClient(hostURL, apiKey, apiSecret)

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
		roomType := c.FormValue("roomType", "public")

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

		// If LiveKit client is available, create a real room
		if vc.livekit != nil {
			room, err := vc.livekit.CreateRoom(roomId)
			if err != nil {
				log.Printf("Error creating room %s: %v", roomId, err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create room",
				})
			}
			log.Printf("Created LiveKit room: %s (ID: %s, Type: %s)", roomName, room.Name, roomType)
		} else {
			// Log room creation without LiveKit
			log.Printf("Created mock room: %s (ID: %s, Type: %s)", roomName, roomId, roomType)
		}

		// Redirect to the new room
		return c.Redirect("/room/" + roomId)
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
	randomStr := generateRandomString(4)
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

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
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
