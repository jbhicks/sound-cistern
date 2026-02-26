---
name: soundcloud-api-tracks
description: Soundcloud Tracks API - fetch track details, streaming, and track management
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: soundcloud-api
---

# Soundcloud Tracks API

## Overview

The Tracks API provides access to track details, streaming URLs, and related tracks.

## Endpoints

### Get Track by ID

```
GET /tracks/{track_id}
```

**Request:**
```http
GET https://api.soundcloud.com/tracks/2246814242 HTTP/1.1
Authorization: Bearer {access_token}
Accept: application/json; charset=utf-8
```

**Response:**
```json
{
  "kind": "track",
  "id": 2246814242,
  "urn": "soundcloud:tracks:2246814242",
  "created_at": "2026/01/13 17:02:45 +0000",
  "duration": 5006576,
  "commentable": true,
  "comment_count": 34,
  "sharing": "public",
  "tag_list": "\"breakbeat\" \"dubstep\"",
  "streamable": true,
  "embeddable_by": "all",
  "purchase_url": null,
  "purchase_title": null,
  "genre": "Breakbeat",
  "title": "Track Title",
  "description": "Track description",
  "label_name": "",
  "release": null,
  "key_signature": null,
  "isrc": null,
  "bpm": 140,
  "release_year": 2026,
  "release_month": 1,
  "release_day": 13,
  "license": "all-rights-reserved",
  "uri": "https://api.soundcloud.com/tracks/soundcloud:tracks:2246814242",
  "user": {
    "kind": "user",
    "id": 7133251,
    "urn": "soundcloud:users:7133251",
    "username": "Artist Name",
    "permalink": "artist-name",
    "avatar_url": "https://i1.sndcdn.com/avatars-xxx-large.jpg",
    "city": "City",
    "country": "Country"
  },
  "permalink_url": "https://soundcloud.com/user/track",
  "artwork_url": "https://i1.sndcdn.com/artworks-xxx-large.jpg",
  "stream_url": "https://api.soundcloud.com/tracks/soundcloud:tracks:2246814242/preview",
  "download_url": null,
  "waveform_url": "https://wave.sndcdn.com/xxx_m.png",
  "playback_count": 1643,
  "download_count": 0,
  "favoritings_count": 189,
  "reposts_count": 20,
  "downloadable": false,
  "access": "playable",
  "policy": null
}
```

### Stream Track

```
GET /tracks/{track_id}/stream
```

**Request:**
```http
GET https://api.soundcloud.com/tracks/2246814242/stream HTTP/1.1
Authorization: Bearer {access_token}
```

**Response:** 302 Redirect to audio file URL

Note: This redirects to the actual audio file. Handle redirects in your HTTP client.

### Get Related Tracks

```
GET /tracks/{track_id}/related
```

**Request:**
```http
GET https://api.soundcloud.com/tracks/2246814242/related?limit=10 HTTP/1.1
Authorization: Bearer {access_token}
```

**Response:** Array of track objects

### Get User's Tracks

```
GET /me/tracks
GET /users/{user_id}/tracks
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Number of results |
| `offset` | int | Pagination offset |
| `linked_partitioning` | bool | Enable cursor pagination |

## Go Implementation

### Track Model

```go
type SoundcloudTrack struct {
    Kind            string          `json:"kind"`
    ID              int64           `json:"id"`
    URN             string          `json:"urn"`
    CreatedAt       string          `json:"created_at"`
    Duration        int             `json:"duration"`
    Commentable     bool            `json:"commentable"`
    CommentCount    int             `json:"comment_count"`
    Sharing         string          `json:"sharing"`
    TagList         string          `json:"tag_list"`
    Streamable      bool            `json:"streamable"`
    EmbeddableBy    string          `json:"embeddable_by"`
    Genre           string          `json:"genre"`
    Title           string          `json:"title"`
    Description     string          `json:"description"`
    PermalinkURL    string          `json:"permalink_url"`
    ArtworkURL      string          `json:"artwork_url"`
    StreamURL       string          `json:"stream_url"`
    DownloadURL     string          `json:"download_url"`
    WaveformURL     string          `json:"waveform_url"`
    PlaybackCount   int             `json:"playback_count"`
    DownloadCount   int             `json:"download_count"`
    FavoritingsCount int            `json:"favoritings_count"`
    RepostsCount    int             `json:"reposts_count"`
    Downloadable    bool            `json:"downloadable"`
    Access          string          `json:"access"`
    BPM             float64         `json:"bpm"`
    ReleaseYear     int             `json:"release_year"`
    User            SoundcloudUser  `json:"user"`
}

