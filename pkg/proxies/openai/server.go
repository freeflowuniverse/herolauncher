package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/jobsmanager"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Server represents the OpenAI proxy server
type Server struct {
	Factory *Factory
	App     *fiber.App
}

// NewServer creates a new proxy server with the given factory
func NewServer(factory *Factory) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	// Add middleware
	app.Use(logger.New())

	server := &Server{
		Factory: factory,
		App:     app,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures all the API routes for the proxy server
func (s *Server) setupRoutes() {
	// API Routes
	api := s.App.Group("/ai")

	// Models endpoints
	api.Get("/models", s.handleModels)
	api.Get("/models/:model", s.handleGetModel)

	// Chat completions endpoint
	api.Post("/chat/completions", s.handleChatCompletions)

	// Completions endpoint
	api.Post("/completions", s.handleCompletions)

	// Embeddings endpoint
	api.Post("/embeddings", s.handleEmbeddings)

	// Images endpoints
	api.Post("/images/generations", s.handleImagesGenerations)
	api.Post("/images/edits", s.handleImagesEdits)
	api.Post("/images/variations", s.handleImagesVariations)

	// Files endpoints
	api.Get("/files", s.handleListFiles)
	api.Post("/files", s.handleUploadFile)
	api.Get("/files/:file_id", s.handleGetFile)
	api.Delete("/files/:file_id", s.handleDeleteFile)
}

// Start starts the proxy server on the configured port
func (s *Server) Start() error {
	return s.App.Listen(fmt.Sprintf(":%d", s.Factory.Config.Port))
}

// extractAPIKey extracts the API key from the Authorization header
func extractAPIKey(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if auth == "" {
		return "" //TODO: give error, needs to be set
	}

	//TODO: why we have following???
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// registerUserIfNeeded registers a new user if the API key is not already registered
func (s *Server) registerUserIfNeeded(apiKey string) {
	// Check if the user already exists
	_, err := s.Factory.GetUserConfig(apiKey)
	if err == nil {
		// User already exists, nothing to do
		return
	}

	// Create a new user configuration with default settings
	userConfig := UserConfig{
		Budget:      10000,           // Default budget
		ModelGroups: []string{"all"}, // Default access to all models
		OpenAIKey:   "",              // Empty OpenAI key means we'll use the server's default key
	}

	// Add the user to the factory
	s.Factory.AddUserConfig(apiKey, userConfig)
}

// createJob creates a new job for tracking the request
func createJob(c *fiber.Ctx, apiKey string, endpoint string, requestBody interface{}) (*jobsmanager.Job, error) {
	// Create a new job
	job := jobsmanager.NewJob()
	job.ParamsType = jobsmanager.ParamsTypeAI
	job.Topic = "openai-proxy"
	job.CircleID = "ai"

	// Serialize request body to JSON
	paramsBytes, err := json.Marshal(map[string]interface{}{
		"endpoint":   endpoint,
		"api_key":    apiKey,
		"request":    requestBody,
		"client_ip":  c.IP(),
		"user_agent": c.Get("User-Agent"),
	})
	if err != nil {
		return nil, fmt.Errorf("error serializing request: %w", err)
	}

	job.Params = string(paramsBytes)
	job.Status = jobsmanager.JobStatusActive
	job.TimeStart = time.Now().Unix()

	// Save the job to store it
	if err := job.Save(); err != nil {
		return nil, fmt.Errorf("error saving job: %w", err)
	}

	return job, nil
}

// finishJob updates the job with the result or error
func finishJob(job *jobsmanager.Job, result interface{}, err error) {
	if err != nil {
		job.Finish(jobsmanager.JobStatusError, "", err)
		return
	}

	// Serialize the result to JSON
	resultBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		job.Finish(jobsmanager.JobStatusError, "", jsonErr)
		return
	}

	job.Finish(jobsmanager.JobStatusDone, string(resultBytes), nil)
}

// handleModels handles GET /v1/models
func (s *Server) handleModels(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Create a job to track this request
	job, err := createJob(c, apiKey, "models.list", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/models"
	if url == "/v1/models" {
		url = "https://api.openai.com/v1/models"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Authorization", "Bearer "+s.Factory.GetOpenAIKey(apiKey))

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	return c.Status(resp.StatusCode).JSON(result)
}

// handleGetModel handles GET /v1/models/:model
func (s *Server) handleGetModel(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	modelID := c.Params("model")

	// Create a job to track this request
	job, err := createJob(c, apiKey, "models.get", map[string]string{
		"model_id": modelID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/models/" + modelID
	if strings.HasPrefix(url, "/v1/models/") {
		url = "https://api.openai.com/v1/models/" + modelID
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Authorization", "Bearer "+s.Factory.GetOpenAIKey(apiKey))

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	return c.Status(resp.StatusCode).JSON(result)
}

// handleChatCompletions handles POST /v1/chat/completions
func (s *Server) handleChatCompletions(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Log the raw request body for debugging
	rawBody := c.Body()
	log.Printf("Received chat completions request with API key: %s", apiKey)
	log.Printf("Raw request body: %s", string(rawBody))

	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Extract model from request
	model, _ := requestBody["model"].(string)

	// Check user budget and model access
	if !s.Factory.CanAccessModel(apiKey, model) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access to this model is not allowed for your API key",
		})
	}

	// Create a job to track this request
	job, err := createJob(c, apiKey, "chat.completions", requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/chat/completions"
	if url == "/v1/chat/completions" {
		url = "https://api.openai.com/v1/chat/completions"
	}

	// Convert the request body back to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body: " + err.Error(),
		})
	}

	log.Printf("Forwarding request to OpenAI at URL: %s", url)
	log.Printf("Request body being sent to OpenAI: %s", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Content-Type", "application/json")
	openaiKey := s.Factory.GetOpenAIKey(apiKey)
	req.Header.Set("Authorization", "Bearer "+openaiKey)
	log.Printf("Using OpenAI key: %s (first 5 chars)", openaiKey[:5])

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error from OpenAI: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error reading response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading OpenAI response: " + err.Error(),
		})
	}

	log.Printf("Response from OpenAI (status %d): %s", resp.StatusCode, string(respBody))

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		finishJob(job, nil, err)
		log.Printf("Error parsing OpenAI response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	// Update user budget based on usage if available
	if usage, ok := result["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			// For simplicity, we're using 1 token = 1 budget unit
			_ = s.Factory.DecreaseBudget(apiKey, uint32(totalTokens))
		}
	}

	return c.Status(resp.StatusCode).JSON(result)
}

