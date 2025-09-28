package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/jbhicks/sound-cistern/models"
	"github.com/jbhicks/sound-cistern/pkg/logging"
)

// UsersNew renders the users form
func UsersNew(c buffalo.Context) error {
	u := models.User{}
	c.Set("user", u)
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("users/new.plush.html"))
	}
	// For direct page loads, render the full page with persistent header
	return c.Render(http.StatusOK, r.HTML("users/new_full.plush.html"))
}

// UsersCreate registers a new user with the application.
func UsersCreate(c buffalo.Context) error {
	u := &models.User{}
	if err := c.Bind(u); err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := u.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		// Log failed registration attempt
		logging.SecurityEvent(c, "registration_failed", "failure", "validation_errors", logging.Fields{
			"email":             u.Email,
			"validation_errors": verrs.Error(),
		})

		c.Set("user", u)
		c.Set("errors", verrs)
		if c.Request().Header.Get("HX-Request") == "true" {
			// For HTMX, re-render the form with errors, using the htmx layout
			// Ensure the htmx-target is the form itself or a container that includes the form and error messages.
			return c.Render(http.StatusOK, rHTMX.HTML("users/new.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("users/new.plush.html"))
	}

	// Log successful user registration
	logging.UserAction(c, u.Email, "register", "User registration successful", logging.Fields{
		"user_id":   u.ID.String(),
		"user_role": u.Role,
	})

	c.Session().Set("current_user_id", u.ID)
	c.Flash().Add("success", "Welcome to my-go-saas-template!")

	if c.Request().Header.Get("HX-Request") == "true" {
		// After successful creation, HTMX might expect a redirect or a content swap.
		// Setting HX-Redirect header will cause the browser to redirect.
		c.Response().Header().Set("HX-Redirect", "/")
		return c.Render(http.StatusOK, nil) // Or an empty response
	}
	return c.Redirect(http.StatusFound, "/")
}

// ProfileSettings shows the user profile settings page
func ProfileSettings(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	c.Set("user", user)

	// If user is admin, provide role options
	if user.Role == "admin" {
		roleOptions := map[string]string{
			"user":  "User",
			"admin": "Administrator",
		}
		c.Set("roleOptions", roleOptions)
	}

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("users/profile.plush.html"))
	}
	// For direct page loads, render the full page with persistent header
	return c.Render(http.StatusOK, r.HTML("users/profile_full.plush.html"))
}

// ProfileUpdate updates the user's profile information
func ProfileUpdate(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)

	// Create a copy to avoid modifying the session user
	updatedUser := &models.User{}
	*updatedUser = *user

	// Bind only the profile fields we want to update
	if err := c.Bind(updatedUser); err != nil {
		return errors.WithStack(err)
	}

	// Preserve the password hash and other sensitive fields
	updatedUser.ID = user.ID
	updatedUser.Email = user.Email // Don't allow email changes in profile
	updatedUser.PasswordHash = user.PasswordHash
	updatedUser.CreatedAt = user.CreatedAt
	updatedUser.Password = "" // Clear password fields for profile updates
	updatedUser.PasswordConfirmation = ""

	// Only allow role changes for admins
	if user.Role != "admin" {
		updatedUser.Role = user.Role // Preserve original role for non-admins
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndUpdate(updatedUser)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", updatedUser)
		c.Set("errors", verrs)
		if c.Request().Header.Get("HX-Request") == "true" {
			return c.Render(http.StatusOK, rHTMX.HTML("users/profile.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("users/profile.plush.html"))
	}

	c.Flash().Add("success", "Profile updated successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/profile")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/profile")
}

// AccountSettings shows the user account settings page
func AccountSettings(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)
	c.Set("user", user)
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("users/account.plush.html"))
	}
	// For direct page loads, render the full page with persistent header
	return c.Render(http.StatusOK, r.HTML("users/account_full.plush.html"))
}

// AccountUpdate updates the user's account settings
func AccountUpdate(c buffalo.Context) error {
	user := c.Value("current_user").(*models.User)

	// Get the current password for verification
	currentPassword := c.Param("current_password")
	newPassword := c.Param("new_password")
	confirmPassword := c.Param("confirm_password")

	tx := c.Value("tx").(*pop.Connection)

	// If changing password, verify current password first
	if newPassword != "" {
		if currentPassword == "" {
			c.Flash().Add("danger", "Current password is required to change password")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Verify current password
		err := user.VerifyPassword(currentPassword)
		if err != nil {
			c.Flash().Add("danger", "Current password is incorrect")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Check password confirmation
		if newPassword != confirmPassword {
			c.Flash().Add("danger", "New passwords do not match")
			c.Set("user", user)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		// Create a copy for updating
		updatedUser := &models.User{}
		*updatedUser = *user

		// Set the new password fields
		updatedUser.Password = newPassword
		updatedUser.PasswordConfirmation = confirmPassword

		// Hash the new password
		ph, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return errors.WithStack(err)
		}
		updatedUser.PasswordHash = string(ph)

		verrs, err := tx.ValidateAndUpdate(updatedUser)
		if err != nil {
			return errors.WithStack(err)
		}

		if verrs.HasAny() {
			c.Set("user", user)
			c.Set("errors", verrs)
			return c.Render(http.StatusOK, r.HTML("users/account.plush.html"))
		}

		c.Flash().Add("success", "Password updated successfully!")
	} else {
		c.Flash().Add("info", "No changes were made")
	}

	return c.Redirect(http.StatusFound, "/account")
}

// SetCurrentUser attempts to find a user based on the current_user_id
// in the session. If one is found it is set on the context.
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			// Debug logging for tests
			if c.Value("test_mode") != nil {
				logging.Debug("SetCurrentUser: Found session user ID", logging.Fields{
					"session_user_id": uid,
				})
			}

			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				// If user not found, clear the session and continue
				// This handles cases where user was deleted but session still exists
				if c.Value("test_mode") != nil {
					logging.Debug("SetCurrentUser: User not found in DB", logging.Fields{
						"session_user_id": uid,
						"error":           err.Error(),
					})
				}
				c.Session().Delete("current_user_id")
			} else {
				if c.Value("test_mode") != nil {
					logging.Debug("SetCurrentUser: Found user in DB", logging.Fields{
						"user_id": u.ID.String(),
						"email":   u.Email,
						"role":    u.Role,
					})
				}
				c.Set("current_user", u)
			}
		} else {
			if c.Value("test_mode") != nil {
				logging.Debug("SetCurrentUser: No current_user_id in session", logging.Fields{})
			}
		}
		return next(c)
	}
}

// Authorize require a user be logged in before accessing a route
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		// Check if current_user was set by SetCurrentUser middleware
		user, ok := c.Value("current_user").(*models.User)
		if !ok || user == nil {
			if c.Value("test_mode") != nil {
				if !ok {
					logging.Debug("Authorize: current_user not found in context or wrong type", logging.Fields{})
				} else {
					logging.Debug("Authorize: current_user is nil", logging.Fields{})
				}
			}

			c.Session().Set("redirectURL", c.Request().URL.String())

			err := c.Session().Save()
			if err != nil {
				return errors.WithStack(err)
			}

			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(http.StatusFound, "/auth/new")
		}

		if c.Value("test_mode") != nil {
			logging.Debug("Authorize: User authorized", logging.Fields{
				"user_id": user.ID.String(),
				"email":   user.Email,
				"role":    user.Role,
			})
		}
		return next(c)
	}
}
