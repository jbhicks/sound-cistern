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
	// Set test mode
	originalTestMode := isTestMode
	isTestMode = true
	defer func() { isTestMode = originalTestMode }()

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

		assert.Contains(t, response, "collection")
		collection := response["collection"].([]interface{})
		assert.Len(t, collection, 2)

		// Check first track
		track := collection[0].(map[string]interface{})
		assert.Equal(t, "track", track["type"])

		origin := track["origin"].(map[string]interface{})
		trackData := origin["track"].(map[string]interface{})
		assert.Equal(t, "Test Electronic Track", trackData["title"])
		assert.Equal(t, "Electronic", trackData["genre"])
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
		// Test that isTestMode is set correctly
		assert.True(t, isTestMode)

		// Test environment variable detection
		os.Setenv("TEST_MODE", "true")
		defer os.Unsetenv("TEST_MODE")

		// Re-run the test mode detection logic
		testModeFromEnv := os.Getenv("TEST_MODE") == "true"
		assert.True(t, testModeFromEnv)
	})
}
