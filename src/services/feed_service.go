package services

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jbhicks/sound-cistern/src/models"
)

// FeedService handles feed caching and filtering
type FeedService struct {
	DB *pop.Connection
}

// NewFeedService creates a new service
func NewFeedService(db *pop.Connection) *FeedService {
	return &FeedService{DB: db}
}

// CacheFeed caches the feed for a user
func (fs *FeedService) CacheFeed(userID string, tracks []interface{}) error {
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return err
	}

	// Convert tracks to JSON
	tracksJSON, err := json.Marshal(tracks)
	if err != nil {
		return err
	}

	// Update or create feed entry
	feed := &models.Feed{}
	err = fs.DB.Where("user_id = ?", userUUID).First(feed)
	if err != nil {
		// Create new feed
		feed = &models.Feed{
			ID:        uuid.Must(uuid.NewV4()),
			UserID:    userUUID,
			Tracks:    string(tracksJSON),
			UpdatedAt: time.Now(),
		}
		return fs.DB.Create(feed)
	} else {
		// Update existing feed
		feed.Tracks = string(tracksJSON)
		feed.UpdatedAt = time.Now()
		return fs.DB.Update(feed)
	}
}

// GetCachedFeed gets cached feed for user
func (fs *FeedService) GetCachedFeed(userID string) ([]interface{}, error) {
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return nil, err
	}

	feed := &models.Feed{}
	err = fs.DB.Where("user_id = ?", userUUID).First(feed)
	if err != nil {
		return []interface{}{}, nil // Return empty if no cache
	}

	var tracks []interface{}
	err = json.Unmarshal([]byte(feed.Tracks), &tracks)
	if err != nil {
		return []interface{}{}, err
	}

	return tracks, nil
}

// FilterTracks filters tracks based on criteria
func (fs *FeedService) FilterTracks(tracks []interface{}, criteria map[string]interface{}) []interface{} {
	var filtered []interface{}
	for _, track := range tracks {
		trackMap := track.(map[string]interface{})
		if fs.matchesCriteria(trackMap, criteria) {
			filtered = append(filtered, track)
		}
	}
	return filtered
}

// matchesCriteria checks if track matches filter criteria
func (fs *FeedService) matchesCriteria(track map[string]interface{}, criteria map[string]interface{}) bool {
	if minLength, ok := criteria["min_length"].(float64); ok {
		if track["length"].(float64) < minLength {
			return false
		}
	}
	if maxLength, ok := criteria["max_length"].(float64); ok {
		if track["length"].(float64) > maxLength {
			return false
		}
	}
	if genres, ok := criteria["genres"].([]interface{}); ok {
		trackGenre := track["genre"].(string)
		found := false
		for _, g := range genres {
			if g.(string) == trackGenre {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if query, ok := criteria["query"].(string); ok {
		title := track["title"].(string)
		if !contains(title, query) {
			return false
		}
	}
	return true
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || contains(s[1:], substr))
}