// handleCompletions handles POST /v1/completions
func (s *Server) handleCompletions(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Log the raw request body for debugging
	rawBody := c.Body()
	log.Printf("Received completions request with API key: %s", apiKey)
	log.Printf("Raw request body: %s", string(rawBody))

	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Extract model from request
	model, _ := requestBody["model"].(string)

	// Check user budget and model access
	if !s.Factory.CanAccessModel(apiKey, model) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access to this model is not allowed for your API key",
		})
	}

	// Create a job to track this request
	job, err := createJob(c, apiKey, "completions", requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/completions"
	if url == "/v1/completions" {
		url = "https://api.openai.com/v1/completions"
	}

	// Convert the request body back to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body: " + err.Error(),
		})
	}

	log.Printf("Forwarding completions request to OpenAI at URL: %s", url)
	log.Printf("Request body being sent to OpenAI: %s", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Content-Type", "application/json")
	openaiKey := s.Factory.GetOpenAIKey(apiKey)
	req.Header.Set("Authorization", "Bearer "+openaiKey)
	log.Printf("Using OpenAI key: %s (first 5 chars)", openaiKey[:5])

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error from OpenAI: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error reading response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading OpenAI response: " + err.Error(),
		})
	}

	log.Printf("Response from OpenAI (status %d): %s", resp.StatusCode, string(respBody))

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		finishJob(job, nil, err)
		log.Printf("Error parsing OpenAI response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	// Update user budget based on usage if available
	if usage, ok := result["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			// For simplicity, we're using 1 token = 1 budget unit
			_ = s.Factory.DecreaseBudget(apiKey, uint32(totalTokens))
		}
	}

	return c.Status(resp.StatusCode).JSON(result)
}

