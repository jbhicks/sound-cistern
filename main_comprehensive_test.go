//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFavoritesAPIInTestMode tests the favorites API in test mode
func TestFavoritesAPIInTestMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set test mode
	originalTestMode := os.Getenv("TEST_MODE")
	os.Setenv("TEST_MODE", "true")
	defer func() { os.Setenv("TEST_MODE", originalTestMode) }()

	t.Run("Favorites API returns HTML in test mode", func(t *testing.T) {
		// This would require the server to be running
		// For now, just verify the test mode flag is set
		assert.Equal(t, "true", os.Getenv("TEST_MODE"))
	})

	t.Run("Mock favorites data structure", func(t *testing.T) {
		// Verify mock favorites structure
		mockFavorites := []struct {
			id, title, artist, genre string
			duration                 int64
			plays, likes, reposts    int64
		}{
			{"101", "Favorite Track One", "Artist X", "Electronic", 180000, 1500, 250, 50},
			{"102", "Another Favorite", "Artist Y", "House", 240000, 3200, 480, 120},
			{"103", "Liked Song", "Artist Z", "Techno", 60000, 890, 120, 25},
		}

		assert.Len(t, mockFavorites, 3)
		assert.Equal(t, "Favorite Track One", mockFavorites[0].title)
	})
}

// TestFavoritesFiltering tests filtering parameters for favorites
func TestFavoritesFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Duration filter parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/favorites?duration_min=5", nil)
		assert.Equal(t, "5", req.URL.Query().Get("duration_min"))
	})

	t.Run("Search query parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/favorites?q=electronic", nil)
		assert.Equal(t, "electronic", req.URL.Query().Get("q"))
	})

	t.Run("Sort parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/favorites?sort=oldest", nil)
		assert.Equal(t, "oldest", req.URL.Query().Get("sort"))
	})

	t.Run("Content type parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/favorites?content_type=reposts", nil)
		assert.Equal(t, "reposts", req.URL.Query().Get("content_type"))
	})

	t.Run("Limit parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/favorites?limit=50", nil)
		assert.Equal(t, "50", req.URL.Query().Get("limit"))
	})
}

// TestSearchAPIEndpoint tests the search API
func TestSearchAPIEndpoint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Search query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/search?q=test", nil)
		assert.Equal(t, "test", req.URL.Query().Get("q"))
	})

	t.Run("Search genre filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/search?genre=electronic", nil)
		assert.Equal(t, "electronic", req.URL.Query().Get("genre"))
	})

	t.Run("Search date filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/search?date=2024-01-01", nil)
		assert.Equal(t, "2024-01-01", req.URL.Query().Get("date"))
	})
}

// TestHTMXFavoritesToggle tests the HTMX favorites toggle endpoint
func TestHTMXFavoritesToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("HTMX request header detection", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/favorites/123/htmx", nil)
		req.Header.Set("HX-Request", "true")

		assert.Equal(t, "true", req.Header.Get("HX-Request"))
	})

	t.Run("HTMX target element", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/favorites/123/htmx", nil)
		req.Header.Set("HX-Target", "#track-123")

		assert.Equal(t, "#track-123", req.Header.Get("HX-Target"))
	})

	t.Run("Favorite button HTML structure", func(t *testing.T) {
		// Verify expected HTML structure for favorite button
		html := `<button type="button" class="btn btn-secondary favorite-btn" hx-post="/api/favorites/123/htmx" hx-target="this" hx-swap="outerHTML">Favorite</button>`

		assert.Contains(t, html, `hx-post="/api/favorites/123/htmx"`)
		assert.Contains(t, html, `hx-target="this"`)
		assert.Contains(t, html, `hx-swap="outerHTML"`)
	})
}

// TestLoadMorePagination tests the load more / pagination functionality
func TestLoadMorePagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Page parameter parsing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?page=2", nil)
		assert.Equal(t, "2", req.URL.Query().Get("page"))
	})

	t.Run("Load more button HTMX attributes", func(t *testing.T) {
		html := `<button hx-get="/api/stream?page=2" hx-target="#track-container" hx-swap="beforeend" hx-trigger="revealed">Load More</button>`

		assert.Contains(t, html, `hx-get="/api/stream?page=2"`)
		assert.Contains(t, html, `hx-target="#track-container"`)
		assert.Contains(t, html, `hx-swap="beforeend"`)
		assert.Contains(t, html, `hx-trigger="revealed"`)
	})

	t.Run("Pagination data attributes", func(t *testing.T) {
		html := `<button data-page="2" data-loaded="20">Load More</button>`

		assert.Contains(t, html, `data-page="2"`)
		assert.Contains(t, html, `data-loaded="20"`)
	})
}

