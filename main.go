package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "github.com/jbhicks/sound-cistern/pb_migrations"
	"github.com/jbhicks/sound-cistern/views"
)

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

// generateRandomString generates a random string for state parameter
func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random string: %v", err)
	}
	return strings.ReplaceAll(base64.URLEncoding.EncodeToString(b), "=", "")[:length]
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	app := pocketbase.New()

	var publicDir string = "./pb_public"

	isGoRun := true
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	jsvm.MustRegister(app, jsvm.Config{})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Security headers middleware
		e.Router.Use(middleware.Secure())

		// CSRF protection
		e.Router.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLength: 32,
			TokenLookup: "form:csrf_token",
			Skipper: func(c echo.Context) bool {
				// Skip CSRF for PocketBase admin routes
				return strings.HasPrefix(c.Path(), "/_") || strings.HasPrefix(c.Path(), "/api/admins/")
			},
		}))

		// Static file serving
		e.Router.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   publicDir,
			Browse: false,
			HTML5:  true,
		}))

		// Health check endpoint
		e.Router.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		})

		// Home page
		e.Router.GET("/", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Home",
				Description: "Your Soundcloud feed aggregator",
				CurrentPath: "/",
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.Home(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Stream page
		e.Router.GET("/stream", func(c echo.Context) error {
			data := views.StreamData{
				PageData: views.PageData{
					Title:       "Stream",
					Description: "Chronological feed of your Soundcloud tracks",
					CurrentPath: "/stream",
				},
				Tracks: []views.Track{}, // Start with empty tracks, will be loaded via client-side API call
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.Stream(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Sign in page
		e.Router.GET("/signin", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Sign In",
				Description: "Sign in to your account",
				CurrentPath: "/signin",
			}

			return views.SignIn(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Sign up page
		e.Router.GET("/signup", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Sign Up",
				Description: "Create your account",
				CurrentPath: "/signup",
			}

			return views.SignUp(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Blog index page (without enhanced features)
		e.Router.GET("/blog", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Blog",
				Description: "Latest posts and articles",
				CurrentPath: "/blog",
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

			data := views.PageData{
				Title:       post.Title,
				Description: post.Excerpt,
				CurrentPath: "/blog/" + slug,
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.BlogShow(data, post).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Soundcloud OAuth endpoints
		e.Router.GET("/auth/soundcloud", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord == nil {
				return c.Redirect(http.StatusTemporaryRedirect, "/signin?error=oauth_requires_auth")
			}

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

			// Store state and code verifier in oauth_states collection
			oauthStatesCollection, err := app.Dao().FindCollectionByNameOrId("oauth_states")
			if err != nil {
				log.Printf("Failed to find oauth_states collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// Create state record
			stateRecord := models.NewRecord(oauthStatesCollection)
			stateRecord.Set("state", state)
			stateRecord.Set("code_verifier", codeVerifier)
			stateRecord.Set("user_id", authRecord.Id)
			stateRecord.Set("expires_at", time.Now().Add(10*time.Minute).Format(time.RFC3339))

			if err := app.Dao().SaveRecord(stateRecord); err != nil {
				log.Printf("Failed to store OAuth state: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to store OAuth state",
				})
			}

			// Build Soundcloud OAuth URL with PKCE
			authURL := url.URL{
				Scheme: "https",
				Host:   "api.soundcloud.com",
				Path:   "/connect",
				RawQuery: url.Values{
					"response_type":         {"code"},
					"client_id":             {clientID},
					"redirect_uri":          {redirectURI},
					"scope":                 {"non-expiring"},
					"state":                 {state},
					"code_challenge":        {codeChallenge},
					"code_challenge_method": {"S256"},
				}.Encode(),
			}

			log.Printf("Redirecting user %s to Soundcloud OAuth with state: %s", authRecord.Id, state)
			return c.Redirect(http.StatusTemporaryRedirect, authURL.String())
		}, apis.RequireRecordAuth())

		// Soundcloud OAuth callback endpoint
		e.Router.GET("/auth/soundcloud/callback", func(c echo.Context) error {
			// Get query parameters from Soundcloud callback
			code := c.QueryParam("code")
			state := c.QueryParam("state")
			errorParam := c.QueryParam("error")

			// Handle OAuth errors from Soundcloud
			if errorParam != "" {
				log.Printf("OAuth error from Soundcloud: %s", errorParam)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=oauth_failed")
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

			// Check if state has expired
			expiresAtStr := stateRecord.GetString("expires_at")
			if expiresAtStr != "" {
				expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
				if err == nil && time.Now().After(expiresAt) {
					log.Printf("OAuth state expired for state: %s", state)
					// Clean up expired state record
					if err := app.Dao().DeleteRecord(stateRecord); err != nil {
						log.Printf("Warning: Failed to clean up expired state record: %v", err)
					}
					return c.Redirect(http.StatusTemporaryRedirect, "/?error=state_expired")
				}
			}

			// Get the code verifier
			codeVerifier := stateRecord.GetString("code_verifier")
			if codeVerifier == "" {
				log.Printf("Missing code verifier for state: %s", state)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=missing_verifier")
			}

			// Get the user ID from the state record
			userID := stateRecord.GetString("user_id")
			if userID == "" {
				log.Printf("Missing user ID for state: %s", state)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=missing_user")
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
				"https://api.soundcloud.com/oauth2/token",
				tokenData,
			)
			if err != nil {
				log.Printf("Failed to exchange authorization code: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=token_exchange_failed")
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Token exchange failed with status: %d", resp.StatusCode)
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

			// Check if Soundcloud user already exists
			authCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("Failed to find soundcloud_users collection: %v", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/?error=server_error")
			}

			existingUser, err := app.Dao().FindFirstRecordByFilter(
				authCollection.Id,
				"soundcloud_id = {:soundcloud_id}",
				map[string]any{"soundcloud_id": userInfo.Username},
			)

			if err != nil {
				// Create new Soundcloud user record
				soundcloudUser := models.NewRecord(authCollection)
				soundcloudUser.Set("soundcloud_id", userInfo.Username)
				soundcloudUser.Set("user_id", userID)
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

				log.Printf("Created new Soundcloud user: %s for PocketBase user: %s", userInfo.Username, userID)
			} else {
				// Update existing Soundcloud user record
				existingUser.Set("user_id", userID)
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

				log.Printf("Updated existing Soundcloud user: %s for PocketBase user: %s", userInfo.Username, userID)
			}

			// Clean up the state record
			if err := app.Dao().DeleteRecord(stateRecord); err != nil {
				log.Printf("Warning: Failed to clean up state record: %v", err)
			}

			// Redirect to success page
			log.Printf("OAuth flow completed successfully for Soundcloud user: %s", userInfo.Username)
			return c.Redirect(http.StatusTemporaryRedirect, "/stream")
		}, apis.ActivityLogger(app))

		// Public assets
		e.Router.GET("/public/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), false))

		// API routes (protected)
		e.Router.GET("/api/user", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"id":    authRecord.Id,
				"email": authRecord.GetString("email"),
				"name":  authRecord.GetString("first_name") + " " + authRecord.GetString("last_name"),
			})
		}, apis.RequireRecordAuth())

		// Stream endpoint - fetch user's Soundcloud tracks chronologically
		e.Router.GET("/api/stream", func(c echo.Context) error {
			// This endpoint requires authentication via apis.RequireRecordAuth()

			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

			// Find Soundcloud user record associated with this user
			soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
			if err != nil {
				log.Printf("Failed to find soundcloud_users collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// Look for Soundcloud user linked to this PocketBase user
			soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
				soundcloudUsersCollection.Id,
				"user_id = {:user_id}",
				map[string]any{"user_id": authRecord.Id},
			)
			if err != nil {
				log.Printf("No Soundcloud user found for user %s: %v", authRecord.Id, err)
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "No Soundcloud account linked",
				})
			}

			// Check if access token is still valid
			accessToken := soundcloudUser.GetString("access_token")
			expiresAtStr := soundcloudUser.GetString("expires_at")

			if accessToken == "" {
				log.Printf("No access token found for Soundcloud user")
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Soundcloud access token missing",
				})
			}

			// Check token expiration
			if expiresAtStr != "" {
				expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
				if err == nil && time.Now().After(expiresAt) {
					log.Printf("Access token expired for Soundcloud user")
					// TODO: Implement token refresh logic
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Soundcloud access token expired",
					})
				}
			}

			// Fetch activities from Soundcloud API
			client := &http.Client{Timeout: 15 * time.Second}

			// Try to get user's activities first
			activitiesURL := "https://api.soundcloud.com/me/activities?limit=50"
			req, err := http.NewRequest("GET", activitiesURL, nil)
			if err != nil {
				log.Printf("Failed to create activities request: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to fetch activities",
				})
			}

			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Accept", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to fetch activities from Soundcloud: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to fetch activities from Soundcloud",
				})
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Soundcloud API returned status: %d", resp.StatusCode)
				if resp.StatusCode == http.StatusUnauthorized {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"error": "Invalid or expired Soundcloud token",
					})
				}
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Soundcloud API error",
				})
			}

			// Parse activities response
			var activitiesResponse struct {
				Collection []struct {
					Type   string `json:"type"`
					Origin struct {
						Track struct {
							ID           int64  `json:"id"`
							Title        string `json:"title"`
							Description  string `json:"description"`
							Duration     int64  `json:"duration"`
							Genre        string `json:"genre"`
							CreatedAt    string `json:"created_at"`
							PermalinkURL string `json:"permalink_url"`
							ArtworkURL   string `json:"artwork_url"`
							User         struct {
								ID       int64  `json:"id"`
								Username string `json:"username"`
							} `json:"user"`
						} `json:"track"`
					} `json:"origin"`
					CreatedAt string `json:"created_at"`
				} `json:"collection"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&activitiesResponse); err != nil {
				log.Printf("Failed to decode activities response: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to parse activities response",
				})
			}

			// Get soundcloud_tracks collection for caching
			soundcloudTracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
			if err != nil {
				log.Printf("Failed to find soundcloud_tracks collection: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Database configuration error",
				})
			}

			// Process and store tracks from activities
			tracks := make([]map[string]interface{}, 0, len(activitiesResponse.Collection))

			for _, activity := range activitiesResponse.Collection {
				// Only process track-related activities
				if activity.Type != "track" && activity.Type != "track-repost" && activity.Type != "track-station" {
					continue
				}

				track := activity.Origin.Track

				// Parse the creation date from activity
				createdAt, err := time.Parse(time.RFC3339, activity.CreatedAt)
				if err != nil {
					// Try alternative format
					createdAt, err = time.Parse("2006/01/02 15:04:05 +0000", activity.CreatedAt)
					if err != nil {
						log.Printf("Failed to parse activity created_at: %s, error: %v", activity.CreatedAt, err)
						createdAt = time.Now() // fallback to current time
					}
				}

				// Check if track already exists in database
				existingTrack, err := app.Dao().FindFirstRecordByFilter(
					soundcloudTracksCollection.Id,
					"soundcloud_id = {:track_id}",
					map[string]any{"track_id": fmt.Sprintf("%d", track.ID)},
				)

				var trackRecord *models.Record
				if err != nil {
					// Create new track record
					trackRecord = models.NewRecord(soundcloudTracksCollection)
					trackRecord.Set("user_id", soundcloudUser.Id)
					trackRecord.Set("soundcloud_id", fmt.Sprintf("%d", track.ID))
					trackRecord.Set("title", track.Title)
					trackRecord.Set("length", track.Duration)
					trackRecord.Set("genre", track.Genre)
					trackRecord.Set("post_time", createdAt)
					trackRecord.Set("artist_name", track.User.Username)
					trackRecord.Set("permalink_url", track.PermalinkURL)
					trackRecord.Set("artwork_url", track.ArtworkURL)

					if err := app.Dao().SaveRecord(trackRecord); err != nil {
						log.Printf("Warning: Failed to cache track %d: %v", track.ID, err)
					}
				} else {
					// Update existing track if needed
					existingTrack.Set("title", track.Title)
					existingTrack.Set("length", track.Duration)
					existingTrack.Set("genre", track.Genre)
					existingTrack.Set("post_time", createdAt)
					existingTrack.Set("artist_name", track.User.Username)
					existingTrack.Set("permalink_url", track.PermalinkURL)
					existingTrack.Set("artwork_url", track.ArtworkURL)

					if err := app.Dao().SaveRecord(existingTrack); err != nil {
						log.Printf("Warning: Failed to update track %d: %v", track.ID, err)
					}
					trackRecord = existingTrack
				}

				// Check if favorited
				favoritesCollection, _ := app.Dao().FindCollectionByNameOrId("favorites")
				_, err = app.Dao().FindFirstRecordByFilter(
					favoritesCollection.Id,
					"user_id = {:user_id} && track_id = {:track_id}",
					map[string]any{"user_id": authRecord.Id, "track_id": trackRecord.Id},
				)
				isFavorited := err == nil

				// Add to response
				artworkURL := track.ArtworkURL
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-large.jpg"
				}

				tracks = append(tracks, map[string]interface{}{
					"track_id":       fmt.Sprintf("%d", track.ID),
					"track_title":    track.Title,
					"artist_name":    track.User.Username,
					"track_duration": track.Duration,
					"artwork_url":    artworkURL,
					"created_at":     createdAt.Format(time.RFC3339),
					"permalink_url":  track.PermalinkURL,
					"is_favorited":   isFavorited,
				})
			}

			// Sort tracks by creation date (newest first)
			for i := 0; i < len(tracks)-1; i++ {
				for j := i + 1; j < len(tracks); j++ {
					timeI, _ := time.Parse(time.RFC3339, tracks[i]["created_at"].(string))
					timeJ, _ := time.Parse(time.RFC3339, tracks[j]["created_at"].(string))
					if timeJ.After(timeI) {
						tracks[i], tracks[j] = tracks[j], tracks[i]
					}
				}
			}

			log.Printf("Successfully fetched %d tracks for stream", len(tracks))
			return c.JSON(http.StatusOK, map[string]interface{}{
				"tracks": tracks,
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

				artworkURL := record.GetString("artwork_url")
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-large.jpg"
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
		e.Router.GET("/favorites", func(c echo.Context) error {
			data := views.FavoritesData{
				PageData: views.PageData{
					Title:       "Favorites",
					Description: "Your favorited tracks",
					CurrentPath: "/favorites",
				},
				Favorites: []views.Track{}, // Load via API
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.Favorites(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

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
				artworkURL := trackRecord.GetString("artwork_url")
				if artworkURL == "" {
					artworkURL = "https://i1.sndcdn.com/avatars-000000000000000000000000000000-default-large.jpg"
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

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
