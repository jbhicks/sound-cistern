package actions

import (
	"github.com/jbhicks/sound-cistern/models"
)

// Tests for blog functionality
// Test_BlogShow verifies the blog post details page displays correctly
func (as *ActionSuite) Test_BlogShow() {
	// Create a test admin user
	user := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	user.Password = "password"
	user.PasswordConfirmation = "password"
	verrs, err := user.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Create a published test post
	post := &models.Post{
		Title:     "Test Blog Post",
		Slug:      "test-blog-post",
		Content:   "This is a test blog post content with more details.",
		Excerpt:   "This is a test excerpt.",
		Published: true,
		AuthorID:  user.ID,
	}
	verrs, err = as.DB.ValidateAndCreate(post)
	as.NoError(err)
	as.False(verrs.HasAny())

	res := as.HTML("/blog/%s", post.Slug).Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Test Blog Post")
	as.Contains(body, "This is a test blog post content with more details")
	as.Contains(body, "Admin User")
}

func (as *ActionSuite) Test_BlogShow_NotFound() {
	// Test accessing non-existent post
	req := as.HTML("/blog/non-existent-post")
	res := req.Get()
	as.Equal(404, res.Code)
}

func (as *ActionSuite) Test_AdminPostsIndex_RequiresAuth() {
	// Test admin posts index without authentication
	req := as.HTML("/admin/posts")
	res := req.Get()
	as.Equal(302, res.Code) // Should redirect to login
}

func (as *ActionSuite) Test_AdminPostsIndex_RequiresAdminRole() {
	// Create a regular user
	user := &models.User{
		Email:     "user@test.com",
		FirstName: "Regular",
		LastName:  "User",
		Role:      "user",
	}
	user.Password = "password"
	user.PasswordConfirmation = "password"
	verrs, err := user.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Set user session directly following Buffalo patterns
	as.Session.Set("current_user_id", user.ID)

	// Test admin posts index as regular user
	req := as.HTML("/admin/posts")
	res := req.Get()
	as.Equal(302, res.Code) // Should redirect to dashboard
}

func (as *ActionSuite) Test_AdminPostsIndex_WithAdmin() {
	// Create an admin user
	admin := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	admin.Password = "password"
	admin.PasswordConfirmation = "password"
	verrs, err := admin.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Set admin session directly following Buffalo patterns
	as.Session.Set("current_user_id", admin.ID)

	// Test admin posts index
	req := as.HTML("/admin/posts")
	res := req.Get()
	as.Equal(200, res.Code)

	body := res.Body.String()
	as.Contains(body, "Blog Posts")
}

func (as *ActionSuite) Test_AdminPostsCreate() {
	// Create an admin user
	admin := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	admin.Password = "password"
	admin.PasswordConfirmation = "password"
	verrs, err := admin.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Set admin session directly following Buffalo patterns
	as.Session.Set("current_user_id", admin.ID)

	// Create a new post
	postData := &models.Post{
		Title:     "New Test Post",
		Content:   "This is the content of the new test post.",
		Published: true,
	}

	res := as.HTML("/admin/posts").Post(postData)
	as.Equal(302, res.Code) // Should redirect after creation

	// Verify post was created
	post := &models.Post{}
	err = as.DB.Where("title = ?", "New Test Post").First(post)
	as.NoError(err)
	as.Equal("New Test Post", post.Title)
	as.Equal("new-test-post", post.Slug) // Should auto-generate slug
	as.True(post.Published)
	as.Equal(admin.ID, post.AuthorID)
}