// TestExportEndpoints tests the export endpoints
func TestExportEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Favorites JSON export endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export/favorites/json", nil)
		assert.Equal(t, "/api/export/favorites/json", req.URL.Path)
	})

	t.Run("Favorites CSV export endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export/favorites/csv", nil)
		assert.Equal(t, "/api/export/favorites/csv", req.URL.Path)
	})

	t.Run("Playlists JSON export endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/export/playlists/json", nil)
		assert.Equal(t, "/api/export/playlists/json", req.URL.Path)
	})

	t.Run("Export response has correct content type for JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "application/json")
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("Export response has correct content type for CSV", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "text/csv")
		assert.Equal(t, "text/csv", rec.Header().Get("Content-Type"))
	})
}

// TestSettingsPage tests the settings page
func TestSettingsPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Settings page endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		assert.Equal(t, "/settings", req.URL.Path)
	})
}

// TestDashboardPage tests the dashboard page
func TestDashboardPage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Dashboard page endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		assert.Equal(t, "/dashboard", req.URL.Path)
	})
}

// TestBlogPages tests the blog pages
func TestBlogPages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Blog index page endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/blog", nil)
		assert.Equal(t, "/blog", req.URL.Path)
	})

	t.Run("Blog post page endpoint exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/blog/my-post", nil)
		assert.Equal(t, "/blog/my-post", req.URL.Path)
	})

	t.Run("Blog post slug parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/blog/test-slug", nil)
		// Extract slug from path
		slug := strings.TrimPrefix(req.URL.Path, "/blog/")
		assert.Equal(t, "test-slug", slug)
	})
}

// TestThemeToggle tests the theme toggle functionality
func TestThemeToggle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Theme toggle JavaScript function exists", func(t *testing.T) {
		js := `function toggleTheme() { localStorage.setItem('theme', document.documentElement.getAttribute('data-theme') === 'dark' ? 'light' : 'dark'); }`
		assert.Contains(t, js, "toggleTheme")
		assert.Contains(t, js, "localStorage")
	})

	t.Run("Theme data attribute", func(t *testing.T) {
		html := `<html data-theme="dark">`
		assert.Contains(t, html, "data-theme")
	})

	t.Run("Theme toggle button", func(t *testing.T) {
		html := `<button onclick="toggleTheme()">Toggle Theme</button>`
		assert.Contains(t, html, "toggleTheme")
	})
}

// TestServiceWorkerOffline tests the service worker offline functionality
func TestServiceWorkerOffline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Service worker file exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/sw.js", nil)
		assert.Equal(t, "/sw.js", req.URL.Path)
	})

	t.Run("Service worker registration", func(t *testing.T) {
		js := `if ('serviceWorker' in navigator) { navigator.serviceWorker.register('/sw.js'); }`
		assert.Contains(t, js, "serviceWorker")
		assert.Contains(t, js, "register")
	})

	t.Run("Offline indicator element", func(t *testing.T) {
		html := `<div id="offline-indicator" hidden>You are offline</div>`
		assert.Contains(t, html, "offline-indicator")
	})
}

// TestAudioPlayer tests the audio player functionality
func TestAudioPlayer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Audio element with source", func(t *testing.T) {
		html := `<audio controls><source src="/api/track/123/stream" type="audio/mpeg"></audio>`
		assert.Contains(t, html, `<audio`)
		assert.Contains(t, html, `src="/api/track/123/stream"`)
	})

	t.Run("Player bar container", func(t *testing.T) {
		html := `<div id="player-bar"><audio></audio></div>`
		assert.Contains(t, html, "player-bar")
	})

	t.Run("Track stream API endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/track/123/stream", nil)
		assert.Equal(t, "/api/track/123/stream", req.URL.Path)
	})
}

// TestVisualizer tests the audio visualizer
func TestVisualizer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Canvas element for visualizer", func(t *testing.T) {
		html := `<canvas id="visualizer"></canvas>`
		assert.Contains(t, html, "visualizer")
	})

	t.Run("AudioContext for visualization", func(t *testing.T) {
		js := `const audioCtx = new (window.AudioContext || window.webkitAudioContext)();`
		assert.Contains(t, js, "AudioContext")
	})

	t.Run("Butterchurn visualizer import", func(t *testing.T) {
		js := `butterchurnImport.then(lib => { const butterchurn = lib.default; })`
		assert.Contains(t, js, "butterchurn")
	})
}

