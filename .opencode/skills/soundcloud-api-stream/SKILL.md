---
name: soundcloud-api-stream
description: Soundcloud Stream/Activities API - user's activity feed and stream endpoints
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: soundcloud-api
---

# Soundcloud Stream/Activities API

## Overview

The Activities API provides the user's stream/feed - tracks from artists they follow, reposts, and their own uploads.

## Endpoints

### Get User's Activity Stream

```
GET /me/activities
```

**Request:**
```http
GET https://api.soundcloud.com/me/activities?limit=50 HTTP/1.1
Authorization: Bearer {access_token}
Accept: application/json; charset=utf-8
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Number of results (default: 50, max: 200) |
| `offset` | int | Pagination offset |
| `linked_partitioning` | bool | Enable cursor-based pagination |
| `filters` | string | Filter activities by type |

**Response:**
```json
{
  "filters": {
    "date_from": "",
    "date_to": "",
    "duration_max": "",
    "duration_min": "",
    "favorites": false,
    "genre": "",
    "genres": [],
    "search": "",
    "sort": ""
  },
  "pagination": {
    "has_more": true,
    "limit": 50,
    "page": 1,
    "total": 150,
    "total_pages": 3
  },
  "tracks": [
    {
      "id": 2267275367,
      "title": "Track Title",
      "duration": 4180558,
      "permalink_url": "https://soundcloud.com/user/track",
      "artwork_url": "https://i1.sndcdn.com/artworks-xxx-large.jpg",
      "genre": "Electronic",
      "created_at": "2026-02-16T00:37:06Z",
      "is_repost": false,
      "user": {
        "id": 123456,
        "username": "Artist Name",
        "avatar_url": "https://i1.sndcdn.com/avatars-xxx-large.jpg"
      }
    }
  ]
}
```

### Activity Types

The stream includes various activity types:
- `track` - Original track upload
- `track-repost` - User reposted someone else's track
- `playlist` - Playlist creation
- `playlist-repost` - Reposted playlist

## Go Implementation

### Fetch Activities

```go
type ActivitiesResponse struct {
    Filters    Filters              `json:"filters"`
    Pagination Pagination           `json:"pagination"`
    Tracks     []SoundcloudTrack    `json:"tracks"`
}

type Filters struct {
    DateFrom    string   `json:"date_from"`
    DateTo      string   `json:"date_to"`
    DurationMax string   `json:"duration_max"`
    DurationMin string   `json:"duration_min"`
    Favorites   bool     `json:"favorites"`
    Genre       string   `json:"genre"`
    Genres      []string `json:"genres"`
    Search      string   `json:"search"`
    Sort        string   `json:"sort"`
}

type Pagination struct {
    HasMore    bool `json:"has_more"`
    Limit      int  `json:"limit"`
    Page       int  `json:"page"`
    Total      int  `json:"total"`
    TotalPages int  `json:"total_pages"`
}

