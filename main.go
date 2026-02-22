package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"github.com/jbhicks/sound-cistern/handlers"
	_ "github.com/jbhicks/sound-cistern/pb_migrations"
	"github.com/jbhicks/sound-cistern/views"
	"github.com/jbhicks/sound-cistern/views/components"
)

// Test mode flag - set via environment variable or query parameter (?test_mode=true)
var isTestMode bool = os.Getenv("TEST_MODE") == "true"

// generateCodeVerifier generates a random code verifier for PKCE
func generateCodeVerifier() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate code verifier: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// generateCodeChallenge creates a SHA256 hash of the code verifier for PKCE
func generateCodeChallenge(codeVerifier string) string {
	h := sha256.New()
	h.Write([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// upgradeArtworkURL replaces Soundcloud's default 'large' (100x100) with 't500x500' for better quality
func upgradeArtworkURL(url string) string {
	if url == "" {
		return url
	}
	// Replace any small size with largest available (t500x500)
	url = strings.Replace(url, "-t50x50", "-t500x500", -1)
	url = strings.Replace(url, "-t120x120", "-t500x500", -1)
	url = strings.Replace(url, "-t200x200", "-t500x500", -1)
	url = strings.Replace(url, "-t250x250", "-t500x500", -1)
	url = strings.Replace(url, "-large.", "-t500x500.", 1)
	return url
}

// generateRandomString generates a random string for state parameter
func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	return strings.ReplaceAll(base64.URLEncoding.EncodeToString(b), "=", "")[:length]
}

// createTestUser creates or retrieves a test user for testing
func createTestUser(app *pocketbase.PocketBase) *models.Record {
	usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		log.Printf("Failed to find users collection: %v", err)
		return nil
	}

	// Try to find existing test user
	testUser, err := app.Dao().FindFirstRecordByFilter(
		usersCollection.Id,
		"email = 'test@example.com'",
	)

	if err == nil {
		// Test user exists, return it
		return testUser
	}

	// Create new test user
	testUser = models.NewRecord(usersCollection)
	testUser.Set("email", "test@example.com")
	testUser.Set("username", "testuser")
	testUser.Set("first_name", "Test")
	testUser.Set("last_name", "User")
	testUser.Set("password", "testpassword123")
	testUser.Set("tokenKey", "") // Clear to trigger auto-generation

	if err := app.Dao().SaveRecord(testUser); err != nil {
		// Try to find existing user if creation failed
		existingUser, findErr := app.Dao().FindFirstRecordByFilter(
			usersCollection.Id,
			"email = 'test@example.com'",
		)
		if findErr == nil {
			return existingUser
		}
		log.Printf("Failed to create test user: %v", err)
		return nil
	}

	// Create associated Soundcloud user for testing
	soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
	if err == nil {
		soundcloudUser := models.NewRecord(soundcloudUsersCollection)
		soundcloudUser.Set("soundcloud_id", "testuser")
		soundcloudUser.Set("user_id", testUser.Id)
		soundcloudUser.Set("access_token", "mock_access_token_12345")
		soundcloudUser.Set("expires_at", time.Now().Add(24*time.Hour).Format(time.RFC3339))

		if err := app.Dao().SaveRecord(soundcloudUser); err != nil {
			log.Printf("Failed to create test Soundcloud user: %v", err)
		}
	}

	log.Printf("Created test user: %s", testUser.Id)
	return testUser
}

// testAuthMiddleware automatically authenticates requests in test mode
func testAuthMiddleware(app *pocketbase.PocketBase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isTestMode {
				return next(c)
			}

			// Skip auth for admin routes and health checks
			if strings.HasPrefix(c.Path(), "/_") || c.Path() == "/health" {
				return next(c)
			}

			// In test mode, inject test user as auth record if not already set
			if c.Get(apis.ContextAuthRecordKey) == nil {
				usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err == nil {
					// Find any existing user or use the first one
					users, _ := app.Dao().FindRecordsByFilter(
						usersCollection.Id,
						"1=1",
						"created",
						1,
						0,
						nil,
					)
					if len(users) > 0 {
						c.Set(apis.ContextAuthRecordKey, users[0])
					}
				}
			}

			return next(c)
		}
	}
}

// authRedirectMiddleware redirects unauthenticated users to login page
func authRedirectMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}

// soundcloudAuthMiddleware redirects users without Soundcloud auth to login page
func soundcloudAuthMiddleware(app *pocketbase.PocketBase) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// In test mode, skip auth checks
			if isTestMode {
				return next(c)
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			// Check if user has Soundcloud auth
			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("Failed to find soundcloud_users collection: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			_, err = app.Dao().FindFirstRecordByFilter(
				soundcloudUsersCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				// No Soundcloud auth, redirect to login
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}

			return next(c)
		}
	}
}

// mockSoundcloudMiddleware intercepts Soundcloud API calls in test mode
func mockSoundcloudMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isTestMode {
				return next(c)
			}

			// Intercept external HTTP calls to Soundcloud API
			if strings.Contains(c.Request().URL.String(), "api.soundcloud.com") {
				switch {
				case strings.Contains(c.Request().URL.Path, "/me/activities"):
					return mockSoundcloudActivitiesResponse(c)
				case strings.Contains(c.Request().URL.Path, "/me"):
					return mockSoundcloudUserResponse(c)
				case strings.Contains(c.Request().URL.Path, "/oauth2/token"):
					return mockSoundcloudTokenResponse(c)
				case strings.Contains(c.Request().URL.Path, "/tracks/") && !strings.Contains(c.Request().URL.Path, "/tracks/soundcloud:"):
					// Mock track endpoint - return transcodings for audio streaming
					return mockSoundcloudTrackResponse(c)
				case strings.Contains(c.Request().URL.Path, "/media/playable"):
					// Mock transcoding URL endpoint
					return mockTranscodingResponse(c)
				}
			}

			// Also mock the widget URLs
			if strings.Contains(c.Request().URL.String(), "w.soundcloud.com/player") {
				return c.String(http.StatusOK, "Mock player")
			}

			return next(c)
		}
	}
}

// Mock track response with transcodings for streaming
func mockSoundcloudTrackResponse(c echo.Context) error {
	trackID := c.PathParam("id")
	if trackID == "" {
		// Try to extract from URL path
		trackID = strings.TrimPrefix(c.Request().URL.Path, "/tracks/")
	}

	response := map[string]interface{}{
		"id":         trackID,
		"title":      "Mock Track",
		"duration":   180000,
		"stream_url": "https://example.com/stream",
		"media": map[string]interface{}{
			"transcodings": []map[string]interface{}{
				{
					"preset":   "mp3_320",
					"duration": 180,
					"url":      "https://api.soundcloud.com/media/soundcloud:tracks:" + trackID + "/mp3_320/stream",
				},
				{
					"preset":   "mp3_256",
					"duration": 180,
					"url":      "https://api.soundcloud.com/media/soundcloud:tracks:" + trackID + "/mp3_256/stream",
				},
				{
					"preset":   "mp3_128",
					"duration": 180,
					"url":      "https://api.soundcloud.com/media/soundcloud:tracks:" + trackID + "/mp3_128/stream",
				},
			},
		},
	}
	return c.JSON(http.StatusOK, response)
}

// Mock transcoding URL response - returns the actual stream URL
func mockTranscodingResponse(c echo.Context) error {
	response := map[string]interface{}{
		"url": "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3",
	}
	return c.JSON(http.StatusOK, response)
}

// Mock Soundcloud API responses for testing
func mockSoundcloudActivitiesResponse(c echo.Context) error {
	response := getMockActivities()
	return c.JSON(http.StatusOK, response)
}

func getMockActivities() map[string]interface{} {
	return map[string]interface{}{
		"filters": map[string]interface{}{
			"date_from":    "",
			"date_to":      "",
			"duration_max": "",
			"duration_min": "",
			"favorites":    false,
			"genre":        "",
			"genres":       []interface{}{},
			"search":       "",
			"sort":         "",
		},
		"pagination": map[string]interface{}{
			"has_more":    false,
			"limit":       20,
			"page":        1,
			"total":       7,
			"total_pages": 1,
		},
		"tracks": []map[string]interface{}{
			{
				"artist_name":    "DJ PFunk",
				"artwork_url":    "https://i1.sndcdn.com/artworks-HVGxK1ZyjyfIrh73-ZfBskw-t500x500.jpg",
				"created_at":     "2026-02-16T00:37:06Z",
				"genre":          "Breakbeat",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/paul-barnard-647247988/dj-pfunk-mix-frequencies?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 4180558,
				"track_id":       "2267275367",
				"track_title":    "DJ PFUNK Mix \"Frequencies\"",
			},
			{
				"artist_name":    "Juicy Junglist",
				"artwork_url":    "https://i1.sndcdn.com/artworks-EpKQidrlhQveaMFI-FS3Bfg-t500x500.jpg",
				"created_at":     "2026-02-15T23:47:11Z",
				"genre":          "BASS",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/juicyjunglist/love-burn-2026-juicy-junglist?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 5674083,
				"track_id":       "2267256026",
				"track_title":    "Love Burn 2026 | Juicy Junglist Live at Incendia (Breaks, Bass & UKG)",
			},
			{
				"artist_name":    "The Owl",
				"artwork_url":    "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg",
				"created_at":     "2026-02-15T21:19:29Z",
				"genre":          "Electronic",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/owl_the/owl016-b1-nite-hawk-ride-me?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 118126,
				"track_id":       "2267192987",
				"track_title":    "OWL016   B1. NITE HAWK - Ride Me  (Low Res Snippet)",
			},
			{
				"artist_name":    "SEVEN SEVEN DEUCE RECORDS & ENTERTAINMENT",
				"artwork_url":    "https://i1.sndcdn.com/artworks-uxUS371ew85S3xZs-vQj8mw-t500x500.jpg",
				"created_at":     "2026-02-15T20:18:09Z",
				"genre":          "",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/user-772789439/rufus-du-sol-always-rob?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 378279,
				"track_id":       "2031598860",
				"track_title":    "Rufus Du Sol - Always (Rob Cokeless Remix)",
			},
			{
				"artist_name":    "Nurse Noise",
				"artwork_url":    "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg",
				"created_at":     "2026-02-15T20:12:30Z",
				"genre":          " ",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/nursenoise/setting-loving-boundaries-self?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 624744,
				"track_id":       "2267160449",
				"track_title":    "Setting Loving Boundaries: Self-Regulation for Parents",
			},
			{
				"artist_name":    "🧱Brick Haus",
				"artwork_url":    "https://i1.sndcdn.com/artworks-5zhOEoo4EtToGgYd-BRZBXg-t500x500.jpg",
				"created_at":     "2026-02-15T19:22:13Z",
				"genre":          "Booty Breaks",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/brickhausyyc/sidequest-the-way-she-dance?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 149420,
				"track_id":       "2267138519",
				"track_title":    "SIDEQUEST - The Way She Dance (Brick Haus Edit)",
			},
			{
				"artist_name":    "Tracklistings",
				"artwork_url":    "https://i1.sndcdn.com/artworks-fBACkkg8ainELkaB-NfXKdw-t500x500.jpg",
				"created_at":     "2026-02-15T18:18:28Z",
				"genre":          "TL PREMIERE",
				"is_favorited":   false,
				"permalink_url":  "https://soundcloud.com/tracklistings3-0/tl-premiere-2136?utm_medium=api&utm_campaign=social_sharing&utm_source=id_316535",
				"track_duration": 301871,
				"track_id":       "2267055014",
				"track_title":    "TL PREMIERE : True Self - Late Nite Enthusiast (Saigg Remix) [B L A D E]",
			},
		},
	}
}

