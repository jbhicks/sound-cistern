---
name: soundcloud-oauth
description: Complete guide for Soundcloud OAuth 2.0/2.1 integration and API authentication in Go applications
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: soundcloud-api-oauth
---

# Soundcloud OAuth Authentication Skill

## Overview
This skill provides comprehensive patterns for implementing Soundcloud OAuth 2.0/2.1 authentication in Go applications, including authorization code flow, client credentials flow, token management, and API integration.

## Authentication Flows

### OAuth 2.1 Authorization Code Flow (User Authentication)
Use when your application needs to access user-specific data on their behalf.

#### Configuration Setup
```go
package soundcloud

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "time"
)

type Config struct {
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    RedirectURI  string `json:"redirect_uri"`
    BaseURL      string `json:"base_url"`
}

func NewConfig(clientID, clientSecret, redirectURI string) *Config {
    return &Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        RedirectURI:  redirectURI,
        BaseURL:      "https://api.soundcloud.com",
    }
}
```

#### Generate Authorization URL
```go
func (c *Config) GetAuthURL(state, codeVerifier string) string {
    params := url.Values{
        "response_type":         {"code"},
        "client_id":            {c.ClientID},
        "redirect_uri":         {c.RedirectURI},
        "scope":                {"user-read email"},
        "state":                {state},
        "code_challenge":        {generateCodeChallenge(codeVerifier)},
        "code_challenge_method": {"S256"},
    }
    
    return fmt.Sprintf("https://api.soundcloud.com/authorize?%s", params.Encode())
}

func generateCodeVerifier() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return base64.RawURLEncoding.EncodeToString(bytes)
}

func generateCodeChallenge(verifier string) string {
    // Generate SHA256 hash of verifier
    hash := sha256.Sum256([]byte(verifier))
    return base64.RawURLEncoding.EncodeToString(hash[:])
}
```

#### Handle OAuth Callback
```go
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    Scope        string `json:"scope"`
}

func (c *Config) HandleCallback(code, codeVerifier, state string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"authorization_code"},
        "client_id":     {c.ClientID},
        "client_secret": {c.ClientSecret},
        "redirect_uri":  {c.RedirectURI},
        "code":          {code},
        "code_verifier": {codeVerifier},
    }
    
    resp, err := http.PostForm("https://api.soundcloud.com/oauth/token", data)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code for token: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token exchange failed: %s", string(body))
    }
    
    var token TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }
    
    return &token, nil
}
```

### Client Credentials Flow (Application Authentication)
Use when your application needs to access public Soundcloud data without user context.

```go
func (c *Config) GetClientCredentialsToken() (*TokenResponse, error) {
    // Create Basic Auth header
    credentials := base64.StdEncoding.EncodeToString(
        []byte(fmt.Sprintf("%s:%s", c.ClientID, c.ClientSecret)),
    )
    
    data := url.Values{
        "grant_type": {"client_credentials"},
    }
    
    req, err := http.NewRequest("POST", "https://api.soundcloud.com/oauth/token", strings.NewReader(data.Encode()))
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Authorization", fmt.Sprintf("Basic %s", credentials))
    req.Header.Set("Accept", "application/json")
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to get client credentials token: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("client credentials failed: %s", string(body))
    }
    
    var token TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }
    
    return &token, nil
}
```

## PocketBase Integration

