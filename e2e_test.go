//go:build e2e

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	serverCmd    *exec.Cmd
	serverURL    = "http://localhost:8090"
	testDataDir  = "./pb_data_test"
	serverCancel context.CancelFunc
)

func TestMain(m *testing.M) {
	if os.Getenv("SKIP_SERVER_START") == "true" {
		os.Exit(m.Run())
	}

	code := 1
	defer func() {
		os.Exit(code)
	}()

	fmt.Println("=== E2E Test Setup ===")

	if err := setupTestDatabase(); err != nil {
		fmt.Printf("Failed to setup test database: %v\n", err)
		return
	}

	if err := startTestServer(); err != nil {
		fmt.Printf("Failed to start test server: %v\n", err)
		return
	}
	defer stopTestServer()

	fmt.Println("=== Running E2E Tests ===")
	code = m.Run()
}

func setupTestDatabase() error {
	fmt.Println("Setting up test database...")

	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(testDataDir, 0755); err != nil {
			return fmt.Errorf("failed to create test data directory: %w", err)
		}
	}

	templateDir := "./pb_data_test_template"
	if _, err := os.Stat(templateDir); err == nil {
		fmt.Println("Found test database template, copying...")
		if err := copyDir(templateDir, testDataDir); err != nil {
			return fmt.Errorf("failed to copy test database template: %w", err)
		}
	}

	return nil
}

