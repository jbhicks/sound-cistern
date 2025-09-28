package actions

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/jbhicks/sound-cistern/models"
	"github.com/jbhicks/sound-cistern/pkg/logging"
)

// AuthLanding shows a landing page to login
func AuthLanding(c buffalo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("auth/landing.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("auth/landing.plush.html"))
}

// AuthNew loads the signin page
func AuthNew(c buffalo.Context) error {
	c.Set("user", models.User{})
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("auth/new.plush.html"))
	}
	// For direct page loads, render the full page with persistent header
	return c.Render(http.StatusOK, r.HTML("auth/new_full.plush.html"))
}

// AuthCreate attempts to log the user in with an existing account.
func AuthCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	// Preserve the plaintext password before database query
	plaintextPassword := u.Password

	tx := c.Value("tx").(*pop.Connection)

	// find a user with the email
	err := tx.Where("email = ?", strings.ToLower(strings.TrimSpace(u.Email))).First(u)

	// helper function to handle bad attempts
	bad := func() error {
		// Always perform a dummy bcrypt operation to prevent timing attacks
		bcrypt.CompareHashAndPassword([]byte("$2a$10$dummy.hash.to.prevent.timing.attacks"), []byte("dummy"))

		// Log failed login attempt
		logging.SecurityEvent(c, "login_failed", "failure", "invalid_credentials", logging.Fields{
			"email": strings.ToLower(strings.TrimSpace(u.Email)),
		})

		verrs := validate.NewErrors()
		verrs.Add("email", "invalid email/password")

		c.Set("errors", verrs)
		c.Set("user", &models.User{}) // Don't leak the submitted email

		if c.Request().Header.Get("HX-Request") == "true" {
			return c.Render(http.StatusUnauthorized, rHTMX.HTML("auth/new.plush.html"))
		}
		return c.Render(http.StatusUnauthorized, r.HTML("auth/new.plush.html"))
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// couldn't find an user with the supplied email address.
			return bad()
		}
		return errors.WithStack(err)
	}

	// confirm that the given password matches the hashed password from the db
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plaintextPassword))
	if err != nil {
		return bad()
	}

	// Log successful login
	logging.UserAction(c, u.Email, "login", "User logged in successfully", logging.Fields{
		"user_id":   u.ID.String(),
		"user_role": u.Role,
	})

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome Back!")

	redirectURL := "/"
	if redir, ok := c.Session().Get("redirectURL").(string); ok && redir != "" {
		redirectURL = redir
	}
	// Always clear the redirect URL after use, even if empty
	c.Session().Delete("redirectURL")

	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", redirectURL)
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, redirectURL)
}

// AuthDestroy clears the session and logs a user out
func AuthDestroy(c buffalo.Context) error {
	// Get user info before clearing session for logging
	var userID string
	if uid := c.Session().Get("current_user_id"); uid != nil {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	c.Session().Clear()

	// Log logout event
	if userID != "" {
		logging.UserAction(c, userID, "logout", "User logged out", logging.Fields{})
	}

	c.Flash().Add("success", "You have been logged out!")
	if c.Request().Header.Get("HX-Request") == "true" {
		// Instead of relying on HX-Refresh or HX-Redirect,
		// render a small HTML snippet that forces a client-side redirect.
		// We use the rHTMX engine which uses the htmx.plush.html layout (which is just <%= yield %>)
		// so only the script will be sent.
		return c.Render(http.StatusOK, rHTMX.HTML("auth/force_redirect.plush.html"))
	}
	return c.Redirect(http.StatusFound, "/")
}
