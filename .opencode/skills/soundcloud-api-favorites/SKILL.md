---
name: soundcloud-api-favorites
description: Soundcloud Favorites/Likes API endpoints - fetch, sync, and manage user favorites
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: soundcloud-api
---

# Soundcloud Favorites/Likes API

## Overview

The Favorites API provides access to a user's liked/favorited tracks. Soundcloud is transitioning from `/favorites` to `/likes/tracks` nomenclature.

## Endpoints

### Get Authenticated User's Favorites

```
GET /me/favorites
GET /me/likes/tracks (newer)
```

**Request:**
```http
GET https://api.soundcloud.com/me/favorites HTTP/1.1
Authorization: Bearer {access_token}
Accept: application/json; charset=utf-8
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Number of results (default: 50, max: 200) |
| `linked_partitioning` | bool | Enable pagination via `next_href` |
| `offset` | int | Offset for pagination (deprecated, use cursor) |

**Response:**
```json
[
  {
    "kind": "track",
    "id": 2246814242,
    "urn": "soundcloud:tracks:2246814242",
    "created_at": "2026/01/13 17:02:45 +0000",
    "duration": 5006576,
    "commentable": true,
    "sharing": "public",
    "streamable": true,
    "genre": "Electronic",
    "title": "Track Title",
    "description": "Track description",
    "permalink_url": "https://soundcloud.com/user/track",
    "artwork_url": "https://i1.sndcdn.com/artworks-xxx-large.jpg",
    "stream_url": "https://api.soundcloud.com/tracks/soundcloud:tracks:xxx/preview",
    "user": {
      "kind": "user",
      "id": 7133251,
      "username": "Artist Name",
      "permalink": "artist-name",
      "avatar_url": "https://i1.sndcdn.com/avatars-xxx-large.jpg",
      "city": "City",
      "country": "Country"
    },
    "playback_count": 1643,
    "favoritings_count": 189,
    "user_favorite": true
  }
]
```

**Key Differences from Activities:**
- Returns a **plain array**, not a wrapped object with `tracks` key
- Each track has `user_favorite: true`
- No pagination metadata wrapper (use `linked_partitioning` for cursor)

### Get User's Favorites by User ID

```
GET /users/{user_id}/favorites (deprecated)
GET /users/{user_id}/likes/tracks (newer)
```

**Request:**
```http
GET https://api.soundcloud.com/users/141564746/likes/tracks HTTP/1.1
Authorization: Bearer {access_token}
```

### Like/Favorite a Track

```
PUT /me/favorites/{track_id}
PUT /me/likes/tracks/{track_id} (newer)
```

**Request:**
```http
PUT https://api.soundcloud.com/me/favorites/2246814242 HTTP/1.1
Authorization: Bearer {access_token}
```

**Response:** 201 Created with track object

### Unlike/Unfavorite a Track

```
DELETE /me/favorites/{track_id}
DELETE /me/likes/tracks/{track_id} (newer)
```

**Request:**
```http
DELETE https://api.soundcloud.com/me/favorites/2246814242 HTTP/1.1
Authorization: Bearer {access_token}
```

**Response:** 200 OK or 204 No Content

## Go Implementation

### Fetch Favorites

```go
func (c *SoundcloudClient) GetFavorites(limit int) ([]SoundcloudTrack, error) {
    params := url.Values{}
    params.Set("limit", strconv.Itoa(limit))
    
    resp, err := c.doRequest("GET", "/me/favorites", params)
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

### Like a Track

```go
func (c *SoundcloudClient) LikeTrack(trackID int64) error {
    path := fmt.Sprintf("/me/favorites/%d", trackID)
    
    req, err := http.NewRequest("PUT", c.baseURL+path, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
        return handleSoundcloudError(resp)
    }
    
    return nil
}
```

### Unlike a Track

```go
func (c *SoundcloudClient) UnlikeTrack(trackID int64) error {
    path := fmt.Sprintf("/me/favorites/%d", trackID)
    
    req, err := http.NewRequest("DELETE", c.baseURL+path, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
        return handleSoundcloudError(resp)
    }
    
    return nil
}
```

## Mock Data

### Mock Favorites Response

```go
func getMockFavorites() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "kind":         "track",
            "id":           float64(2246814242),
            "urn":          "soundcloud:tracks:2246814242",
            "created_at":   "2026/01/13 17:02:45 +0000",
            "duration":     float64(5006576),
            "sharing":      "public",
            "streamable":   true,
            "genre":        "",
            "title":        "BobbyBuzZ - MAD SCIENTIST Serato stems live remix set",
            "permalink_url": "https://soundcloud.com/dj-bobby-buzz/stems-mix",
            "artwork_url":   "https://i1.sndcdn.com/artworks-0jamvqjNnoPKBi3S-NI2Cog-large.jpg",
            "stream_url":    "https://api.soundcloud.com/tracks/soundcloud:tracks:2246814242/preview",
            "user": map[string]interface{}{
                "avatar_url": "https://i1.sndcdn.com/avatars-ZqQI7w45r30zdKdp-suOV9A-large.jpg",
                "id":         float64(7133251),
                "username":   "BobbyBuzZ-TheoryonRecords",
                "permalink":  "dj-bobby-buzz",
                "city":       "Coral Springs",
                "country":    "United States",
            },
            "playback_count":    float64(1643),
            "favoritings_count": float64(189),
            "user_favorite":     true,
        },
    }
}

func mockSoundcloudFavoritesResponse(c echo.Context) error {
    return c.JSON(http.StatusOK, getMockFavorites())
}
```

## Sync Pattern

### Initial Load + Progressive Sync

```go
// Initial load from database (fast)
func loadFavoritesFromDB(app *pocketbase.PocketBase, userID string) ([]Favorite, error) {
    records, err := app.Dao().FindRecordsByFilter(
        "favorites",
        "user_id = {:user_id}",
        "-created",
        100, 0,
        map[string]any{"user_id": userID},
    )
    // ... process records
}

// Sync from Soundcloud API (background)
func syncFavoritesFromSoundcloud(app *pocketbase.PocketBase, userID string, accessToken string) error {
    client := NewSoundcloudClient(accessToken)
    
    tracks, err := client.GetFavorites(200)
    if err != nil {
        return err
    }
    
    // Get existing favorites to compare
    existing, _ := app.Dao().FindRecordsByFilter(
        "favorites",
        "user_id = {:user_id}",
        "", 0, 0,
        map[string]any{"user_id": userID},
    )
    
    existingIDs := make(map[string]bool)
    for _, r := range existing {
        existingIDs[r.GetString("soundcloud_id")] = true
    }
    
    // Add new favorites
    tracksCollection, _ := app.Dao().FindCollectionByNameOrId("soundcloud_tracks")
    favoritesCollection, _ := app.Dao().FindCollectionByNameOrId("favorites")
    
    for _, track := range tracks {
        // Ensure track exists in soundcloud_tracks
        trackRecord := upsertTrack(app, tracksCollection, track)
        
        // Add to favorites if not exists
        if !existingIDs[strconv.FormatInt(track.ID, 10)] {
            fav := models.NewRecord(favoritesCollection)
            fav.Set("user_id", userID)
            fav.Set("track_id", trackRecord.Id)
            fav.Set("soundcloud_id", track.ID)
            fav.Set("synced_at", time.Now())
            app.Dao().SaveRecord(fav)
        }
    }
    
    return nil
}
```

### API Endpoint Pattern

```go
// GET /api/favorites - Load from DB, trigger background sync
e.Router.GET("/api/favorites", func(c echo.Context) error {
    authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
    
    // 1. Load from database (immediate response)
    favorites := loadFavoritesFromDB(app, authRecord.Id)
    
    // 2. Trigger background sync (non-blocking)
    go func() {
        soundcloudUser := getSoundcloudUser(app, authRecord.Id)
        syncFavoritesFromSoundcloud(app, authRecord.Id, soundcloudUser.AccessToken)
    }()
    
    return c.JSON(http.StatusOK, map[string]interface{}{
        "favorites": favorites,
        "syncing": true,
    })
})

// POST /api/favorites/sync - Force full sync
e.Router.POST("/api/favorites/sync", func(c echo.Context) error {
    authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
    soundcloudUser := getSoundcloudUser(app, authRecord.Id)
    
    if err := syncFavoritesFromSoundcloud(app, authRecord.Id, soundcloudUser.AccessToken); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }
    
    favorites := loadFavoritesFromDB(app, authRecord.Id)
    return c.JSON(http.StatusOK, map[string]interface{}{
        "favorites": favorites,
        "synced": true,
    })
})
```

## Testing

### Unit Test

```go
func TestGetFavorites(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/me/favorites", r.URL.Path)
        assert.Equal(t, "Bearer test_token", r.Header.Get("Authorization"))
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode([]map[string]interface{}{
            {"id": float64(123), "title": "Test Track", "user_favorite": true},
        })
    }))
    defer server.Close()
    
    client := NewSoundcloudClient("test_token")
    client.baseURL = server.URL
    
    tracks, err := client.GetFavorites(50)
    require.NoError(t, err)
    assert.Len(t, tracks, 1)
    assert.Equal(t, int64(123), tracks[0].ID)
}
```