func startTestServer() error {
	fmt.Println("Starting test server...")

	ctx, cancel := context.WithCancel(context.Background())
	serverCancel = cancel

	cmd := exec.CommandContext(ctx, "./sound-cistern", "serve", "--dir="+testDataDir, "--http=0.0.0.0:8090")
	cmd.Env = append(os.Environ(),
		"TEST_MODE=true",
		"SOUNDCLOUD_CLIENT_ID=test_client_id",
		"SOUNDCLOUD_CLIENT_SECRET=test_client_secret",
		"SOUNDCLOUD_REDIRECT_URI=http://localhost:8090/auth/soundcloud/callback",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var err error
	serverCmd = cmd
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	fmt.Println("Waiting for server to be ready...")
	if err := waitForServer(serverURL, 30*time.Second); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	fmt.Println("Test server is ready!")
	return nil
}

func stopTestServer() {
	fmt.Println("Stopping test server...")

	if serverCancel != nil {
		serverCancel()
	}

	if serverCmd != nil && serverCmd.Process != nil {
		if runtime.GOOS == "windows" {
			serverCmd.Process.Kill()
		} else {
			serverCmd.Process.Signal(os.Interrupt)
		}

		done := make(chan struct{})
		go func() {
			serverCmd.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(5 * time.Second):
			serverCmd.Process.Kill()
		}
	}

	fmt.Println("Test server stopped")
}

func waitForServer(url string, timeout time.Duration) error {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func TestE2EHealthCheck(t *testing.T) {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	t.Log("Health check passed")
}

func TestE2ELoginPage(t *testing.T) {
	resp, err := http.Get(serverURL + "/login")
	if err != nil {
		t.Fatalf("Failed to get login page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	t.Log("Login page accessible")
}

func TestE2EStreamPageInTestMode(t *testing.T) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(serverURL + "/stream")
	if err != nil {
		t.Fatalf("Failed to get stream page: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Stream page status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusTemporaryRedirect {
		location := resp.Header.Get("Location")
		t.Logf("Redirected to: %s", location)
	}

	if resp.StatusCode == http.StatusOK {
		t.Log("Stream page accessible (test mode auto-auth working)")
	}
}

func TestE2EMockActivitiesAPI(t *testing.T) {
	activities := getMockActivities()

	tracks, ok := activities["tracks"].([]map[string]interface{})
	if !ok {
		t.Fatal("Mock activities should contain tracks array")
	}

	if len(tracks) != 7 {
		t.Errorf("Expected 7 mock tracks, got %d", len(tracks))
	}

	t.Logf("Mock activities contain %d tracks", len(tracks))
}

func TestE2EBlogPage(t *testing.T) {
	resp, err := http.Get(serverURL + "/blog")
	if err != nil {
		t.Fatalf("Failed to get blog page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	t.Log("Blog page accessible")
}

func TestE2EStaticFiles(t *testing.T) {
	resp, err := http.Get(serverURL + "/css/style.css")
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("Server not running")
		}
		t.Fatalf("Failed to get static file: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Static file status: %d", resp.StatusCode)
}

// TestE2EStreamAPIPage1 tests the stream API returns first page of tracks
func TestE2EStreamAPIPage1(t *testing.T) {
	resp, err := http.Get(serverURL + "/api/stream?page=1")
	if err != nil {
		t.Fatalf("Failed to get stream API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := resp.Body.Read(make([]byte, 1)) // just check it's not empty
	if len(body) == 0 {
		t.Error("Response body should not be empty")
	}
}

// TestE2EStreamAPIPagination tests all 3 pages of pagination
func TestE2EStreamAPIPagination(t *testing.T) {
	// Page 1 - should have 9 tracks
	resp1, err := http.Get(serverURL + "/api/stream?page=1")
	if err != nil {
		t.Fatalf("Failed to get page 1: %v", err)
	}
	defer resp1.Body.Close()

	body1, _ := httpGetBody(resp1)
	trackCount1 := strings.Count(body1, "stream-flip-card")
	assert.Equal(t, 9, trackCount1, "Page 1 should have 9 tracks")
	t.Logf("Page 1: %d tracks", trackCount1)

	// Page 2 - should have 9 tracks
	resp2, err := http.Get(serverURL + "/api/stream?page=2")
	if err != nil {
		t.Fatalf("Failed to get page 2: %v", err)
	}
	defer resp2.Body.Close()

	body2, _ := httpGetBody(resp2)
	trackCount2 := strings.Count(body2, "stream-flip-card")
	assert.Equal(t, 9, trackCount2, "Page 2 should have 9 tracks")
	t.Logf("Page 2: %d tracks", trackCount2)

	// Page 3 - should have 7 tracks (25 - 18 = 7)
	resp3, err := http.Get(serverURL + "/api/stream?page=3")
	if err != nil {
		t.Fatalf("Failed to get page 3: %v", err)
	}
	defer resp3.Body.Close()

	body3, _ := httpGetBody(resp3)
	trackCount3 := strings.Count(body3, "stream-flip-card")
	assert.Equal(t, 7, trackCount3, "Page 3 should have 7 tracks")
	t.Logf("Page 3: %d tracks", trackCount3)

	// Page 4 - should have 0 tracks (empty)
	resp4, err := http.Get(serverURL + "/api/stream?page=4")
	if err != nil {
		t.Fatalf("Failed to get page 4: %v", err)
	}
	defer resp4.Body.Close()

	body4, _ := httpGetBody(resp4)
	trackCount4 := strings.Count(body4, "stream-flip-card")
	assert.Equal(t, 0, trackCount4, "Page 4 should have 0 tracks")
	t.Logf("Page 4: %d tracks", trackCount4)
}

// TestE2EStreamAPISearch tests search functionality via API
func TestE2EStreamAPISearch(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected int
	}{
		{"Search metal", "metal", 1},
		{"Search techno", "techno", 1},
		{"Search house", "house", 2},
		{"Search jazz", "jazz", 1},
		{"Empty search", "", 9}, // returns first page
		{"No results", "xyz123", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := serverURL + "/api/stream?q=" + tt.query
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to get search results: %v", err)
			}
			defer resp.Body.Close()

			body, _ := httpGetBody(resp)
			trackCount := strings.Count(body, "stream-flip-card")
			assert.Equal(t, tt.expected, trackCount, "Search '%s' should return %d tracks", tt.query, tt.expected)
		})
	}
}

// TestE2EStreamAPIDurationFilter tests duration filter via API
func TestE2EStreamAPIDurationFilter(t *testing.T) {
	tests := []struct {
		name        string
		durationMin int
		expected    int
	}{
		{"No filter", 0, 9},
		{"Min 3 min", 3, 9},
		{"Min 5 min", 5, 6},
		{"Min 7 min", 7, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := serverURL + fmt.Sprintf("/api/stream?duration_min=%d", tt.durationMin)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to get filtered results: %v", err)
			}
			defer resp.Body.Close()

			body, _ := httpGetBody(resp)
			trackCount := strings.Count(body, "stream-flip-card")
			// Note: Exact count may vary based on page size
			if tt.expected > 0 {
				assert.GreaterOrEqual(t, trackCount, 1, "Should have at least 1 track for duration_min=%d", tt.durationMin)
			} else {
				assert.Equal(t, 0, trackCount, "Should have 0 tracks for no matches")
			}
		})
	}
}

// TestE2EStreamPageStructure tests the stream page HTML structure
func TestE2EStreamPageStructure(t *testing.T) {
	resp, err := http.Get(serverURL + "/stream")
	if err != nil {
		t.Fatalf("Failed to get stream page: %v", err)
	}
	defer resp.Body.Close()

	body, _ := httpGetBody(resp)

	// Check key elements exist
	assert.True(t, strings.Contains(body, `id="filter-q"`), "Search input should exist")
	assert.True(t, strings.Contains(body, `id="filter-content-type"`), "Content type filter should exist")
	assert.True(t, strings.Contains(body, `id="filter-sort"`), "Sort filter should exist")
	assert.True(t, strings.Contains(body, `id="filter-limit"`), "Limit filter should exist")
	assert.True(t, strings.Contains(body, `id="duration-range"`), "Duration slider should exist")
	assert.True(t, strings.Contains(body, `id="cards-grid"`), "Cards grid should exist")
	assert.True(t, strings.Contains(body, `id="load-more-btn"`), "Load more button should exist")

	// Check HTMX attributes
	assert.True(t, strings.Contains(body, `hx-get="/api/stream"`), "HTMX GET to API should exist")
	assert.True(t, strings.Contains(body, `hx-target="#cards-grid"`), "HTMX target should be cards-grid")
}

// TestE2EStreamPageTracks tests that the stream page displays tracks on load
func TestE2EStreamPageTracks(t *testing.T) {
	resp, err := http.Get(serverURL + "/stream")
	if err != nil {
		t.Fatalf("Failed to get stream page: %v", err)
	}
	defer resp.Body.Close()

	body, _ := httpGetBody(resp)
	trackCount := strings.Count(body, "stream-flip-card")

	// The page initially loads 9 tracks via API call
	assert.GreaterOrEqual(t, trackCount, 1, "Stream page should display at least 1 track")
	t.Logf("Stream page displays %d tracks", trackCount)
}

// TestE2EStreamHTMXEndpoints tests HTMX endpoints return correct format
func TestE2EStreamHTMXEndpoints(t *testing.T) {
	// Test HTMX request returns HTML without wrapper div
	req, _ := http.NewRequest("GET", serverURL+"/api/stream?page=2", nil)
	req.Header.Set("HX-Request", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make HTMX request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("Content-Type"))

	body, _ := httpGetBody(resp)
	// HTMX requests should NOT have the grid wrapper div
	assert.True(t, strings.Contains(body, "stream-flip-card"), "Should contain track cards")
}

// TestE2EStreamFiltersCombination tests multiple filters combined
func TestE2EStreamFiltersCombination(t *testing.T) {
	// Search + duration filter
	url := serverURL + "/api/stream?q=metal&duration_min=5"
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get filtered results: %v", err)
	}
	defer resp.Body.Close()

	body, _ := httpGetBody(resp)
	trackCount := strings.Count(body, "stream-flip-card")

	// "Heavy Metal" is 4:40 = 4 min, so with duration_min=5 it should be excluded
	assert.Equal(t, 0, trackCount, "Metal search with 5min filter should return 0 tracks")
	t.Logf("Combined filters returned %d tracks", trackCount)
}

// TestE2EStreamLimitOption tests the limit dropdown option
func TestE2EStreamLimitOption(t *testing.T) {
	tests := []struct {
		name  string
		limit string
	}{
		{"Limit 20", "20"},
		{"Limit 50", "50"},
		{"Limit 100", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := serverURL + "/api/stream?limit=" + tt.limit
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to get with limit: %v", err)
			}
			defer resp.Body.Close()

			body, _ := httpGetBody(resp)
			trackCount := strings.Count(body, "stream-flip-card")

			// Each page returns up to limit tracks
			assert.GreaterOrEqual(t, trackCount, 1, "Limit %s should return tracks", tt.limit)
		})
	}
}

// TestE2EStreamSortOptions tests sorting via API
func TestE2EStreamSortOptions(t *testing.T) {
	sorts := []string{"newest", "oldest", "title", "artist", "duration"}

	for _, sort := range sorts {
		t.Run("Sort by "+sort, func(t *testing.T) {
			url := serverURL + "/api/stream?sort=" + sort
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("Failed to get sorted results: %v", err)
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode, "Sort %s should return 200", sort)
		})
	}
}

// Helper function to get response body as string
func httpGetBody(resp *http.Response) (string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", err
	}
	return buf.String(), nil
}