### OAuth Flow in PocketBase Routes
```go
package main

import (
    "crypto/rand"
    "encoding/base64"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
)

// Store OAuth state in session or database
func generateState() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return base64.URLEncoding.EncodeToString(bytes)
}

// Initiate OAuth flow
e.Router.GET("/auth/soundcloud", func(c echo.Context) error {
    config := soundcloud.NewConfig(
        os.Getenv("SOUNDCLOUD_CLIENT_ID"),
        os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
        os.Getenv("SOUNDCLOUD_REDIRECT_URI"),
    )
    
    state := generateState()
    codeVerifier := generateCodeVerifier()
    
    // Store state and code_verifier in session/temp storage
    session := getSession(c)
    session.Set("oauth_state", state)
    session.Set("code_verifier", codeVerifier)
    
    authURL := config.GetAuthURL(state, codeVerifier)
    
    return c.Redirect(http.StatusFound, authURL)
})

// Handle OAuth callback
e.Router.GET("/auth/soundcloud/callback", func(c echo.Context) error {
    config := soundcloud.NewConfig(
        os.Getenv("SOUNDCLOUD_CLIENT_ID"),
        os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
        os.Getenv("SOUNDCLOUD_REDIRECT_URI"),
    )
    
    session := getSession(c)
    
    // Verify state parameter
    expectedState := session.GetString("oauth_state")
    actualState := c.QueryParam("state")
    code := c.QueryParam("code")
    
    if expectedState == "" || actualState != expectedState {
        return apis.NewBadRequestError("Invalid state parameter", nil)
    }
    
    codeVerifier := session.GetString("code_verifier")
    if codeVerifier == "" {
        return apis.NewBadRequestError("Missing code verifier", nil)
    }
    
    // Exchange authorization code for access token
    token, err := config.HandleCallback(code, codeVerifier, expectedState)
    if err != nil {
        return apis.NewBadRequestError("Failed to exchange authorization code", err)
    }
    
    // Get user info from Soundcloud
    userInfo, err := getUserInfo(token.AccessToken)
    if err != nil {
        return apis.NewBadRequestError("Failed to get user info", err)
    }
    
    // Create or update user in PocketBase
    user, err := createOrUpdateSoundcloudUser(app, userInfo, token)
    if err != nil {
        return apis.NewInternalServerError("Failed to create user", err)
    }
    
    // Clear OAuth session data
    session.Delete("oauth_state")
    session.Delete("code_verifier")
    
    // Authenticate user in PocketBase
    return authenticatePocketBaseUser(c, user)
})
```

### User Information Management
```go
type SoundcloudUser struct {
    ID          int64  `json:"id"`
    Username    string  `json:"username"`
    DisplayName string  `json:"display_name"`
    Email       string  `json:"email"`
    AvatarURL   string  `json:"avatar_url"`
    Country     string  `json:"country"`
    City        string  `json:"city"`
    Description string  `json:"description"`
}

func getUserInfo(accessToken string) (*SoundcloudUser, error) {
    req, err := http.NewRequest("GET", "https://api.soundcloud.com/me", nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
    
    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get user info: %s", string(body))
    }
    
    var user SoundcloudUser
    if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
        return nil, err
    }
    
    return &user, nil
}

func createOrUpdateSoundcloudUser(app *pocketbase.PocketBase, userInfo *SoundcloudUser, token *soundcloud.TokenResponse) (*models.Record, error) {
    usersCollection, err := app.Dao().FindCollectionByNameOrId("users")
    if err != nil {
        return nil, err
    }
    
    // Check if user already exists
    existingUser, err := app.Dao().FindFirstRecordByFilter(
        "users",
        "soundcloud_id = {:soundcloud_id}",
        dbx.Params{"soundcloud_id": userInfo.ID},
    )
    
    var user *models.Record
    
    if err == nil && existingUser != nil {
        // Update existing user
        user = existingUser
        user.Set("display_name", userInfo.DisplayName)
        user.Set("avatar_url", userInfo.AvatarURL)
        user.Set("last_login", time.Now())
    } else {
        // Create new user
        user = models.NewRecord(usersCollection)
        user.Set("soundcloud_id", userInfo.ID)
        user.Set("username", userInfo.Username)
        user.Set("display_name", userInfo.DisplayName)
        user.Set("email", userInfo.Email)
        user.Set("avatar_url", userInfo.AvatarURL)
        user.Set("auth_provider", "soundcloud")
        user.Set("created", time.Now())
        user.Set("last_login", time.Now())
    }
    
    // Store tokens securely
    user.Set("access_token", token.AccessToken)
    user.Set("refresh_token", token.RefreshToken)
    user.Set("token_expires", time.Now().Add(time.Duration(token.ExpiresIn)*time.Second))
    
    if err := app.Dao().SaveRecord(user); err != nil {
        return nil, err
    }
    
    return user, nil
}
```

## Token Management

