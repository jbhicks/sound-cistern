package views

import "time"

type Track struct {
	TrackID       string    `json:"track_id"`
	TrackTitle    string    `json:"track_title"`
	ArtistName    string    `json:"artist_name"`
	TrackDuration int64     `json:"track_duration"`
	ArtworkURL    string    `json:"artwork_url"`
	CreatedAt     time.Time `json:"created_at"`
	PermalinkURL  string    `json:"permalink_url"`
	IsFavorited   bool      `json:"is_favorited"`
}
