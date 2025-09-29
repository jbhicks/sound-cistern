package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	client := &http.Client{Timeout: 10 * time.Second}

	// Exchange code for access token
	tokenURL := "https://api.soundcloud.com/oauth2/token"
	data := fmt.Sprintf("client_id=%s&client_secret=%s&redirect_uri=%s&grant_type=authorization_code&code=%s",
		s.ClientID, s.ClientSecret, s.RedirectURI, code)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("token exchange failed: %d", res.StatusCode)
	}

	var tokenResponse map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("no access token in response")
	}

	// Fetch user info
	userInfo, err := s.fetchUserInfo(accessToken)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token": accessToken,
		"user_info":    userInfo,
	}, nil
}

// fetchUserInfo fetches user information from Soundcloud
func (s *SoundcloudService) fetchUserInfo(accessToken string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://api.soundcloud.com/me", nil)
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
		return nil, fmt.Errorf("user info API error: %d", res.StatusCode)
	}
	var userInfo map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&userInfo)
	return userInfo, err
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
