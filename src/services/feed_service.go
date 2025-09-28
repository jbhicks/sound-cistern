package services

// FeedService handles feed caching and filtering
type FeedService struct {
}

// NewFeedService creates a new service
func NewFeedService() *FeedService {
	return &FeedService{}
}

// CacheFeed caches the feed for a user
func (fs *FeedService) CacheFeed(userID string, tracks []interface{}) error {
	// TODO: Save to database
	return nil
}

// GetCachedFeed gets cached feed for user
func (fs *FeedService) GetCachedFeed(userID string) ([]interface{}, error) {
	// TODO: Load from database
	return []interface{}{}, nil
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
