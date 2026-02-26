package handlers

import (
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

func ListPlaylists(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, []map[string]string{})
	}
}

func CreatePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, map[string]string{"id": "new"})
	}
}

func GetPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, map[string]string{"id": c.PathParam("id")})
	}
}

func AddTrackToPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "added"})
	}
}

func RemoveTrackFromPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "removed"})
	}
}

func DeletePlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "deleted"})
	}
}

func GetSharedPlaylist(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.Redirect(302, "/playlists")
	}
}

func ExportFavoritesJSON(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, []map[string]string{})
	}
}

func ExportFavoritesCSV(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(200, "")
	}
}

func ExportPlaylistsJSON(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(200, []map[string]string{})
	}
}

func GetUserRSSFeed(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(200, "")
	}
}

func GetSharedRSSFeed(app *pocketbase.PocketBase) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(200, "")
	}
}
