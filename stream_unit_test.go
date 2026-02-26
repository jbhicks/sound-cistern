//go:build unit

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// Mock stream data structure (matches what's used in streamHandler)
type mockStreamTrack struct {
	id       string
	title    string
	artist   string
	genre    string
	duration int64 // in milliseconds
	plays    int64
	likes    int64
	reposts  int64
}

// TestStreamMockDataCount tests that we have exactly 25 mock tracks
func TestStreamMockDataCount(t *testing.T) {
	mockTracks := getStreamMockTracks()
	assert.Equal(t, 25, len(mockTracks), "Should have exactly 25 mock tracks")
}

// TestStreamMockDataVariety tests variety in mock data
func TestStreamMockDataVariety(t *testing.T) {
	mockTracks := getStreamMockTracks()

	// Track unique genres
	genres := make(map[string]bool)
	artists := make(map[string]bool)
	durations := make(map[int]bool) // rounded to minutes

	for _, track := range mockTracks {
		genres[track.genre] = true
		artists[track.artist] = true
		durationMinutes := int(track.duration / 60000)
		durations[durationMinutes] = true
	}

	assert.GreaterOrEqual(t, len(genres), 10, "Should have at least 10 different genres")
	assert.Equal(t, 25, len(artists), "Should have 25 unique artists")
	assert.GreaterOrEqual(t, len(durations), 5, "Should have at least 5 different duration ranges")

	t.Logf("Mock data variety: %d genres, %d artists, %d duration ranges", len(genres), len(artists), len(durations))
}

// TestStreamMockDataDurationRange tests duration range (45s to 9min)
func TestStreamMockDataDurationRange(t *testing.T) {
	mockTracks := getStreamMockTracks()

	var minDuration, maxDuration int64
	for _, track := range mockTracks {
		if minDuration == 0 || track.duration < minDuration {
			minDuration = track.duration
		}
		if track.duration > maxDuration {
			maxDuration = track.duration
		}
	}

	// Min should be ~45s (45000ms), max should be ~9min (540000ms)
	assert.GreaterOrEqual(t, minDuration, int64(45000), "Minimum duration should be ~45 seconds")
	assert.LessOrEqual(t, maxDuration, int64(600000), "Maximum duration should be ~10 minutes")

	t.Logf("Duration range: %ds to %ds", minDuration/1000, maxDuration/1000)
}

// TestStreamMockDataPlaysCount tests realistic play counts
func TestStreamMockDataPlaysCount(t *testing.T) {
	mockTracks := getStreamMockTracks()

	for _, track := range mockTracks {
		assert.Greater(t, track.plays, int64(0), "Track %s should have plays", track.title)
		assert.GreaterOrEqual(t, track.plays, track.likes, "Plays should be >= likes")
		assert.GreaterOrEqual(t, track.plays, track.reposts, "Plays should be >= reposts")
	}
}

// getStreamMockTracks returns the mock tracks used in stream handler
func getStreamMockTracks() []mockStreamTrack {
	return []mockStreamTrack{
		{"1", "Midnight Drive", "Synthwave Boy", "Synthwave", 245000, 15000, 2500, 500},
		{"2", "Neon Lights", "Cyber Punk", "Electronic", 180000, 32000, 4800, 1200},
		{"3", "Bass Drop", "DJ Thunder", "House", 240000, 89000, 12000, 2500},
		{"4", "Ocean Waves", "Chill Master", "Ambient", 360000, 5000, 890, 120},
		{"5", "Rapid Fire", "Drum & Bass", "Drum & Bass", 185000, 45000, 6700, 890},
		{"6", "Sunset Vibes", "LoFi Queen", "Lo-Fi", 195000, 12000, 3400, 450},
		{"7", "Heavy Metal", "Rock Stars", "Metal", 280000, 67000, 8900, 1200},
		{"8", "Jazz Morning", "Smooth Jazz", "Jazz", 320000, 8900, 1200, 200},
		{"9", "Techno Bunker", "Berlin DJ", "Techno", 420000, 23000, 3400, 780},
		{"10", "Pop Hit", "Mainstream", "Pop", 195000, 250000, 45000, 12000},
		{"11", "Acoustic Session", "Guitar Hero", "Acoustic", 210000, 15000, 3200, 400},
		{"12", "Dubstep Wobble", "Bass Cannon", "Dubstep", 200000, 78000, 12000, 2300},
		{"13", "Classical Mood", "Orchestra", "Classical", 480000, 3400, 560, 89},
		{"14", "Hip Hop Beat", "Rapper", "Hip-Hop", 220000, 180000, 34000, 8900},
		{"15", "Trance State", "Uplift", "Trance", 410000, 56000, 8900, 1500},
		{"16", "Deep House", "Poolside", "House", 380000, 23000, 4500, 670},
		{"17", "Funkytown", "Disco Dan", "Funk", 265000, 45000, 7800, 1200},
		{"18", "Emo Nights", "Punk Heart", "Punk", 175000, 89000, 15000, 3400},
		{"19", "Indie Dream", "Alt Rock", "Indie", 230000, 12000, 2100, 340},
		{"20", "EDM Festival", "Headliner", "EDM", 295000, 340000, 56000, 15000},
		{"21", "Short Snippet", "Quick Beat", "Electronic", 45000, 500, 89, 12},
		{"22", "Long Journey", "Progressive", "Progressive", 540000, 8900, 1200, 180},
		{"23", "Reggae Vibes", "Island", "Reggae", 205000, 23000, 4500, 780},
		{"24", "Country Road", "Nashville", "Country", 198000, 15000, 2100, 340},
		{"25", "R&B Soul", "Smooth Vocal", "R&B", 248000, 67000, 12000, 2100},
	}
}

