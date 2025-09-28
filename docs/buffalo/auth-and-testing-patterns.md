# Buffalo Authentication and Testing Patterns

## Official Buffalo Authentication Pattern

Based on the official Buffalo documentation, here are the key patterns we should follow:

### SetCurrentUser Middleware
```go
func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			err := tx.Find(u, uid)
			if err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
		}
		return next(c)
	}
}
```

### Authorize Middleware
```go
func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Flash().Add("danger", "You must be authorized to see that page")
			return c.Redirect(302, "/")
		}
		return next(c)
	}
}
```

### App Setup with Middleware
```go
app.Use(SetCurrentUser)
app.Use(Authorize)
// Skip Authorize for public routes
app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthNew, AuthCreate)
```

### Testing Patterns

#### Session Management in Tests
```go
func (as *ActionSuite) Test_HomeHandler_LoggedIn() {
	u := &models.User{
		Email:    "mark@example.com",
		Password: "password",
		PasswordConfirmation: "password",
	}
	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())
	
	// Set user ID in session
	as.Session.Set("current_user_id", u.ID)
	
	res := as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Sign Out")
	
	// Clear session for logout test
	as.Session.Clear()
	res = as.HTML("/").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Sign In")
}
```

#### User Creation Tests
```go
func (as *ActionSuite) Test_Users_Create() {
	count, err := as.DB.Count("users")
	as.NoError(err)
	as.Equal(0, count)
	
	u := &models.User{
		Email:                "mark@example.com",
		Password:             "password",
		PasswordConfirmation: "password",
	}
	
	res := as.HTML("/users").Post(u)
	as.Equal(302, res.Code)
	
	count, err = as.DB.Count("users")
	as.NoError(err)
	as.Equal(1, count)
}
```

## Key Differences from Our Current Implementation

1. **Error Handling**: Buffalo's official SetCurrentUser middleware returns the error directly instead of logging warnings and continuing
2. **Test Environment**: No special handling for test environment in middleware - Buffalo's test suite handles this properly
3. **Middleware Order**: SetCurrentUser should be applied before Authorize, and public routes should skip Authorize
4. **Session Testing**: Use `as.Session.Set()` and `as.Session.Clear()` for session manipulation in tests

## Template Patterns

Templates should check for `current_user` availability:
```html
<%= if (current_user) { %>
  <h1><%= current_user.email %></h1>
  <a href="/signout" data-method="delete">sign out</a>
<% } else { %>
  <a href="/signin" class="btn btn-primary">sign in</a>
  <a href="/users/new" class="btn btn-success">register</a>
<% } %>
```