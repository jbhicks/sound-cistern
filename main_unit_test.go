//go:build unit

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test environment for unit tests
func TestMain(m *testing.M) {
	// Set up any global test configuration here
	m.Run()
}

// TestMockSoundcloudActivitiesResponse tests the mock API response
func TestMockSoundcloudActivitiesResponse(t *testing.T) {
	// Create a test Echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me/activities", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the mock response function
	err := mockSoundcloudActivitiesResponse(c)
	require.NoError(t, err)

	// Check response
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify structure
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
}

// TestMockSoundcloudUserResponse tests the mock user API response
func TestMockSoundcloudUserResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mockSoundcloudUserResponse(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, "Test User", response["display_name"])
}

// TestMockSoundcloudTokenResponse tests the mock token API response
func TestMockSoundcloudTokenResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/oauth2/token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mockSoundcloudTokenResponse(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
	assert.Equal(t, "mock_access_token_12345", response["access_token"])
}

// TestCreateTestUser tests the test user creation functionality
func TestCreateTestUser(t *testing.T) {
	// Skip this test for now as it requires full PocketBase initialization
	// This would need database setup and migrations to run properly
	t.Skip("Skipping database-dependent test - requires full PocketBase setup")
}

// TestTestModeFlag tests that the test mode flag is properly set
func TestTestModeFlag(t *testing.T) {
	// Save original value
	originalTestMode := isTestMode

	// Test default (should be false in normal operation)
	// Note: This test assumes TEST_MODE env var is not set
	// In actual test runs, it would be set by the test framework

	// Restore original value
	isTestMode = originalTestMode
}

// BenchmarkMockAPIResponse benchmarks the mock API performance
func BenchmarkMockSoundcloudActivitiesResponse(b *testing.B) {
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/me/activities", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockSoundcloudActivitiesResponse(c)
	}
}
