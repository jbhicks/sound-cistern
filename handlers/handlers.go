package handlers

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/jbhicks/sound-cistern/views"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
)

func AnalyticsPage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		data := views.PageData{
			Title:       "Listening Analytics",
			Description: "Your listening habits and statistics",
			CurrentPath: "/analytics",
			User:        authRecord,
		}
		return views.Analytics(data).Render(c.Request().Context(), c.Response())
	}
}

func PlaylistsPage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		data := views.PageData{
			Title:       "Playlists",
			Description: "Your playlists",
			CurrentPath: "/playlists",
			User:        authRecord,
		}
		return views.Playlists(data, []views.PlaylistInfo{}).Render(c.Request().Context(), c.Response())
	}
}

func PlaylistShowPage(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		data := views.PageData{
			Title:       "Playlist",
			Description: "Playlist details",
			CurrentPath: "/playlists",
			User:        authRecord,
		}
		playlist := views.PlaylistDetail{
			ID: c.PathParam("id"),
		}
		return views.PlaylistShow(data, playlist).Render(c.Request().Context(), c.Response())
	}
}

// generateShareToken produces a random 12-character alphanumeric token.
func generateShareToken() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	token := make([]byte, 12)
	for i := range token {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[n.Int64()]
	}
	return string(token), nil
}

// trackCountForPlaylist returns the number of tracks in the given playlist.
func trackCountForPlaylist(app *pocketbase.PocketBase, playlistID string) int {
	records, err := app.Dao().FindRecordsByFilter(
		"playlist_tracks",
		"playlist_id = {:playlist_id}",
		"position",
		-1,
		0,
		map[string]any{"playlist_id": playlistID},
	)
	if err != nil {
		return 0
	}
	return len(records)
}

// fetchPlaylistTracks is shared between GetPlaylist and GetSharedPlaylist.
func fetchPlaylistTracks(app *pocketbase.PocketBase, playlistID string) ([]map[string]interface{}, error) {
	ptRecords, err := app.Dao().FindRecordsByFilter(
		"playlist_tracks",
		"playlist_id = {:playlist_id}",
		"position",
		-1,
		0,
		map[string]any{"playlist_id": playlistID},
	)
	if err != nil {
		return nil, err
	}

	tracks := make([]map[string]interface{}, 0, len(ptRecords))
	for _, pt := range ptRecords {
		trackRecord, err := app.Dao().FindRecordById("soundcloud_tracks", pt.GetString("track_id"))
		if err != nil {
			continue
		}
		tracks = append(tracks, map[string]interface{}{
			"track_id":          trackRecord.GetString("soundcloud_id"),
			"track_title":       trackRecord.GetString("title"),
			"artist_name":       trackRecord.GetString("artist_name"),
			"artwork_url":       trackRecord.GetString("artwork_url"),
			"track_duration":    trackRecord.GetInt("length"),
			"genre":             trackRecord.GetString("genre"),
			"playback_count":    trackRecord.GetInt("playback_count"),
			"favoritings_count": trackRecord.GetInt("favoritings_count"),
			"permalink_url":     trackRecord.GetString("permalink_url"),
			"pt_id":             pt.Id,
		})
	}
	return tracks, nil
}

// ListPlaylists returns all playlists for the authenticated user.
func ListPlaylists(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		records, err := app.Dao().FindRecordsByFilter(
			"playlists",
			"user_id = {:user_id}",
			"-created",
			-1,
			0,
			map[string]any{"user_id": authRecord.Id},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch playlists"})
		}

		playlists := make([]map[string]interface{}, 0, len(records))
		for _, r := range records {
			playlists = append(playlists, map[string]interface{}{
				"id":          r.Id,
				"name":        r.GetString("name"),
				"share_token": r.GetString("share_token"),
				"track_count": trackCountForPlaylist(app, r.Id),
				"created":     r.Created,
				"updated":     r.Updated,
			})
		}
		return c.JSON(http.StatusOK, playlists)
	}
}

