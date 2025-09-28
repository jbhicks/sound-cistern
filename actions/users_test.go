package actions

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jbhicks/sound-cistern/models"
)

func (as *ActionSuite) Test_Users_New() {
	res := as.HTML("/users/new").Get()
	as.Equal(http.StatusOK, res.Code)
}

func (as *ActionSuite) Test_Users_Create() {
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("mark-%d@example.com", timestamp)
	u := &models.User{
		Email:                email,
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Mark",
		LastName:             "Smith",
	}

	res := as.HTML("/users").Post(u)
	as.Equal(http.StatusFound, res.Code)

	// Verify the redirect location
	location := res.Header().Get("Location")
	as.Equal("/", location, "Should redirect to home page after successful user creation")

	// Test that we can authenticate with the created user immediately
	// This implicitly tests that the user was created and can be found
	authData := &models.User{
		Email:    email,
		Password: "password",
	}

	authRes := as.HTML("/auth").Post(authData)
	as.Equal(http.StatusFound, authRes.Code, "Should be able to authenticate with newly created user")
}

func (as *ActionSuite) Test_ProfileSettings_LoggedIn() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("profile-test-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Profile",
		LastName:             "Test",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Now login with the same user
	loginData := &models.User{
		Email:    signupData.Email,
		Password: "password",
	}

	// Login via auth endpoint
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Access profile settings while logged in
	res := as.HTML("/profile").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Profile Settings")
}

func (as *ActionSuite) Test_ProfileSettings_RequiresAuth() {
	res := as.HTML("/profile").Get()
	as.Equal(http.StatusFound, res.Code) // Should redirect to signin
}

func (as *ActionSuite) Test_ProfileUpdate_LoggedIn() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("profile-update-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Update",
		LastName:             "Test",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Now login with the same user
	loginData := &models.User{
		Email:    signupData.Email,
		Password: "password",
	}

	// Login via auth endpoint
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Update profile data
	updateData := &models.User{
		FirstName: "UpdatedFirst",
		LastName:  "UpdatedLast",
	}

	res := as.HTML("/profile").Post(updateData)
	as.Equal(http.StatusFound, res.Code) // Should redirect after successful update

	// Verify the profile was updated by checking the profile page
	profileRes := as.HTML("/profile").Get()
	as.Equal(http.StatusOK, profileRes.Code)
	as.Contains(profileRes.Body.String(), "UpdatedFirst")
	as.Contains(profileRes.Body.String(), "UpdatedLast")
}

func (as *ActionSuite) Test_AccountSettings_LoggedIn() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("account-test-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Account",
		LastName:             "Test",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Now login with the same user
	loginData := &models.User{
		Email:    signupData.Email,
		Password: "password",
	}

	// Login via auth endpoint
	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Assert we can see the account settings page with user data
	res := as.HTML("/account").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Account Settings")
	as.Contains(res.Body.String(), signupData.Email)
}

func (as *ActionSuite) Test_AccountSettings_RequiresAuth() {
	res := as.HTML("/account").Get()
	as.Equal(http.StatusFound, res.Code) // Should redirect to signin
}

func (as *ActionSuite) Test_AccountSettings_HTMX_Partial() {
	timestamp := time.Now().UnixNano()

	// Create and login user
	signupData := &models.User{
		Email:                fmt.Sprintf("htmx-test-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "HTMX",
		LastName:             "Test",
	}

	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	loginData := &models.User{
		Email:    signupData.Email,
		Password: "password",
	}

	loginRes := as.HTML("/auth").Post(loginData)
	as.Equal(http.StatusFound, loginRes.Code)

	// Test HTMX request (should return partial content)
	req := as.HTML("/account")
	req.Headers["HX-Request"] = "true"
	htmxRes := req.Get()

	as.Equal(http.StatusOK, htmxRes.Code)
	as.Contains(htmxRes.Body.String(), "Account Settings")
	// HTMX response should NOT contain navigation (it's a partial)
	as.NotContains(htmxRes.Body.String(), "Buffalo SaaS")
	as.NotContains(htmxRes.Body.String(), "<nav")

	// Test regular request (should return full page)
	regularRes := as.HTML("/account").Get()

	as.Equal(http.StatusOK, regularRes.Code)
	as.Contains(regularRes.Body.String(), "Account Settings")
	// Regular response SHOULD contain navigation (it's a full page)
	as.Contains(regularRes.Body.String(), "Buffalo SaaS")
	as.Contains(regularRes.Body.String(), "<nav")
}

// Debug test to see what validation errors are happening
func (as *ActionSuite) Test_Debug_User_Creation() {
	timestamp := time.Now().Unix()

	// First test direct database creation
	u := &models.User{
		Email:                fmt.Sprintf("direct-test-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Direct",
		LastName:             "Test",
	}

	// Test direct database creation
	tx := as.DB
	verrs, err := u.Create(tx)

	as.T().Logf("Direct Create() errors: %v, validation errors: %v", err, verrs.String())
	as.NoError(err)
	if verrs.HasAny() {
		as.T().Logf("Validation errors from direct Create(): %v", verrs.String())
	}
	as.False(verrs.HasAny(), "Expected no validation errors, got: %v", verrs.String())

	// Now test web interface with different email
	signupData := &models.User{
		Email:                fmt.Sprintf("debug-test-%d@example.com", timestamp+1),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Debug",
		LastName:             "Test",
	}

	// Create user via web interface
	res := as.HTML("/users").Post(signupData)

	// Print the response code and body to debug
	as.T().Logf("Web Response Code: %d", res.Code)
	if res.Code != http.StatusFound {
		as.T().Logf("Expected 302 but got %d", res.Code)
		bodyStr := res.Body.String()
		if strings.Contains(bodyStr, "text-red-600") {
			as.T().Log("Found error styling in response - there are validation errors")
		} else {
			as.T().Log("No error styling found - validation might be passing but redirect failing")
		}
	}
}
