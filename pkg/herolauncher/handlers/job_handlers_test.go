package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freeflowuniverse/herolauncher/pkg/herojobs"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHeroJobsClient is a mock implementation of the HeroJobs client
type MockHeroJobsClient struct {
	mock.Mock
}

// Connect mocks the Connect method
func (m *MockHeroJobsClient) Connect() error {
	args := m.Called()
	return args.Error(0)
}

// Close mocks the Close method
func (m *MockHeroJobsClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// SubmitJob mocks the SubmitJob method
func (m *MockHeroJobsClient) SubmitJob(job *herojobs.Job) (*herojobs.Job, error) {
	args := m.Called(job)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*herojobs.Job), args.Error(1)
}

// GetJob mocks the GetJob method
func (m *MockHeroJobsClient) GetJob(jobID string) (*herojobs.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*herojobs.Job), args.Error(1)
}

// DeleteJob mocks the DeleteJob method
func (m *MockHeroJobsClient) DeleteJob(jobID string) error {
	args := m.Called(jobID)
	return args.Error(0)
}

// ListJobs mocks the ListJobs method
func (m *MockHeroJobsClient) ListJobs(circleID, topic string) ([]string, error) {
	args := m.Called(circleID, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// QueueSize mocks the QueueSize method
func (m *MockHeroJobsClient) QueueSize(circleID, topic string) (int64, error) {
	args := m.Called(circleID, topic)
	return args.Get(0).(int64), args.Error(1)
}

// QueueEmpty mocks the QueueEmpty method
func (m *MockHeroJobsClient) QueueEmpty(circleID, topic string) error {
	args := m.Called(circleID, topic)
	return args.Error(0)
}

// QueueGet mocks the QueueGet method
func (m *MockHeroJobsClient) QueueGet(circleID, topic string) (*herojobs.Job, error) {
	args := m.Called(circleID, topic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*herojobs.Job), args.Error(1)
}

// CreateJob mocks the CreateJob method
func (m *MockHeroJobsClient) CreateJob(circleID, topic, sessionKey, heroScript, rhaiScript string) (*herojobs.Job, error) {
	args := m.Called(circleID, topic, sessionKey, heroScript, rhaiScript)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*herojobs.Job), args.Error(1)
}

// setupTest initializes a test environment with a mock client
func setupTest() (*JobHandler, *MockHeroJobsClient, *fiber.App) {
	mockClient := new(MockHeroJobsClient)
	handler := &JobHandler{
		client: mockClient,
	}

	app := fiber.New()

	// Register routes
	api := app.Group("/api")
	jobs := api.Group("/jobs")
	jobs.Post("/create", handler.createJob)
	jobs.Get("/queue/get", handler.queueGet)
	jobs.Post("/queue/empty", handler.queueEmpty)
	jobs.Post("/submit", handler.submitJob)
	jobs.Get("/get/:jobid", handler.getJob)
	jobs.Delete("/delete/:jobid", handler.deleteJob)
	jobs.Get("/list", handler.listJobs)
	jobs.Get("/queue/size", handler.queueSize)

	return handler, mockClient, app
}

// createTestRequest creates a test request with the given method, path, and body
func createTestRequest(method, path string, body io.Reader) (*http.Request, error) {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// TestQueueEmpty tests the queueEmpty handler
func TestQueueEmpty(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		circleID       string
		topic          string
		connectError   error
		emptyError     error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   nil,
			emptyError:     nil,
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"status":"success","message":"Queue for circle test-circle and topic test-topic emptied successfully"}`,
		},
		{
			name:           "Connection Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   errors.New("connection error"),
			emptyError:     nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to connect to HeroJobs server: connection error"}`,
		},
		{
			name:           "Empty Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   nil,
			emptyError:     errors.New("empty error"),
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to empty queue: empty error"}`,
		},
		{
			name:           "Empty Circle ID",
			circleID:       "",
			topic:          "test-topic",
			connectError:   nil,
			emptyError:     nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Circle ID is required"}`,
		},
		{
			name:           "Empty Topic",
			circleID:       "test-circle",
			topic:          "",
			connectError:   nil,
			emptyError:     nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Topic is required"}`,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock client for each test
			mockClient := new(MockHeroJobsClient)
			
			// Setup mock expectations - Connect is always called in the handler
			mockClient.On("Connect").Return(tc.connectError)
			
			// QueueEmpty and Close are only called if Connect succeeds and parameters are valid
			if tc.connectError == nil && tc.circleID != "" && tc.topic != "" {
				mockClient.On("QueueEmpty", tc.circleID, tc.topic).Return(tc.emptyError)
				mockClient.On("Close").Return(nil)
			} else {
				// Close is still called via defer even if we return early
				mockClient.On("Close").Return(nil).Maybe()
			}
			
			// Create a new handler with the mock client
			handler := &JobHandler{
				client: mockClient,
			}
			
			// Create a new app for each test
			app := fiber.New()
			api := app.Group("/api")
			jobs := api.Group("/jobs")
			jobs.Post("/queue/empty", handler.queueEmpty)
			
			// Create request body
			reqBody := map[string]string{
				"circleid": tc.circleID,
				"topic":    tc.topic,
			}
			reqBodyBytes, err := json.Marshal(reqBody)
			assert.NoError(t, err)
			
			// Create test request
			req, err := createTestRequest(http.MethodPost, "/api/jobs/queue/empty", bytes.NewReader(reqBodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// Perform the request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			
			// Check status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			
			// Check response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedBody, string(body))
			
			// Verify that all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

// TestQueueGet tests the queueGet handler
func TestQueueGet(t *testing.T) {
	// Create a test job
	testJob := &herojobs.Job{
		JobID:    "test-job-id",
		CircleID: "test-circle",
		Topic:    "test-topic",
	}
	
	// Test cases
	tests := []struct {
		name           string
		circleID       string
		topic          string
		connectError   error
		getError       error
		getResponse    *herojobs.Job
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   nil,
			getError:       nil,
			getResponse:    testJob,
			expectedStatus: fiber.StatusOK,
			// Include all fields in the response, even empty ones
			expectedBody:   `{"jobid":"test-job-id","circleid":"test-circle","topic":"test-topic","error":"","heroscript":"","result":"","rhaiscript":"","sessionkey":"","status":"","time_end":0,"time_scheduled":0,"time_start":0,"timeout":0}`,
		},
		{
			name:           "Connection Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   errors.New("connection error"),
			getError:       nil,
			getResponse:    nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to connect to HeroJobs server: connection error"}`,
		},
		{
			name:           "Get Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			connectError:   nil,
			getError:       errors.New("get error"),
			getResponse:    nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to get job from queue: get error"}`,
		},
		{
			name:           "Empty Circle ID",
			circleID:       "",
			topic:          "test-topic",
			connectError:   nil,
			getError:       nil,
			getResponse:    nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Circle ID is required"}`,
		},
		{
			name:           "Empty Topic",
			circleID:       "test-circle",
			topic:          "",
			connectError:   nil,
			getError:       nil,
			getResponse:    nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Topic is required"}`,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock client for each test
			mockClient := new(MockHeroJobsClient)
			
			// Setup mock expectations - Connect is always called in the handler
			mockClient.On("Connect").Return(tc.connectError)
			
			// QueueGet and Close are only called if Connect succeeds and parameters are valid
			if tc.connectError == nil && tc.circleID != "" && tc.topic != "" {
				mockClient.On("QueueGet", tc.circleID, tc.topic).Return(tc.getResponse, tc.getError)
				mockClient.On("Close").Return(nil)
			} else {
				// Close is still called via defer even if we return early
				mockClient.On("Close").Return(nil).Maybe()
			}
			
			// Create a new handler with the mock client
			handler := &JobHandler{
				client: mockClient,
			}
			
			// Create a new app for each test
			app := fiber.New()
			api := app.Group("/api")
			jobs := api.Group("/jobs")
			jobs.Get("/queue/get", handler.queueGet)
			
			// Create test request
			path := fmt.Sprintf("/api/jobs/queue/get?circleid=%s&topic=%s", tc.circleID, tc.topic)
			req, err := createTestRequest(http.MethodGet, path, nil)
			assert.NoError(t, err)
			
			// Perform the request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			
			// Check status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			
			// Check response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedBody, string(body))
			
			// Verify that all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

// TestCreateJob tests the createJob handler
func TestCreateJob(t *testing.T) {
	// Create a test job
	testJob := &herojobs.Job{
		JobID:    "test-job-id",
		CircleID: "test-circle",
		Topic:    "test-topic",
	}
	
	// Test cases
	tests := []struct {
		name           string
		circleID       string
		topic          string
		sessionKey     string
		heroScript     string
		rhaiScript     string
		connectError   error
		createError    error
		createResponse *herojobs.Job
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			circleID:       "test-circle",
			topic:          "test-topic",
			sessionKey:     "test-key",
			heroScript:     "test-hero-script",
			rhaiScript:     "test-rhai-script",
			connectError:   nil,
			createError:    nil,
			createResponse: testJob,
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"jobid":"test-job-id","circleid":"test-circle","topic":"test-topic","error":"","heroscript":"","result":"","rhaiscript":"","sessionkey":"","status":"","time_end":0,"time_scheduled":0,"time_start":0,"timeout":0}`,
		},
		{
			name:           "Connection Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			sessionKey:     "test-key",
			heroScript:     "test-hero-script",
			rhaiScript:     "test-rhai-script",
			connectError:   errors.New("connection error"),
			createError:    nil,
			createResponse: nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to connect to HeroJobs server: connection error"}`,
		},
		{
			name:           "Create Error",
			circleID:       "test-circle",
			topic:          "test-topic",
			sessionKey:     "test-key",
			heroScript:     "test-hero-script",
			rhaiScript:     "test-rhai-script",
			connectError:   nil,
			createError:    errors.New("create error"),
			createResponse: nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to create job: create error"}`,
		},
		{
			name:           "Empty Circle ID",
			circleID:       "",
			topic:          "test-topic",
			sessionKey:     "test-key",
			heroScript:     "test-hero-script",
			rhaiScript:     "test-rhai-script",
			connectError:   nil,
			createError:    nil,
			createResponse: nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Circle ID is required"}`,
		},
		{
			name:           "Empty Topic",
			circleID:       "test-circle",
			topic:          "",
			sessionKey:     "test-key",
			heroScript:     "test-hero-script",
			rhaiScript:     "test-rhai-script",
			connectError:   nil,
			createError:    nil,
			createResponse: nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Topic is required"}`,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock client for each test
			mockClient := new(MockHeroJobsClient)
			
			// Setup mock expectations - Connect is always called in the handler
			mockClient.On("Connect").Return(tc.connectError)
			
			// CreateJob and Close are only called if Connect succeeds and parameters are valid
			if tc.connectError == nil && tc.circleID != "" && tc.topic != "" {
				mockClient.On("CreateJob", tc.circleID, tc.topic, tc.sessionKey, tc.heroScript, tc.rhaiScript).Return(tc.createResponse, tc.createError)
				mockClient.On("Close").Return(nil)
			} else {
				// Close is still called via defer even if we return early
				mockClient.On("Close").Return(nil).Maybe()
			}
			
			// Create a new handler with the mock client
			handler := &JobHandler{
				client: mockClient,
			}
			
			// Create a new app for each test
			app := fiber.New()
			api := app.Group("/api")
			jobs := api.Group("/jobs")
			jobs.Post("/create", handler.createJob)
			
			// Create request body
			reqBody := map[string]string{
				"circleid":    tc.circleID,
				"topic":       tc.topic,
				"sessionkey":  tc.sessionKey,
				"heroscript":  tc.heroScript,
				"rhaiscript":  tc.rhaiScript,
			}
			reqBodyBytes, err := json.Marshal(reqBody)
			assert.NoError(t, err)
			
			// Create test request
			req, err := createTestRequest(http.MethodPost, "/api/jobs/create", bytes.NewReader(reqBodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// Perform the request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			
			// Check status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			
			// Check response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedBody, string(body))
			
			// Verify that all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

// TestSubmitJob tests the submitJob handler
func TestSubmitJob(t *testing.T) {
	// Create a test job
	testJob := &herojobs.Job{
		JobID:    "test-job-id",
		CircleID: "test-circle",
		Topic:    "test-topic",
	}
	
	// Test cases
	tests := []struct {
		name           string
		job            *herojobs.Job
		connectError   error
		submitError    error
		submitResponse *herojobs.Job
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success",
			job:            testJob,
			connectError:   nil,
			submitError:    nil,
			submitResponse: testJob,
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"jobid":"test-job-id","circleid":"test-circle","topic":"test-topic","error":"","heroscript":"","result":"","rhaiscript":"","sessionkey":"","status":"","time_end":0,"time_scheduled":0,"time_start":0,"timeout":0}`,
		},
		{
			name:           "Connection Error",
			job:            testJob,
			connectError:   errors.New("connection error"),
			submitError:    nil,
			submitResponse: nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to connect to HeroJobs server: connection error"}`,
		},
		{
			name:           "Submit Error",
			job:            testJob,
			connectError:   nil,
			submitError:    errors.New("submit error"),
			submitResponse: nil,
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"Failed to submit job: submit error"}`,
		},
		{
			name:           "Empty Job",
			job:            nil,
			connectError:   nil,
			submitError:    nil,
			submitResponse: nil,
			expectedStatus: fiber.StatusBadRequest,
			expectedBody:   `{"error":"Failed to parse job data: unexpected end of JSON input"}`,
		},
	}
	
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new mock client for each test
			mockClient := new(MockHeroJobsClient)
			
			// Setup mock expectations - Connect is always called in the handler
			mockClient.On("Connect").Return(tc.connectError)
			
			// SubmitJob and Close are only called if Connect succeeds and job is not nil
			if tc.connectError == nil && tc.job != nil {
				mockClient.On("SubmitJob", tc.job).Return(tc.submitResponse, tc.submitError)
				mockClient.On("Close").Return(nil)
			} else {
				// Close is still called via defer even if we return early
				mockClient.On("Close").Return(nil).Maybe()
			}
			
			// Create a new handler with the mock client
			handler := &JobHandler{
				client: mockClient,
			}
			
			// Create a new app for each test
			app := fiber.New()
			api := app.Group("/api")
			jobs := api.Group("/jobs")
			jobs.Post("/submit", handler.submitJob)
			
			// Create request body
			var reqBodyBytes []byte
			var err error
			if tc.job != nil {
				reqBodyBytes, err = json.Marshal(tc.job)
				assert.NoError(t, err)
			}
			
			// Create test request
			req, err := createTestRequest(http.MethodPost, "/api/jobs/submit", bytes.NewReader(reqBodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			
			// Perform the request
			resp, err := app.Test(req)
			assert.NoError(t, err)
			
			// Check status code
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			
			// Check response body
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, tc.expectedBody, string(body))
			
			// Verify that all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}