// CreatePlaylist creates a new playlist for the authenticated user.
func CreatePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)

		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil || body.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
		}

		playlistsCollection, err := app.Dao().FindCollectionByNameOrId("playlists")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
		}

		record := models.NewRecord(playlistsCollection)
		record.Set("name", body.Name)
		record.Set("user_id", authRecord.Id)

		if err := app.Dao().SaveRecord(record); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create playlist"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":          record.Id,
			"name":        record.GetString("name"),
			"share_token": record.GetString("share_token"),
			"track_count": 0,
			"created":     record.Created,
			"updated":     record.Updated,
		})
	}
}

// GetPlaylist returns a single playlist with its tracks.
func GetPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		tracks, err := fetchPlaylistTracks(app, playlistID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch tracks"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":          playlist.Id,
			"name":        playlist.GetString("name"),
			"share_token": playlist.GetString("share_token"),
			"tracks":      tracks,
		})
	}
}

// AddTrackToPlaylist adds a track to the given playlist.
func AddTrackToPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		var body struct {
			TrackID string `json:"track_id"`
		}
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil || body.TrackID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "track_id is required"})
		}

		// Find the soundcloud_tracks record by soundcloud_id
		trackRecord, err := app.Dao().FindFirstRecordByFilter(
			"soundcloud_tracks",
			"soundcloud_id = {:track_id}",
			map[string]any{"track_id": body.TrackID},
		)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Track not found"})
		}

		// Check for duplicate
		dup, _ := app.Dao().FindFirstRecordByFilter(
			"playlist_tracks",
			"playlist_id = {:playlist_id} && track_id = {:track_id}",
			map[string]any{"playlist_id": playlistID, "track_id": trackRecord.Id},
		)
		if dup != nil {
			return c.JSON(http.StatusOK, map[string]string{"status": "already_added"})
		}

		// Determine next position
		existingTracks, _ := app.Dao().FindRecordsByFilter(
			"playlist_tracks",
			"playlist_id = {:playlist_id}",
			"-position",
			1,
			0,
			map[string]any{"playlist_id": playlistID},
		)
		position := 0
		if len(existingTracks) > 0 {
			position = existingTracks[0].GetInt("position") + 1
		}

		ptCollection, err := app.Dao().FindCollectionByNameOrId("playlist_tracks")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
		}

		ptRecord := models.NewRecord(ptCollection)
		ptRecord.Set("playlist_id", playlistID)
		ptRecord.Set("track_id", trackRecord.Id)
		ptRecord.Set("position", position)
		ptRecord.Set("added_by", authRecord.Id)

		if err := app.Dao().SaveRecord(ptRecord); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to add track"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "added"})
	}
}

// RemoveTrackFromPlaylist removes a track from the given playlist.
// The :track_id path param may be a soundcloud_id or playlist_track record ID.
func RemoveTrackFromPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")
		trackID := c.PathParam("track_id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		// Try playlist_track record ID first
		ptRecord, err := app.Dao().FindFirstRecordByFilter(
			"playlist_tracks",
			"playlist_id = {:playlist_id} && id = {:pt_id}",
			map[string]any{"playlist_id": playlistID, "pt_id": trackID},
		)
		if err != nil {
			// Fall back: find by soundcloud_id
			scTrack, err2 := app.Dao().FindFirstRecordByFilter(
				"soundcloud_tracks",
				"soundcloud_id = {:track_id}",
				map[string]any{"track_id": trackID},
			)
			if err2 != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Track not found in playlist"})
			}
			ptRecord, err = app.Dao().FindFirstRecordByFilter(
				"playlist_tracks",
				"playlist_id = {:playlist_id} && track_id = {:track_id}",
				map[string]any{"playlist_id": playlistID, "track_id": scTrack.Id},
			)
			if err != nil {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Track not found in playlist"})
			}
		}

		if err := app.Dao().DeleteRecord(ptRecord); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to remove track"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "removed"})
	}
}

// DeletePlaylist deletes a playlist and all its playlist_tracks records.
func DeletePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		// Delete playlist_tracks first (cascade safety)
		ptRecords, _ := app.Dao().FindRecordsByFilter(
			"playlist_tracks",
			"playlist_id = {:playlist_id}",
			"position",
			-1,
			0,
			map[string]any{"playlist_id": playlistID},
		)
		for _, pt := range ptRecords {
			_ = app.Dao().DeleteRecord(pt)
		}

		if err := app.Dao().DeleteRecord(playlist); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete playlist"})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// RenamePlaylist renames an existing playlist.
func RenamePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(c.Request().Body).Decode(&body); err != nil || body.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
		}

		playlist.Set("name", body.Name)
		if err := app.Dao().SaveRecord(playlist); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to rename playlist"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":          playlist.Id,
			"name":        playlist.GetString("name"),
			"share_token": playlist.GetString("share_token"),
			"track_count": trackCountForPlaylist(app, playlist.Id),
			"created":     playlist.Created,
			"updated":     playlist.Updated,
		})
	}
}

// SharePlaylist generates (or returns existing) a share token for a playlist.
func SharePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		playlistID := c.PathParam("id")

		playlist, err := app.Dao().FindRecordById("playlists", playlistID)
		if err != nil || playlist.GetString("user_id") != authRecord.Id {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		token := playlist.GetString("share_token")
		if token == "" {
			token, err = generateShareToken()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
			}
			playlist.Set("share_token", token)
			if err := app.Dao().SaveRecord(playlist); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save share token"})
			}
		}

		return c.JSON(http.StatusOK, map[string]string{
			"share_token": token,
			"share_url":   "/share/" + token,
		})
	}
}

// GetSharedPlaylist returns a public playlist by its share token (no auth required).
func GetSharedPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.PathParam("token")
		if token == "" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}

		playlist, err := app.Dao().FindFirstRecordByFilter(
			"playlists",
			"share_token = {:token}",
			map[string]any{"token": token},
		)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Playlist not found"})
		}

		tracks, err := fetchPlaylistTracks(app, playlist.Id)
		if err != nil {
			tracks = []map[string]interface{}{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":     playlist.Id,
			"name":   playlist.GetString("name"),
			"tracks": tracks,
		})
	}
}

// exportFavoriteTrack is a shared struct for favorites export.
type exportFavoriteTrack struct {
	TrackID          string `json:"track_id"`
	TrackTitle       string `json:"track_title"`
	ArtistName       string `json:"artist_name"`
	Genre            string `json:"genre"`
	TrackDuration    int    `json:"track_duration"`
	ArtworkURL       string `json:"artwork_url"`
	PermalinkURL     string `json:"permalink_url"`
	PlaybackCount    int    `json:"playback_count"`
	FavoritingsCount int    `json:"favoritings_count"`
	FavoritedAt      string `json:"favorited_at"`
}

// fetchFavoriteTracks retrieves the authenticated user's favorited tracks.
func fetchFavoriteTracks(app *pocketbase.PocketBase, authRecord *models.Record) ([]exportFavoriteTrack, error) {
	favRecords, err := app.Dao().FindRecordsByFilter(
		"favorites",
		"user_id = {:userId}",
		"-created",
		0,
		0,
		map[string]any{"userId": authRecord.Id},
	)
	if err != nil {
		return nil, err
	}

	out := make([]exportFavoriteTrack, 0, len(favRecords))
	for _, fav := range favRecords {
		trackRecord, err := app.Dao().FindRecordById("soundcloud_tracks", fav.GetString("track_id"))
		if err != nil {
			continue
		}
		out = append(out, exportFavoriteTrack{
			TrackID:          trackRecord.GetString("soundcloud_id"),
			TrackTitle:       trackRecord.GetString("title"),
			ArtistName:       trackRecord.GetString("artist_name"),
			Genre:            trackRecord.GetString("genre"),
			TrackDuration:    trackRecord.GetInt("length"),
			ArtworkURL:       trackRecord.GetString("artwork_url"),
			PermalinkURL:     trackRecord.GetString("permalink_url"),
			PlaybackCount:    trackRecord.GetInt("playback_count"),
			FavoritingsCount: trackRecord.GetInt("favoritings_count"),
			FavoritedAt:      fav.Created.Time().Format(time.RFC3339),
		})
	}
	return out, nil
}

func ExportFavoritesJSON(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok || authRecord == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		tracks, err := fetchFavoriteTracks(app, authRecord)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch favorites"})
		}

		data, err := json.MarshalIndent(tracks, "", "  ")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encode JSON"})
		}

		c.Response().Header().Set("Content-Disposition", `attachment; filename="sound-cistern-favorites.json"`)
		c.Response().Header().Set("Content-Type", "application/json")
		return c.Blob(http.StatusOK, "application/json", data)
	}
}

