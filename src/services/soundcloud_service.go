package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SoundcloudService handles Soundcloud API interactions
type SoundcloudService struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// NewSoundcloudService creates a new service
func NewSoundcloudService(clientID, clientSecret, redirectURI string) *SoundcloudService {
	return &SoundcloudService{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}
}

// GetAuthURL returns the Soundcloud OAuth URL
func (s *SoundcloudService) GetAuthURL() string {
	return fmt.Sprintf("https://soundcloud.com/connect?client_id=%s&redirect_uri=%s&response_type=code", s.ClientID, s.RedirectURI)
}

// HandleCallback exchanges code for access token and creates user
func (s *SoundcloudService) HandleCallback(code string) (interface{}, error) {
	// TODO: Exchange code for token, fetch user info
	return map[string]interface{}{}, nil
}

// FetchUserFeed fetches user's feed from Soundcloud
func (s *SoundcloudService) FetchUserFeed(accessToken string) ([]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.soundcloud.com/me/tracks", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("API error: %d", res.StatusCode)
	}
	var tracks []interface{}
	err = json.NewDecoder(res.Body).Decode(&tracks)
	return tracks, err
}
