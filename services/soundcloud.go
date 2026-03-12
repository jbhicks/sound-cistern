package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// GetValidToken returns a valid access token, refreshing if needed.
// Returns the token and a boolean indicating if re-authentication is needed.
func GetValidToken(app *pocketbase.PocketBase, authRecord *models.Record) (string, bool, error) {
	soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
	if err != nil {
		return "", false, fmt.Errorf("database error: %v", err)
	}

	soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
		soundcloudUsersCollection.Id,
		"user_id = {:user_id}",
		map[string]any{"user_id": authRecord.Id},
	)
	if err != nil {
		return "", true, fmt.Errorf("soundcloud account not linked")
	}

	accessToken := soundcloudUser.GetString("access_token")
	refreshToken := soundcloudUser.GetString("refresh_token")

	if accessToken == "" && refreshToken == "" {
		return "", true, fmt.Errorf("no access token - re-authentication required")
	}

	// Check if token is expired and needs refresh
	expiresAtStr := soundcloudUser.GetString("expires_at")
	log.Printf("[Token] Current state - accessToken present: %v, expiresAt: %s", accessToken != "", expiresAtStr)

	needsRefresh := accessToken == ""
	log.Printf("[Token] needsRefresh initial: %v", needsRefresh)

	if !needsRefresh && expiresAtStr != "" {
		// Handle alternative format with space instead of T
		expiresAtStr = strings.ReplaceAll(expiresAtStr, " ", "T")
		expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			log.Printf("[Token] Failed to parse expires_at '%s': %v", expiresAtStr, err)
		} else {
			threshold := expiresAt.Add(-5 * time.Minute)
			needsRefresh = time.Now().After(threshold)
			log.Printf("[Token] Token expires at %s, threshold %s, now %s, needs refresh: %v",
				expiresAt.Format(time.RFC3339), threshold.Format(time.RFC3339), time.Now().Format(time.RFC3339), needsRefresh)
		}
	}

	if needsRefresh && refreshToken != "" {
		log.Printf("[Token] Refreshing expired token")
		result, err := RefreshTokenFull(refreshToken)
		if err != nil {
			log.Printf("[Token] Refresh failed: %v - re-authentication required", err)
			return "", true, fmt.Errorf("token refresh failed: %v", err)
		}
		accessToken = result.AccessToken
		soundcloudUser.Set("access_token", result.AccessToken)
		// Always save the rotated refresh token — SoundCloud invalidates the old one
		if result.RefreshToken != "" {
			soundcloudUser.Set("refresh_token", result.RefreshToken)
		}
		// Update expiry using the returned expires_in, falling back to 1 hour
		expiresIn := result.ExpiresIn
		if expiresIn <= 0 {
			expiresIn = 3600
		}
		soundcloudUser.Set("expires_at", time.Now().Add(time.Duration(expiresIn)*time.Second).Format(time.RFC3339))
		if err := app.Dao().SaveRecord(soundcloudUser); err != nil {
			log.Printf("Warning: Failed to save refreshed token: %v", err)
		}
		log.Printf("[Token] Token refreshed successfully, expires in %ds", expiresIn)
	}

	return accessToken, false, nil
}

