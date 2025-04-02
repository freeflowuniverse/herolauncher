package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestSetup represents the common test setup
type TestSetup struct {
	App    *fiber.App
	Assert *assert.Assertions
}

// NewTestSetup creates a new test setup
func NewTestSetup(t *testing.T) *TestSetup {
	return &TestSetup{
		App:    fiber.New(),
		Assert: assert.New(t),
	}
}

// PerformRequest performs an HTTP request and returns the response
func (ts *TestSetup) PerformRequest(method, path string, body interface{}) *http.Response {
	// Convert body to JSON if it's not nil
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	// Create a new HTTP request
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	resp, _ := ts.App.Test(req)
	return resp
}

// AssertStatusCode asserts that the response has the expected status code
func (ts *TestSetup) AssertStatusCode(resp *http.Response, expected int) {
	ts.Assert.Equal(expected, resp.StatusCode, "Expected status code %d but got %d", expected, resp.StatusCode)
}

// ParseResponseBody parses the response body into the given struct
func (ts *TestSetup) ParseResponseBody(resp *http.Response, v interface{}) {
	defer resp.Body.Close()
	ts.Assert.NoError(json.NewDecoder(resp.Body).Decode(v), "Failed to parse response body")
}
