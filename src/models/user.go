package models

import (
	"github.com/gofrs/uuid"
	"time"
)

// User represents a Soundcloud user
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	SoundcloudID string    `json:"soundcloud_id" db:"soundcloud_id"`
	AccessToken  string    `json:"access_token" db:"access_token"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Users is a slice of User
type Users []User