func SyncTracks(app *pocketbase.PocketBase, authRecord *models.Record, targetLimit int) (int, int, error) {
	if targetLimit <= 0 {
		targetLimit = 100
	}

	soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
	if err != nil {
		return 0, 0, fmt.Errorf("database error: %v", err)
	}

	soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
		soundcloudUsersCollection.Id,
		"user_id = {:user_id}",
		map[string]any{"user_id": authRecord.Id},
	)
	if err != nil {
		return 0, 0, fmt.Errorf("soundcloud account not linked")
	}

	accessToken, needsReauth, err := GetValidToken(app, authRecord)
	if err != nil {
		return 0, 0, err
	}
	if needsReauth {
		return 0, 0, fmt.Errorf("re-authentication required")
	}

	client := &http.Client{Timeout: 60 * time.Second}
	allActivities := []map[string]interface{}{}
	offset := 0
	apiLimit := 50

	for len(allActivities) < targetLimit {
		url := fmt.Sprintf("https://api.soundcloud.com/me/activities?limit=%d&offset=%d", apiLimit, offset)
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 {
			if resp != nil {
				resp.Body.Close()
			}
			break
		}

		var activities map[string]interface{}
		if json.NewDecoder(resp.Body).Decode(&activities) != nil {
			resp.Body.Close()
			break
		}
		resp.Body.Close()

		collection, ok := activities["collection"].([]interface{})
		if !ok || len(collection) == 0 {
			break
		}

		for _, item := range collection {
			if activity, ok := item.(map[string]interface{}); ok {
				allActivities = append(allActivities, activity)
			}
		}

		offset += apiLimit
		if len(collection) < apiLimit {
			break
		}
	}

	log.Printf("[Sync] Total collected: %d", len(allActivities))

	tracksCollection, _ := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")

	// Pre-fetch all existing soundcloud_ids for this user in ONE query
	existingIDs := make(map[string]bool)
	existingRecords, _ := app.Dao().FindRecordsByFilter(
		tracksCollection.Id,
		"user_id = {:user_id}",
		"-created",
		10000, // reasonable limit
		0,
		map[string]any{"user_id": soundcloudUser.Id},
	)
	for _, rec := range existingRecords {
		existingIDs[rec.GetString("soundcloud_id")] = true
	}
	log.Printf("[Sync] Found %d existing tracks in database", len(existingIDs))

	savedCount := 0

	for _, activity := range allActivities {
		origin, hasOrigin := activity["origin"].(map[string]interface{})
		if !hasOrigin {
			continue
		}

		var soundcloudID string
		if id, ok := origin["id"].(float64); ok {
			soundcloudID = fmt.Sprintf("%.0f", id)
		} else if id, ok := origin["id"].(string); ok {
			soundcloudID = id
		} else {
			continue
		}

		// Check against pre-fetched set instead of DB query
		if existingIDs[soundcloudID] {
			continue
		}
		// Mark as seen so we don't check again
		existingIDs[soundcloudID] = true

		trackRecord := models.NewRecord(tracksCollection)
		trackRecord.Set("user_id", soundcloudUser.Id)
		trackRecord.Set("soundcloud_id", soundcloudID)

		if title, ok := origin["title"].(string); ok {
			trackRecord.Set("title", title)
		}
		if duration, ok := origin["duration"].(float64); ok {
			trackRecord.Set("length", int64(duration))
		}
		if user, ok := origin["user"].(map[string]interface{}); ok {
			if username, ok := user["username"].(string); ok {
				trackRecord.Set("artist_name", username)
			}
		}
		if genre, ok := origin["genre"].(string); ok {
			trackRecord.Set("genre", genre)
		}
		if permalink, ok := origin["permalink_url"].(string); ok {
			trackRecord.Set("permalink_url", permalink)
		}
		if artwork, ok := origin["artwork_url"].(string); ok {
			trackRecord.Set("artwork_url", artwork)
		}
		if createdAt, ok := activity["created_at"].(string); ok {
			trackRecord.Set("post_time", createdAt)
		}
		if bpm, ok := origin["bpm"].(float64); ok {
			trackRecord.Set("bpm", bpm)
		}

		if app.Dao().SaveRecord(trackRecord) == nil {
			savedCount++
		}
	}

	log.Printf("[Sync] Saved %d new tracks (total fetched: %d)", savedCount, len(allActivities))
	return savedCount, len(allActivities), nil
}

// TokenRefreshResult holds all fields returned by SoundCloud on a token refresh.
// SoundCloud rotates refresh tokens, so the caller must persist RefreshToken too.
type TokenRefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds
}

func RefreshAccessToken(refreshToken string) (string, error) {
	result, err := RefreshTokenFull(refreshToken)
	if err != nil {
		return "", err
	}
	return result.AccessToken, nil
}

func RefreshTokenFull(refreshToken string) (*TokenRefreshResult, error) {
	tokenURL := "https://api.soundcloud.com/oauth2/token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("SOUNDCLOUD_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("SOUNDCLOUD_CLIENT_SECRET"))
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return nil, err
	}

	return &TokenRefreshResult{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		ExpiresIn:    tokenData.ExpiresIn,
	}, nil
}

