package handlers

import (
	"net/http"
	"strconv"

	"github.com/jbhicks/sound-cistern/middleware"
	"github.com/jbhicks/sound-cistern/services"
	"github.com/labstack/echo/v5"
)

type Handlers struct {
	mockData *services.MockDataService
}

func NewHandlers() *Handlers {
	return &Handlers{
		mockData: services.NewMockDataService(),
	}
}

func (h *Handlers) StreamAPI(isTestMode bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		durationMin := 0
		if dm := c.QueryParam("duration_min"); dm != "" {
			if parsed, err := strconv.Atoi(dm); err == nil {
				durationMin = parsed
			}
		}
		searchQuery := c.QueryParam("q")

		if isTestMode {
			html := h.mockData.GetStreamTracks(durationMin, searchQuery)
			return c.HTML(http.StatusOK, html)
		}

		// Real implementation would go here
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Not implemented in production"})
	}
}

func (h *Handlers) FavoritesAPI(isTestMode bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		durationMin := 0
		if dm := c.QueryParam("duration_min"); dm != "" {
			if parsed, err := strconv.Atoi(dm); err == nil {
				durationMin = parsed
			}
		}
		searchQuery := c.QueryParam("q")

		if isTestMode {
			html := h.mockData.GetFavoritesTracks(durationMin, searchQuery)
			return c.HTML(http.StatusOK, html)
		}

		// Real implementation would go here
		return c.JSON(http.StatusNotImplemented, map[string]string{"error": "Not implemented in production"})
	}
}

func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func NotFound(c echo.Context) error {
	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "Not found",
	})
}

var RateLimiter = middleware.NewRateLimiter(100, 60)
var ValidateParams = middleware.ValidateQueryParams("q", "duration_min", "duration_max", "sort", "limit", "content_type")