type SoundcloudUser struct {
    Kind         string `json:"kind"`
    ID           int64  `json:"id"`
    URN          string `json:"urn"`
    Username     string `json:"username"`
    Permalink    string `json:"permalink"`
    AvatarURL    string `json:"avatar_url"`
    City         string `json:"city"`
    Country      string `json:"country"`
    FollowersCount int  `json:"followers_count"`
    TrackCount   int    `json:"track_count"`
}
```

### Get Track

```go
func (c *SoundcloudClient) GetTrack(trackID int64) (*SoundcloudTrack, error) {
    path := fmt.Sprintf("/tracks/%d", trackID)
    
    resp, err := c.doRequest("GET", path, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, handleSoundcloudError(resp)
    }
    
    var track SoundcloudTrack
    if err := json.NewDecoder(resp.Body).Decode(&track); err != nil {
        return nil, err
    }
    
    return &track, nil
}
```

### Stream Track

```go
func (c *SoundcloudClient) GetStreamURL(trackID int64) (string, error) {
    path := fmt.Sprintf("/tracks/%d/stream", trackID)
    
    // Don't follow redirects - we want the redirect URL
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }
    
    req, _ := http.NewRequest("GET", c.baseURL+path, nil)
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusTemporaryRedirect {
        return "", fmt.Errorf("stream not available")
    }
    
    location := resp.Header.Get("Location")
    if location == "" {
        return "", fmt.Errorf("no stream URL in response")
    }
    
    return location, nil
}
```

### Get Related Tracks

```go
func (c *SoundcloudClient) GetRelatedTracks(trackID int64, limit int) ([]SoundcloudTrack, error) {
    path := fmt.Sprintf("/tracks/%d/related", trackID)
    
    params := url.Values{}
    params.Set("limit", strconv.Itoa(limit))
    
    resp, err := c.doRequest("GET", path, params)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, handleSoundcloudError(resp)
    }
    
    var tracks []SoundcloudTrack
    if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
        return nil, err
    }
    
    return tracks, nil
}
```

### Get User's Tracks

```go
func (c *SoundcloudClient) GetUserTracks(userID int64, limit int) ([]SoundcloudTrack, error) {
    path := fmt.Sprintf("/users/%d/tracks", userID)
    
    params := url.Values{}
    params.Set("limit", strconv.Itoa(limit))
    
    resp, err := c.doRequest("GET", path, params)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, handleSoundcloudError(resp)
    }
    
    var tracks []SoundcloudTrack
    if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
        return nil, err
    }
    
    return tracks, nil
}

func (c *SoundcloudClient) GetMyTracks(limit int) ([]SoundcloudTrack, error) {
    params := url.Values{}
    params.Set("limit", strconv.Itoa(limit))
    
    resp, err := c.doRequest("GET", "/me/tracks", params)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, handleSoundcloudError(resp)
    }
    
    var tracks []SoundcloudTrack
    if err := json.NewDecoder(resp.Body).Decode(&tracks); err != nil {
        return nil, err
    }
    
    return tracks, nil
}
```

## Database Schema

### soundcloud_tracks Collection

```go
// pb_migrations/1696000002_create_soundcloud_tracks.go
collection.Fields.Add(
    &core.TextField{
        Name:     "soundcloud_id",
        Required: true,
    },
    &core.TextField{
        Name:     "user_id", // links to soundcloud_users
        Required: true,
    },
    &core.TextField{
        Name: "title",
    },
    &core.TextField{
        Name: "artist_name",
    },
    &core.TextField{
        Name: "genre",
    },
    &core.TextField{
        Name: "tag_list",
    },
    &core.NumberField{
        Name: "duration", // milliseconds
    },
    &core.NumberField{
        Name: "length", // seconds (calculated from duration)
    },
    &core.TextField{
        Name: "artwork_url",
    },
    &core.TextField{
        Name: "permalink_url",
    },
    &core.TextField{
        Name: "stream_url",
    },
    &core.TextField{
        Name: "user_avatar_url",
    },
    &core.BoolField{
        Name:    "is_repost",
        Default: false,
    },
    &core.NumberField{
        Name: "playback_count",
    },
    &core.NumberField{
        Name: "favoritings_count",
    },
    &core.NumberField{
        Name: "comment_count",
    },
    &core.NumberField{
        Name: "reposts_count",
    },
    &core.NumberField{
        Name: "bpm",
    },
    &core.NumberField{
        Name: "release_year",
    },
    &core.BoolField{
        Name:    "downloadable",
        Default: false,
    },
    &core.DateTimeField{
        Name: "soundcloud_created_at",
    },
)
```

## Testing

### Unit Test

```go
func TestGetTrack(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/tracks/123", r.URL.Path)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":         float64(123),
            "title":      "Test Track",
            "duration":   float64(180000),
            "streamable": true,
            "user": map[string]interface{}{
                "id":       float64(456),
                "username": "Test Artist",
            },
        })
    }))
    defer server.Close()
    
    client := NewSoundcloudClient("test_token")
    client.baseURL = server.URL
    
    track, err := client.GetTrack(123)
    require.NoError(t, err)
    assert.Equal(t, int64(123), track.ID)
    assert.Equal(t, "Test Track", track.Title)
}
```