func GetStreamURL(accessToken, trackID string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, _ := http.NewRequest("GET", "https://api.soundcloud.com/tracks/"+trackID, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("track fetch returned status %d", resp.StatusCode)
	}

	var trackData map[string]interface{}
	if json.NewDecoder(resp.Body).Decode(&trackData) != nil {
		return "", fmt.Errorf("failed to decode track response")
	}

	streamEndpoint := "https://api.soundcloud.com/tracks/soundcloud:tracks:" + trackID + "/stream"
	streamReq, _ := http.NewRequest("GET", streamEndpoint, nil)
	streamReq.Header.Set("Authorization", "Bearer "+accessToken)

	clientNoRedirect := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	streamResp, err := clientNoRedirect.Do(streamReq)
	if err != nil {
		return "", err
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode == 302 {
		location := streamResp.Header.Get("Location")
		if location != "" {
			return location, nil
		}
	}

	return "", fmt.Errorf("no stream URL available")
}

// SyncedFavorite is a single favorited track returned by SyncFavorites.
type SyncedFavorite struct {
	TrackID          string
	SoundcloudID     string
	Title            string
	ArtistName       string
	ArtworkURL       string
	PermalinkURL     string
	DownloadURL      string
	Downloadable     bool
	Duration         int64
	Genre            string
	PlaybackCount    int64
	FavoritingsCount int64
	BPM              float64
}

// SyncFavorites fetches the user's liked tracks from SoundCloud, upserts them
// into soundcloud_tracks, and ensures each has a row in the favorites table.
// Returns the full list of the user's favorites after the sync.
func SyncFavorites(app *pocketbase.PocketBase, authRecord *models.Record) ([]SyncedFavorite, error) {
	accessToken, needsReauth, err := GetValidToken(app, authRecord)
	if err != nil {
		return nil, err
	}
	if needsReauth {
		return nil, fmt.Errorf("re-authentication required")
	}

	soundcloudUsersCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_users")
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	soundcloudUser, err := app.Dao().FindFirstRecordByFilter(
		soundcloudUsersCollection.Id,
		"user_id = {:user_id}",
		map[string]any{"user_id": authRecord.Id},
	)
	if err != nil {
		return nil, fmt.Errorf("soundcloud account not linked")
	}

	tracksCollection, err := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}
	favoritesCollection, err := app.Dao().FindCollectionByNameOrId("favorites")
	if err != nil {
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Pre-fetch existing track IDs for this user
	existingTrackIDs := make(map[string]string) // soundcloud_id → pb record ID
	existingTrackRecords, _ := app.Dao().FindRecordsByFilter(
		tracksCollection.Id,
		"user_id = {:user_id}",
		"", 0, 0,
		map[string]any{"user_id": soundcloudUser.Id},
	)
	for _, r := range existingTrackRecords {
		existingTrackIDs[r.GetString("soundcloud_id")] = r.Id
	}

	// Pre-fetch existing favorites for this user
	existingFavs := make(map[string]bool) // pb track record ID → favorited
	existingFavRecords, _ := app.Dao().FindRecordsByFilter(
		favoritesCollection.Id,
		"user_id = {:user_id}",
		"", 0, 0,
		map[string]any{"user_id": authRecord.Id},
	)
	for _, r := range existingFavRecords {
		existingFavs[r.GetString("track_id")] = true
	}

	// Paginate through /me/likes/tracks
	client := &http.Client{Timeout: 60 * time.Second}
	nextURL := "https://api.soundcloud.com/me/likes/tracks?limit=200&linked_partitioning=1"
	var allTracks []map[string]interface{}

	for nextURL != "" {
		req, _ := http.NewRequest("GET", nextURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("soundcloud request failed: %v", err)
		}
		if resp.StatusCode == 401 {
			resp.Body.Close()
			return nil, fmt.Errorf("re-authentication required")
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("soundcloud returned status %d", resp.StatusCode)
		}

		// Response is a collection wrapper: {"collection": [...], "next_href": "..."}
		var page struct {
			Collection []map[string]interface{} `json:"collection"`
			NextHref   string                   `json:"next_href"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
			resp.Body.Close()
			break
		}
		resp.Body.Close()

		allTracks = append(allTracks, page.Collection...)
		nextURL = page.NextHref
		log.Printf("[SyncFavorites] Fetched %d likes so far (next: %v)", len(allTracks), nextURL != "")
	}

	log.Printf("[SyncFavorites] Total likes fetched from SoundCloud: %d", len(allTracks))

	// Upsert each track and favorite
	results := make([]SyncedFavorite, 0, len(allTracks))
	for _, t := range allTracks {
		var scID string
		switch v := t["id"].(type) {
		case float64:
			scID = fmt.Sprintf("%.0f", v)
		case string:
			scID = v
		default:
			continue
		}

		title, _ := t["title"].(string)
		genre, _ := t["genre"].(string)
		permalink, _ := t["permalink_url"].(string)
		artwork, _ := t["artwork_url"].(string)
		downloadURL, _ := t["download_url"].(string)
		downloadable, _ := t["downloadable"].(bool)
		duration, _ := t["duration"].(float64)
		playbackCount, _ := t["playback_count"].(float64)
		favoritingsCount, _ := t["favoritings_count"].(float64)
		bpm, _ := t["bpm"].(float64)
		artistName := ""
		if user, ok := t["user"].(map[string]interface{}); ok {
			artistName, _ = user["username"].(string)
		}

		// Upgrade artwork to 500x500
		if artwork != "" {
			artwork = strings.NewReplacer(
				"-large.", "-t500x500.",
				"-t67x67.", "-t500x500.",
				"-t300x300.", "-t500x500.",
			).Replace(artwork)
		}

		// Upsert into soundcloud_tracks
		var pbTrackID string
		if existing, ok := existingTrackIDs[scID]; ok {
			pbTrackID = existing
			// Update mutable fields on the existing record
			if rec, err := app.Dao().FindRecordById(tracksCollection.Id, pbTrackID); err == nil {
				rec.Set("title", title)
				rec.Set("artist_name", artistName)
				rec.Set("artwork_url", artwork)
				rec.Set("permalink_url", permalink)
				rec.Set("length", int64(duration))
				rec.Set("genre", genre)
				rec.Set("playback_count", int64(playbackCount))
				rec.Set("favoritings_count", int64(favoritingsCount))
				rec.Set("bpm", bpm)
				rec.Set("downloadable", downloadable)
				rec.Set("download_url", downloadURL)
				_ = app.Dao().SaveRecord(rec)
			}
		} else {
			rec := models.NewRecord(tracksCollection)
			rec.Set("user_id", soundcloudUser.Id)
			rec.Set("soundcloud_id", scID)
			rec.Set("title", title)
			rec.Set("artist_name", artistName)
			rec.Set("artwork_url", artwork)
			rec.Set("permalink_url", permalink)
			rec.Set("length", int64(duration))
			rec.Set("genre", genre)
			rec.Set("playback_count", int64(playbackCount))
			rec.Set("favoritings_count", int64(favoritingsCount))
			rec.Set("bpm", bpm)
			rec.Set("downloadable", downloadable)
			rec.Set("download_url", downloadURL)
			if err := app.Dao().SaveRecord(rec); err != nil {
				log.Printf("[SyncFavorites] Failed to save track %s: %v", scID, err)
				continue
			}
			pbTrackID = rec.Id
			existingTrackIDs[scID] = pbTrackID
		}

		// Upsert into favorites
		if !existingFavs[pbTrackID] {
			fav := models.NewRecord(favoritesCollection)
			fav.Set("user_id", authRecord.Id)
			fav.Set("track_id", pbTrackID)
			if err := app.Dao().SaveRecord(fav); err != nil {
				log.Printf("[SyncFavorites] Failed to save favorite for track %s: %v", scID, err)
			} else {
				existingFavs[pbTrackID] = true
			}
		}

		results = append(results, SyncedFavorite{
			TrackID:          pbTrackID,
			SoundcloudID:     scID,
			Title:            title,
			ArtistName:       artistName,
			ArtworkURL:       artwork,
			PermalinkURL:     permalink,
			DownloadURL:      downloadURL,
			Downloadable:     downloadable,
			Duration:         int64(duration),
			Genre:            genre,
			PlaybackCount:    int64(playbackCount),
			FavoritingsCount: int64(favoritingsCount),
			BPM:              bpm,
		})
	}

	log.Printf("[SyncFavorites] Done — %d favorites synced", len(results))
	return results, nil
}
