//go:build integration

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testUserID = "7jhe1fvqy9j5n53" // Longshot - the authenticated user
	serverURL  = "https://soundcistern.jbhicks.dev"
)

func getAuthCookie(t *testing.T) *http.Cookie {
	if os.Getenv("TEST_MODE") == "true" {
		return &http.Cookie{Name: "pb_auth", Value: ""}
	}

	token := os.Getenv("TEST_JWT_TOKEN")
	if token == "" {
		t.Skip("TEST_JWT_TOKEN not set - set with: export TEST_JWT_TOKEN='<fresh pb_auth cookie from browser>'")
		return nil
	}

	cookie := &http.Cookie{
		Name:     "pb_auth",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   !strings.Contains(serverURL, "localhost"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}

	return cookie
}

func TestAuthenticatedEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("Home page with auth redirects to stream", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, serverURL+"/", nil)
		require.NoError(t, err)
		req.AddCookie(getAuthCookie(t))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Home page with auth should redirect")

		location := resp.Header.Get("Location")
		assert.Equal(t, "/stream", location, "Should redirect to /stream")
	})

	t.Run("Stream page with auth", func(t *testing.T) {
		if os.Getenv("TEST_MODE") == "true" {
			t.Skip("Skip in TEST_MODE - RequireRecordAuth doesn't work with test middleware")
		}
		req, err := http.NewRequest(http.MethodGet, serverURL+"/stream", nil)
		require.NoError(t, err)
		req.AddCookie(getAuthCookie(t))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Stream page should return 200 with auth")
	})

	t.Run("Stream page displays tracks", func(t *testing.T) {
		if os.Getenv("TEST_MODE") == "true" {
			t.Skip("Skip in TEST_MODE - RequireRecordAuth doesn't work with test middleware")
		}
		req, err := http.NewRequest(http.MethodGet, serverURL+"/stream", nil)
		require.NoError(t, err)
		req.AddCookie(getAuthCookie(t))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		html := string(body)
		t.Logf("Response status: %d, body length: %d", resp.StatusCode, len(html))
		previewLen := 200
		if len(html) < previewLen {
			previewLen = len(html)
		}
		t.Logf("Body preview: %s", html[:previewLen])
		assert.Contains(t, html, "track-container", "Stream page should have track container")

		hasTracks := strings.Contains(html, "track-item") || strings.Contains(html, "data-track-id")
		assert.True(t, hasTracks, "Stream page should display tracks (found neither track-item nor data-track-id)")
	})

	t.Run("API user endpoint with auth", func(t *testing.T) {
		if os.Getenv("TEST_MODE") == "true" {
			t.Skip("Skip in TEST_MODE - RequireRecordAuth doesn't work with test middleware")
		}
		req, err := http.NewRequest(http.MethodGet, serverURL+"/api/user", nil)
		require.NoError(t, err)
		req.AddCookie(getAuthCookie(t))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "API user endpoint should return 200 with auth")

		var userData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&userData)
		require.NoError(t, err)

		assert.Contains(t, userData, "id", "Response should contain user id")
		t.Logf("User data: %+v", userData)
	})

	t.Run("Home page without auth redirects to stream (default tab)", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, serverURL+"/", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Home page without auth should redirect")

		location := resp.Header.Get("Location")
		assert.Equal(t, "/stream", location, "Should redirect to /stream (default tab)")
	})

	t.Run("Login page with auth should still work", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, serverURL+"/login", nil)
		require.NoError(t, err)
		req.AddCookie(getAuthCookie(t))

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Login page should return 200 with auth")
	})
}

func TestSoundCloudAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("TEST_MODE") == "true" {
		t.Skip("Skip in TEST_MODE - RequireRecordAuth doesn't work with test middleware")
	}

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, serverURL+"/api/stream", nil)
	require.NoError(t, err)
	req.AddCookie(getAuthCookie(t))

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	t.Logf("Stream API response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)
		t.Logf("Stream data: %+v", data)

		assert.Contains(t, data, "tracks", "Response should contain tracks")
		tracks, ok := data["tracks"].([]interface{})
		require.True(t, ok, "tracks should be an array")
		assert.Greater(t, len(tracks), 0, "Should have at least one track")
	} else {
		body := make([]byte, 1024)
		n, _ := resp.Body.Read(body)
		t.Logf("Response body: %s", string(body[:n]))
	}
}

func TestLoginPageAccessible(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, serverURL+"/login", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Login page should be accessible without auth")
}

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, serverURL+"/health", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health map[string]string
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)
	assert.Equal(t, "healthy", health["status"])
}

func TestStreamPageDisplaysTracks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser test in short mode")
	}

	token := os.Getenv("TEST_JWT_TOKEN")
	if token == "" {
		t.Skip("TEST_JWT_TOKEN not set")
	}

	domain := strings.Replace(serverURL, "https://", "", 1)

	cookie := &http.Cookie{
		Name:     "pb_auth",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&url.URL{Scheme: "https", Host: domain}, []*http.Cookie{cookie})
	client := &http.Client{Jar: jar}

	resp, err := client.Get(serverURL + "/api/stream")
	require.NoError(t, err)
	defer resp.Body.Close()

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)

	tracks, ok := data["tracks"].([]interface{})
	require.True(t, ok, "tracks should be an array")

	assert.Greater(t, len(tracks), 0, "API should return tracks")

	t.Logf("API returned %d tracks", len(tracks))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	if serverURLEnv := os.Getenv("TEST_SERVER_URL"); serverURLEnv != "" {
		serverURL = serverURLEnv
	}
}
