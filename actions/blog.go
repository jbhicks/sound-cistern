package actions

import (
	"fmt"
	"io"
	"github.com/jbhicks/sound-cistern/models"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
)

// BlogIndex displays all published posts
func BlogIndex(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := []models.Post{}

	// Get published posts ordered by created_at desc
	if err := tx.Where("published = ?", true).Order("created_at desc").All(&posts); err != nil {
		return err
	}

	// Load authors for each post
	for i := range posts {
		if err := tx.Load(&posts[i], "Author"); err != nil {
			return err
		}
	}

	c.Set("posts", posts)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(200, r.HTML("blog/index.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(200, r.HTML("blog/index_full.plush.html"))
}

// BlogShow displays a single post by slug
func BlogShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	slug := c.Param("slug")
	post := &models.Post{}

	// Find published post by slug
	if err := tx.Where("slug = ? AND published = ?", slug, true).First(post); err != nil {
		return c.Error(404, err)
	}

	// Load author
	if err := tx.Load(post, "Author"); err != nil {
		return err
	}

	c.Set("post", post)

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == "true" {
		return c.Render(200, r.HTML("blog/show.plush.html"))
	}

	// Direct access - render full page with navigation
	return c.Render(200, r.HTML("blog/show_full.plush.html"))
}

// AdminPostsIndex lists all posts for admin with search and filtering
func AdminPostsIndex(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	posts := &models.Posts{}

	// Build query with search and filter parameters
	query := tx.Q()

	// Get search and status parameters with defaults
	search := c.Param("search")
	status := c.Param("status")

	// Handle search parameter
	if search != "" {
		query = query.Where("title ILIKE ? OR content ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Handle status filter
	if status != "" {
		if status == "published" {
			query = query.Where("published = ?", true)
		} else if status == "draft" {
			query = query.Where("published = ?", false)
		}
	}

	// Get posts ordered by created_at desc
	if err := query.Order("created_at desc").All(posts); err != nil {
		return err
	}

	// Load authors for each post
	for i := range *posts {
		if err := tx.Load(&(*posts)[i], "Author"); err != nil {
			return err
		}
	}

	// Set template variables
	c.Set("posts", posts)
	c.Set("search", search)
	c.Set("status", status)
	return c.Render(200, r.HTML("admin/posts/index.plush.html"))
}

// AdminPostsShow displays a single post for admin
func AdminPostsShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	id, err := strconv.Atoi(c.Param("post_id"))
	if err != nil {
		return c.Error(404, err)
	}

	post := &models.Post{}
	if err := tx.Find(post, id); err != nil {
		return c.Error(404, err)
	}

	// Load author
	if err := tx.Load(post, "Author"); err != nil {
		return err
	}

	c.Set("post", post)
	return c.Render(200, r.HTML("admin/posts/show.plush.html"))
}

// AdminPostsNew displays the form for creating a new post
func AdminPostsNew(c buffalo.Context) error {
	post := &models.Post{}
	c.Set("post", post)
	return c.Render(200, r.HTML("admin/posts/new.plush.html"))
}

// AdminPostsCreate creates a new post
func AdminPostsCreate(c buffalo.Context) error {
	post := &models.Post{}

	if err := c.Bind(post); err != nil {
		return err
	}

	// Set author to current user
	currentUser := c.Value("current_user").(*models.User)
	post.AuthorID = currentUser.ID

	// Handle file upload if present
	if c.Request().MultipartForm != nil && c.Request().MultipartForm.File["image"] != nil {
		fileHeaders := c.Request().MultipartForm.File["image"]
		if len(fileHeaders) > 0 {
			fileHeader := fileHeaders[0]
			file, err := fileHeader.Open()
			if err == nil {
				defer file.Close()

				// Create uploads directory if it doesn't exist
				uploadDir := filepath.Join("public", "uploads")
				if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
					return err
				}

				// Generate unique filename
				filename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
				filePath := filepath.Join(uploadDir, filename)

				// Save the file
				dst, err := os.Create(filePath)
				if err != nil {
					return err
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					return err
				}

				// Set the image path relative to public directory
				post.Image = "/uploads/" + filename
			}
		}
	}

	// Generate slug and excerpt before validation
	post.GenerateSlug()
	post.GenerateExcerpt()

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	verrs, err := tx.ValidateAndCreate(post)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("admin/posts/new.plush.html"))
	}

	c.Flash().Add("success", "Post was created successfully")
	return c.Redirect(302, "/admin/posts/%d", post.ID)
}

