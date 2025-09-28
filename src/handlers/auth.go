package handlers

import (
	"github.com/gobuffalo/buffalo"
)

// AuthHandler handles authentication
type AuthHandler struct {
	SoundcloudService interface{}
}

// NewAuthHandler creates a new handler
func NewAuthHandler(ss interface{}) *AuthHandler {
	return &AuthHandler{SoundcloudService: ss}
}

// SoundcloudLogin initiates Soundcloud OAuth
func (ah *AuthHandler) SoundcloudLogin(c buffalo.Context) error {
	// TODO: Get auth URL
	url := "https://soundcloud.com/connect"
	return c.Redirect(302, url)
}

// SoundcloudCallback handles the callback
func (ah *AuthHandler) SoundcloudCallback(c buffalo.Context) error {
	// TODO: Handle callback
	return c.Redirect(302, "/feed")
}
