---
name: pocketbase-templ
description: Complete guide for PocketBase development with Templ type-safe templates in Go
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: pocketbase-templ-htmx
---

# PocketBase + Templ Development Skill

## Overview
This skill provides comprehensive guidance for developing web applications using PocketBase as the backend framework and Templ for type-safe HTML templates in Go. This combination provides excellent performance, type safety, and developer experience.

## Core Architecture Patterns

### Project Structure
```
your-project/
├── main.go              # PocketBase application entry point
├── views/               # Templ template files (*.templ)
│   ├── layout.templ     # Base layout wrapper
│   ├── home.templ       # Homepage template
│   ├── components/      # Reusable components
│   └── partials/       # Template fragments
├── pb_migrations/       # Database migrations (Go files)
├── pb_data/            # SQLite database and runtime data
├── public/             # Static assets (CSS, JS, images)
└── go.mod              # Go dependencies
```

### PocketBase Application Setup
```go
package main

import (
    "context"
    "net/http"
    
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
    "github.com/jbhicks/sound-cistern/views"  // Import generated views
)

func main() {
    app := pocketbase.New()

    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        // Static file serving
        e.Router.GET("/public/*", apis.StaticDirectoryHandler(os.DirFS("./public"), false))

        // Home page with Templ
        e.Router.GET("/", func(c echo.Context) error {
            data := views.PageData{
                Title:       "Home",
                Description: "Welcome to our app",
                CurrentPath: "/",
            }

            authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
            if authRecord != nil {
                data.User = authRecord
            }

            return views.Home(data).Render(c.Request().Context(), c.Response().Writer)
        }, apis.ActivityLogger(app))

        return nil
    })

    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

## Templ Template Development

### Template File Structure
Always create `.templ` files in the `views/` directory with proper package declaration:

```templ
package views

import (
    "github.com/pocketbase/pocketbase/models"
)

// PageData represents common data passed to templates
type PageData struct {
    Title       string
    Description string
    CurrentPath string
    User        *models.Record
}

// Layout wraps all pages with common structure
templ Layout(data PageData, content templ.Component) {
    <!DOCTYPE html>
    <html>
        <head>
            <title>{ data.Title }</title>
            <meta name="description" content={ data.Description } />
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
        </head>
        <body>
            @Header(data)
            <main>
                @content
            </main>
            @Footer(data)
        </body>
    </html>
}

// Home page template
templ Home(data PageData) {
    @Layout(data) {
        <div class="container">
            <h1>Welcome to our app</h1>
            @if data.User != nil {
                <p>Hello, { data.User.GetString("email") }!</p>
            } else {
                <p>Please <a href="/signin">sign in</a></p>
            }
        </div>
    }
}
```

### Component Reuse Patterns
Create reusable components for consistent UI:

```templ
// components/button.templ
package components

// Button creates a reusable button component
templ Button(text string, class string, onClick string) {
    <button 
        class={ "btn", class }
        @if onClick != "" {
            onclick={ onClick }
        }
    >
        { text }
    </button>
}

