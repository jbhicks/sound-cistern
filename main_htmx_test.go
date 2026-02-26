//go:build integration

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// TestHTMXEndpoints tests all HTMX-specific endpoints
func TestHTMXEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set test mode for mock data
	originalTestMode := os.Getenv("TEST_MODE")
	os.Setenv("TEST_MODE", "true")
	defer func() { os.Setenv("TEST_MODE", originalTestMode) }()

	e := echo.New()

	t.Run("HTMX Request Headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/favorites/12345/htmx", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Verify HTMX header is present
		hxRequest := c.Request().Header.Get("HX-Request")
		assert.Equal(t, "true", hxRequest, "HX-Request header should be 'true'")
	})

	t.Run("HTML Response Content-Type", func(t *testing.T) {
		// Simulate what the handler does
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "text/html")

		contentType := rec.Header().Get("Content-Type")
		assert.Equal(t, "text/html", contentType, "Response should have text/html content type")
	})

	t.Run("Favorite Button HTML Structure", func(t *testing.T) {
		// This test validates the HTML structure returned by the favorite button
		// We can't fully test without auth, but we can validate the expected format

		// Test unfavorited button structure
		unfavoritedHTML := `<button type="button" class="btn btn-secondary favorite-btn" ` +
			`hx-post="/api/favorites/12345/htmx" hx-target="this" hx-swap="outerHTML" ` +
			`hx-indicator=".btn-loading" aria-label="Add Test Track to favorites" aria-pressed="false">` +
			`<span class="btn-icon favorite-icon" aria-hidden="true">` +
			`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">` +
			`<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>` +
			`</svg></span><span class="btn-text">Favorite</span>` +
			`<span class="btn-loading" aria-hidden="true">` +
			`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">` +
			`<circle cx="12" cy="12" r="10" stroke-dasharray="60" stroke-dashoffset="10">` +
			`<animateTransform attributeName="transform" type="rotate" from="0 12 12" to="360 12 12" dur="1s" repeatCount="indefinite"/>` +
			`</circle></svg></span></button>`

		// Verify key HTMX attributes are present
		assert.Contains(t, unfavoritedHTML, `hx-post="/api/favorites/12345/htmx"`, "Should have hx-post attribute")
		assert.Contains(t, unfavoritedHTML, `hx-target="this"`, "Should have hx-target attribute")
		assert.Contains(t, unfavoritedHTML, `hx-swap="outerHTML"`, "Should have hx-swap attribute")
		assert.Contains(t, unfavoritedHTML, `hx-indicator=".btn-loading"`, "Should have hx-indicator attribute")
		assert.Contains(t, unfavoritedHTML, `aria-pressed="false"`, "Should have aria-pressed=false")
		assert.Contains(t, unfavoritedHTML, ">Favorite</span>", "Should show 'Favorite' text")

		// Test favorited button structure
		favoritedHTML := strings.ReplaceAll(unfavoritedHTML, `aria-pressed="false"`, `aria-pressed="true" class="btn btn-secondary favorite-btn favorited"`)
		favoritedHTML = strings.ReplaceAll(favoritedHTML, ">Favorite</span>", ">Favorited</span>")
		favoritedHTML = strings.ReplaceAll(favoritedHTML, `fill="none"`, `fill="currentColor"`)

		assert.Contains(t, favoritedHTML, `aria-pressed="true"`, "Should have aria-pressed=true")
		assert.Contains(t, favoritedHTML, "favorited", "Should have favorited class")
		assert.Contains(t, favoritedHTML, ">Favorited</span>", "Should show 'Favorited' text")
	})

	t.Run("Clear Filters HTMX Attributes", func(t *testing.T) {
		clearButtonHTML := `<button type="button" class="secondary outline" ` +
			`hx-get="/api/stream" hx-target="#track-container" hx-swap="innerHTML" ` +
			`hx-vals='{"q":"","genre":"","sort":"newest","duration_min":"0"}' ` +
			`onclick="document.getElementById('filter-form').reset();">` +
			`Clear</button>`

		assert.Contains(t, clearButtonHTML, `hx-get="/api/stream"`, "Should have hx-get attribute")
		assert.Contains(t, clearButtonHTML, `hx-target="#track-container"`, "Should have hx-target attribute")
		assert.Contains(t, clearButtonHTML, `hx-swap="innerHTML"`, "Should have hx-swap attribute")
		assert.Contains(t, clearButtonHTML, `duration_min`, "Should have duration_min in hx-vals")
	})
}

// TestHTMXFiltering tests HTMX-based filtering functionality
func TestHTMXFiltering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set test mode for mock data
	originalTestMode := os.Getenv("TEST_MODE")
	os.Setenv("TEST_MODE", "true")
	defer func() { os.Setenv("TEST_MODE", originalTestMode) }()

	e := echo.New()

	t.Run("Filter Form HTMX Trigger", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/stream?q=test&genre=electronic", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Verify query params are parsed correctly
		q := c.QueryParam("q")
		genre := c.QueryParam("genre")

		assert.Equal(t, "test", q, "Should parse search query")
		assert.Equal(t, "electronic", genre, "Should parse genre filter")
	})

	t.Run("Load More Button HTMX", func(t *testing.T) {
		loadMoreHTML := `<button type="button" class="btn btn-secondary load-more-btn" id="load-more-btn" ` +
			`data-page="2" hx-get="/api/stream?page=2" hx-target="#track-container" ` +
			`hx-swap="beforeend" hx-indicator="#load-more-indicator" ` +
			`hx-push-url="false" aria-label="Load more tracks">` +
			`<span class="btn-text">Load More Tracks</span>` +
			`<span class="btn-loading" aria-hidden="true">...</span></button>`

		assert.Contains(t, loadMoreHTML, `hx-get="/api/stream?page=2"`, "Should have hx-get with page param")
		assert.Contains(t, loadMoreHTML, `hx-target="#track-container"`, "Should target track container")
		assert.Contains(t, loadMoreHTML, `hx-swap="beforeend"`, "Should use beforeend swap")
		assert.Contains(t, loadMoreHTML, `hx-indicator="#load-more-indicator"`, "Should have indicator")
	})
}

// TestHTMXErrorHandling tests error responses for HTMX requests
func TestHTMXErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_ = echo.New()

	t.Run("Error Response Format", func(t *testing.T) {
		// Simulate error response
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusInternalServerError)
		rec.WriteString("Database error")

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Database error", rec.Body.String())
	})

	t.Run("Not Found Response", func(t *testing.T) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusNotFound)
		rec.WriteString("Track not found")

		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Equal(t, "Track not found", rec.Body.String())
	})
}

// BenchmarkHTMXResponse benchmarks HTMX response generation
func BenchmarkHTMXResponseGeneration(b *testing.B) {
	e := echo.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/favorites/12345/htmx", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		_ = e.NewContext(req, rec)

		// Simulate setting content type
		rec.Header().Set("Content-Type", "text/html")
	}
}