func (as *ActionSuite) Test_BlogIndex() {
	// Create a test admin user
	user := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	user.Password = "password"
	user.PasswordConfirmation = "password"
	verrs, err := user.Create(as.DB)
	as.NoError(err)
	if verrs.HasAny() {
		as.T().Logf("User validation errors: %v", verrs.Errors)
	}
	as.False(verrs.HasAny())

	// Create a published test post
	post := &models.Post{
		Title:     "Test Blog Post",
		Slug:      "test-blog-post",
		Content:   "This is a test blog post content.",
		Excerpt:   "This is a test excerpt",
		Published: true,
		AuthorID:  user.ID,
	}
	verrs, err = as.DB.ValidateAndCreate(post)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Create an unpublished test post (should not appear)
	unpublishedPost := &models.Post{
		Title:     "Unpublished Post",
		Slug:      "unpublished-post",
		Content:   "This post should not appear.",
		Published: false,
		AuthorID:  user.ID,
	}
	verrs, err = as.DB.ValidateAndCreate(unpublishedPost)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test blog index page
	req := as.HTML("/blog/")
	res := req.Get()
	as.Equal(200, res.Code)

	// Debug: Check what's in the database
	testPosts := &models.Posts{}
	err = as.DB.All(testPosts)
	as.NoError(err)
	as.T().Logf("Total posts in database: %d", len(*testPosts))

	publishedPosts := &models.Posts{}
	err = as.DB.Where("published = ?", true).All(publishedPosts)
	as.NoError(err)
	as.T().Logf("Published posts in database: %d", len(*publishedPosts))

	// Check that published post appears
	body := res.Body.String()
	as.Contains(body, "Test Blog Post")
	as.Contains(body, "This is a test excerpt")

	// Check that unpublished post does not appear
	as.NotContains(body, "Unpublished Post")
}

// Test that admin post pages have proper navigation headers
func (as *ActionSuite) Test_AdminPostPagesHaveNavigation() {
	// Create a test admin user
	user := &models.User{
		Email:     "admin@test.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
	}
	user.Password = "password"
	user.PasswordConfirmation = "password"
	verrs, err := user.Create(as.DB)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Create session for admin user
	sess := as.Session
	sess.Set("current_user_id", user.ID)

	// Create a test post for show/edit pages
	post := &models.Post{
		Title:     "Test Post",
		Slug:      "test-post",
		Content:   "Test content",
		Excerpt:   "Test excerpt",
		Published: true,
		AuthorID:  user.ID,
	}
	verrs, err = as.DB.ValidateAndCreate(post)
	as.NoError(err)
	as.False(verrs.HasAny())

	// Test admin posts index page
	res := as.HTML("/admin/posts").Get()
	as.Equal(200, res.Code)
	body := res.Body.String()
	as.Contains(body, `<nav class="container-fluid dashboard-nav">`, "Admin posts index should have navigation header")
	as.Contains(body, `<strong>My Go SaaS</strong>`, "Navigation should include site logo")
	as.Contains(body, `Admin Dashboard`, "Navigation should have admin dashboard link")

	// Test admin posts new page
	res = as.HTML("/admin/posts/new").Get()
	as.Equal(200, res.Code)
	body = res.Body.String()
	as.Contains(body, `<nav class="container-fluid dashboard-nav">`, "Admin posts new should have navigation header")
	as.Contains(body, `<strong>My Go SaaS</strong>`, "Navigation should include site logo")
	as.Contains(body, `Admin Dashboard`, "Navigation should have admin dashboard link")

	// Test admin posts show page
	res = as.HTML("/admin/posts/%d", post.ID).Get()
	as.Equal(200, res.Code)
	body = res.Body.String()
	as.Contains(body, `<nav class="container-fluid dashboard-nav">`, "Admin posts show should have navigation header")
	as.Contains(body, `<strong>My Go SaaS</strong>`, "Navigation should include site logo")
	as.Contains(body, `Admin Dashboard`, "Navigation should have admin dashboard link")

	// Test admin posts edit page
	res = as.HTML("/admin/posts/%d/edit", post.ID).Get()
	as.Equal(200, res.Code)
	body = res.Body.String()
	as.Contains(body, `<nav class="container-fluid dashboard-nav">`, "Admin posts edit should have navigation header")
	as.Contains(body, `<strong>My Go SaaS</strong>`, "Navigation should include site logo")
	as.Contains(body, `Admin Dashboard`, "Navigation should have admin dashboard link")
}
