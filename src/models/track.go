package models

import (
	"github.com/gofrs/uuid"
	"time"
)

// Track represents a Soundcloud track
type Track struct {
	ID           uuid.UUID `json:"id" db:"id"`
	UserID       uuid.UUID `json:"user_id" db:"user_id"`
	SoundcloudID string    `json:"soundcloud_id" db:"soundcloud_id"`
	Title        string    `json:"title" db:"title"`
	Length       int       `json:"length" db:"length"`
	Genre        string    `json:"genre" db:"genre"`
	PostTime     time.Time `json:"post_time" db:"post_time"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Tracks is a slice of Track
type Tracks []Track
