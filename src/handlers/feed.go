package handlers

import (
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
)

// FeedHandler handles feed and filter
type FeedHandler struct {
	FeedService interface{}
}

// NewFeedHandler creates a new handler
func NewFeedHandler(fs interface{}) *FeedHandler {
	return &FeedHandler{FeedService: fs}
}

// GetFeed gets the user's feed
func (fh *FeedHandler) GetFeed(c buffalo.Context) error {
	// TODO: Get cached feed
	tracks := []interface{}{}
	return c.Render(200, render.JSON(tracks))
}

// FilterFeed filters the feed
func (fh *FeedHandler) FilterFeed(c buffalo.Context) error {
	// TODO: Filter feed
	filtered := []interface{}{}
	return c.Render(200, render.JSON(filtered))
}
