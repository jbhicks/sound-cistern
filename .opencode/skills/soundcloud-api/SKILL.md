---
name: soundcloud-api
description: Complete Soundcloud API reference with endpoint documentation, response schemas, and implementation patterns
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: soundcloud-api
---

# Soundcloud API Reference

## Overview

This skill provides comprehensive documentation for the Soundcloud Public API, including all endpoints, response schemas, and implementation patterns. Each endpoint category has its own sub-skill for detailed implementation guidance.

## Base URLs

- **API v1**: `https://api.soundcloud.com`
- **API v2**: `https://api-v2.soundcloud.com`
- **OAuth**: `https://secure.soundcloud.com` (auth) / `https://api.soundcloud.com` (token)

## Authentication

All Soundcloud API endpoints require authentication via Bearer token:

```
Authorization: Bearer {access_token}
```

See `soundcloud-oauth` skill for complete OAuth flow implementation.

## API Endpoint Categories

### Me (Authenticated User)

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/me` | GET | Get authenticated user info | `soundcloud-api-me` |
| `/me/activities` | GET | Get user's activity stream | `soundcloud-api-stream` |
| `/me/favorites` | GET | Get user's favorited tracks | `soundcloud-api-favorites` |
| `/me/likes/tracks` | GET | Get user's liked tracks (newer) | `soundcloud-api-favorites` |
| `/me/tracks` | GET | Get user's uploaded tracks | `soundcloud-api-tracks` |
| `/me/playlists` | GET | Get user's playlists | `soundcloud-api-playlists` |
| `/me/followings` | GET | Get users the authenticated user follows | `soundcloud-api-follows` |
| `/me/followers` | GET | Get users following the authenticated user | `soundcloud-api-follows` |

### Users

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/users/{id}` | GET | Get user by ID | `soundcloud-api-users` |
| `/users/{id}/tracks` | GET | Get user's tracks | `soundcloud-api-tracks` |
| `/users/{id}/favorites` | GET | Get user's favorites (deprecated) | `soundcloud-api-favorites` |
| `/users/{id}/likes/tracks` | GET | Get user's liked tracks | `soundcloud-api-favorites` |
| `/users/{id}/playlists` | GET | Get user's playlists | `soundcloud-api-playlists` |

### Tracks

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/tracks/{id}` | GET | Get track by ID | `soundcloud-api-tracks` |
| `/tracks/{id}/stream` | GET | Stream track (requires auth) | `soundcloud-api-stream` |
| `/tracks/{id}/related` | GET | Get related tracks | `soundcloud-api-tracks` |

### Playlists

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/playlists/{id}` | GET | Get playlist by ID | `soundcloud-api-playlists` |

### Search (API v2)

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/search` | GET | Search tracks, users, playlists | `soundcloud-api-search` |
| `/search/tracks` | GET | Search tracks only | `soundcloud-api-search` |
| `/search/users` | GET | Search users only | `soundcloud-api-search` |

### Likes & Favorites

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/me/favorites/{id}` | PUT | Like/favorite a track | `soundcloud-api-favorites` |
| `/me/favorites/{id}` | DELETE | Unlike/unfavorite a track | `soundcloud-api-favorites` |

### Reposts

| Endpoint | Method | Description | Sub-skill |
|----------|--------|-------------|-----------|
| `/me/likes/tracks/{id}` | PUT | Like a track (newer) | `soundcloud-api-favorites` |
| `/me/likes/tracks/{id}` | DELETE | Unlike a track (newer) | `soundcloud-api-favorites` |

## Common Response Schemas

### Track Object

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
  "tag_list": "\"tag1\" \"tag2\"",
  "streamable": true,
  "embeddable_by": "all",
  "genre": "Electronic",
  "title": "Track Title",
  "description": "Track description",
  "permalink_url": "https://soundcloud.com/user/track",
  "artwork_url": "https://i1.sndcdn.com/artworks-xxx-large.jpg",
  "stream_url": "https://api.soundcloud.com/tracks/soundcloud:tracks:xxx/preview",
  "user": {
    "kind": "user",
    "id": 7133251,
    "urn": "soundcloud:users:7133251",
    "username": "Artist Name",
    "permalink": "artist-name",
    "avatar_url": "https://i1.sndcdn.com/avatars-xxx-large.jpg",
    "city": "City",
    "country": "Country",
    "followers_count": 4302,
    "track_count": 155
  },
  "playback_count": 1643,
  "favoritings_count": 189,
  "reposts_count": 20,
  "user_favorite": false,
  "user_playback_count": 1,
  "downloadable": false,
  "access": "playable"
}
```

### User Object

```json
{
  "kind": "user",
  "id": 141564746,
  "urn": "soundcloud:users:141564746",
  "username": "username",
  "display_name": "Display Name",
  "permalink": "permalink",
  "avatar_url": "https://i1.sndcdn.com/avatars-xxx-large.jpg",
  "city": "City",
  "country": "Country",
  "description": "User bio",
  "track_count": 50,
  "followers_count": 7333,
  "followings_count": 1099,
  "public_favorites_count": 25,
  "playlist_count": 10
}
```

### Activities Response (Stream)

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
    "limit": 20,
    "page": 1,
    "total": 100,
    "total_pages": 5
  },
  "tracks": [
    { /* Track object */ }
  ]
}
```