### Refresh Token Implementation
```go
func (c *Config) RefreshToken(refreshToken string) (*TokenResponse, error) {
    data := url.Values{
        "grant_type":    {"refresh_token"},
        "client_id":     {c.ClientID},
        "client_secret": {c.ClientSecret},
        "refresh_token": {refreshToken},
    }
    
    resp, err := http.PostForm("https://api.soundcloud.com/oauth/token", data)
    if err != nil {
        return nil, fmt.Errorf("failed to refresh token: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("token refresh failed: %s", string(body))
    }
    
    var token TokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
        return nil, fmt.Errorf("failed to decode refresh response: %w", err)
    }
    
    return &token, nil
}

// Auto-refresh middleware
func tokenRefreshMiddleware(app *pocketbase.PocketBase) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
            if authRecord == nil {
                return next(c)
            }
            
            // Check if token needs refresh
            expiresAt := authRecord.GetDateTime("token_expires")
            if expiresAt.Before(time.Now().Add(5*time.Minute)) {
                refreshToken := authRecord.GetString("refresh_token")
                if refreshToken != "" {
                    config := soundcloud.NewConfig(
                        os.Getenv("SOUNDCLOUD_CLIENT_ID"),
                        os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
                        os.Getenv("SOUNDCLOUD_REDIRECT_URI"),
                    )
                    
                    newToken, err := config.RefreshToken(refreshToken)
                    if err == nil {
                        authRecord.Set("access_token", newToken.AccessToken)
                        authRecord.Set("refresh_token", newToken.RefreshToken)
                        authRecord.Set("token_expires", time.Now().Add(time.Duration(newToken.ExpiresIn)*time.Second))
                        app.Dao().SaveRecord(authRecord)
                    }
                }
            }
            
            return next(c)
        }
    }
}
```

### Token Validation
```go
func (c *Config) ValidateToken(accessToken string) (bool, error) {
    req, err := http.NewRequest("GET", "https://api.soundcloud.com/me", nil)
    if err != nil {
        return false, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    return resp.StatusCode == http.StatusOK, nil
}
```

## Database Schema

### Soundcloud User Collection Migration
```go
// pb_migrations/1696000004_add_soundcloud_users.go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
)

func init() {
    core.OnMigrate().Register(func(db core.DB) error {
        collection := &core.Collection{
            Name: "soundcloud_users",
            Type: core.CollectionTypeBase,
        }

        collection.Fields.Add(
            &core.TextField{
                Name:     "soundcloud_id",
                Required: true,
                Unique:   true,
            },
            &core.TextField{
                Name:     "username",
                Required: true,
            },
            &core.TextField{
                Name:     "display_name",
            },
            &core.TextField{
                Name:     "email",
            },
            &core.TextField{
                Name:     "avatar_url",
            },
            &core.TextField{
                Name:     "access_token",
                Required: true,
                Max:      1000,
            },
            &core.TextField{
                Name:     "refresh_token",
                Max:      1000,
            },
            &core.DateTimeField{
                Name:     "token_expires",
            },
            &core.TextField{
                Name:     "auth_provider",
                Default:   "soundcloud",
            },
            &core.DateTimeField{
                Name:     "last_login",
            },
        )

        // Set access rules
        viewRule := "id = @request.auth.id"
        collection.ViewRule = &viewRule
        
        updateRule := "id = @request.auth.id"
        collection.UpdateRule = &updateRule

        return app.Dao().SaveCollection(collection)
    }, nil)
}
```

## API Integration

### Soundcloud API Client
```go
type SoundcloudClient struct {
    config     *Config
    accessToken string
    httpClient *http.Client
}

func NewSoundcloudClient(config *Config, accessToken string) *SoundcloudClient {
    return &SoundcloudClient{
        config:     config,
        accessToken: accessToken,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (sc *SoundcloudClient) GetCurrentUser() (*SoundcloudUser, error) {
    return getUserInfo(sc.accessToken)
}

func (sc *SoundcloudClient) GetUserTracks(userID int64, limit int) ([]SoundcloudTrack, error) {
    url := fmt.Sprintf("https://api.soundcloud.com/users/%d/tracks?limit=%d", userID, limit)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.accessToken))
    
    resp, err := sc.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get tracks: %s", string(body))
    }
    
    var tracks []SoundcloudTrack
    if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
        return nil, err
    }
    
    return tracks, nil
}

type SoundcloudTrack struct {
    ID          int64     `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Duration    int       `json:"duration"`
    CreatedAt   time.Time `json:"created_at"`
    ArtworkURL  string    `json:"artwork_url"`
    StreamURL   string    `json:"stream_url"`
    DownloadURL string    `json:"download_url"`
    Genre       string    `json:"genre"`
    Tags        []string  `json:"tag_list"`
    Private     bool      `json:"private"`
}
```

### Feed Aggregation
```go
func (sc *SoundcloudClient) GetUserFeed(userID int64) ([]SoundcloudTrack, error) {
    // Get tracks user follows
    follows, err := sc.GetUserFollowings(userID)
    if err != nil {
        return nil, err
    }
    
    var allTracks []SoundcloudTrack
    seen := make(map[int64]bool)
    
    // Get recent tracks from followed users
    for _, user := range follows {
        tracks, err := sc.GetUserTracks(user.ID, 10)
        if err != nil {
            continue // Skip failed requests
        }
        
        for _, track := range tracks {
            if !seen[track.ID] {
                allTracks = append(allTracks, track)
                seen[track.ID] = true
            }
        }
    }
    
    // Sort by creation date
    sort.Slice(allTracks, func(i, j int) bool {
        return allTracks[i].CreatedAt.After(allTracks[j].CreatedAt)
    })
    
    return allTracks, nil
}