// AdminPostsEdit displays the form for editing a post
func AdminPostsEdit(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	id, err := strconv.Atoi(c.Param("post_id"))
	if err != nil {
		return c.Error(404, err)
	}

	post := &models.Post{}
	if err := tx.Find(post, id); err != nil {
		return c.Error(404, err)
	}

	c.Set("post", post)
	return c.Render(200, r.HTML("admin/posts/edit.plush.html"))
}

// AdminPostsUpdate updates a post
func AdminPostsUpdate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	id, err := strconv.Atoi(c.Param("post_id"))
	if err != nil {
		return c.Error(404, err)
	}

	post := &models.Post{}
	if err := tx.Find(post, id); err != nil {
		return c.Error(404, err)
	}

	if err := c.Bind(post); err != nil {
		return err
	}

	// Handle file upload if present
	if c.Request().MultipartForm != nil && c.Request().MultipartForm.File["image"] != nil {
		fileHeaders := c.Request().MultipartForm.File["image"]
		if len(fileHeaders) > 0 {
			fileHeader := fileHeaders[0]
			file, err := fileHeader.Open()
			if err == nil {
				defer file.Close()

				// Create uploads directory if it doesn't exist
				uploadDir := filepath.Join("public", "uploads")
				if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
					return err
				}

				// Generate unique filename
				filename := fmt.Sprintf("%d_%s", time.Now().Unix(), fileHeader.Filename)
				filePath := filepath.Join(uploadDir, filename)

				// Save the file
				dst, err := os.Create(filePath)
				if err != nil {
					return err
				}
				defer dst.Close()

				if _, err := io.Copy(dst, file); err != nil {
					return err
				}

				// Set the image path relative to public directory
				post.Image = "/uploads/" + filename
			}
		}
	}

	// Generate excerpt if not provided
	post.GenerateExcerpt()

	verrs, err := tx.ValidateAndUpdate(post)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		c.Set("post", post)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("admin/posts/edit.plush.html"))
	}

	c.Flash().Add("success", "Post was updated successfully")
	return c.Redirect(302, "/admin/posts/%d", post.ID)
}

// AdminPostsDelete deletes a post
func AdminPostsDelete(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	id, err := strconv.Atoi(c.Param("post_id"))
	if err != nil {
		return c.Error(404, err)
	}

	post := &models.Post{}
	if err := tx.Find(post, id); err != nil {
		return c.Error(404, err)
	}

	if err := tx.Destroy(post); err != nil {
		return err
	}

	c.Flash().Add("success", "Post was deleted successfully")
	return c.Redirect(302, "/admin/posts")
}

// AdminPostsBulk handles bulk operations on posts
func AdminPostsBulk(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return fmt.Errorf("no transaction found")
	}

	bulkAction := c.Param("bulk_action")

	// Get post IDs from form data
	req := c.Request()
	if err := req.ParseForm(); err != nil {
		return err
	}
	postIDs := req.Form["post_ids"]

	if len(postIDs) == 0 {
		c.Flash().Add("error", "No posts selected")
		return c.Redirect(302, "/admin/posts")
	}

	switch bulkAction {
	case "publish":
		posts := &models.Posts{}
		if err := tx.Where("id IN (?)", postIDs).All(posts); err != nil {
			return err
		}

		for i := range *posts {
			(*posts)[i].Published = true
			if err := tx.Save(&(*posts)[i]); err != nil {
				return err
			}
		}
		c.Flash().Add("success", fmt.Sprintf("Published %d post(s)", len(postIDs)))

	case "unpublish":
		posts := &models.Posts{}
		if err := tx.Where("id IN (?)", postIDs).All(posts); err != nil {
			return err
		}

		for i := range *posts {
			(*posts)[i].Published = false
			if err := tx.Save(&(*posts)[i]); err != nil {
				return err
			}
		}
		c.Flash().Add("success", fmt.Sprintf("Unpublished %d post(s)", len(postIDs)))

	case "delete":
		posts := &models.Posts{}
		if err := tx.Where("id IN (?)", postIDs).All(posts); err != nil {
			return err
		}

		for _, post := range *posts {
			if err := tx.Destroy(&post); err != nil {
				return err
			}
		}
		c.Flash().Add("success", fmt.Sprintf("Deleted %d post(s)", len(postIDs)))

	default:
		c.Flash().Add("error", "Invalid bulk action")
	}

	return c.Redirect(302, "/admin/posts")
}
