package actions

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop/v6"
	"github.com/jbhicks/sound-cistern/models"
	"github.com/jbhicks/sound-cistern/pkg/logging"
	"github.com/jbhicks/sound-cistern/src/services"
)

// SoundcloudAuth initiates Soundcloud OAuth login
func SoundcloudAuth(c buffalo.Context) error {
	// Get configuration from environment
	clientID := envy.Get("SOUNDCLOUD_CLIENT_ID", "")
	clientSecret := envy.Get("SOUNDCLOUD_CLIENT_SECRET", "")
	redirectURI := envy.Get("SOUNDCLOUD_REDIRECT_URI", "http://jbhicks.dev/auth/callback")

	if clientID == "" || clientSecret == "" {
		return c.Error(http.StatusInternalServerError, errors.New("Soundcloud OAuth not configured"))
	}

	// Create service and get auth URL
	soundcloudService := services.NewSoundcloudService(clientID, clientSecret, redirectURI)
	authURL := soundcloudService.GetAuthURL()

	// Redirect to Soundcloud for authentication
	return c.Redirect(http.StatusFound, authURL)
}

// SoundcloudCallback handles Soundcloud OAuth callback
func SoundcloudCallback(c buffalo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.Error(http.StatusBadRequest, errors.New("authorization code required"))
	}

	// Get configuration from environment
	clientID := envy.Get("SOUNDCLOUD_CLIENT_ID", "")
	clientSecret := envy.Get("SOUNDCLOUD_CLIENT_SECRET", "")
	redirectURI := envy.Get("SOUNDCLOUD_REDIRECT_URI", "http://jbhicks.dev/auth/callback")

	if clientID == "" || clientSecret == "" {
		return c.Error(http.StatusInternalServerError, errors.New("Soundcloud OAuth not configured"))
	}

	// Create service and handle callback
	soundcloudService := services.NewSoundcloudService(clientID, clientSecret, redirectURI)
	result, err := soundcloudService.HandleCallback(code)
	if err != nil {
		logging.Error("Soundcloud callback failed", err)
		return c.Error(http.StatusInternalServerError, errors.New("authentication failed"))
	}

	resultMap := result.(map[string]interface{})
	accessToken := resultMap["access_token"].(string)
	userInfo := resultMap["user_info"].(map[string]interface{})

	// Store access token in session for this demo
	// In production, you'd want to store this in the database associated with the user
	c.Session().Set("soundcloud_access_token", accessToken)
	c.Session().Set("soundcloud_user_id", userInfo["id"])

	userID := fmt.Sprintf("%v", userInfo["id"])
	logging.Info("Soundcloud authentication successful", logging.Fields{"user_id": userID})

	// Redirect to feed page
	return c.Redirect(http.StatusFound, "/feed")
}

// FeedIndex displays the user's Soundcloud feed
func FeedIndex(c buffalo.Context) error {
	// Get access token from session
	accessToken, ok := c.Session().Get("soundcloud_access_token").(string)
	if !ok || accessToken == "" {
		// Redirect to auth if not authenticated
		return c.Redirect(http.StatusFound, "/auth/soundcloud")
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return c.Error(http.StatusInternalServerError, errors.New("database connection not available"))
	}

	// Get current user
	currentUser := c.Value("current_user")
	if currentUser == nil {
		return c.Error(http.StatusUnauthorized, errors.New("user not authenticated"))
	}
	user := currentUser.(*models.User)

	// Create services
	clientID := envy.Get("SOUNDCLOUD_CLIENT_ID", "")
	clientSecret := envy.Get("SOUNDCLOUD_CLIENT_SECRET", "")
	redirectURI := envy.Get("SOUNDCLOUD_REDIRECT_URI", "http://jbhicks.dev/auth/callback")

	soundcloudService := services.NewSoundcloudService(clientID, clientSecret, redirectURI)
	feedService := services.NewFeedService(tx)

	// Try to get cached feed first
	cachedTracks, err := feedService.GetCachedFeed(user.ID.String())
	if err != nil {
		logging.Error("Error getting cached feed", err, logging.Fields{"user_id": user.ID.String()})
	}

	var tracks []interface{}
	if len(cachedTracks) > 0 {
		tracks = cachedTracks
		logging.Info("Using cached feed", logging.Fields{"user_id": user.ID.String(), "track_count": len(tracks)})
	} else {
		// Fetch fresh feed from Soundcloud
		freshTracks, err := soundcloudService.FetchUserFeed(accessToken)
		if err != nil {
			logging.Error("Error fetching feed from Soundcloud", err, logging.Fields{"user_id": user.ID.String()})
			return c.Error(http.StatusInternalServerError, errors.New("failed to fetch feed"))
		}

		tracks = freshTracks

		// Cache the feed
		if err := feedService.CacheFeed(user.ID.String(), tracks); err != nil {
			logging.Error("Error caching feed", err, logging.Fields{"user_id": user.ID.String()})
		}

		logging.Info("Fetched fresh feed", logging.Fields{"user_id": user.ID.String(), "track_count": len(tracks)})
	}

	// Set data for template
	c.Set("tracks", tracks)
	c.Set("user", user)

	return c.Render(http.StatusOK, r.HTML("feed/index.html"))
}

// FeedFilter filters the feed based on criteria
func FeedFilter(c buffalo.Context) error {
	// Get access token from session
	accessToken, ok := c.Session().Get("soundcloud_access_token").(string)
	if !ok || accessToken == "" {
		return c.Error(http.StatusUnauthorized, errors.New("not authenticated"))
	}

	// Get database connection
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return c.Error(http.StatusInternalServerError, errors.New("database connection not available"))
	}

	// Get current user
	currentUser := c.Value("current_user")
	if currentUser == nil {
		return c.Error(http.StatusUnauthorized, errors.New("user not authenticated"))
	}
	user := currentUser.(*models.User)

	// Parse filter criteria from request body
	var criteria map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&criteria); err != nil {
		return c.Error(http.StatusBadRequest, errors.New("invalid filter criteria"))
	}

	// Create feed service
	feedService := services.NewFeedService(tx)

	// Get cached feed
	tracks, err := feedService.GetCachedFeed(user.ID.String())
	if err != nil {
		logging.Error("Error getting cached feed for filtering", err, logging.Fields{"user_id": user.ID.String()})
		return c.Error(http.StatusInternalServerError, errors.New("failed to get feed"))
	}

	if len(tracks) == 0 {
		// If no cached feed, redirect to feed page to fetch fresh data
		return c.Redirect(http.StatusFound, "/feed")
	}

	// Filter tracks based on criteria
	filteredTracks := feedService.FilterTracks(tracks, criteria)

	logging.Info("Feed filtered", logging.Fields{
		"user_id":        user.ID.String(),
		"original_count": len(tracks),
		"filtered_count": len(filteredTracks),
	})

	// Return filtered tracks as JSON
	return c.Render(http.StatusOK, r.JSON(filteredTracks))
}