// handleEmbeddings handles POST /v1/embeddings
func (s *Server) handleEmbeddings(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Log the raw request body for debugging
	rawBody := c.Body()
	log.Printf("Received embeddings request with API key: %s", apiKey)
	log.Printf("Raw request body: %s", string(rawBody))

	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Extract model from request
	model, _ := requestBody["model"].(string)

	// Check user budget and model access
	if !s.Factory.CanAccessModel(apiKey, model) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "access to this model is not allowed for your API key",
		})
	}

	// Create a job to track this request
	job, err := createJob(c, apiKey, "embeddings", requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/embeddings"
	if url == "/v1/embeddings" {
		url = "https://api.openai.com/v1/embeddings"
	}

	// Convert the request body back to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body: " + err.Error(),
		})
	}

	log.Printf("Forwarding embeddings request to OpenAI at URL: %s", url)
	log.Printf("Request body being sent to OpenAI: %s", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Content-Type", "application/json")
	openaiKey := s.Factory.GetOpenAIKey(apiKey)
	req.Header.Set("Authorization", "Bearer "+openaiKey)
	log.Printf("Using OpenAI key: %s (first 5 chars)", openaiKey[:5])

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error from OpenAI: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error reading response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading OpenAI response: " + err.Error(),
		})
	}

	log.Printf("Response from OpenAI (status %d): %s", resp.StatusCode, string(respBody))

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		finishJob(job, nil, err)
		log.Printf("Error parsing OpenAI response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	// Update user budget based on usage if available
	if usage, ok := result["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			// For simplicity, we're using 1 token = 1 budget unit
			_ = s.Factory.DecreaseBudget(apiKey, uint32(totalTokens))
		}
	}

	return c.Status(resp.StatusCode).JSON(result)
}

// handleImagesGenerations handles POST /v1/images/generations
func (s *Server) handleImagesGenerations(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Log the raw request body for debugging
	rawBody := c.Body()
	log.Printf("Received images generations request with API key: %s", apiKey)
	log.Printf("Raw request body: %s", string(rawBody))

	// Parse the request body
	var requestBody map[string]interface{}
	if err := json.Unmarshal(rawBody, &requestBody); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Create a job to track this request
	job, err := createJob(c, apiKey, "images.generations", requestBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Check budget for image generation
	// For image generation we use a standard cost based on size and quantity
	n := 1 // Default to 1 image
	if nValue, ok := requestBody["n"].(float64); ok {
		n = int(nValue)
	}

	costPerImage := uint32(1000) // Base cost - adjust as needed
	totalCost := costPerImage * uint32(n)

	userConfig, err := s.Factory.GetUserConfig(apiKey)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid API key",
		})
	}

	if userConfig.Budget < totalCost {
		finishJob(job, nil, fmt.Errorf("insufficient budget"))
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient budget for this request",
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/images/generations"
	if url == "/v1/images/generations" {
		url = "https://api.openai.com/v1/images/generations"
	}

	// Convert the request body back to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to marshal request body: " + err.Error(),
		})
	}

	log.Printf("Forwarding images generations request to OpenAI at URL: %s", url)
	log.Printf("Request body being sent to OpenAI: %s", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Content-Type", "application/json")
	openaiKey := s.Factory.GetOpenAIKey(apiKey)
	req.Header.Set("Authorization", "Bearer "+openaiKey)
	log.Printf("Using OpenAI key: %s (first 5 chars)", openaiKey[:5])

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error from OpenAI: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		finishJob(job, nil, err)
		log.Printf("Error reading response body: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error reading OpenAI response: " + err.Error(),
		})
	}

	log.Printf("Response from OpenAI (status %d): %s", resp.StatusCode, string(respBody))

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		finishJob(job, nil, err)
		log.Printf("Error parsing OpenAI response: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	// Deduct the cost if successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_ = s.Factory.DecreaseBudget(apiKey, totalCost)
	}

	return c.Status(resp.StatusCode).JSON(result)
}