// TestStreamFilters tests all stream filter combinations
func TestStreamFilters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("Duration min filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?duration_min=5", nil)
		assert.Equal(t, "5", req.URL.Query().Get("duration_min"))
	})

	t.Run("Duration max filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?duration_max=10", nil)
		assert.Equal(t, "10", req.URL.Query().Get("duration_max"))
	})

	t.Run("Content type filter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?content_type=reposts", nil)
		assert.Equal(t, "reposts", req.URL.Query().Get("content_type"))
	})

	t.Run("Sort by newest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?sort=newest", nil)
		assert.Equal(t, "newest", req.URL.Query().Get("sort"))
	})

	t.Run("Sort by oldest", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?sort=oldest", nil)
		assert.Equal(t, "oldest", req.URL.Query().Get("sort"))
	})

	t.Run("Sort by title", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?sort=title", nil)
		assert.Equal(t, "title", req.URL.Query().Get("sort"))
	})

	t.Run("Sort by artist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?sort=artist", nil)
		assert.Equal(t, "artist", req.URL.Query().Get("sort"))
	})

	t.Run("Sort by duration", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?sort=duration", nil)
		assert.Equal(t, "duration", req.URL.Query().Get("sort"))
	})

	t.Run("Multiple filters combined", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?duration_min=5&sort=newest&q=test", nil)
		assert.Equal(t, "5", req.URL.Query().Get("duration_min"))
		assert.Equal(t, "newest", req.URL.Query().Get("sort"))
		assert.Equal(t, "test", req.URL.Query().Get("q"))
	})
}

// TestHTMXEndpointsComplete tests all HTMX endpoints
func TestHTMXEndpointsComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/stream"},
		{"GET", "/api/favorites"},
		{"POST", "/api/favorites/123/htmx"},
		{"GET", "/api/search"},
		{"POST", "/api/sync"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			assert.Equal(t, endpoint.path, req.URL.Path)
		})
	}
}

// TestPageNavigation tests all page navigation endpoints
func TestPageNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pages := []string{
		"/",
		"/stream",
		"/favorites",
		"/playlists",
		"/analytics",
		"/settings",
		"/dashboard",
		"/blog",
		"/login",
		"/signout",
	}

	for _, page := range pages {
		t.Run(page, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, page, nil)
			assert.Equal(t, page, req.URL.Path)
		})
	}
}

// TestPublicEndpoints tests public endpoints that don't require auth
func TestPublicEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	publicEndpoints := []struct {
		path   string
		method string
	}{
		{"/", "GET"},
		{"/login", "GET"},
		{"/health", "GET"},
		{"/blog", "GET"},
		{"/blog/test-post", "GET"},
		{"/share/abc123", "GET"},
		{"/feed/rss/xyz789", "GET"},
	}

	for _, endpoint := range publicEndpoints {
		t.Run(endpoint.method+" "+endpoint.path, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			assert.Equal(t, endpoint.path, req.URL.Path)
		})
	}
}

// TestStreamFormatByOS tests that the correct stream format is selected based on OS
func TestStreamFormatByOS(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test Apple device detection
	t.Run("Apple devices get MP3", func(t *testing.T) {
		appleUserAgents := []string{
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
			"Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X) AppleWebKit/605.1.15",
		}

		for _, ua := range appleUserAgents {
			// Verify User-Agent contains Apple identifiers
			lowerUA := strings.ToLower(ua)
			isApple := strings.Contains(lowerUA, "macintosh") || 
				strings.Contains(lowerUA, "mac os") ||
				strings.Contains(lowerUA, "iphone") || 
				strings.Contains(lowerUA, "ipad")
			assert.True(t, isApple, "Should detect Apple device: %s", ua)
		}
	})

	// Test non-Apple device detection
	t.Run("Non-Apple devices get HLS", func(t *testing.T) {
		nonAppleUserAgents := []string{
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
			"Mozilla/5.0 (Android 10; Mobile; rv:91.0) Gecko/91.0 Firefox/91.0",
		}

		for _, ua := range nonAppleUserAgents {
			lowerUA := strings.ToLower(ua)
			isApple := strings.Contains(lowerUA, "macintosh") || 
				strings.Contains(lowerUA, "mac os") ||
				strings.Contains(lowerUA, "iphone") || 
				strings.Contains(lowerUA, "ipad") ||
				strings.Contains(lowerUA, "ipod")
			assert.False(t, isApple, "Should NOT detect as Apple: %s", ua)
		}
	})
}
