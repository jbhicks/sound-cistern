package models

import (
	"github.com/gofrs/uuid"
	"time"
)

// Feed represents a user's cached feed
type Feed struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Tracks    string    `json:"tracks" db:"tracks"` // JSON array of track IDs
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Feeds is a slice of Feed
type Feeds []Feed
