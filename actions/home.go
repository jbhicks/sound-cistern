package actions

import (
	"github.com/jbhicks/sound-cistern/models"
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler serves the public landing page
func HomeHandler(c buffalo.Context) error {
	userID := c.Session().Get("current_user_id")
	if userID != nil {
		c.Set("user_logged_in", true)
		if user := c.Value("current_user"); user != nil {
			c.Set("current_user", user)
		}
	} else {
		c.Set("user_logged_in", false)
		c.Set("current_user", nil)
	}

	htmxRequest := IsHTMX(c.Request())
	c.LogField("is_htmx_request_in_handler_for_home", htmxRequest)
	c.Set("isHTMXRequest", htmxRequest) // Still useful for _index_content.plush.html if it has conditionals

	if htmxRequest {
		// For ALL HTMX requests to home (including the initial hx-trigger='load'),
		// render only the content part using rHTMX.
		return c.Render(http.StatusOK, rHTMX.HTML("home/_index_content.plush.html"))
	}

	// For a full page load (non-HTMX), just render the main index page.
	// The <main> tag in index.plush.html will now have hx-trigger="load"
	// which will immediately make another request (this time an HTMX one) to this same handler
	// to fetch and inject the _index_content.plush.html.

	// No longer pre-rendering content to a string or setting initialPageContent
	// c.Set("initialPageContent", ...)

	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// DashboardHandler serves the protected dashboard for authenticated users
func DashboardHandler(c buffalo.Context) error {
	// Get current_user
	currentUser, ok := c.Value("current_user").(*models.User)
	if !ok || currentUser == nil {
		// This should ideally not happen if AuthMiddleware is working
		return c.Redirect(http.StatusSeeOther, "/")
	}

	// You can pass additional data to the template if needed
	c.Set("user", currentUser) // This is the same as current_user, but explicit for template

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("home/dashboard.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(http.StatusOK, r.HTML("home/dashboard_full.plush.html"))
}