func (sc *SoundcloudClient) GetUserFollowings(userID int64) ([]SoundcloudUser, error) {
    url := fmt.Sprintf("https://api.soundcloud.com/users/%d/followings", userID)
    
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.accessToken))
    
    resp, err := sc.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get followings: %s", string(body))
    }
    
    var users []SoundcloudUser
    if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
        return nil, err
    }
    
    return users, nil
}
```

## PocketBase Routes for Soundcloud Integration

### Feed Management Endpoints
```go
// Sync user's Soundcloud feed
e.Router.POST("/api/soundcloud/sync", func(c echo.Context) error {
    authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
    if authRecord == nil {
        return apis.NewUnauthorizedError("Authentication required", nil)
    }
    
    if authRecord.GetString("auth_provider") != "soundcloud" {
        return apis.NewBadRequestError("User not authenticated via Soundcloud", nil)
    }
    
    accessToken := authRecord.GetString("access_token")
    soundcloudID := authRecord.GetInt64("soundcloud_id")
    
    client := soundcloud.NewSoundcloudClient(config, accessToken)
    
    // Get user's feed from Soundcloud
    tracks, err := client.GetUserFeed(soundcloudID)
    if err != nil {
        return apis.NewBadRequestError("Failed to sync Soundcloud feed", err)
    }
    
    // Store tracks in PocketBase
    err = storeTracksInPocketBase(app, tracks, authRecord.Id)
    if err != nil {
        return apis.NewInternalServerError("Failed to store tracks", err)
    }
    
    // Trigger HTMX update if applicable
    if isHTMXRequest(c) {
        return views.SyncSuccess(len(tracks)).Render(c.Request().Context(), c.Response().Writer)
    }
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "message": "Sync completed",
        "tracks_synced": len(tracks),
    })
})