func ExportFavoritesCSV(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok || authRecord == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		tracks, err := fetchFavoriteTracks(app, authRecord)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch favorites"})
		}

		c.Response().Header().Set("Content-Disposition", `attachment; filename="sound-cistern-favorites.csv"`)
		c.Response().Header().Set("Content-Type", "text/csv")

		w := csv.NewWriter(c.Response())
		// Write header row
		if err := w.Write([]string{
			"track_id", "track_title", "artist_name", "genre",
			"duration_ms", "permalink_url", "plays", "likes", "favorited_at",
		}); err != nil {
			return err
		}
		// Write data rows
		for _, t := range tracks {
			if err := w.Write([]string{
				t.TrackID,
				t.TrackTitle,
				t.ArtistName,
				t.Genre,
				itoa(t.TrackDuration),
				t.PermalinkURL,
				itoa(t.PlaybackCount),
				itoa(t.FavoritingsCount),
				t.FavoritedAt,
			}); err != nil {
				return err
			}
		}
		w.Flush()
		return w.Error()
	}
}

// exportPlaylistTrack holds track data nested inside a playlist export.
type exportPlaylistTrack struct {
	TrackID          string `json:"track_id"`
	TrackTitle       string `json:"track_title"`
	ArtistName       string `json:"artist_name"`
	Genre            string `json:"genre"`
	TrackDuration    int    `json:"track_duration"`
	ArtworkURL       string `json:"artwork_url"`
	PermalinkURL     string `json:"permalink_url"`
	PlaybackCount    int    `json:"playback_count"`
	FavoritingsCount int    `json:"favoritings_count"`
	Position         int    `json:"position"`
}

// exportPlaylist is a playlist with its tracks for export.
type exportPlaylist struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	TrackCount  int                   `json:"track_count"`
	Created     string                `json:"created"`
	Tracks      []exportPlaylistTrack `json:"tracks"`
}

func ExportPlaylistsJSON(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok || authRecord == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		playlistRecords, err := app.Dao().FindRecordsByFilter(
			"playlists",
			"user_id = {:userId}",
			"-created",
			0,
			0,
			map[string]any{"userId": authRecord.Id},
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch playlists"})
		}

		result := make([]exportPlaylist, 0, len(playlistRecords))
		for _, pl := range playlistRecords {
			// Fetch tracks for this playlist ordered by position
			ptRecords, err := app.Dao().FindRecordsByFilter(
				"playlist_tracks",
				"playlist_id = {:playlistId}",
				"position",
				0,
				0,
				map[string]any{"playlistId": pl.Id},
			)
			if err != nil {
				ptRecords = nil
			}

			tracks := make([]exportPlaylistTrack, 0, len(ptRecords))
			for _, pt := range ptRecords {
				trackRecord, err := app.Dao().FindRecordById("soundcloud_tracks", pt.GetString("track_id"))
				if err != nil {
					continue
				}
				tracks = append(tracks, exportPlaylistTrack{
					TrackID:          trackRecord.GetString("soundcloud_id"),
					TrackTitle:       trackRecord.GetString("title"),
					ArtistName:       trackRecord.GetString("artist_name"),
					Genre:            trackRecord.GetString("genre"),
					TrackDuration:    trackRecord.GetInt("length"),
					ArtworkURL:       trackRecord.GetString("artwork_url"),
					PermalinkURL:     trackRecord.GetString("permalink_url"),
					PlaybackCount:    trackRecord.GetInt("playback_count"),
					FavoritingsCount: trackRecord.GetInt("favoritings_count"),
					Position:         pt.GetInt("position"),
				})
			}

			result = append(result, exportPlaylist{
				ID:          pl.Id,
				Name:        pl.GetString("name"),
				Description: pl.GetString("description"),
				TrackCount:  pl.GetInt("track_count"),
				Created:     pl.Created.Time().Format(time.RFC3339),
				Tracks:      tracks,
			})
		}

		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encode JSON"})
		}

		c.Response().Header().Set("Content-Disposition", `attachment; filename="sound-cistern-playlists.json"`)
		c.Response().Header().Set("Content-Type", "application/json")
		return c.Blob(http.StatusOK, "application/json", data)
	}
}

