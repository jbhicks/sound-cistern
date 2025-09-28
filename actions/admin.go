package actions

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/pkg/errors"

	"github.com/jbhicks/sound-cistern/models"
	"github.com/jbhicks/sound-cistern/pkg/logging"
)

// AdminRequired middleware ensures only admins can access admin routes
func AdminRequired(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user, ok := c.Value("current_user").(*models.User)
		if !ok || user == nil {
			if c.Value("test_mode") != nil {
				if !ok {
					logging.Debug("AdminRequired: current_user not found in context or wrong type", logging.Fields{})
				} else {
					logging.Debug("AdminRequired: current_user is nil", logging.Fields{})
				}
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		if user.Role != "admin" {
			// Log unauthorized admin access attempt
			logging.SecurityEvent(c, "unauthorized_admin_access", "failure", "insufficient_privileges", logging.Fields{
				"user_id": user.ID.String(),
				"email":   user.Email,
				"role":    user.Role,
			})

			if c.Value("test_mode") != nil {
				logging.Debug("AdminRequired: User is not admin", logging.Fields{
					"role": user.Role,
				})
			}
			c.Flash().Add("danger", "Access denied. Administrator privileges required.")
			return c.Redirect(http.StatusFound, "/dashboard")
		}

		// Log successful admin access
		logging.UserAction(c, user.ID.String(), "admin_access", "User accessed admin area", logging.Fields{
			"email": user.Email,
		})

		if c.Value("test_mode") != nil {
			logging.Debug("AdminRequired: Admin access granted", logging.Fields{
				"user_id": user.ID.String(),
				"email":   user.Email,
			})
		}
		return next(c)
	}
}

// AdminDashboard shows the admin dashboard
func AdminDashboard(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	// Get user statistics
	userCount, err := tx.Count("users")
	if err != nil {
		return errors.WithStack(err)
	}

	adminCount, err := tx.Where("role = ?", "admin").Count("users")
	if err != nil {
		return errors.WithStack(err)
	}

	// Get post statistics
	postCount, err := tx.Count("posts")
	if err != nil {
		return errors.WithStack(err)
	}

	c.Set("userCount", userCount)
	c.Set("adminCount", adminCount)
	c.Set("regularUserCount", userCount-adminCount)
	c.Set("postCount", postCount)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, r.HTML("admin/dashboard.plush.html"))
	}

	// Direct access - render full page with navigation (same template since it includes nav)
	return c.Render(http.StatusOK, r.HTML("admin/dashboard.plush.html"))
}

// AdminUsers lists all users for admin management
func AdminUsers(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	users := []models.User{}
	q := tx.PaginateFromParams(c.Params())

	if err := q.Order("created_at desc").All(&users); err != nil {
		return errors.WithStack(err)
	}

	c.Set("users", users)
	c.Set("pagination", q.Paginator)

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("admin/users.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("admin/users.plush.html"))
}

// AdminUserShow shows a specific user for admin editing
func AdminUserShow(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	c.Set("user", user)

	// Provide role options - Buffalo SelectTag expects slice of maps with "value" and "label" keys
	roleOptions := []map[string]interface{}{
		{"value": "user", "label": "User"},
		{"value": "admin", "label": "Administrator"},
	}
	c.Set("roleOptions", roleOptions)

	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(http.StatusOK, rHTMX.HTML("admin/user_edit.plush.html"))
	}
	return c.Render(http.StatusOK, r.HTML("admin/user_edit.plush.html"))
}

// AdminUserUpdate updates a user as admin
func AdminUserUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Create a copy for updates
	updatedUser := &models.User{}
	*updatedUser = *user

	// Bind form data
	if err := c.Bind(updatedUser); err != nil {
		return errors.WithStack(err)
	}

	// Preserve sensitive fields that shouldn't be changed via this form
	updatedUser.ID = user.ID
	updatedUser.PasswordHash = user.PasswordHash
	updatedUser.CreatedAt = user.CreatedAt
	updatedUser.Password = ""
	updatedUser.PasswordConfirmation = ""

	verrs, err := tx.ValidateAndUpdate(updatedUser)
	if err != nil {
		return errors.WithStack(err)
	}

	if verrs.HasAny() {
		c.Set("user", updatedUser)
		c.Set("errors", verrs)

		// Provide role options for re-render - Buffalo SelectTag expects slice of maps
		roleOptions := []map[string]interface{}{
			{"value": "user", "label": "User"},
			{"value": "admin", "label": "Administrator"},
		}
		c.Set("roleOptions", roleOptions)

		if c.Request().Header.Get("HX-Request") == "true" {
			return c.Render(http.StatusOK, rHTMX.HTML("admin/user_edit.plush.html"))
		}
		return c.Render(http.StatusOK, r.HTML("admin/user_edit.plush.html"))
	}

	// Log admin user update
	adminUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, adminUser.ID.String(), "admin_update_user", fmt.Sprintf("Admin updated user %s", updatedUser.Email), logging.Fields{
		"admin_email":    adminUser.Email,
		"target_user_id": updatedUser.ID.String(),
		"target_email":   updatedUser.Email,
		"updated_role":   updatedUser.Role,
		"previous_role":  user.Role,
	})

	c.Flash().Add("success", "User updated successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/users")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/users")
}

// AdminUserDelete deletes a user (admin only)
func AdminUserDelete(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	// Prevent deletion of the current admin user
	currentUser := c.Value("current_user").(*models.User)
	if user.ID == currentUser.ID {
		c.Flash().Add("danger", "You cannot delete your own account.")
		return c.Redirect(http.StatusFound, "/admin/users")
	}

	if err := tx.Destroy(user); err != nil {
		return errors.WithStack(err)
	}

	// Log admin user deletion
	adminUser := c.Value("current_user").(*models.User)
	logging.UserAction(c, adminUser.ID.String(), "admin_delete_user", fmt.Sprintf("Admin deleted user %s", user.Email), logging.Fields{
		"admin_email":     adminUser.Email,
		"deleted_user_id": user.ID.String(),
		"deleted_email":   user.Email,
		"deleted_role":    user.Role,
	})

	c.Flash().Add("success", "User deleted successfully!")
	if c.Request().Header.Get("HX-Request") == "true" {
		c.Response().Header().Set("HX-Redirect", "/admin/users")
		return c.Render(http.StatusOK, nil)
	}
	return c.Redirect(http.StatusFound, "/admin/users")
}
