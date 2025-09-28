package actions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jbhicks/sound-cistern/models"
)

func (as *ActionSuite) Test_Auth_Signin() {
	res := as.HTML("/auth/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), `<a href="/auth/new/">Sign In</a>`)
}

func (as *ActionSuite) Test_Auth_New() {
	res := as.HTML("/auth/new").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Sign In")
}

func (as *ActionSuite) Test_Auth_Create() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("mark-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Mark",
		LastName:             "Smith",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	tcases := []struct {
		Email       string
		Password    string
		Status      int
		RedirectURL string

		Identifier string
	}{
		{signupData.Email, signupData.Password, http.StatusFound, "/", "Valid"},
		{"noexist@example.com", "password", http.StatusUnauthorized, "", "Email Invalid"},
		{signupData.Email, "invalidPassword", http.StatusUnauthorized, "", "Password Invalid"},
	}

	for _, tcase := range tcases {
		as.Run(tcase.Identifier, func() {
			res := as.HTML("/auth").Post(&models.User{
				Email:    tcase.Email,
				Password: tcase.Password,
			})

			as.Equal(tcase.Status, res.Code)
			as.Equal(tcase.RedirectURL, res.Location())
		})
	}
}

func (as *ActionSuite) Test_Auth_Redirect() {
	timestamp := time.Now().UnixNano()

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("redirect-%d@example.com", timestamp),
		Password:             "password",
		PasswordConfirmation: "password",
		FirstName:            "Redirect",
		LastName:             "Test",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	tcases := []struct {
		redirectURL    interface{}
		resultLocation string

		identifier string
	}{
		{"/some/url", "/some/url", "RedirectURL defined"},
		{nil, "/", "RedirectURL nil"},
		{"", "/", "RedirectURL empty"},
	}

	for _, tcase := range tcases {
		as.Run(tcase.identifier, func() {
			as.Session.Set("redirectURL", tcase.redirectURL)

			res := as.HTML("/auth").Post(signupData)

			as.Equal(http.StatusFound, res.Code)
			as.Equal(res.Location(), tcase.resultLocation)
		})
	}
}

func (as *ActionSuite) Test_Auth_Create_Password_Preservation() {
	timestamp := time.Now().UnixNano()

	// This test specifically verifies that the plaintext password is preserved
	// during the authentication process and not overwritten by the database query.
	// This would catch the bug where tx.First(u) overwrites the Password field.

	// Create a user through the signup endpoint (which works)
	signupData := &models.User{
		Email:                fmt.Sprintf("test.password.preservation-%d@example.com", timestamp),
		Password:             "secretpassword123",
		PasswordConfirmation: "secretpassword123",
		FirstName:            "Test",
		LastName:             "User",
	}

	// Create user via web interface to ensure it's properly committed
	signupRes := as.HTML("/users").Post(signupData)
	as.Equal(http.StatusFound, signupRes.Code)

	// Now attempt to login with the correct password
	// This should succeed if the password is properly preserved during auth
	loginUser := &models.User{
		Email:    signupData.Email,
		Password: "secretpassword123", // Same password used during creation
	}

	res := as.HTML("/auth").Post(loginUser)

	// Should redirect to home page on successful authentication
	as.Equal(http.StatusFound, res.Code, "Authentication should succeed with correct password")
	as.Equal("/", res.Location(), "Should redirect to home page after successful login")

	// Verify session was set
	sessionUserID := as.Session.Get("current_user_id")
	as.NotNil(sessionUserID, "Session should contain current_user_id after successful login")
}