// Get user's stored tracks
e.Router.GET("/api/tracks", func(c echo.Context) error {
    authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
    if authRecord == nil {
        return apis.NewUnauthorizedError("Authentication required", nil)
    }
    
    tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
    if err != nil {
        return err
    }
    
    page := 1
    if p := c.QueryParam("page"); p != "" {
        if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
            page = parsed
        }
    }
    
    var records []*core.Record
    err = app.Dao().RecordQuery(tracksCollection).
        AndWhere(dbx.HashExp{"owner": authRecord.Id}).
        OrderBy("created DESC").
        Limit(20).
        Offset((page - 1) * 20).
        All(&records)
    if err != nil {
        return err
    }
    
    if isHTMXRequest(c) {
        return views.TracksList(records).Render(c.Request().Context(), c.Response().Writer)
    }
    
    data := views.TracksPageData{
        PageData: views.PageData{
            Title:       "My Tracks",
            Description: "Your Soundcloud tracks",
            CurrentPath: "/tracks",
            User:        authRecord,
        },
        Tracks: records,
        Page:   page,
    }
    
    return views.TracksPage(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Error Handling

### OAuth Error Responses
```go
type OAuthError struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description"`
}

func (c *Config) HandleOAuthError(callbackURL string, err error) error {
    errorParams := url.Values{
        "error":             {"access_denied"},
        "error_description": {err.Error()},
    }
    
    redirectURL := fmt.Sprintf("%s?%s", callbackURL, errorParams.Encode())
    return fmt.Errorf("OAuth error: %s", redirectURL)
}
```

### Common Error Handling Patterns
```go
func handleSoundcloudError(err error, c echo.Context) error {
    if strings.Contains(err.Error(), "invalid_client") {
        return apis.NewBadRequestError("Invalid Soundcloud credentials", nil)
    }
    
    if strings.Contains(err.Error(), "invalid_grant") {
        return apis.NewBadRequestError("Invalid or expired authorization code", nil)
    }
    
    if strings.Contains(err.Error(), "unauthorized_client") {
        return apis.NewBadRequestError("Unauthorized Soundcloud client", nil)
    }
    
    return apis.NewInternalServerError("Soundcloud API error", err)
}
```

## Security Best Practices

### OAuth Security Implementation
1. **Always use HTTPS** for all OAuth endpoints
2. **Validate state parameter** to prevent CSRF attacks
3. **Use PKCE** (Proof Key for Code Exchange) for public clients
4. **Store tokens securely** in encrypted database fields
5. **Implement token expiration** handling and refresh logic
6. **Use short-lived access tokens** with refresh token rotation

### Environment Configuration
```go
type AppConfig struct {
    SoundcloudClientID     string `env:"SOUNDCLOUD_CLIENT_ID,required"`
    SoundcloudClientSecret string `env:"SOUNDCLOUD_CLIENT_SECRET,required"`
    SoundcloudRedirectURI  string `env:"SOUNDCLOUD_REDIRECT_URI,required"`
}

func LoadConfig() (*AppConfig, error) {
    var cfg AppConfig
    if err := env.Parse(&cfg); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    return &cfg, nil
}
```

### Secure Token Storage
```go
func encryptToken(token string) (string, error) {
    key := []byte(os.Getenv("ENCRYPTION_KEY"))
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    nonce := make([]byte, gcm.NonceSize())
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptToken(encryptedToken string) (string, error) {
    key := []byte(os.Getenv("ENCRYPTION_KEY"))
    ciphertext, _ := base64.StdEncoding.DecodeString(encryptedToken)
    
    block, _ := aes.NewCipher(key)
    gcm, _ := cipher.NewGCM(block)
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    
    plaintext, _ := gcm.Open(nil, nonce, ciphertext, nil)
    return string(plaintext), nil
}
```

## Testing

### Unit Test OAuth Flow
```go
func TestOAuthFlow(t *testing.T) {
    // Mock HTTP client
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/oauth/token":
            if r.Method == "POST" {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{
                    "access_token": "test_token",
                    "refresh_token": "test_refresh",
                    "expires_in": 3600,
                    "scope": "user-read"
                }`))
            }
        case "/me":
            if r.Header.Get("Authorization") == "Bearer test_token" {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusOK)
                w.Write([]byte(`{
                    "id": 12345,
                    "username": "testuser",
                    "display_name": "Test User",
                    "email": "test@example.com"
                }`))
            }
        }
    }))
    defer server.Close()
    
    // Test token exchange
    config := NewConfig("test_client", "test_secret", "http://localhost/callback")
    token, err := config.HandleCallback("test_code", "test_verifier", "test_state")
    
    assert.NoError(t, err)
    assert.Equal(t, "test_token", token.AccessToken)
    assert.Equal(t, "test_refresh", token.RefreshToken)
}
```

## Deployment Configuration

### Environment Variables
```bash
# Production environment
SOUNDCLOUD_CLIENT_ID=your_production_client_id
SOUNDCLOUD_CLIENT_SECRET=your_production_client_secret
SOUNDCLOUD_REDIRECT_URI=https://yourapp.com/auth/soundcloud/callback

# Development environment
SOUNDCLOUD_CLIENT_ID=your_development_client_id
SOUNDCLOUD_CLIENT_SECRET=your_development_client_secret
SOUNDCLOUD_REDIRECT_URI=http://localhost:8090/auth/soundcloud/callback
```

### Redirect URI Configuration
1. **Register redirect URI** in Soundcloud Developer Console
2. **Use HTTPS** in production (HTTP allowed for development)
3. **Match exact URI** (no trailing slashes)
4. **Update environment** variables for different deployment targets

### Rate Limiting Considerations
- **50 tokens per 12 hours** per application
- **30 tokens per hour** per IP address
- **Implement token caching** and reuse strategies
- **Handle rate limit errors** gracefully with retry logic

---

This skill provides comprehensive patterns for implementing secure, robust Soundcloud OAuth integration in Go applications, covering both user authentication and application-level API access scenarios.