// handleImagesEdits handles POST /v1/images/edits
func (s *Server) handleImagesEdits(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Create a job to track this request
	job, err := createJob(c, apiKey, "images.edits", map[string]string{
		"info": "multipart form data - see raw request",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// For this example, we'll charge a fixed amount
	costPerEdit := uint32(2000) // Base cost - adjust as needed

	userConfig, err := s.Factory.GetUserConfig(apiKey)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid API key",
		})
	}

	if userConfig.Budget < costPerEdit {
		finishJob(job, nil, fmt.Errorf("insufficient budget"))
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient budget for this request",
		})
	}

	// For multipart requests, we need to handle them differently
	// This is a simplified implementation
	finishJob(job, map[string]string{"status": "processed"}, nil)

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "image edits endpoint not fully implemented in this proxy",
	})
}

// handleImagesVariations handles POST /v1/images/variations
func (s *Server) handleImagesVariations(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Create a job to track this request
	job, err := createJob(c, apiKey, "images.variations", map[string]string{
		"info": "multipart form data - see raw request",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// For this example, we'll charge a fixed amount
	costPerVariation := uint32(1500) // Base cost - adjust as needed

	userConfig, err := s.Factory.GetUserConfig(apiKey)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "invalid API key",
		})
	}

	if userConfig.Budget < costPerVariation {
		finishJob(job, nil, fmt.Errorf("insufficient budget"))
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "insufficient budget for this request",
		})
	}

	// For multipart requests, we need to handle them differently
	// This is a simplified implementation
	finishJob(job, map[string]string{"status": "processed"}, nil)

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "image variations endpoint not fully implemented in this proxy",
	})
}

// handleListFiles handles GET /v1/files
func (s *Server) handleListFiles(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Create a job to track this request
	job, err := createJob(c, apiKey, "files.list", nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/files"
	if url == "/v1/files" {
		url = "https://api.openai.com/v1/files"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Authorization", "Bearer "+s.Factory.GetOpenAIKey(apiKey))

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	return c.Status(resp.StatusCode).JSON(result)
}

// handleUploadFile handles POST /v1/files
func (s *Server) handleUploadFile(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	// Create a job to track this request
	job, err := createJob(c, apiKey, "files.upload", map[string]string{
		"info": "multipart form data - see raw request",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// For multipart requests, we need to handle them differently
	// This is a simplified implementation
	finishJob(job, map[string]string{"status": "processed"}, nil)

	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "file upload endpoint not fully implemented in this proxy",
	})
}

// handleGetFile handles GET /v1/files/:file_id
func (s *Server) handleGetFile(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	fileID := c.Params("file_id")

	// Create a job to track this request
	job, err := createJob(c, apiKey, "files.get", map[string]string{
		"file_id": fileID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/files/" + fileID
	if strings.HasPrefix(url, "/v1/files/") {
		url = "https://api.openai.com/v1/files/" + fileID
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Authorization", "Bearer "+s.Factory.GetOpenAIKey(apiKey))

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	return c.Status(resp.StatusCode).JSON(result)
}

// handleDeleteFile handles DELETE /v1/files/:file_id
func (s *Server) handleDeleteFile(c *fiber.Ctx) error {
	apiKey := extractAPIKey(c)
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "missing API key",
		})
	}

	// Register the user if they don't exist yet
	s.registerUserIfNeeded(apiKey)

	fileID := c.Params("file_id")

	// Create a job to track this request
	job, err := createJob(c, apiKey, "files.delete", map[string]string{
		"file_id": fileID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error creating job: " + err.Error(),
		})
	}

	// Forward request to OpenAI
	url := s.Factory.Config.OpenAIBaseURL + "/v1/files/" + fileID
	if strings.HasPrefix(url, "/v1/files/") {
		url = "https://api.openai.com/v1/files/" + fileID
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create request: " + err.Error(),
		})
	}

	// Set headers from original request
	req.Header.Set("Authorization", "Bearer "+s.Factory.GetOpenAIKey(apiKey))

	// Perform the request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error from OpenAI: " + err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		finishJob(job, nil, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "error parsing OpenAI response: " + err.Error(),
		})
	}

	// Finish the job with the result
	finishJob(job, result, nil)

	return c.Status(resp.StatusCode).JSON(result)
}