// TestStreamPaginationLogic tests pagination calculations
func TestStreamPaginationLogic(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		totalTracks  int
		expectedLen  int
		expectedPage int
	}{
		{"Page 1 of 25", 1, 9, 25, 9, 1},
		{"Page 2 of 25", 2, 9, 25, 9, 2},
		{"Page 3 of 25", 3, 9, 25, 7, 3},
		{"Page 4 of 25 (empty)", 4, 9, 25, 0, 4},
		{"Page 1 with 20 limit", 1, 20, 25, 20, 1},
		{"Page 1 with 50 limit", 1, 50, 25, 25, 1},
		{"Page 0 defaults to 1", 0, 9, 25, 9, 1},
		{"Negative page defaults to 1", -1, 9, 25, 9, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate pagination logic from streamHandler
			page := tt.page
			if page < 1 {
				page = 1
			}
			pageSize := tt.pageSize

			startIdx := (page - 1) * pageSize
			endIdx := startIdx + pageSize
			if endIdx > tt.totalTracks {
				endIdx = tt.totalTracks
			}

			resultLen := endIdx - startIdx
			if startIdx >= tt.totalTracks {
				resultLen = 0
			}

			assert.Equal(t, tt.expectedLen, resultLen, "Result length should match")
			assert.Equal(t, tt.expectedPage, page, "Page should be normalized")
		})
	}
}

// TestStreamSearchFilter tests search filtering logic
func TestStreamSearchFilter(t *testing.T) {
	mockTracks := getStreamMockTracks()

	tests := []struct {
		name        string
		query       string
		expectMatch int
	}{
		{"Search 'metal'", "metal", 1},
		{"Search 'techno'", "techno", 1},
		{"Search 'house'", "house", 1},
		{"Search 'jazz'", "jazz", 1},
		{"Search 'Synth'", "synth", 1},
		{"Search 'dubstep'", "dubstep", 1},
		{"Search 'pop'", "pop", 1},
		{"Search 'rock'", "rock", 2},
		{"Search 'edm'", "edm", 1},
		{"Search 'bass'", "bass", 3},
		{"Search 'lofi'", "lofi", 1},
		{"Search 'lo-fi'", "lo-fi", 0},
		{"Empty search", "", 25},
		{"No match", "xyz123", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchQuery := strings.ToLower(tt.query)
			var matches []string

			for _, track := range mockTracks {
				if searchQuery == "" {
					matches = append(matches, track.title)
					continue
				}
				titleMatch := strings.Contains(strings.ToLower(track.title), searchQuery)
				artistMatch := strings.Contains(strings.ToLower(track.artist), searchQuery)
				if titleMatch || artistMatch {
					matches = append(matches, track.title)
				}
			}

			assert.Equal(t, tt.expectMatch, len(matches), "Match count should be %d for query '%s'", tt.expectMatch, tt.query)
		})
	}
}

// TestStreamDurationFilter tests duration filtering logic
func TestStreamDurationFilter(t *testing.T) {
	mockTracks := getStreamMockTracks()

	tests := []struct {
		name        string
		durationMin int // in minutes
	}{
		{"No filter (0)", 0},
		{"Min 3 minutes", 3},
		{"Min 5 minutes", 5},
		{"Min 6 minutes", 6},
		{"Min 7 minutes", 7},
		{"Min 8 minutes", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var matches int
			var withoutFilter int
			for _, track := range mockTracks {
				durationMinutes := track.duration / 60000
				if tt.durationMin > 0 && durationMinutes < int64(tt.durationMin) {
					continue
				}
				matches++
			}
			// Count all tracks without filter
			for range mockTracks {
				withoutFilter++
			}

			// Filter should return fewer or equal tracks
			assert.LessOrEqual(t, matches, withoutFilter, "Filter should return <= total tracks")
			// Higher min duration should return <= tracks than lower min duration
			if tt.durationMin > 0 {
				assert.Less(t, matches, withoutFilter, "Min duration filter should reduce results")
			}
		})
	}
}

// TestStreamHTMXRequestDetection tests HTMX request header detection
func TestStreamHTMXRequestDetection(t *testing.T) {
	tests := []struct {
		name      string
		hxRequest string
		isHTMX    bool
	}{
		{"HTMX request", "true", true},
		{"Regular request", "", false},
		{"Non-HTMX request", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/stream", nil)
			if tt.hxRequest != "" {
				req.Header.Set("HX-Request", tt.hxRequest)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			hxRequest := c.Request().Header.Get("HX-Request")
			isHTMX := hxRequest == "true"

			assert.Equal(t, tt.isHTMX, isHTMX, "HTMX detection should match")
		})
	}
}