// Usage in other templates:
@components.Button("Click me", "btn-primary", "handleClick()")
```

### Template Data Handling
Pass structured data to templates with proper typing:

```go
// In your route handler
e.Router.GET("/posts/:id", func(c echo.Context) error {
    postId := c.PathParam("id")
    
    post, err := app.Dao().FindRecordById("posts", postId)
    if err != nil {
        return apis.NewNotFoundError("Post not found", err)
    }

    // Get author information
    author, _ := app.Dao().FindRecordById("users", post.GetString("author"))
    
    data := views.PostPageData{
        PageData: views.PageData{
            Title:       post.GetString("title"),
            Description: post.GetString("excerpt"),
            CurrentPath: "/posts/" + postId,
            User:        getCurrentUser(c),
        },
        Post: views.Post{
            ID:         post.Id,
            Title:      post.GetString("title"),
            Content:    post.GetString("content"),
            AuthorName: author.GetString("name"),
            CreatedAt:  post.Created.Time(),
        },
    }

    return views.PostDetail(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Database Integration Patterns

### Collection Management in Migrations
Create collections with proper schema and access rules:

```go
// pb_migrations/1696000000_create_posts.go
package migrations

import (
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/types"
)

func init() {
    core.OnMigrate().Register(func(db core.DB) error {
        collection := &core.Collection{
            Name: "posts",
            Type: core.CollectionTypeBase,
        }

        collection.Fields.Add(
            &core.TextField{
                Name:     "title",
                Required: true,
                Min:      3,
                Max:      255,
            },
            &core.TextField{
                Name:     "slug",
                Required: true,
                Unique:   true,
                Pattern:  "^[a-z0-9-]+$",
            },
            &core.EditorField{
                Name:     "content",
                Required: true,
                MaxSize:  5000000, // 5MB
            },
            &core.TextField{
                Name:     "excerpt",
                Max:      500,
            },
            &core.RelationField{
                Name:         "author",
                Required:     true,
                CollectionId: "users",
                MaxSelect:    types.Pointer(1),
            },
            &core.BoolField{
                Name: "published",
            },
        )

        // Set access rules
        listRule := "published = true || author = @request.auth.id"
        collection.ListRule = &listRule
        
        viewRule := "published = true || author = @request.auth.id"
        collection.ViewRule = &viewRule
        
        createRule := "@request.auth.id != '' && @request.auth.verified = true"
        collection.CreateRule = &createRule
        
        updateRule := "author = @request.auth.id"
        collection.UpdateRule = &updateRule
        
        deleteRule := "author = @request.auth.id"
        collection.DeleteRule = &deleteRule

        return app.Dao().SaveCollection(collection)
    }, nil)
}
```

### Database Queries in Route Handlers
Use PocketBase's query builder for safe database operations:

```go
// Get published posts with pagination
e.Router.GET("/posts", func(c echo.Context) error {
    page := 1
    if p := c.QueryParam("page"); p != "" {
        if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
            page = parsed
        }
    }

    postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
    if err != nil {
        return err
    }

    var records []*core.Record
    err = app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.HashExp{"published": true}).
        OrderBy("created DESC").
        Limit(20).
        Offset((page - 1) * 20).
        All(&records)
    if err != nil {
        return err
    }

    posts := make([]views.Post, 0, len(records))
    for _, record := range records {
        author, _ := app.Dao().FindRecordById("users", record.GetString("author"))
        
        posts = append(posts, views.Post{
            ID:         record.Id,
            Title:      record.GetString("title"),
            Slug:       record.GetString("slug"),
            Excerpt:    record.GetString("excerpt"),
            AuthorName: author.GetString("name"),
            CreatedAt:  record.Created.Time(),
        })
    }

    data := views.PostsPageData{
        PageData: views.PageData{
            Title:       "Posts",
            Description: "Latest posts",
            CurrentPath: "/posts",
            User:        getCurrentUser(c),
        },
        Posts: posts,
        Page:  page,
    }

    return views.PostsList(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Event Hooks and Business Logic

### Record Lifecycle Hooks
Implement validation and business logic using PocketBase hooks:

```go
// Auto-generate slug from title
app.OnRecordBeforeCreateRequest("posts").BindFunc(func(e *core.RecordRequestEvent) error {
    title := e.Record.GetString("title")
    if title != "" && e.Record.GetString("slug") == "" {
        // Generate URL-friendly slug
        slug := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
        slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
        
        e.Record.Set("slug", slug)
    }
    return e.Next()
})

// Validate post data
app.OnRecordValidateRequest("posts").BindFunc(func(e *core.RecordRequestEvent) error {
    title := e.Record.GetString("title")
    if len(title) < 3 {
        return apis.NewBadRequestError("Title must be at least 3 characters", nil)
    }

    // Check for duplicate slug
    slug := e.Record.GetString("slug")
    if slug != "" {
        existing, err := app.Dao().FindFirstRecordByFilter(
            "posts",
            "slug = {:slug} && id != {:id}",
            dbx.Params{
                "slug": slug,
                "id":   e.Record.Id,
            },
        )
        if err == nil && existing != nil {
            return apis.NewBadRequestError("Slug already exists", nil)
        }
    }

    return e.Next()
})

// Set author from authenticated user
app.OnRecordBeforeCreateRequest("posts").BindFunc(func(e *core.RecordRequestEvent) error {
    if e.Record.GetString("author") == "" && e.RequestInfo.Auth != nil {
        e.Record.Set("author", e.RequestInfo.Auth.Id)
    }
    return e.Next()
})
```

## Authentication Integration

### Protected Routes
Implement authentication checks with proper user data:

```go
// Require authentication middleware
e.Router.GET("/dashboard", func(c echo.Context) error {
    authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
    if authRecord == nil {
        return apis.NewUnauthorizedError("Authentication required", nil)
    }

    data := views.PageData{
        Title:       "Dashboard",
        Description: "Your personal dashboard",
        CurrentPath: "/dashboard",
        User:        authRecord,
    }

    return views.Dashboard(data).Render(c.Request().Context(), c.Response().Writer)
}, apis.RequireRecordAuth()) // Built-in auth middleware

// Helper function to get current user
func getCurrentUser(c echo.Context) *models.Record {
    if authRecord, ok := c.Get(apis.ContextAuthRecordKey).(*models.Record); ok {
        return authRecord
    }
    return nil
}
```

### User Authentication Forms
Create login/signup forms with HTMX integration:

```templ
// auth.templ
templ SignInForm() {
    <form hx-post="/api/auth/signin" hx-target="#auth-container" hx-swap="innerHTML">
        <div class="form-group">
            <label for="email">Email</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div class="form-group">
            <label for="password">Password</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Sign In</button>
    </form>
}
```

## Development Workflow

### Template Generation Workflow
Always regenerate templates after making changes:

```bash
# Generate all templates
make templ

# Or run manually
templ generate

# Watch for changes during development
templ generate --watch
```

### Development Server Setup
Use PocketBase's development mode for auto-reload:

```bash
# Development with hot reload
make dev

# This runs: ./sound-cistern serve --dev
```

### Error Handling Patterns
Implement proper error handling in route handlers:

```go
e.Router.GET("/posts/:id", func(c echo.Context) error {
    postId := c.PathParam("id")
    
    post, err := app.Dao().FindRecordById("posts", postId)
    if err != nil {
        // Return 404 for missing records
        return apis.NewNotFoundError("Post not found", err)
    }

    // Check if user can view this post
    if !post.GetBool("published") {
        authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
        if authRecord == nil || post.GetString("author") != authRecord.Id {
            return apis.NewForbiddenError("Access denied", nil)
        }
    }

    // Success - render template
    data := views.PostPageData{...}
    return views.PostDetail(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Best Practices

### Security Considerations
1. **Always validate input** in record hooks before database operations
2. **Use PocketBase's access rules** to control data access
3. **Sanitize user input** before rendering in templates
4. **Use parameterized queries** with the query builder
5. **Implement proper authentication** checks on protected routes

### Performance Optimization
1. **Use template components** for reusable UI elements
2. **Optimize database queries** with proper indexing and filtering
3. **Enable compression** for static assets
4. **Cache frequently accessed data** where appropriate
5. **Use pagination** for large data sets

### Code Organization
1. **Separate concerns**: routes in main.go, templates in views/, logic in hooks
2. **Use consistent naming** for templates and components
3. **Document complex hooks** with clear comments
4. **Follow Go conventions** for error handling and package structure
5. **Keep templates focused** on presentation, not business logic

## Common Patterns and Examples

### List with Pagination
```go
// Calculate pagination
page := 1
limit := 20
offset := (page - 1) * limit

// Query with pagination
var records []*core.Record
err = app.Dao().RecordQuery(collection).
    OrderBy("created DESC").
    Limit(limit).
    Offset(offset).
    All(&records)

// Get total count for pagination UI
var total int64
err = app.Dao().RecordQuery(collection).
    Count(&total)
```

### Form Handling with Validation
```go
e.Router.POST("/posts", func(c echo.Context) error {
    // Parse form data
    var formData struct {
        Title   string `form:"title" validate:"required,min=3"`
        Content string `form:"content" validate:"required"`
        Draft   bool   `form:"draft"`
    }
    
    if err := c.Bind(&formData); err != nil {
        return apis.NewBadRequestError("Invalid form data", err)
    }

    // Create new record
    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    record := core.NewRecord(postsCollection)
    record.Set("title", formData.Title)
    record.Set("content", formData.Content)
    record.Set("published", !formData.Draft)
    record.Set("author", getCurrentUser(c).Id)

    if err := app.Dao().SaveRecord(record); err != nil {
        return apis.NewBadRequestError("Failed to create post", err)
    }

    // Redirect or return response
    return c.Redirect(302, "/posts/"+record.Id)
})
```

## Troubleshooting

### Template Compilation Errors
- Run `make templ` after editing `.templ` files
- Check for syntax errors in template files
- Verify all imports are correct in template files

### Database Issues
- Check that migrations run properly on app startup
- Verify collection names match in queries
- Use PocketBase admin UI at `/_/` to inspect data

### Authentication Problems
- Ensure auth middleware is applied to protected routes
- Check that auth records are properly retrieved from context
- Verify access rules in collection definitions

---

This skill provides the foundation for building robust web applications with PocketBase and Templ. Follow these patterns to create maintainable, type-safe web applications with excellent performance and developer experience.