func (c *SoundcloudClient) GetActivities(limit int) (*ActivitiesResponse, error) {
    params := url.Values{}
    params.Set("limit", strconv.Itoa(limit))
    
    resp, err := c.doRequest("GET", "/me/activities", params)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, handleSoundcloudError(resp)
    }
    
    var activities ActivitiesResponse
    if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
        return nil, err
    }
    
    return &activities, nil
}
```

### Paginated Fetch

```go
func (c *SoundcloudClient) GetAllActivities(maxPages int) ([]SoundcloudTrack, error) {
    var allTracks []SoundcloudTrack
    page := 1
    
    for page <= maxPages {
        params := url.Values{}
        params.Set("limit", "200")
        params.Set("page", strconv.Itoa(page))
        
        resp, err := c.doRequest("GET", "/me/activities", params)
        if err != nil {
            return nil, err
        }
        
        var activities ActivitiesResponse
        if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
            resp.Body.Close()
            return nil, err
        }
        resp.Body.Close()
        
        allTracks = append(allTracks, activities.Tracks...)
        
        if !activities.Pagination.HasMore {
            break
        }
        page++
        
        // Rate limiting
        time.Sleep(100 * time.Millisecond)
    }
    
    return allTracks, nil
}
```

## Mock Data

### Mock Activities Response

```go
func getMockActivities() map[string]interface{} {
    return map[string]interface{}{
        "filters": map[string]interface{}{
            "date_from":    "",
            "date_to":      "",
            "duration_max": "",
            "duration_min": "",
            "favorites":    false,
            "genre":        "",
            "genres":       []interface{}{},
            "search":       "",
            "sort":         "",
        },
        "pagination": map[string]interface{}{
            "has_more":    false,
            "limit":       20,
            "page":        1,
            "total":       7,
            "total_pages": 1,
        },
        "tracks": []map[string]interface{}{
            {
                "id":            float64(2267275367),
                "title":         "DJ PFUNK Mix \"Frequencies\"",
                "duration":      float64(4180558),
                "permalink_url": "https://soundcloud.com/paul-barnard-647247988/dj-pfunk-mix-frequencies",
                "artwork_url":   "https://i1.sndcdn.com/artworks-HVGxK1ZyjyfIrh73-ZfBskw-large.jpg",
                "genre":         "Breakbeat",
                "created_at":    "2026-02-16T00:37:06Z",
                "is_repost":     false,
                "user": map[string]interface{}{
                    "id":         float64(123456),
                    "username":   "DJ PFunk",
                    "avatar_url": "https://i1.sndcdn.com/avatars-HVGxK1ZyjyfIrh73-ZfBskw-large.jpg",
                },
            },
            // ... more tracks
        },
    }
}

func mockSoundcloudActivitiesResponse(c echo.Context) error {
    return c.JSON(http.StatusOK, getMockActivities())
}
```

## Sync Pattern

### Auto-sync on Login

```go
// Triggered after OAuth callback
func autoSyncActivities(app *pocketbase.PocketBase, userID, accessToken string) error {
    client := NewSoundcloudClient(accessToken)
    
    activities, err := client.GetActivities(200)
    if err != nil {
        return err
    }
    
    tracksCollection, _ := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
    soundcloudUser := getSoundcloudUserRecord(app, userID)
    
    for _, track := range activities.Tracks {
        upsertTrack(app, tracksCollection, track, soundcloudUser.Id)
    }
    
    return nil
}
```

### Upsert Track

```go
func upsertTrack(app *pocketbase.PocketBase, collection *models.Collection, track SoundcloudTrack, soundcloudUserID string) *models.Record {
    soundcloudID := strconv.FormatInt(track.ID, 10)
    
    existing, _ := app.Dao().FindFirstRecordByFilter(
        collection.Id,
        "soundcloud_id = {:id}",
        map[string]any{"id": soundcloudID},
    )
    
    var record *models.Record
    if existing != nil {
        record = existing
    } else {
        record = models.NewRecord(collection)
        record.Set("soundcloud_id", soundcloudID)
    }
    
    record.Set("user_id", soundcloudUserID)
    record.Set("title", track.Title)
    record.Set("artist_name", track.User.Username)
    record.Set("duration", track.Duration)
    record.Set("genre", track.Genre)
    record.Set("artwork_url", track.ArtworkURL)
    record.Set("permalink_url", track.PermalinkURL)
    record.Set("stream_url", track.StreamURL)
    record.Set("user_avatar_url", track.User.AvatarURL)
    record.Set("is_repost", track.IsRepost)
    record.Set("playback_count", track.PlaybackCount)
    record.Set("favoritings_count", track.FavoritingsCount)
    
    app.Dao().SaveRecord(record)
    return record
}
```

## Testing

### Unit Test

```go
func TestGetActivities(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/me/activities", r.URL.Path)
        assert.Equal(t, "50", r.URL.Query().Get("limit"))
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "filters":    map[string]interface{}{},
            "pagination": map[string]interface{}{"has_more": false, "total": 1},
            "tracks": []map[string]interface{}{
                {"id": float64(123), "title": "Test Track"},
            },
        })
    }))
    defer server.Close()
    
    client := NewSoundcloudClient("test_token")
    client.baseURL = server.URL
    
    activities, err := client.GetActivities(50)
    require.NoError(t, err)
    assert.Len(t, activities.Tracks, 1)
}
```