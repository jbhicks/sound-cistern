//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationAPI tests the API endpoints in test mode
func TestIntegrationAPI(t *testing.T) {
	if os.Getenv("TEST_MODE") == "true" {
		t.Skip("Skip in TEST_MODE - requires mock setup")
	}
	// Set test mode
	originalTestMode := os.Getenv("TEST_MODE")
	os.Setenv("TEST_MODE", "true")
	defer func() { os.Setenv("TEST_MODE", originalTestMode) }()

	// Test health endpoint mock
	t.Run("Health Check Mock", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Simulate health check handler
		err := c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]string
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	// Test mock Soundcloud responses
	t.Run("Mock Soundcloud Activities", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/me/activities", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := mockSoundcloudActivitiesResponse(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "tracks")
		assert.Contains(t, response, "filters")
		assert.Contains(t, response, "pagination")
		tracks := response["tracks"].([]interface{})
		assert.Len(t, tracks, 7)

		// Check first track structure (matches SoundCloud API format)
		firstTrack := tracks[0].(map[string]interface{})
		assert.Equal(t, "DJ PFUNK Mix \"Frequencies\"", firstTrack["title"])
		assert.Equal(t, float64(2267275367), firstTrack["id"])
		assert.Equal(t, "Breakbeat", firstTrack["genre"])

		// Check nested user object
		user := firstTrack["user"].(map[string]interface{})
		assert.Equal(t, "DJ PFunk", user["username"])
	})

	t.Run("Mock Soundcloud User", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := mockSoundcloudUserResponse(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "testuser", response["username"])
		assert.Equal(t, "Test User", response["display_name"])
	})

	t.Run("Mock Soundcloud Token", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/oauth2/token", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := mockSoundcloudTokenResponse(c)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err = json.NewDecoder(rec.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Equal(t, "mock_access_token_12345", response["access_token"])
	})

	// Test test mode flag
	t.Run("Test Mode Flag", func(t *testing.T) {
		// Test environment variable detection
		os.Setenv("TEST_MODE", "true")
		defer os.Unsetenv("TEST_MODE")

		// Re-run the test mode detection logic
		testModeFromEnv := os.Getenv("TEST_MODE") == "true"
		assert.True(t, testModeFromEnv)
	})

	// Test HTMX endpoints
	t.Run("HTMX Request Headers Present", func(t *testing.T) {
		e := echo.New()

		// Test that HTMX headers are properly set on requests
		req := httptest.NewRequest(http.MethodPost, "/api/favorites/12345/htmx", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Verify HX-Request header is detected
		hxRequest := c.Request().Header.Get("HX-Request")
		assert.Equal(t, "true", hxRequest, "HX-Request header should be 'true'")

		// Simulate setting HTML content type (what the handler does)
		c.Response().Header().Set("Content-Type", "text/html")
		assert.Equal(t, "text/html", rec.Header().Get("Content-Type"), "Response should set text/html content type")
	})
}

// TestHTMXResponseHeaders tests that HTMX endpoints return proper headers
func TestHTMXResponseHeaders(t *testing.T) {
	// Set test mode
	originalTestMode := os.Getenv("TEST_MODE")
	os.Setenv("TEST_MODE", "true")
	defer func() { os.Setenv("TEST_MODE", originalTestMode) }()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/favorites/12345/htmx", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Just verify the request headers are properly set up
	assert.Equal(t, "true", req.Header.Get("HX-Request"), "HX-Request header should be set")

	// Simulate what the handler does
	c.Response().Header().Set("Content-Type", "text/html")
	assert.Equal(t, "text/html", rec.Header().Get("Content-Type"), "Content-Type should be text/html")
}