// formatRSSDuration converts milliseconds to a human-readable "M:SS" string.
func formatRSSDuration(ms int) string {
	s := ms / 1000
	m := s / 60
	sec := s % 60
	return fmt.Sprintf("%d:%02d", m, sec)
}

// buildRSSFeed generates RSS 2.0 XML for the given tracks and channel metadata.
func buildRSSFeed(username, selfURL string, tracks []exportFavoriteTrack) string {
	baseURL := "https://soundcistern.jbhicks.dev"

	var items strings.Builder
	for _, t := range tracks {
		link := t.PermalinkURL
		if link == "" {
			link = baseURL
		}
		artworkTag := ""
		if t.ArtworkURL != "" {
			artworkTag = fmt.Sprintf("\n      <itunes:image href=%q/>", t.ArtworkURL)
		}
		pubDate := t.FavoritedAt
		items.WriteString(fmt.Sprintf(`
    <item>
      <title>%s</title>
      <link>%s</link>
      <description>%s · %s</description>
      <guid isPermaLink="true">%s</guid>
      <pubDate>%s</pubDate>%s
    </item>`,
			xmlEscape(t.TrackTitle+" — "+t.ArtistName),
			xmlEscape(link),
			xmlEscape(t.ArtistName),
			xmlEscape(formatRSSDuration(t.TrackDuration)),
			xmlEscape(link),
			pubDate,
			artworkTag,
		))
	}

	lastBuildDate := time.Now().UTC().Format(time.RFC1123Z)

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
  xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"
  xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Sound Cistern — %s's Favorites</title>
    <link>%s</link>
    <description>Your SoundCloud favorites from Sound Cistern</description>
    <lastBuildDate>%s</lastBuildDate>
    <atom:link href=%q rel="self" type="application/rss+xml"/>%s
  </channel>
</rss>`,
		xmlEscape(username),
		baseURL,
		lastBuildDate,
		selfURL,
		items.String(),
	)
}

// xmlEscape escapes special XML characters in a string.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// GetUserRSSFeed serves an RSS 2.0 feed of the authenticated user's favorites.
// The public feed URL uses the user's PocketBase ID as the token.
func GetUserRSSFeed(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record)
		if !ok || authRecord == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		tracks, err := fetchFavoriteTracks(app, authRecord)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch favorites"})
		}

		// Resolve the user's SoundCloud display name if available
		username := authRecord.GetString("email")
		scUser, err := app.Dao().FindFirstRecordByFilter(
			"soundcloud_users",
			"user_id = {:user_id}",
			map[string]any{"user_id": authRecord.Id},
		)
		if err == nil && scUser != nil {
			if u := scUser.GetString("username"); u != "" {
				username = u
			}
		}

		// Public feed URL uses the user's PocketBase ID as the share token
		selfURL := "https://soundcistern.jbhicks.dev/feed/rss/" + authRecord.Id

		xml := buildRSSFeed(username, selfURL, tracks)
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		// Expose the public RSS URL so the frontend can display it
		c.Response().Header().Set("X-RSS-Public-URL", selfURL)
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(xml))
	}
}

// GetSharedRSSFeed serves a public RSS feed identified by the user's PocketBase ID (share token).
func GetSharedRSSFeed(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.PathParam("share_token")
		if token == "" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		// The token is the user's PocketBase ID
		userRecord, err := app.Dao().FindRecordById("users", token)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		tracks, err := fetchFavoriteTracks(app, userRecord)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch favorites"})
		}

		username := userRecord.GetString("email")
		scUser, err := app.Dao().FindFirstRecordByFilter(
			"soundcloud_users",
			"user_id = {:user_id}",
			map[string]any{"user_id": userRecord.Id},
		)
		if err == nil && scUser != nil {
			if u := scUser.GetString("username"); u != "" {
				username = u
			}
		}

		selfURL := "https://soundcistern.jbhicks.dev/feed/rss/" + token

		xml := buildRSSFeed(username, selfURL, tracks)
		c.Response().Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		return c.Blob(http.StatusOK, "application/rss+xml; charset=utf-8", []byte(xml))
	}
}

// itoa converts an int to its decimal string representation.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
