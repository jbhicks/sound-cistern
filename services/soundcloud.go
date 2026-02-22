package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

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

	accessToken := soundcloudUser.GetString("access_token")
	refreshToken := soundcloudUser.GetString("refresh_token")

	if accessToken == "" && refreshToken == "" {
		return 0, 0, fmt.Errorf("no access token")
	}

	if accessToken == "" && refreshToken != "" {
		accessToken, err = refreshAccessToken(refreshToken)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to refresh token: %v", err)
		}
		soundcloudUser.Set("access_token", accessToken)
		app.Dao().SaveRecord(soundcloudUser)
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

		if app.Dao().SaveRecord(trackRecord) == nil {
			savedCount++
		}
	}

	log.Printf("[Sync] Saved %d new tracks (total fetched: %d)", savedCount, len(allActivities))
	return savedCount, len(allActivities), nil
}

func refreshAccessToken(refreshToken string) (string, error) {
	tokenURL := "https://api.soundcloud.com/oauth2/token"
	data := url.Values{}
	data.Set("client_id", os.Getenv("SOUNDCLOUD_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("SOUNDCLOUD_CLIENT_SECRET"))
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
		return "", err
	}

	return tokenData.AccessToken, nil
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
