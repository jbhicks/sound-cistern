package actions

import (
	"fmt"
	"net/http"

	"github.com/jbhicks/sound-cistern/models"
)

// Helper function to create and login a user following Buffalo patterns
func (as *ActionSuite) createAndLoginUser(email, role string) *models.User {
	user := &models.User{
		Email:                email,
		FirstName:            "Test",
		LastName:             "User",
		Role:                 role,
		Password:             "password123",
		PasswordConfirmation: "password123",
	}

	verrs, err := user.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Set user session directly following Buffalo patterns
	as.Session.Set("current_user_id", user.ID)

	return user
}

func (as *ActionSuite) Test_AdminRoutes_RequireAuthentication() {
	// Test admin routes without any authentication - each route individually
	res := as.HTML("/admin/").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")

	res = as.HTML("/admin/dashboard").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")

	res = as.HTML("/admin/users").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/auth/new")
}

func (as *ActionSuite) Test_AdminRoutes_RequireAdminRole() {
	// Create and login regular user
	_ = as.createAndLoginUser("user@example.com", "user")

	// Test admin routes with regular user - each route individually
	res := as.HTML("/admin/").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")

	res = as.HTML("/admin/dashboard").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")

	res = as.HTML("/admin/users").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")
}

func (as *ActionSuite) Test_AdminDashboard_Success() {
	// Create and login admin user
	_ = as.createAndLoginUser("admin@example.com", "admin")

	// Create additional users for statistics
	user1 := &models.User{
		Email:                "user1@example.com",
		FirstName:            "User",
		LastName:             "One",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err := user1.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	user2 := &models.User{
		Email:                "user2@example.com",
		FirstName:            "User",
		LastName:             "Two",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err = user2.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test admin dashboard access
	res := as.HTML("/admin/dashboard").Get()
	as.Equal(http.StatusOK, res.Code)

	// Should display user statistics
	body := res.Body.String()
	as.Contains(body, "3") // Total users (admin + 2 regular users)
	as.Contains(body, "1") // Admin count
	as.Contains(body, "2") // Regular user count
}

func (as *ActionSuite) Test_AdminUsers_Success() {
	// Create and login admin user
	_ = as.createAndLoginUser("admin@example.com", "admin")

	// Create additional test user
	user1 := &models.User{
		Email:                "user1@example.com",
		FirstName:            "Test",
		LastName:             "User",
		Role:                 "user",
		Password:             "password123",
		PasswordConfirmation: "password123",
	}
	verrs, err := user1.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test admin users list
	res := as.HTML("/admin/users").Get()

	// If it's a redirect, let's see where it's going
	if res.Code == http.StatusFound {
		location := res.Header().Get("Location")
		as.Fail("Expected 200 OK but got redirect to: " + location)
		return
	}

	as.Equal(http.StatusOK, res.Code)

	// Should show both users
	body := res.Body.String()
	as.Contains(body, "admin@example.com")
	as.Contains(body, "user1@example.com")
}

func (as *ActionSuite) Test_AdminUsers_Pagination() {
	// Create and login admin user
	_ = as.createAndLoginUser("admin@example.com", "admin")

	// Create multiple users for pagination testing
	for i := 1; i <= 25; i++ {
		user := &models.User{
			Email:                fmt.Sprintf("user%d@example.com", i),
			FirstName:            fmt.Sprintf("User%d", i),
			LastName:             "Test",
			Role:                 "user",
			Password:             "password123",
			PasswordConfirmation: "password123",
		}

		verrs, err := user.Create(as.DB)
		as.NoError(err)
		as.False(verrs.HasAny())
	}

	// Test first page
	res := as.HTML("/admin/users").Get()
	as.Equal(http.StatusOK, res.Code)

	// Test second page
	res = as.HTML("/admin/users?page=2").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_AdminRequired_WithAdminUser() {
	// Create and login admin user
	_ = as.createAndLoginUser("admin@example.com", "admin")

	// Test access to admin dashboard
	res := as.HTML("/admin/dashboard").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_AdminRequired_WithRegularUser() {
	// Create and login regular user
	_ = as.createAndLoginUser("user@example.com", "user")

	// Try to access admin dashboard
	res := as.HTML("/admin/dashboard").Get()

	// Should redirect to dashboard with access denied
	as.Equal(http.StatusFound, res.Code)
	as.Contains(res.Header().Get("Location"), "/dashboard")
}

func (as *ActionSuite) Test_AdminDashboard_HTMX() {
	// Create and login admin user
	_ = as.createAndLoginUser("admin@example.com", "admin")

	// Make HTMX request to admin dashboard
	req := as.HTML("/admin/dashboard")
	req.Headers["HX-Request"] = "true"
	res := req.Get()

	as.Equal(http.StatusOK, res.Code)
}