func mockSoundcloudUserResponse(c echo.Context) error {
	response := map[string]interface{}{
		"id":           987654321,
		"username":     "testuser",
		"display_name": "Test User",
		"avatar_url":   "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg",
	}

	return c.JSON(http.StatusOK, response)
}

func mockSoundcloudTokenResponse(c echo.Context) error {
	response := map[string]interface{}{
		"access_token":  "mock_access_token_12345",
		"refresh_token": "mock_refresh_token_67890",
		"expires_in":    3600,
		"scope":         "non-expiring",
	}

	return c.JSON(http.StatusOK, response)
}

func main() {
	// getHighQualityStreamURL fetches the best quality stream URL from SoundCloud transcodings
	getHighQualityStreamURL := func(accessToken, trackID string) string {
		// In test mode, return mock stream URL
		if isTestMode {
			return "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"
		}

		client := &http.Client{Timeout: 30 * time.Second}

		log.Printf("[Stream] Fetching transcodings for track %s", trackID)

		req, _ := http.NewRequest("GET", "https://api.soundcloud.com/tracks/"+trackID, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Stream] Error fetching track: %v", err)
			return ""
		}
		defer resp.Body.Close()

		log.Printf("[Stream] Track endpoint status: %d", resp.StatusCode)

		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			log.Printf("[Stream] Token may be expired or invalid - got %d", resp.StatusCode)
			// Can't refresh here - need to return empty and let caller handle refresh
			return ""
		}

		if resp.StatusCode != 200 {
			log.Printf("[Stream] Track fetch returned status %d", resp.StatusCode)
			return ""
		}

		var trackData map[string]interface{}
		if json.NewDecoder(resp.Body).Decode(&trackData) != nil {
			return ""
		}

		media, ok := trackData["media"].(map[string]interface{})
		if !ok {
			log.Printf("[Stream] No 'media' field in track response")
			return ""
		}

		transcodings, ok := media["transcodings"].([]interface{})
		if !ok || len(transcodings) == 0 {
			log.Printf("[Stream] No transcodings found for track %s", trackID)
			log.Printf("[Stream] Media field content: %+v", media)
			return ""
		}

		log.Printf("[Stream] Found %d transcodings for track %s", len(transcodings), trackID)

		var bestTranscodingURL string
		var bestPreset string

		for _, t := range transcodings {
			tc, ok := t.(map[string]interface{})
			if !ok {
				continue
			}

			preset, _ := tc["preset"].(string)
			url, _ := tc["url"].(string)

			log.Printf("[Stream] Transcoding: preset=%s, url=%s", preset, url)

			if url == "" || preset == "" {
				continue
			}

			if bestTranscodingURL == "" {
				bestTranscodingURL = url
				bestPreset = preset
			} else if strings.HasPrefix(preset, "mp3") && !strings.HasPrefix(bestPreset, "mp3") {
				bestTranscodingURL = url
				bestPreset = preset
			} else if strings.HasPrefix(preset, "mp3") && strings.HasPrefix(bestPreset, "mp3") {
				currentBitrate := 0
				bestBitrate := 0
				fmt.Sscanf(preset, "mp3_%d", &currentBitrate)
				fmt.Sscanf(bestPreset, "mp3_%d", &bestBitrate)
				if currentBitrate > bestBitrate {
					bestTranscodingURL = url
					bestPreset = preset
				}
			}
		}

		if bestTranscodingURL == "" {
			return ""
		}

		streamReq, _ := http.NewRequest("GET", bestTranscodingURL, nil)
		streamReq.Header.Set("Authorization", "Bearer "+accessToken)

		log.Printf("[Stream] Requesting stream URL from: %s", bestTranscodingURL)
		streamResp, err := client.Do(streamReq)
		if err != nil {
			log.Printf("[Stream] Stream URL request failed: %v", err)
			return ""
		}
		if streamResp.StatusCode != 200 {
			log.Printf("[Stream] Stream URL request returned: %d", streamResp.StatusCode)
			return ""
		}
		defer streamResp.Body.Close()

		var streamData map[string]interface{}
		if json.NewDecoder(streamResp.Body).Decode(&streamData) != nil {
			return ""
		}

		if url, ok := streamData["url"].(string); ok && url != "" {
			log.Printf("[Stream] Got stream URL: %s", url)
			return url
		}

		log.Printf("[Stream] No URL in stream response: %+v", streamData)
		return ""
	}

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	app := pocketbase.New()

	// In test mode, ensure test user exists before routes are set up
	if isTestMode {
		app.OnAfterBootstrap().Add(func(e *core.BootstrapEvent) error {
			createTestUser(app)
			return nil
		})
	}

	var publicDir string = "./public"

	isGoRun := true
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	jsvm.MustRegister(app, jsvm.Config{})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Test mode middleware (always added - checks isTestMode internally)
		e.Router.Pre(testAuthMiddleware(app))
		e.Router.Pre(mockSoundcloudMiddleware())

		e.Router.Use(middleware.Secure())

		// Auth context loading middleware - reads pb_auth cookie and sets auth record
		e.Router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cookie, err := c.Cookie("pb_auth")
				if err != nil || cookie == nil {
					return next(c)
				}

				tokenString := cookie.Value
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(app.Settings().RecordAuthToken.Secret), nil
				})

				if err != nil || !token.Valid {
					return next(c)
				}

				claims, ok := token.Claims.(jwt.MapClaims)
				if !ok {
					return next(c)
				}

				userID, ok := claims["id"].(string)
				if !ok || userID == "" {
					return next(c)
				}

				usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err != nil {
					return next(c)
				}

				user, err := app.Dao().FindRecordById(usersCollection.Id, userID)
				if err != nil {
					return next(c)
				}

				c.Set(apis.ContextAuthRecordKey, user)
				return next(c)
			}
		})

		// CSRF protection (skip in test mode for easier testing)
		if !isTestMode {
			e.Router.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
				TokenLength: 32,
				TokenLookup: "form:csrf_token",
				Skipper: func(c echo.Context) bool {
					// Skip CSRF for PocketBase admin routes, auth refresh, and API endpoints
					return strings.HasPrefix(c.Path(), "/_") || strings.HasPrefix(c.Path(), "/api/admins/") || strings.HasPrefix(c.Path(), "/api/auth/") || strings.HasPrefix(c.Path(), "/api/sync")
				},
			}))
		}

		// Static file serving (no HTML5 fallback - prevents index.html errors on unknown paths)
		e.Router.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   publicDir,
			Browse: false,
			HTML5:  false,
		}))

		// Health check endpoint
		e.Router.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		})

		// Home page - redirect to stream (default tab)
		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, "/stream")
		}, apis.ActivityLogger(app))

		// Proto page - design experiments
		e.Router.GET("/proto", func(c echo.Context) error {
			// In test mode, skip auth check
			if !isTestMode {
				authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
				if authRecord == nil {
					return c.Redirect(http.StatusTemporaryRedirect, "/login")
				}
				soundcloudUsersCollection, _ := app.Dao().FindCollectionByNameOrId("soundcloud_users")
				if soundcloudUsersCollection != nil {
					_, err := app.Dao().FindFirstRecordByFilter(
						soundcloudUsersCollection.Id,
						"user_id = {:user_id}",
						map[string]any{"user_id": authRecord.Id},
					)
					if err != nil {
						return c.Redirect(http.StatusTemporaryRedirect, "/login")
					}
				}
			}

			tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				log.Printf("Failed to find soundcloud_tracks collection: %v", err)
			}

			var tracks []views.Track
			if tracksCollection != nil {
				records, err := app.Dao().FindRecordsByFilter(
					tracksCollection.Id,
					"stream_url != ''",
					"-created",
					10,
					0,
					map[string]any{},
				)
				if err == nil && len(records) > 0 {
					tracks = make([]views.Track, 0, len(records))
					for _, record := range records {
						artworkURL := upgradeArtworkURL(record.GetString("artwork_url"))
						if artworkURL == "" {
							artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg"
						}
						streamURL := record.GetString("stream_url")
						if streamURL == "" {
							streamURL = "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"
						}

						tracks = append(tracks, views.Track{
							TrackID:          record.GetString("soundcloud_id"),
							TrackTitle:       record.GetString("title"),
							ArtistName:       record.GetString("artist_name"),
							Genre:            record.GetString("genre"),
							TrackDuration:    int64(record.GetInt("length")),
							ArtworkURL:       artworkURL,
							StreamURL:        streamURL,
							PermalinkURL:     record.GetString("permalink_url"),
							PlaybackCount:    int64(record.GetInt("playback_count")),
							FavoritingsCount: int64(record.GetInt("favoritings_count")),
							CommentCount:     int64(record.GetInt("comment_count")),
							RepostsCount:     int64(record.GetInt("reposts_count")),
						})
					}
				} else {
					log.Printf("Proto handler: no records found, err: %v", err)
				}
			}

			if len(tracks) == 0 {
				tracks = []views.Track{
					{
						TrackID:       "1",
						TrackTitle:    "Midnight Drive",
						ArtistName:    "Synthwave Master",
						ArtworkURL:    "https://picsum.photos/seed/track1/500/500",
						TrackDuration: 245,
						StreamURL:     "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3",
					},
				}
			}

			data := views.ProtoData{
				PageData: views.PageData{
					Title:       "Prototype Experiments",
					Description: "10 A3-inspired design variations for track cards",
					CurrentPath: "/proto",
				},
				Tracks: tracks,
			}

			return views.Proto(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Helper function to fetch fresh tracks from SoundCloud API
		fetchSoundCloudTracks := func(accessToken string, limit int) ([]views.Track, error) {
			client := &http.Client{Timeout: 30 * time.Second}
			req, err := http.NewRequest("GET", fmt.Sprintf("https://api.soundcloud.com/me/activities?limit=%d", limit), nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return nil, fmt.Errorf("SoundCloud API error: %d", resp.StatusCode)
			}

			var activities map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
				return nil, err
			}

			var tracks []views.Track
			if collection, ok := activities["collection"].([]interface{}); ok {
				for _, item := range collection {
					if activity, ok := item.(map[string]interface{}); ok {
						origin, hasOrigin := activity["origin"].(map[string]interface{})
						if !hasOrigin {
							continue
						}
						trackType, _ := origin["kind"].(string)
						if trackType != "track" && trackType != "track-repost" {
							continue
						}

						trackID := fmt.Sprintf("%v", origin["id"])
						title, _ := origin["title"].(string)
						artistName, _ := origin["user"].(map[string]interface{})["username"].(string)
						genre, _ := origin["genre"].(string)
						durationMs, _ := origin["duration"].(float64)
						artworkURL, _ := origin["artwork_url"].(string)
						if artworkURL != "" {
							artworkURL = strings.Replace(artworkURL, "-t50x50", "-t500x500", 1)
						}
						permalinkURL, _ := origin["permalink_url"].(string)
						streamURL := fmt.Sprintf("/api/track/%s/stream", trackID)
						playbackCount, _ := origin["playback_count"].(float64)
						favoritingsCount, _ := origin["favoritings_count"].(float64)
						repostsCount, _ := origin["reposts_count"].(float64)

						tracks = append(tracks, views.Track{
							TrackID:          trackID,
							TrackTitle:       title,
							ArtistName:       artistName,
							Genre:            genre,
							TrackDuration:    int64(durationMs),
							ArtworkURL:       artworkURL,
							StreamURL:        streamURL,
							PermalinkURL:     "https://soundcloud.com/" + permalinkURL,
							PlaybackCount:    int64(playbackCount),
							FavoritingsCount: int64(favoritingsCount),
							RepostsCount:     int64(repostsCount),
						})
					}
				}
			}
			return tracks, nil
		}

		// Stream page
		e.Router.GET("/stream", func(c echo.Context) error {
			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			data := views.StreamData{
				PageData: views.PageData{
					Title:       "Stream",
					Description: "Chronological feed of your Soundcloud tracks",
					CurrentPath: "/stream",
					TestMode:    isTestMode,
				},
				Tracks:      []views.Track{},
				ActiveTab:   "stream",
				ViewMode:    "grid",
				Filters:     map[string]interface{}{},
				IsLoggedIn:  "false",
				MinDuration: int64(0),
				MaxDuration: int64(94),
			}

			if authRecord != nil {
				data.IsLoggedIn = "true"
				data.User = authRecord

				settingsCollection, err := app.Dao().FindCollectionByNameOrId("user_settings")
				if err == nil {
					settings, err := app.Dao().FindFirstRecordByFilter(
						settingsCollection.Id,
						"user_id = {:user_id}",
						map[string]any{"user_id": authRecord.Id},
					)
					if err == nil {
						data.ActiveTab = settings.GetString("active_tab")
						if data.ActiveTab == "" {
							data.ActiveTab = "stream"
						}
						data.ViewMode = settings.GetString("view_mode")
						if data.ViewMode == "" {
							data.ViewMode = "grid"
						}
						data.Filters = settings.Get("filters").(map[string]interface{})
						if data.Filters == nil {
							data.Filters = map[string]interface{}{}
						}
					}
				}

				// Pre-load tracks server-side for initial render - fetch fresh from SoundCloud
				soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
				if err == nil {
					soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
						soundcloudUsersCollection.Id,
						"user_id = {:user_id}",
						map[string]any{"user_id": authRecord.Id},
					)
					if err == nil {
						accessToken := soundcloudUser.GetString("access_token")
						// Try to get fresh tracks from SoundCloud API
						tracks, err := fetchSoundCloudTracks(accessToken, 20)
						if err == nil && len(tracks) > 0 {
							log.Printf("[Stream] Fetched %d fresh tracks from SoundCloud API", len(tracks))
							data.Tracks = tracks
						} else {
							if err != nil {
								log.Printf("[Stream] Failed to fetch from SoundCloud: %v", err)
							}
							// Fall back to DB
							log.Printf("[Stream] Falling back to DB tracks")
							soundcloudTracksCollection, _ := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
							if soundcloudTracksCollection != nil {
								records, _ := app.Dao().FindRecordsByFilter(
									soundcloudTracksCollection.Id,
									"user_id = {:user_id}",
									"-created",
									20,
									0,
									map[string]any{"user_id": soundcloudUser.Id},
								)
								if len(records) > 0 {
									favoritesCollection, _ := app.Dao().FindCollectionByNameOrId("favorites")
									dbTracks := make([]views.Track, 0, len(records))
									for _, record := range records {
										_, err := app.Dao().FindFirstRecordByFilter(
											favoritesCollection.Id,
											"user_id = {:user_id} && track_id = {:track_id}",
											map[string]any{"user_id": authRecord.Id, "track_id": record.Id},
										)
										isFavorited := err == nil

										artworkURL := upgradeArtworkURL(record.GetString("artwork_url"))
										if artworkURL == "" {
											artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg"
										}

										dbTracks = append(dbTracks, views.Track{
											TrackID:          record.GetString("soundcloud_id"),
											TrackTitle:       record.GetString("title"),
											ArtistName:       record.GetString("artist_name"),
											Genre:            record.GetString("genre"),
											TrackDuration:    int64(record.GetInt("length")),
											ArtworkURL:       artworkURL,
											StreamURL:        record.GetString("stream_url"),
											PermalinkURL:     record.GetString("permalink_url"),
											PlaybackCount:    int64(record.GetInt("playback_count")),
											FavoritingsCount: int64(record.GetInt("favoritings_count")),
											RepostsCount:     int64(record.GetInt("reposts_count")),
											IsFavorited:      isFavorited,
										})
									}
									data.Tracks = dbTracks
								}
							}
						}
					}
				}
			}

			return views.Stream(data).Render(c.Request().Context(), c.Response().Writer)
		}, soundcloudAuthMiddleware(app), apis.ActivityLogger(app))

		// Login splash page
		e.Router.GET("/login", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Welcome to Sound Cistern",
				Description: "Connect your Soundcloud account to get started",
				CurrentPath: "/login",
				TestMode:    isTestMode,
			}

			return views.LoginSplash(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Sign out - clear auth cookie and redirect to home
		e.Router.GET("/signout", func(c echo.Context) error {
			c.SetCookie(&http.Cookie{
				Name:     "pb_auth",
				Value:    "",
				Path:     "/",
				Domain:   ".jbhicks.dev",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
			})
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}, apis.ActivityLogger(app))

		// Blog index page (without enhanced features)
		e.Router.GET("/blog", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Blog",
				Description: "Latest posts and articles",
				CurrentPath: "/blog",
				TestMode:    isTestMode,
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
			if err != nil {
				return err
			}

			records, err := app.Dao().FindRecordsByFilter(
				postsCollection.Id,
				"published = true",
				"-created",
				100,
				0,
			)
			if err != nil {
				return err
			}

			posts := make([]views.Post, 0, len(records))
			for _, record := range records {
				posts = append(posts, views.Post{
					ID:         record.Id,
					Title:      record.GetString("title"),
					Slug:       record.GetString("slug"),
					Content:    record.GetString("content"),
					Excerpt:    record.GetString("excerpt"),
					Image:      record.GetString("image"),
					ImageAlt:   record.GetString("image_alt"),
					CreatedAt:  record.Created.Time(),
					AuthorID:   record.GetString("author"),
					AuthorName: "",
				})
			}

			return views.BlogIndex(data, posts).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Blog post show page (without enhanced features)
		e.Router.GET("/blog/:slug", func(c echo.Context) error {
			slug := c.PathParam("slug")

			postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
			if err != nil {
				return err
			}

			record, err := app.Dao().FindFirstRecordByFilter(
				postsCollection.Id,
				"slug = {:slug} && published = true",
				map[string]any{"slug": slug},
			)
			if err != nil {
				return apis.NewNotFoundError("Post not found", err)
			}

			authorName := ""
			authorID := record.GetString("author")
			if authorID != "" {
				authorRecord, err := app.Dao().FindRecordById("users", authorID)
				if err == nil {
					firstName := authorRecord.GetString("first_name")
					lastName := authorRecord.GetString("last_name")
					authorName = firstName + " " + lastName
				}
			}

			post := views.Post{
				ID:         record.Id,
				Title:      record.GetString("title"),
				Slug:       record.GetString("slug"),
				Content:    record.GetString("content"),
				Excerpt:    record.GetString("excerpt"),
				Image:      record.GetString("image"),
				ImageAlt:   record.GetString("image_alt"),
				CreatedAt:  record.Created.Time(),
				AuthorID:   authorID,
				AuthorName: authorName,
			}

			// Fetch previous post (newer, created after this one)
			var prevPost *views.Post
			prevRecords, err := app.Dao().FindRecordsByFilter(
				postsCollection.Id,
				"published = true && created > {:created}",
				"-created",
				1,
				0,
				map[string]any{"created": record.Created.Time()},
			)
			if err == nil && len(prevRecords) > 0 {
				prevRecord := prevRecords[0]
				prevPost = &views.Post{
					ID:        prevRecord.Id,
					Title:     prevRecord.GetString("title"),
					Slug:      prevRecord.GetString("slug"),
					CreatedAt: prevRecord.Created.Time(),
				}
			}

			// Fetch next post (older, created before this one)
			var nextPost *views.Post
			nextRecords, err := app.Dao().FindRecordsByFilter(
				postsCollection.Id,
				"published = true && created < {:created}",
				"created",
				1,
				0,
				map[string]any{"created": record.Created.Time()},
			)
			if err == nil && len(nextRecords) > 0 {
				nextRecord := nextRecords[0]
				nextPost = &views.Post{
					ID:        nextRecord.Id,
					Title:     nextRecord.GetString("title"),
					Slug:      nextRecord.GetString("slug"),
					CreatedAt: nextRecord.Created.Time(),
				}
			}

			data := views.PageData{
				Title:       post.Title,
				Description: post.Excerpt,
				CurrentPath: "/blog/" + slug,
				TestMode:    isTestMode,
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.BlogShow(data, post, prevPost, nextPost).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Soundcloud OAuth endpoints
		e.Router.GET("/auth/soundcloud", func(c echo.Context) error {
			clientID := os.Getenv("SOUNDCLOUD_CLIENT_ID")
			redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI")

			if clientID == "" || redirectURI == "" {
				log.Printf("Missing Soundcloud OAuth configuration")
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "OAuth configuration missing",
				})
			}

			// Generate PKCE code verifier and challenge
			codeVerifier := generateCodeVerifier()
			codeChallenge := generateCodeChallenge(codeVerifier)

			// Generate random state for CSRF protection
			state := generateRandomString(32)

			log.Printf("DEBUG: Starting OAuth flow, state=%s", state)

			// Store state and code verifier in oauth_states collection
			oauthStatesCollection, err := app.Dao().FindCollectionByNameOrId("oauth_states")
			if err != nil {
				log.Printf("DEBUG: Failed to find oauth_states collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// Create state record
			stateRecord := models.NewRecord(oauthStatesCollection)
			stateRecord.Set("state", state)
			stateRecord.Set("code_verifier", codeVerifier)
			// No user_id since no auth required
			stateRecord.Set("expires_at", time.Now().Add(10*time.Minute).Format(time.RFC3339))

			log.Printf("DEBUG: About to save state record, state=%s, code_verifier=%s", state, codeVerifier[:10]+"...")
			if err := app.Dao().SaveRecord(stateRecord); err != nil {
				log.Printf("DEBUG: Failed to store OAuth state: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to store OAuth state",
				})
			}
			log.Printf("DEBUG: State record saved successfully, id=%s", stateRecord.Id)

			// Build Soundcloud OAuth URL with PKCE
			authURL := url.URL{
				Scheme: "https",
				Host:   "secure.soundcloud.com",
				Path:   "/authorize",
				RawQuery: url.Values{
					"response_type":         {"code"},
					"client_id":             {clientID},
					"redirect_uri":          {redirectURI},
					"scope":                 {""},
					"state":                 {state},
					"code_challenge":        {codeChallenge},
					"code_challenge_method": {"S256"},
				}.Encode(),
			}

			authURLString := authURL.String()
			log.Printf("Redirecting to Soundcloud OAuth:")
			log.Printf("  Full URL: %s", authURLString)
			log.Printf("  State: %s", state)
			log.Printf("  Code Challenge: %s", codeChallenge)
			log.Printf("  Code Verifier (stored): %s", codeVerifier)
			return c.Redirect(http.StatusTemporaryRedirect, authURLString)
		}, apis.ActivityLogger(app))

		// Soundcloud OAuth callback endpoint
		e.Router.GET("/auth/soundcloud/callback", func(c echo.Context) error {
			// Get query parameters from Soundcloud callback
			code := c.QueryParam("code")
			state := c.QueryParam("state")
			errorParam := c.QueryParam("error")

			// Handle OAuth errors from Soundcloud
			if errorParam != "" {
				errorDesc := c.QueryParam("error_description")
				log.Printf("❌ OAuth error from Soundcloud: %s", errorParam)
				log.Printf("❌ Error description: %s", errorDesc)
				log.Printf("❌ Full callback URL: %s", c.Request().URL.String())
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=oauth_failed&error_msg="+url.QueryEscape(errorParam))
			}

			if code == "" || state == "" {
				log.Printf("Missing required OAuth parameters: code=%s, state=%s", code, state)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=invalid_callback")
			}

			// Find the state record to verify CSRF protection and get code_verifier
			oauthStatesCollection, err := app.Dao().FindCollectionByNameOrId("oauth_states")
			if err != nil {
				log.Printf("Failed to find oauth_states collection: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=server_error")
			}

			// Find record with the state
			stateRecord, err := app.Dao().FindFirstRecordByFilter(
				oauthStatesCollection.Id,
				"state = {:state}",
				map[string]any{"state": state},
			)
			if err != nil {
				log.Printf("Failed to find OAuth state record: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=invalid_state")
			}

			// Get the code verifier
			codeVerifier := stateRecord.GetString("code_verifier")
			if codeVerifier == "" {
				log.Printf("Missing code verifier for state: %s", state)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=missing_verifier")
			}

			// Exchange authorization code for access token
			clientID := os.Getenv("SOUNDCLOUD_CLIENT_ID")
			clientSecret := os.Getenv("SOUNDCLOUD_CLIENT_SECRET")
			redirectURI := os.Getenv("SOUNDCLOUD_REDIRECT_URI")

			if clientID == "" || clientSecret == "" || redirectURI == "" {
				log.Printf("Missing Soundcloud OAuth configuration")
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=config_missing")
			}

			// Prepare token request
			tokenData := url.Values{
				"grant_type":    {"authorization_code"},
				"client_id":     {clientID},
				"client_secret": {clientSecret},
				"redirect_uri":  {redirectURI},
				"code":          {code},
				"code_verifier": {codeVerifier},
			}

			// Make POST request to Soundcloud token endpoint
			resp, err := http.PostForm(
				"https://secure.soundcloud.com/oauth/token",
				tokenData,
			)
			if err != nil {
				log.Printf("Failed to exchange authorization code: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=token_exchange_failed")
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				log.Printf("❌ Token exchange failed with status: %d", resp.StatusCode)
				log.Printf("❌ Response body: %s", string(body))
				log.Printf("❌ Request data: client_id=%s, redirect_uri=%s, code_verifier=%s",
					clientID[:10]+"...", redirectURI, codeVerifier[:10]+"...")
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=token_exchange_failed")
			}

			// Parse token response
			var tokenResponse struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int64  `json:"expires_in"`
				Scope        string `json:"scope"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
				log.Printf("Failed to decode token response: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=token_decode_failed")
			}

			// Get user information from Soundcloud API
			userReq, err := http.NewRequest(
				"GET",
				"https://api.soundcloud.com/me",
				nil,
			)
			if err != nil {
				log.Printf("Failed to create user info request: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_failed")
			}

			userReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

			client := &http.Client{Timeout: 10 * time.Second}
			userResp, err := client.Do(userReq)
			if err != nil {
				log.Printf("Failed to fetch user info: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_failed")
			}
			defer userResp.Body.Close()

			if userResp.StatusCode != http.StatusOK {
				log.Printf("User info request failed with status: %d", userResp.StatusCode)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_failed")
			}

			// Parse user info response
			var userInfo struct {
				ID          int64  `json:"id"`
				Username    string `json:"username"`
				DisplayName string `json:"display_name"`
				AvatarURL   string `json:"avatar_url"`
			}

			if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
				log.Printf("Failed to decode user info: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_info_decode_failed")
			}

			log.Printf("Soundcloud user info: ID=%d, Username=%s", userInfo.ID, userInfo.Username)

			// Check if Soundcloud user already exists (try both username and numeric ID)
			authCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("Failed to find soundcloud_users collection: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=server_error")
			}

			soundcloudID := fmt.Sprintf("%d", userInfo.ID)
			existingUser, err := app.Dao().FindFirstRecordByFilter(
				authCollection.Id,
				"soundcloud_id = {:soundcloud_id}",
				map[string]any{"soundcloud_id": soundcloudID},
			)

			var user *models.Record
			if err != nil {
				// Soundcloud user does not exist, create new PocketBase user
				usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err != nil {
					log.Printf("Failed to find users collection: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=server_error")
				}

				user = models.NewRecord(usersCollection)
				user.Set("email", userInfo.Username+"@soundcloud.local")
				user.Set("password", "soundcloud_default") // In production, generate secure password
				user.Set("first_name", userInfo.DisplayName)
				user.Set("username", userInfo.Username)

				if err := app.Dao().SaveRecord(user); err != nil {
					log.Printf("Failed to create PocketBase user: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_create_failed")
				}

				log.Printf("Created new PocketBase user: %s for Soundcloud user: %s", user.Id, userInfo.Username)
			} else {
				// Get the associated PocketBase user
				userID := existingUser.GetString("user_id")
				usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
				if err != nil {
					log.Printf("Failed to find users collection: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=server_error")
				}

				user, err = app.Dao().FindRecordById(usersCollection.Id, userID)
				if err != nil {
					log.Printf("Failed to find PocketBase user: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=user_not_found")
				}
			}

			// Create or update Soundcloud user record
			if err != nil {
				// Create new Soundcloud user record
				soundcloudUser := models.NewRecord(authCollection)
				soundcloudUser.Set("soundcloud_id", soundcloudID)
				soundcloudUser.Set("user_id", user.Id)
				soundcloudUser.Set("access_token", tokenResponse.AccessToken)
				if tokenResponse.RefreshToken != "" {
					soundcloudUser.Set("refresh_token", tokenResponse.RefreshToken)
				}
				// Set expiration time
				expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
				soundcloudUser.Set("expires_at", expiresAt.Format(time.RFC3339))

				if err := app.Dao().SaveRecord(soundcloudUser); err != nil {
					log.Printf("Failed to save Soundcloud user: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=save_user_failed")
				}

				log.Printf("Created new Soundcloud user: %s for PocketBase user: %s", userInfo.Username, user.Id)
			} else {
				// Update existing Soundcloud user record
				existingUser.Set("user_id", user.Id)
				existingUser.Set("access_token", tokenResponse.AccessToken)
				if tokenResponse.RefreshToken != "" {
					existingUser.Set("refresh_token", tokenResponse.RefreshToken)
				}
				// Update expiration time
				expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
				existingUser.Set("expires_at", expiresAt.Format(time.RFC3339))

				if err := app.Dao().SaveRecord(existingUser); err != nil {
					log.Printf("Failed to update Soundcloud user: %v", err)
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=update_user_failed")
				}

				log.Printf("Updated existing Soundcloud user: %s for PocketBase user: %s", userInfo.Username, user.Id)
			}

			// Set auth context for the session
			c.Set(apis.ContextAuthRecordKey, user)

			tokenDuration := 30 * 24 * time.Hour // Default 30 days
			if dur := os.Getenv("AUTH_TOKEN_DURATION_HOURS"); dur != "" {
				if hours, err := strconv.Atoi(dur); err == nil && hours > 0 {
					tokenDuration = time.Duration(hours) * time.Hour
				}
			}

			// Generate JWT token for auth cookie
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":   user.Id,
				"type": "authRecord",
				"exp":  time.Now().Add(tokenDuration).Unix(),
			})
			tokenString, err := token.SignedString([]byte(app.Settings().RecordAuthToken.Secret))
			if err != nil {
				log.Printf("Failed to generate auth token: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=token_gen_failed")
			}

			// Set auth cookie - make it work across all jbhicks.dev subdomains
			c.SetCookie(&http.Cookie{
				Name:     "pb_auth",
				Value:    tokenString,
				Path:     "/",
				Domain:   ".jbhicks.dev",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   int(tokenDuration.Seconds()),
			})

			// Clean up the state record
			if err := app.Dao().DeleteRecord(stateRecord); err != nil {
				log.Printf("Warning: Failed to clean up state record: %v", err)
			}

			// Auto-sync tracks from Soundcloud on first login
			log.Printf("[AutoSync] Starting auto-sync for user %s after OAuth", user.Id)
			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("[AutoSync] ERROR: Failed to find soundcloud_users collection: %v", err)
			} else {
				soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
					soundcloudUsersCollection.Id,
					"user_id = {:user_id}",
					map[string]any{"user_id": user.Id},
				)
				if err != nil {
					log.Printf("[AutoSync] ERROR: No soundcloud_user found for user %s: %v", user.Id, err)
				} else if soundcloudUser == nil {
					log.Printf("[AutoSync] ERROR: soundcloudUser is nil for user %s", user.Id)
				} else {
					accessToken := soundcloudUser.GetString("access_token")
					if accessToken == "" {
						log.Printf("[AutoSync] ERROR: Access token is empty for user %s", user.Id)
					} else if isTestMode {
						log.Printf("[AutoSync] Skipping: Test mode is enabled")
					} else {
						log.Printf("[AutoSync] Fetching activities from SoundCloud for user %s", user.Id)
						client := &http.Client{Timeout: 30 * time.Second}
						req, _ := http.NewRequest("GET", "https://api.soundcloud.com/me/activities?limit=100", nil)
						req.Header.Set("Authorization", "Bearer "+accessToken)
						if resp, err := client.Do(req); err != nil {
							log.Printf("[AutoSync] ERROR: Failed to fetch activities from SoundCloud: %v", err)
						} else {
							defer resp.Body.Close()
							if resp.StatusCode != 200 {
								body, _ := io.ReadAll(resp.Body)
								log.Printf("[AutoSync] ERROR: SoundCloud API returned status %d: %s", resp.StatusCode, string(body))
							} else {
								var activities map[string]interface{}
								if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
									log.Printf("[AutoSync] ERROR: Failed to decode activities response: %v", err)
								} else {
									tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
									if err != nil {
										log.Printf("[AutoSync] ERROR: Failed to find soundcloud_tracks collection: %v", err)
									} else if tracksCollection == nil {
										log.Printf("[AutoSync] ERROR: tracksCollection is nil")
									} else {
										savedCount := 0
										existingCount := 0
										errorCount := 0

										// Parse Soundcloud activities response structure
										// The /me/activities endpoint returns { "collection": [ { "type": "track", "origin": { track data } } ] }
										collection, hasCollection := activities["collection"].([]interface{})
										if !hasCollection {
											log.Printf("[AutoSync] WARNING: No 'collection' field in activities response")
										} else {
											for _, item := range collection {
												if activity, ok := item.(map[string]interface{}); ok {
													// Extract track data from the origin field
													origin, hasOrigin := activity["origin"].(map[string]interface{})
													if !hasOrigin {
														continue
													}

													// Get track ID
													var soundcloudID string
													if id, ok := origin["id"].(float64); ok {
														soundcloudID = fmt.Sprintf("%.0f", id)
													} else if id, ok := origin["id"].(string); ok {
														soundcloudID = id
													} else {
														continue
													}

													// Check if track already exists
													existing, _ := app.Dao().FindFirstRecordByFilter(
														tracksCollection.Id,
														"user_id = {:user_id} && soundcloud_id = {:soundcloud_id}",
														map[string]any{"user_id": soundcloudUser.Id, "soundcloud_id": soundcloudID},
													)
													if existing != nil {
														existingCount++
														continue
													}

													// Create track record
													trackRecord := models.NewRecord(tracksCollection)
													trackRecord.Set("user_id", soundcloudUser.Id)
													trackRecord.Set("soundcloud_id", soundcloudID)

													// Title
													if title, ok := origin["title"].(string); ok {
														trackRecord.Set("title", title)
													}

													// Duration
													if duration, ok := origin["duration"].(float64); ok {
														trackRecord.Set("length", int64(duration))
													}

													// Artist name (from user object)
													if user, ok := origin["user"].(map[string]interface{}); ok {
														if username, ok := user["username"].(string); ok {
															trackRecord.Set("artist_name", username)
														}
													}

													// Genre
													if genre, ok := origin["genre"].(string); ok {
														trackRecord.Set("genre", genre)
													}

													// Permalink URL
													if permalink, ok := origin["permalink_url"].(string); ok {
														trackRecord.Set("permalink_url", permalink)
													}

													// Artwork URL
													if artwork, ok := origin["artwork_url"].(string); ok {
														trackRecord.Set("artwork_url", artwork)
													}

													// Stream URL for playback
													if streamURL, ok := origin["stream_url"].(string); ok {
														trackRecord.Set("stream_url", streamURL)
													}

													// Created at (from activity, not origin)
													if createdAt, ok := activity["created_at"].(string); ok {
														trackRecord.Set("post_time", createdAt)
													}

													if app.Dao().SaveRecord(trackRecord) == nil {
														savedCount++
													} else {
														errorCount++
														log.Printf("[AutoSync] ERROR: Failed to save track %s: %v", soundcloudID, err)
													}
												}
											}
											log.Printf("[AutoSync] Complete: saved=%d, existing=%d, errors=%d for user %s", savedCount, existingCount, errorCount, user.Id)
										}
									}
								}
							}
						}
					}
				}
			}

			// Redirect to soundcistern subdomain
			log.Printf("OAuth flow completed successfully for Soundcloud user: %s", userInfo.Username)
			return c.Redirect(http.StatusTemporaryRedirect, "https://soundcistern.jbhicks.dev/")
		}, apis.ActivityLogger(app))

		// API routes (protected)
		e.Router.GET("/api/user", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"id":    authRecord.Id,
				"email": authRecord.GetString("email"),
				"name":  authRecord.GetString("first_name") + " " + authRecord.GetString("last_name"),
			})
		}, apis.RequireRecordAuth())

		// Token refresh endpoint - refresh the JWT using stored SoundCloud refresh token
		// Works even with expired JWT by parsing the cookie manually
		e.Router.POST("/api/auth/refresh", func(c echo.Context) error {
			log.Printf("Refresh endpoint called, cookie: %v", c.Cookies())
			// Parse token from cookie without validation (to get user ID even if expired)
			cookie, err := c.Cookie("pb_auth")
			if err != nil || cookie == nil {
				log.Printf("No cookie found: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "No auth cookie found",
				})
			}

			tokenString := cookie.Value
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(app.Settings().RecordAuthToken.Secret), nil
			})

			// Even if token is expired, we can still get the claims
			if err != nil {
				// Try to extract claims from expired token
				if token != nil {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if userID, ok := claims["id"].(string); ok && userID != "" {
							// Proceed with user ID from expired token
							goto foundUser
						}
					}
				}
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

		foundUser:
			claims, _ := token.Claims.(jwt.MapClaims)
			userID, _ := claims["id"].(string)

			usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database error",
				})
			}

			authRecord, err := app.Dao().FindRecordById(usersCollection.Id, userID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not found",
				})
			}

			// Look up the soundcloud_users record to get the refresh token
			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database error",
				})
			}

			soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
				soundcloudUsersCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": userID},
			)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "No SoundCloud connection found",
				})
			}

			refreshToken := soundcloudUser.GetString("refresh_token")
			if refreshToken == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "No refresh token available",
				})
			}

			// Use the refresh token to get new SoundCloud tokens
			tokenURL := "https://api.soundcloud.com/oauth2/token"
			data := url.Values{}
			data.Set("client_id", os.Getenv("SOUNDCLOUD_CLIENT_ID"))
			data.Set("client_secret", os.Getenv("SOUNDCLOUD_CLIENT_SECRET"))
			data.Set("grant_type", "refresh_token")
			data.Set("refresh_token", refreshToken)

			resp, err := http.PostForm(tokenURL, data)
			if err != nil || resp.StatusCode != 200 {
				log.Printf("Failed to refresh SoundCloud token: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Failed to refresh authentication",
				})
			}
			defer resp.Body.Close()

			var tokenResponse struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int    `json:"expires_in"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
				log.Printf("Failed to decode token response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to parse token response",
				})
			}

			// Update stored tokens and expiration
			soundcloudUser.Set("access_token", tokenResponse.AccessToken)
			if tokenResponse.RefreshToken != "" {
				soundcloudUser.Set("refresh_token", tokenResponse.RefreshToken)
			}
			expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
			soundcloudUser.Set("expires_at", expiresAt.Format(time.RFC3339))
			if err := app.Dao().SaveRecord(soundcloudUser); err != nil {
				log.Printf("Warning: Failed to save updated SoundCloud tokens: %v", err)
			}

			// Generate new JWT
			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"id":   authRecord.Id,
				"type": "authRecord",
				"exp":  time.Now().Add(24 * time.Hour).Unix(),
			})
			tokenString, err = newToken.SignedString([]byte(app.Settings().RecordAuthToken.Secret))
			if err != nil {
				log.Printf("Failed to generate auth token: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to generate authentication token",
				})
			}

			// Set new auth cookie
			c.SetCookie(&http.Cookie{
				Name:     "pb_auth",
				Value:    tokenString,
				Path:     "/",
				Domain:   ".jbhicks.dev",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   86400,
			})

			return c.JSON(http.StatusOK, map[string]string{
				"status": "refreshed",
			})
		}, apis.RequireRecordAuth())

		// Stream endpoint - fetch user's Soundcloud tracks with filtering, sorting, and pagination
		streamHandler := func(c echo.Context) error {
			// In test mode, return mock HTML tracks
			if isTestMode {
				durationMinStr := c.QueryParam("duration_min")
				durationMin := 0
				if d, err := strconv.Atoi(durationMinStr); err == nil {
					durationMin = d
				}

				searchQuery := strings.ToLower(c.QueryParam("q"))

				var htmlBuilder strings.Builder
				htmlBuilder.WriteString("<div class=\"stream-flip-grid\">")
				mockTracks := []struct {
					id, title, artist, genre string
					duration                 int64
					plays, likes, reposts    int64
				}{
					{"1", "Test Track One", "Artist A", "Electronic", 180000, 1500, 250, 50},
					{"2", "Another Track", "Artist B", "House", 240000, 3200, 480, 120},
					{"3", "Short Track", "Artist C", "Techno", 60000, 890, 120, 25},
				}
				for _, t := range mockTracks {
					// Filter by duration (convert ms to minutes)
					durationMinutes := t.duration / 60000
					if durationMin > 0 && durationMinutes < int64(durationMin) {
						continue
					}
					// Filter by search query
					if searchQuery != "" {
						titleMatch := strings.Contains(strings.ToLower(t.title), searchQuery)
						artistMatch := strings.Contains(strings.ToLower(t.artist), searchQuery)
						if !titleMatch && !artistMatch {
							continue
						}
					}
					artworkURL := "https://picsum.photos/seed/" + t.id + "/500/500"
					components.StreamFlipCard(
						t.id,
						t.title,
						t.artist,
						t.genre,
						t.duration,
						artworkURL,
						t.plays,
						t.likes,
						t.reposts,
					).Render(c.Request().Context(), &htmlBuilder)
				}
				htmlBuilder.WriteString("</div>")
				return c.HTML(http.StatusOK, htmlBuilder.String())
			}

			// This endpoint requires authentication via apis.RequireRecordAuth()
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			// Parse query parameters
			log.Printf("[Stream] Request from user %s: q=%s, genre=%s, sort=%s, favorites=%s, page=%s, limit=%s",
				authRecord.Id,
				c.QueryParam("q"),
				c.QueryParam("genre"),
				c.QueryParam("sort"),
				c.QueryParam("favorites"),
				c.QueryParam("page"),
				c.QueryParam("limit"))
			page := 1
			if p := c.QueryParam("page"); p != "" {
				if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
					page = parsed
				}
			}

			limit := 20
			if l := c.QueryParam("limit"); l != "" {
				if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
					limit = parsed
				}
			}

			searchQuery := c.QueryParam("q")
			genreFilter := c.QueryParam("genre")
			sortBy := c.QueryParam("sort") // newest, oldest, title, artist, duration
			dateFrom := c.QueryParam("date_from")
			dateTo := c.QueryParam("date_to")
			favoritesOnly := c.QueryParam("favorites") == "true"

			// Duration filter (in minutes, convert to milliseconds for DB)
			durationMin := c.QueryParam("duration_min")
			durationMax := c.QueryParam("duration_max")

			// Get soundcloud_tracks collection
			soundcloudTracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				log.Printf("Failed to find soundcloud_tracks collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// First find the soundcloud_users record linked to this auth user
			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("Failed to find soundcloud_users collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
				soundcloudUsersCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				log.Printf("No soundcloud_user found for user %s: %v", authRecord.Id, err)
				return c.JSON(http.StatusOK, map[string]interface{}{
					"tracks": []interface{}{},
					"total":  0,
					"page":   page,
					"limit":  limit,
					"genres": []string{},
				})
			}

			// Build filter
			filter := "user_id = {:user_id}"
			params := map[string]any{"user_id": soundcloudUser.Id}

			if searchQuery != "" {
				filter += " && (title ~ {:search} || artist_name ~ {:search} || genre ~ {:search})"
				params["search"] = searchQuery
			}

			if genreFilter != "" && genreFilter != "all" {
				filter += " && genre = {:genre}"
				params["genre"] = genreFilter
			}

			if dateFrom != "" {
				filter += " && post_time >= {:date_from}"
				params["date_from"] = dateFrom
			}

			if dateTo != "" {
				filter += " && post_time <= {:date_to}"
				params["date_to"] = dateTo
			}

			// Duration filter (input in minutes, stored in milliseconds)
			if durationMin != "" {
				if minDur, err := strconv.Atoi(durationMin); err == nil && minDur > 0 {
					filter += " && length >= {:duration_min}"
					params["duration_min"] = minDur * 60 * 1000 // convert minutes to ms
				}
			}

			if durationMax != "" {
				if maxDur, err := strconv.Atoi(durationMax); err == nil && maxDur > 0 {
					filter += " && length <= {:duration_max}"
					params["duration_max"] = maxDur * 60 * 1000 // convert minutes to ms
				}
			}

			// Determine sort order
			sortOrder := "-post_time" // default: newest first
			if sortBy == "oldest" {
				sortOrder = "post_time"
			} else if sortBy == "title" {
				sortOrder = "title"
			} else if sortBy == "artist" {
				sortOrder = "artist_name"
			} else if sortBy == "duration" {
				sortOrder = "length"
			}

			// Count total matching tracks (for pagination)
			var totalCount int64
			countQuery := fmt.Sprintf("SELECT COUNT(*) as count FROM %s WHERE %s", soundcloudTracksCollection.Name, filter)
			if err := app.DB().NewQuery(countQuery).Bind(params).Row(&totalCount); err != nil {
				log.Printf("Warning: Failed to count tracks: %v", err)
			}

			// Query tracks with pagination
			offset := (page - 1) * limit
			records, err := app.Dao().FindRecordsByFilter(
				soundcloudTracksCollection.Id,
				filter,
				sortOrder,
				limit,
				offset,
				params,
			)
			if err != nil {
				log.Printf("Failed to fetch tracks: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to fetch tracks",
				})
			}

			// If favorites only, filter in memory
			if favoritesOnly {
				favoritesCollection, _ := app.Dao().FindCollectionByNameOrId("favorites")
				filteredRecords := make([]*models.Record, 0)
				for _, record := range records {
					_, err := app.Dao().FindFirstRecordByFilter(
						favoritesCollection.Id,
						"user_id = {:user_id} && track_id = {:track_id}",
						map[string]any{"user_id": authRecord.Id, "track_id": record.Id},
					)
					if err == nil {
						filteredRecords = append(filteredRecords, record)
					}
				}
				records = filteredRecords
				totalCount = int64(len(records))
			}

			// Get all unique genres for filter dropdown
			genres := []string{}
			genreQuery := fmt.Sprintf("SELECT DISTINCT genre FROM %s WHERE user_id = {:user_id} AND genre != '' ORDER BY genre", soundcloudTracksCollection.Name)
			genreRows, err := app.DB().NewQuery(genreQuery).Rows()
			if err == nil {
				for genreRows.Next() {
					var genre string
					if genreRows.Scan(&genre); genre != "" {
						genres = append(genres, genre)
					}
				}
				genreRows.Close()
			}

			log.Printf("[Stream] Success: returned %d tracks for user %s", len(records), authRecord.Id)

			var htmlBuilder strings.Builder
			htmlBuilder.WriteString("<div class=\"stream-flip-grid\">")
			for _, record := range records {
				artworkURL := upgradeArtworkURL(record.GetString("artwork_url"))
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg"
				}

				components.StreamFlipCard(
					record.GetString("soundcloud_id"),
					record.GetString("title"),
					record.GetString("artist_name"),
					record.GetString("genre"),
					int64(record.GetInt("length")),
					artworkURL,
					int64(record.GetInt("playback_count")),
					int64(record.GetInt("favoritings_count")),
					int64(record.GetInt("reposts_count")),
				).Render(c.Request().Context(), &htmlBuilder)
			}
			htmlBuilder.WriteString("</div>")

			return c.HTML(http.StatusOK, htmlBuilder.String())
		}
		if isTestMode {
			e.Router.GET("/api/stream", streamHandler)
		} else {
			e.Router.GET("/api/stream", streamHandler, apis.RequireRecordAuth())
		}

		// Sync endpoint - fetch tracks from SoundCloud and save to database
		e.Router.POST("/api/sync", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			var activities map[string]interface{}

			// In test mode, use mock data directly
			if isTestMode {
				activities = getMockActivities()
			} else {
				soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
				if err != nil {
					log.Printf("Failed to find soundcloud_users collection: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
				}

				soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
					soundcloudUsersCollection.Id,
					"user_id = {:user_id}",
					map[string]any{"user_id": authRecord.Id},
				)
				if err != nil {
					log.Printf("[Sync] ERROR: Could not find soundcloud_users for user_id=%s, err=%v", authRecord.Id, err)
					return c.JSON(http.StatusNotFound, map[string]string{"error": "SoundCloud account not linked"})
				}

				accessToken := soundcloudUser.GetString("access_token")
				refreshToken := soundcloudUser.GetString("refresh_token")
				scUserID := soundcloudUser.GetString("soundcloud_id")
				log.Printf("[Sync] Using token for soundcloud_id=%s, has_access_token=%v, has_refresh_token=%v", scUserID, accessToken != "", refreshToken != "")

				if accessToken == "" && refreshToken == "" {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No access token"})
				}

				client := &http.Client{Timeout: 30 * time.Second}
				req, err := http.NewRequest("GET", "https://api.soundcloud.com/me/activities?limit=50", nil)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
				}
				req.Header.Set("Authorization", "Bearer "+accessToken)

				log.Printf("[Sync] Making request to SoundCloud API...")
				resp, err := client.Do(req)
				if err != nil {
					log.Printf("[Sync] ERROR: Failed to fetch SoundCloud activities: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch tracks from SoundCloud"})
				}
				defer resp.Body.Close()
				log.Printf("[Sync] SoundCloud API responded with status: %d", resp.StatusCode)

				if resp.StatusCode != 200 {
					body, _ := io.ReadAll(resp.Body)
					log.Printf("SoundCloud API error: %d - %s", resp.StatusCode, string(body))
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "SoundCloud API error"})
				}

				if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
					log.Printf("Failed to decode SoundCloud response: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse SoundCloud response"})
				}
			}

			tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
			}

			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
			}

			soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
				soundcloudUsersCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "No Soundcloud account linked"})
			}

			savedCount := 0
			var tracks []map[string]interface{}

			// Parse Soundcloud activities response structure
			// The /me/activities endpoint returns { "collection": [ { "type": "track", "origin": { track data } } ] }
			if collection, ok := activities["collection"].([]interface{}); ok {
				for _, item := range collection {
					if activity, ok := item.(map[string]interface{}); ok {
						// Extract track data from the origin field
						origin, hasOrigin := activity["origin"].(map[string]interface{})
						if !hasOrigin {
							continue
						}

						// Get track ID
						var soundcloudID string
						if id, ok := origin["id"].(float64); ok {
							soundcloudID = fmt.Sprintf("%.0f", id)
						} else if id, ok := origin["id"].(string); ok {
							soundcloudID = id
						} else {
							continue
						}

						existing, _ := app.Dao().FindFirstRecordByFilter(
							tracksCollection.Id,
							"user_id = {:user_id} && soundcloud_id = {:soundcloud_id}",
							map[string]any{"user_id": soundcloudUser.Id, "soundcloud_id": soundcloudID},
						)
						if existing != nil {
							continue
						}

						trackRecord := models.NewRecord(tracksCollection)
						trackRecord.Set("user_id", soundcloudUser.Id)
						trackRecord.Set("soundcloud_id", soundcloudID)

						// Title
						if title, ok := origin["title"].(string); ok {
							trackRecord.Set("title", title)
						}

						// Duration
						if duration, ok := origin["duration"].(float64); ok {
							trackRecord.Set("length", int64(duration))
						}

						// Artist name (from user object)
						if user, ok := origin["user"].(map[string]interface{}); ok {
							if username, ok := user["username"].(string); ok {
								trackRecord.Set("artist_name", username)
							}
						}

						// Genre
						if genre, ok := origin["genre"].(string); ok {
							trackRecord.Set("genre", genre)
						}

						// Permalink URL
						if permalink, ok := origin["permalink_url"].(string); ok {
							trackRecord.Set("permalink_url", permalink)
						}

						// Artwork URL
						if artwork, ok := origin["artwork_url"].(string); ok {
							trackRecord.Set("artwork_url", artwork)
						}

						// Created at (from activity, not origin)
						if createdAt, ok := activity["created_at"].(string); ok {
							trackRecord.Set("post_time", createdAt)
						}

						if err := app.Dao().SaveRecord(trackRecord); err != nil {
							log.Printf("Failed to save track: %v", err)
							continue
						}
						savedCount++

						// Add to tracks list for response
						tracks = append(tracks, origin)
					}
				}
			}

			log.Printf("Synced %d tracks for user %s", savedCount, authRecord.Id)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"synced": savedCount,
				"total":  len(tracks),
			})
		}, apis.RequireRecordAuth())

		// Search endpoint - search user's tracks by title, artist, genre
		e.Router.GET("/api/search", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			// Get query parameters
			searchQuery := c.QueryParam("q")
			genreFilter := c.QueryParam("genre")
			dateFilter := c.QueryParam("date")

			// Get soundcloud_tracks collection
			soundcloudTracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				log.Printf("Failed to find soundcloud_tracks collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// Build filter
			filter := "user_id = {:user_id}"
			params := map[string]any{"user_id": authRecord.Id}

			if searchQuery != "" {
				filter += " && (title ~ {:search} || artist_name ~ {:search} || genre ~ {:search})"
				params["search"] = searchQuery
			}

			if genreFilter != "" {
				filter += " && genre = {:genre}"
				params["genre"] = genreFilter
			}

			if dateFilter != "" {
				filter += " && post_time >= {:date}"
				params["date"] = dateFilter
			}

			// Query tracks
			records, err := app.Dao().FindRecordsByFilter(
				soundcloudTracksCollection.Id,
				filter,
				"-post_time",
				1000, // Large limit for search
				0,
				params,
			)
			if err != nil {
				log.Printf("Failed to search tracks: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to search tracks",
				})
			}

			// Build response
			tracks := make([]map[string]interface{}, 0, len(records))
			for _, record := range records {
				// Check if favorited
				favoritesCollection, _ := app.Dao().FindCollectionByNameOrId("favorites")
				_, err = app.Dao().FindFirstRecordByFilter(
					favoritesCollection.Id,
					"user_id = {:user_id} && track_id = {:track_id}",
					map[string]any{"user_id": authRecord.Id, "track_id": record.Id},
				)
				isFavorited := err == nil

				artworkURL := upgradeArtworkURL(record.GetString("artwork_url"))
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg"
				}

				tracks = append(tracks, map[string]interface{}{
					"track_id":       record.GetString("soundcloud_id"),
					"track_title":    record.GetString("title"),
					"artist_name":    record.GetString("artist_name"),
					"track_duration": record.GetInt("length"),
					"artwork_url":    artworkURL,
					"created_at":     record.GetDateTime("post_time").Time().Format(time.RFC3339),
					"permalink_url":  record.GetString("permalink_url"),
					"is_favorited":   isFavorited,
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"tracks": tracks,
			})
		}, apis.RequireRecordAuth())

		// User settings API - GET to load, POST to save
		e.Router.GET("/api/settings", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			settingsCollection, err := app.Dao().FindCollectionByNameOrId("user_settings")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Settings collection not found",
				})
			}

			settings, err := app.Dao().FindFirstRecordByFilter(
				settingsCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				return c.JSON(http.StatusOK, map[string]interface{}{
					"active_tab": "stream",
					"filters":    map[string]interface{}{},
					"view_mode":  "grid",
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"active_tab": settings.GetString("active_tab"),
				"filters":    settings.Get("filters"),
				"view_mode":  settings.GetString("view_mode"),
			})
		}, apis.RequireRecordAuth())

		e.Router.POST("/api/settings", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			var req struct {
				ActiveTab string                 `json:"active_tab"`
				Filters   map[string]interface{} `json:"filters"`
				ViewMode  string                 `json:"view_mode"`
			}
			if err := c.Bind(&req); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Invalid request body",
				})
			}

			settingsCollection, err := app.Dao().FindCollectionByNameOrId("user_settings")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Settings collection not found",
				})
			}

			settings, err := app.Dao().FindFirstRecordByFilter(
				settingsCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)

			if err != nil {
				settings = models.NewRecord(settingsCollection)
				settings.Set("user_id", authRecord.Id)
			}

			if req.ActiveTab != "" {
				settings.Set("active_tab", req.ActiveTab)
			}
			if req.Filters != nil {
				settings.Set("filters", req.Filters)
			}
			if req.ViewMode != "" {
				settings.Set("view_mode", req.ViewMode)
			}

			if err := app.Dao().SaveRecord(settings); err != nil {
				log.Printf("Failed to save user settings: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to save settings",
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{
				"active_tab": settings.GetString("active_tab"),
				"filters":    settings.Get("filters"),
				"view_mode":  settings.GetString("view_mode"),
			})
		}, apis.RequireRecordAuth())

		e.Router.GET("/favorites", func(c echo.Context) error {
			data := views.FavoritesData{
				PageData: views.PageData{
					Title:       "Favorites",
					Description: "Your favorited tracks",
					CurrentPath: "/favorites",
					TestMode:    isTestMode,
				},
				Favorites: []views.Track{}, // Load via API
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.Favorites(data).Render(c.Request().Context(), c.Response().Writer)
		}, soundcloudAuthMiddleware(app), apis.ActivityLogger(app))

		// Favorites toggle endpoint
		e.Router.POST("/api/favorites/toggle", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			var req struct {
				TrackID string `json:"track_id"`
			}
			if err := c.Bind(&req); err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
			}
			if req.TrackID == "" {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Track ID required"})
			}

			// Find soundcloud_tracks record
			soundcloudTracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
			}
			trackRecord, err := app.Dao().FindFirstRecordByFilter(
				soundcloudTracksCollection.Id,
				"soundcloud_id = {:track_id}",
				map[string]any{"track_id": req.TrackID},
			)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Track not found"})
			}

			// Check if favorite exists
			favoritesCollection, err := app.Dao().FindCollectionByNameOrId("favorites")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
			}
			existing, err := app.Dao().FindFirstRecordByFilter(
				favoritesCollection.Id,
				"user_id = {:user_id} && track_id = {:track_id}",
				map[string]any{"user_id": authRecord.Id, "track_id": trackRecord.Id},
			)
			if err == nil {
				// Exists, delete it
				if err := app.Dao().DeleteRecord(existing); err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to unfavorite"})
				}
				return c.JSON(http.StatusOK, map[string]bool{"favorited": false})
			} else {
				// Not exists, create
				favRecord := models.NewRecord(favoritesCollection)
				favRecord.Set("user_id", authRecord.Id)
				favRecord.Set("track_id", trackRecord.Id)
				if err := app.Dao().SaveRecord(favRecord); err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to favorite"})
				}
				return c.JSON(http.StatusOK, map[string]bool{"favorited": true})
			}
		}, apis.RequireRecordAuth())

		// Favorites toggle endpoint for HTMX - returns HTML fragment
		e.Router.POST("/api/favorites/:id/htmx", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			trackID := c.PathParam("id")

			// Find soundcloud_tracks record
			soundcloudTracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				return c.String(http.StatusInternalServerError, "Database error")
			}
			trackRecord, err := app.Dao().FindFirstRecordByFilter(
				soundcloudTracksCollection.Id,
				"soundcloud_id = {:track_id}",
				map[string]any{"track_id": trackID},
			)
			if err != nil {
				return c.String(http.StatusNotFound, "Track not found")
			}

			// Get track title
			trackTitle := trackRecord.GetString("track_title")

			// Check if favorite exists
			favoritesCollection, err := app.Dao().FindCollectionByNameOrId("favorites")
			if err != nil {
				return c.String(http.StatusInternalServerError, "Database error")
			}
			existing, err := app.Dao().FindFirstRecordByFilter(
				favoritesCollection.Id,
				"user_id = {:user_id} && track_id = {:track_id}",
				map[string]any{"user_id": authRecord.Id, "track_id": trackRecord.Id},
			)

			isFavorited := false
			if err == nil {
				// Exists, delete it (unfavorite)
				if err := app.Dao().DeleteRecord(existing); err != nil {
					return c.String(http.StatusInternalServerError, "Failed to unfavorite")
				}
				isFavorited = false
			} else {
				// Not exists, create (favorite)
				favRecord := models.NewRecord(favoritesCollection)
				favRecord.Set("user_id", authRecord.Id)
				favRecord.Set("track_id", trackRecord.Id)
				if err := app.Dao().SaveRecord(favRecord); err != nil {
					return c.String(http.StatusInternalServerError, "Failed to favorite")
				}
				isFavorited = true
			}

			// Return HTML fragment
			c.Response().Header().Set("Content-Type", "text/html")
			return components.FavoriteButton(trackID, trackTitle, isFavorited).Render(c.Request().Context(), c.Response().Writer)
		}, apis.RequireRecordAuth())

		// List favorites endpoint
		e.Router.GET("/api/favorites", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			favoritesCollection, err := app.Dao().FindCollectionByNameOrId("favorites")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
			}

			records, err := app.Dao().FindRecordsByFilter(
				favoritesCollection.Id,
				"user_id = {:user_id}",
				"-created",
				100,
				0,
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch favorites"})
			}

			favorites := make([]map[string]interface{}, 0, len(records))
			for _, record := range records {
				trackRecord, err := app.Dao().FindRecordById("soundcloud_tracks", record.GetString("track_id"))
				if err != nil {
					continue
				}
				artworkURL := upgradeArtworkURL(trackRecord.GetString("artwork_url"))
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-t500x500.jpg"
				}
				favorites = append(favorites, map[string]interface{}{
					"track_id":       trackRecord.GetString("soundcloud_id"),
					"track_title":    trackRecord.GetString("title"),
					"artist_name":    trackRecord.GetString("artist_name"),
					"track_duration": trackRecord.GetInt("length"),
					"artwork_url":    artworkURL,
					"created_at":     record.Created.Time().Format(time.RFC3339),
					"permalink_url":  trackRecord.GetString("permalink_url"),
					"is_favorited":   true,
				})
			}

			return c.JSON(http.StatusOK, map[string]interface{}{"favorites": favorites})
		}, apis.RequireRecordAuth())

		// Analytics page
		e.Router.GET("/analytics", handlers.AnalyticsPage(app), soundcloudAuthMiddleware(app), apis.ActivityLogger(app))

		// Playlist routes (require authentication)
		e.Router.GET("/playlists", handlers.PlaylistsPage(app), soundcloudAuthMiddleware(app), apis.ActivityLogger(app))
		e.Router.GET("/playlists/:id", handlers.PlaylistShowPage(app), soundcloudAuthMiddleware(app), apis.ActivityLogger(app))
		e.Router.GET("/api/playlists", handlers.ListPlaylists(app), apis.RequireRecordAuth())
		e.Router.POST("/api/playlists", handlers.CreatePlaylist(app), apis.RequireRecordAuth())
		e.Router.GET("/api/playlists/:id", handlers.GetPlaylist(app), apis.RequireRecordAuth())
		e.Router.POST("/api/playlists/:id/tracks", handlers.AddTrackToPlaylist(app), apis.RequireRecordAuth())
		e.Router.DELETE("/api/playlists/:id/tracks/:track_id", handlers.RemoveTrackFromPlaylist(app), apis.RequireRecordAuth())
		e.Router.DELETE("/api/playlists/:id", handlers.DeletePlaylist(app), apis.RequireRecordAuth())
		e.Router.GET("/share/:token", handlers.GetSharedPlaylist(app)) // Public

		// Export routes (require authentication)
		e.Router.GET("/api/export/favorites/json", handlers.ExportFavoritesJSON(app), apis.RequireRecordAuth())
		e.Router.GET("/api/export/favorites/csv", handlers.ExportFavoritesCSV(app), apis.RequireRecordAuth())
		e.Router.GET("/api/export/playlists/json", handlers.ExportPlaylistsJSON(app), apis.RequireRecordAuth())

		// RSS feed routes
		e.Router.GET("/feed/rss", handlers.GetUserRSSFeed(app), apis.RequireRecordAuth())
		e.Router.GET("/feed/rss/:share_token", handlers.GetSharedRSSFeed(app)) // Public

		// Stream proxy endpoint - streams audio from SoundCloud using stored auth
		e.Router.GET("/api/track/:id/stream", func(c echo.Context) error {
			var accessToken string
			var refreshToken string

			// In test mode, skip auth check
			if isTestMode {
				// Use first available soundcloud user token in test mode
				soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
				if err == nil {
					records, _ := app.Dao().FindRecordsByFilter(
						soundcloudUsersCollection.Id,
						"access_token != ''",
						"-created",
						1,
						0,
						map[string]any{},
					)
					if len(records) > 0 {
						accessToken = records[0].GetString("access_token")
						refreshToken = records[0].GetString("refresh_token")
					}
				}
			} else {
				authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
				if authRecord == nil {
					return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
				}

				// Find the soundcloud_users record for this user
				soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
				}

				records, err := app.Dao().FindRecordsByFilter(
					soundcloudUsersCollection.Id,
					"user_id = {:user_id}",
					"-created",
					1,
					0,
					map[string]any{"user_id": authRecord.Id},
				)
				if err != nil || len(records) == 0 {
					return c.JSON(http.StatusNotFound, map[string]string{"error": "No SoundCloud account connected"})
				}

				soundcloudUser := records[0]
				accessToken = soundcloudUser.GetString("access_token")
				refreshToken = soundcloudUser.GetString("refresh_token")

				// If no access token, try to refresh
				if accessToken == "" && refreshToken != "" {
					tokenURL := "https://api.soundcloud.com/oauth2/token"
					data := url.Values{}
					data.Set("client_id", os.Getenv("SOUNDCLOUD_CLIENT_ID"))
					data.Set("client_secret", os.Getenv("SOUNDCLOUD_CLIENT_SECRET"))
					data.Set("grant_type", "refresh_token")
					data.Set("refresh_token", refreshToken)

					resp, err := http.PostForm(tokenURL, data)
					if err == nil && resp.StatusCode == 200 {
						var tokenResp struct {
							AccessToken  string `json:"access_token"`
							RefreshToken string `json:"refresh_token"`
							ExpiresIn    int    `json:"expires_in"`
						}
						if json.NewDecoder(resp.Body).Decode(&tokenResp) == nil {
							accessToken = tokenResp.AccessToken
							if tokenResp.RefreshToken != "" {
								soundcloudUser.Set("refresh_token", tokenResp.RefreshToken)
							}
							soundcloudUser.Set("access_token", tokenResp.AccessToken)
							app.Dao().SaveRecord(soundcloudUser)
						}
						resp.Body.Close()
					}
				}
			}

			if accessToken == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No access token available"})
			}

			trackID := c.PathParam("id")
			log.Printf("[Stream] Looking for track with soundcloud_id=%s", trackID)

			// Find the track in our database to get the stream_url
			tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				log.Printf("[Stream] Collection not found: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Track not found"})
			}

			trackRecords, err := app.Dao().FindRecordsByFilter(
				tracksCollection.Id,
				"soundcloud_id = {:track_id}",
				"-post_time",
				1,
				0,
				map[string]any{"track_id": trackID},
			)
			if err != nil {
				log.Printf("[Stream] Query error: %v", err)
			}
			log.Printf("[Stream] Found %d tracks for soundcloud_id=%s", len(trackRecords), trackID)
			if err != nil || len(trackRecords) == 0 {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Track not found in database"})
			}

			// Always get high-quality stream from transcodings (ignore stored stream_url)
			// First, try to refresh token if needed
			if !isTestMode && accessToken != "" && refreshToken != "" {
				// Test if token is valid by making a quick API call
				testClient := &http.Client{Timeout: 5 * time.Second}
				testReq, _ := http.NewRequest("GET", "https://api.soundcloud.com/me", nil)
				testReq.Header.Set("Authorization", "Bearer "+accessToken)
				testResp, err := testClient.Do(testReq)

				if err != nil {
					log.Printf("[Stream] Token test request failed: %v", err)
				} else {
					log.Printf("[Stream] Token test status: %d", testResp.StatusCode)
					testResp.Body.Close()

					if testResp.StatusCode == 401 || testResp.StatusCode == 403 {
						log.Printf("[Stream] Token expired (status %d), refreshing...", testResp.StatusCode)
						// Token expired, refresh it
						tokenURL := "https://api.soundcloud.com/oauth2/token"
						data := url.Values{}
						data.Set("client_id", os.Getenv("SOUNDCLOUD_CLIENT_ID"))
						data.Set("client_secret", os.Getenv("SOUNDCLOUD_CLIENT_SECRET"))
						data.Set("grant_type", "refresh_token")
						data.Set("refresh_token", refreshToken)

						resp, err := http.PostForm(tokenURL, data)
						if err != nil {
							log.Printf("[Stream] Token refresh POST failed: %v", err)
						} else {
							log.Printf("[Stream] Token refresh response status: %d", resp.StatusCode)

							if resp.StatusCode == 200 {
								var tokenResp struct {
									AccessToken  string `json:"access_token"`
									RefreshToken string `json:"refresh_token"`
									ExpiresIn    int    `json:"expires_in"`
								}
								if json.NewDecoder(resp.Body).Decode(&tokenResp) == nil {
									accessToken = tokenResp.AccessToken
									log.Printf("[Stream] Token refreshed! New token: %s...", accessToken[:min(30, len(accessToken))])
								} else {
									log.Printf("[Stream] Token refresh JSON decode failed")
								}
							} else {
								body, _ := io.ReadAll(resp.Body)
								log.Printf("[Stream] Token refresh failed: status=%d body=%s", resp.StatusCode, string(body))
							}
							resp.Body.Close()
						}
					} else if testResp.StatusCode == 200 {
						log.Printf("[Stream] Token is valid")
					}
				}
			}

			streamURL := getHighQualityStreamURL(accessToken, trackID)

			// If no high-quality URL, fall back to stored stream_url
			if streamURL == "" {
				streamURL = trackRecords[0].GetString("stream_url")
			}

			// If still no URL, return error
			if streamURL == "" {
				log.Printf("[Stream] No stream URL for track %s - stored URL is empty and high-quality fetch failed", trackID)
				return c.JSON(http.StatusNotFound, map[string]string{"error": "No stream URL available for this track. Try syncing tracks again."})
			}

			// Proxy the stream
			client := &http.Client{Timeout: 0}
			req, _ := http.NewRequest("GET", streamURL, nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)

			proxyResp, err := client.Do(req)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch stream"})
			}
			defer proxyResp.Body.Close()

			if proxyResp.StatusCode != 200 {
				log.Printf("[Stream] proxyResp.StatusCode: %d, URL: %s", proxyResp.StatusCode, streamURL)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Stream not available", "status": fmt.Sprintf("%d", proxyResp.StatusCode)})
			}

			c.Response().Header().Set("Content-Type", proxyResp.Header.Get("Content-Type"))
			c.Response().Header().Set("Accept-Ranges", "bytes")
			c.Response().WriteHeader(http.StatusOK)

			io.Copy(c.Response(), proxyResp.Body)
			return nil
		})

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