### Favorites Response

```json
[
  { /* Track object with user_favorite: true */ },
  { /* Track object with user_favorite: true */ }
]
```

Note: Unlike `/me/activities`, `/me/favorites` returns a plain array, not a wrapped object.

## Implementation Patterns

### HTTP Client Setup

```go
type SoundcloudClient struct {
    accessToken string
    httpClient  *http.Client
    baseURL     string
}

func NewSoundcloudClient(accessToken string) *SoundcloudClient {
    return &SoundcloudClient{
        accessToken: accessToken,
        httpClient:  &http.Client{Timeout: 30 * time.Second},
        baseURL:     "https://api.soundcloud.com",
    }
}

func (c *SoundcloudClient) doRequest(method, path string, params url.Values) (*http.Response, error) {
    fullURL := c.baseURL + path
    if params != nil {
        fullURL += "?" + params.Encode()
    }
    
    req, err := http.NewRequest(method, fullURL, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", "Bearer "+c.accessToken)
    req.Header.Set("Accept", "application/json; charset=utf-8")
    
    return c.httpClient.Do(req)
}
```

### Error Handling

```go
type SoundcloudError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Link    string `json:"link"`
    Status  string `json:"status"`
}

func handleSoundcloudError(resp *http.Response) error {
    body, _ := io.ReadAll(resp.Body)
    
    var scErr SoundcloudError
    if json.Unmarshal(body, &scErr) == nil && scErr.Message != "" {
        return fmt.Errorf("soundcloud error %d: %s", scErr.Code, scErr.Message)
    }
    
    return fmt.Errorf("soundcloud error %d: %s", resp.StatusCode, string(body))
}
```

### Pagination Pattern

```go
type PaginatedResponse struct {
    Collection   []interface{} `json:"collection"`
    NextHref     string        `json:"next_href"`
    PrevHref     string        `json:"prev_href"`
}

func (c *SoundcloudClient) FetchAllPages(path string) ([]interface{}, error) {
    var allItems []interface{}
    nextURL := c.baseURL + path
    
    for nextURL != "" {
        resp, err := c.doRequest("GET", strings.TrimPrefix(nextURL, c.baseURL), nil)
        if err != nil {
            return nil, err
        }
        
        var page PaginatedResponse
        if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
            resp.Body.Close()
            return nil, err
        }
        resp.Body.Close()
        
        allItems = append(allItems, page.Collection...)
        nextURL = page.NextHref
        
        // Rate limiting
        time.Sleep(100 * time.Millisecond)
    }
    
    return allItems, nil
}
```

## Rate Limits

- OAuth tokens: 50 per 12 hours per app, 30 per hour per IP
- API requests: Varies by endpoint, generally ~600/5min window
- Implement exponential backoff and caching

## Mock Responses for Testing

The following mock functions are available in `main.go` for testing:

| Mock Function | Endpoint | Purpose |
|---------------|----------|---------|
| `mockSoundcloudActivitiesResponse` | `/me/activities` | Returns mock stream data |
| `mockSoundcloudFavoritesResponse` | `/me/favorites` | Returns mock favorites array |
| `mockSoundcloudUserResponse` | `/me` | Returns mock user info |
| `mockSoundcloudTokenResponse` | `/oauth2/token` | Returns mock tokens |

## Sub-skills

Load individual sub-skills for detailed implementation:

- `/skill soundcloud-api-me` - User profile endpoints
- `/skill soundcloud-api-stream` - Activity stream endpoints
- `/skill soundcloud-api-favorites` - Favorites/likes endpoints
- `/skill soundcloud-api-tracks` - Track endpoints
- `/skill soundcloud-api-playlists` - Playlist endpoints
- `/skill soundcloud-api-search` - Search endpoints
- `/skill soundcloud-api-follows` - Following/followers endpoints

## Related Skills

- `soundcloud-oauth` - OAuth authentication flow
- `sound-cistern-dev` - Development server setup