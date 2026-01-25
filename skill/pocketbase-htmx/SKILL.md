---
name: pocketbase-htmx
description: Complete guide for HTMX integration with PocketBase web applications in Go
license: MIT
compatibility: opencode
metadata:
  version: "1.0"
  audience: go-developers
  stack: pocketbase-templ-htmx
---

# PocketBase + HTMX Integration Skill

## Overview
This skill provides comprehensive patterns for integrating HTMX with PocketBase applications to create dynamic, responsive web interfaces without writing JavaScript. HTMX enables AJAX requests, CSS transitions, and real-time updates directly from HTML attributes.

## Core HTMX + PocketBase Architecture

### Project Setup
```html
<!-- public/index.html or in your base template -->
<!DOCTYPE html>
<html>
<head>
    <title>My PocketBase App</title>
    <script src="https://unpkg.com/htmx.org@1.9.12"></script>
    <meta name="htmx-config" content='{"globalViewTransitions": true}'>
</head>
<body>
    <div id="main-content">
        <!-- Content will be loaded here -->
    </div>
</body>
</html>
```

### Basic HTMX Attributes PocketBase Integration
```html
<!-- Simple GET request to PocketBase API -->
<button hx-get="/api/posts" 
        hx-target="#main-content" 
        hx-swap="innerHTML">
    Load Posts
</button>

<!-- POST request with form data -->
<form hx-post="/api/posts" 
      hx-target="#posts-list" 
      hx-swap="beforeend">
    <input type="text" name="title" required>
    <textarea name="content" required></textarea>
    <button type="submit">Create Post</button>
</form>

<!-- DELETE request with confirmation -->
button hx-delete="/api/posts/{id}" 
        hx-target="#post-{id}"
        hx-swap="outerHTML"
        hx-confirm="Are you sure you want to delete this post?">
    Delete
</button>
```

## PocketBase Route Handlers for HTMX

### HTMX-Aware Route Handlers
Create handlers that detect HTMX requests and respond appropriately:

```go
package main

import (
    "net/http"
    "strings"
    
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
)

// Helper to check if request is from HTMX
func isHTMXRequest(c echo.Context) bool {
    return c.Request().Header.Get("HX-Request") == "true"
}

// HTMX response helper
func htmxResponse(c echo.Context, templateContent string, statusCode int) error {
    if isHTMXRequest(c) {
        c.Response().Header().Set("HX-Trigger", "contentUpdated")
        return c.HTML(statusCode, templateContent)
    }
    
    // For non-HTMX requests, return full page
    return renderFullPage(c, templateContent)
}

// Posts list endpoint
e.Router.GET("/api/posts", func(c echo.Context) error {
    postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
    if err != nil {
        return err
    }

    var records []*core.Record
    err = app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.HashExp{"published": true}).
        OrderBy("created DESC").
        Limit(20).
        All(&records)
    if err != nil {
        return err
    }

    if isHTMXRequest(c) {
        // Return partial template for HTMX
        return views.PostsList(records).Render(c.Request().Context(), c.Response().Writer)
    }

    // Return full page for regular requests
    data := views.PageData{
        Title:       "Posts",
        Description: "Latest posts",
        CurrentPath: "/posts",
    }
    return views.PostsPage(data, records).Render(c.Request().Context(), c.Response().Writer)
})

// Create post endpoint
e.Router.POST("/api/posts", func(c echo.Context) error {
    authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*core.Record)
    if authRecord == nil {
        if isHTMXRequest(c) {
            return c.HTML(http.StatusUnauthorized, "<div>Please sign in to create posts</div>")
        }
        return apis.NewUnauthorizedError("Authentication required", nil)
    }

    var formData struct {
        Title   string `form:"title" validate:"required,min=3"`
        Content string `form:"content" validate:"required"`
    }
    
    if err := c.Bind(&formData); err != nil {
        return handleValidationError(c, err)
    }

    // Create new record
    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    record := core.NewRecord(postsCollection)
    record.Set("title", formData.Title)
    record.Set("content", formData.Content)
    record.Set("author", authRecord.Id)
    record.Set("published", true)

    if err := app.Dao().SaveRecord(record); err != nil {
        return apis.NewBadRequestError("Failed to create post", err)
    }

    if isHTMXRequest(c) {
        // Return the new post as HTML fragment
        return views.PostCard(record).Render(c.Request().Context(), c.Response().Writer)
    }

    return c.Redirect(http.StatusFound, "/posts/"+record.Id)
})
```

## Dynamic Content Loading

### Navigation with HTMX
Create seamless navigation without page reloads:

```html
<!-- Navigation component -->
<nav>
    <a href="/" 
       hx-get="/" 
       hx-target="#main-content" 
       hx-push-url="true"
       hx-swap="innerHTML"
       class="nav-link">
        Home
    </a>
    <a href="/posts" 
       hx-get="/posts" 
       hx-target="#main-content" 
       hx-push-url="true"
       hx-swap="innerHTML"
       class="nav-link">
        Posts
    </a>
    <a href="/dashboard" 
       hx-get="/dashboard" 
       hx-target="#main-content" 
       hx-push-url="true"
       hx-swap="innerHTML"
       class="nav-link">
        Dashboard
    </a>
</nav>
```

### Infinite Scroll Pagination
```html
<!-- posts.templ -->
<div id="posts-container">
    @for i, post := range posts {
        @views.PostCard(post)
    }
</div>

@if hasMore {
    <button id="load-more"
            hx-get="/api/posts?page={ currentPage + 1 }"
            hx-target="#posts-container"
            hx-swap="beforeend"
            hx-trigger="revealed"
            hx-indicator="#loading">
        Load More Posts
    </button>
}

<div id="loading" class="htmx-indicator">
    Loading...
</div>
```

```go
// Route handler for infinite scroll
e.Router.GET("/api/posts", func(c echo.Context) error {
    page := 1
    if p := c.QueryParam("page"); p != "" {
        if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
            page = parsed
        }
    }

    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    
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

    // Check if there are more pages
    var count int64
    app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.HashExp{"published": true}).
        Count(&count)
    hasMore := int64(page*20) < count

    if isHTMXRequest(c) {
        // Return just the new posts for infinite scroll
        var html strings.Builder
        for _, record := range records {
            component := views.PostCard(record)
            component.Render(c.Request().Context(), &html)
        }
        
        if hasMore {
            // Include load more button
            loadMoreBtn := views.LoadMoreButton(page + 1)
            loadMoreBtn.Render(c.Request().Context(), &html)
        }
        
        return c.HTML(http.StatusOK, html.String())
    }

    // Full page for initial load
    data := views.PostsPageData{
        Posts:   records,
        Page:    page,
        HasMore: hasMore,
    }
    return views.PostsPage(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Form Handling and Validation

### Real-time Form Validation
```html
<!-- Registration form with live validation -->
<form id="signup-form" 
      hx-post="/api/auth/validate" 
      hx-target="#form-errors" 
      hx-trigger="keyup changed delay:500ms"
      hx-swap="innerHTML">
    
    <div class="form-group">
        <label for="email">Email</label>
        <input type="email" 
               id="email" 
               name="email" 
               hx-post="/api/auth/validate/email"
               hx-target="#email-error"
               hx-trigger="blur"
               required>
        <div id="email-error" class="error-message"></div>
    </div>
    
    <div class="form-group">
        <label for="password">Password</label>
        <input type="password" 
               id="password" 
               name="password" 
               hx-post="/api/auth/validate/password"
               hx-target="#password-error"
               hx-trigger="blur"
               required>
        <div id="password-error" class="error-message"></div>
    </div>
    
    <button type="submit" 
            hx-post="/api/auth/signup"
            hx-target="#signup-form"
            hx-swap="outerHTML">
        Sign Up
    </button>
</form>

<div id="form-errors"></div>
```

```go
// Validation endpoint for live validation
e.Router.POST("/api/auth/validate", func(c echo.Context) error {
    var formData struct {
        Email    string `form:"email"`
        Password string `form:"password"`
    }
    
    if err := c.Bind(&formData); err != nil {
        return c.String(http.StatusBadRequest, "Invalid form data")
    }

    var errors []string
    
    // Validate email
    if formData.Email != "" {
        if !strings.Contains(formData.Email, "@") {
            errors = append(errors, "Invalid email format")
        } else {
            // Check if email already exists
            existing, _ := app.Dao().FindFirstRecordByFilter(
                "users",
                "email = {:email}",
                dbx.Params{"email": formData.Email},
            )
            if existing != nil {
                errors = append(errors, "Email already exists")
            }
        }
    }
    
    // Validate password
    if formData.Password != "" {
        if len(formData.Password) < 8 {
            errors = append(errors, "Password must be at least 8 characters")
        }
    }

    if len(errors) > 0 {
        errorHtml := "<div class=\"validation-errors\">"
        for _, err := range errors {
            errorHtml += fmt.Sprintf("<div class=\"error\">%s</div>", err)
        }
        errorHtml += "</div>"
        return c.HTML(http.StatusBadRequest, errorHtml)
    }

    return c.String(http.StatusOK, "")
})
```

### Form Submission with Feedback
```html
<!-- Post creation form with progress indicator -->
<form id="post-form"
      hx-post="/api/posts"
      hx-target="#post-form"
      hx-swap="outerHTML"
      hx-indicator="#submit-progress">
    
    <div class="form-group">
        <label for="title">Title</label>
        <input type="text" id="title" name="title" required>
    </div>
    
    <div class="form-group">
        <label for="content">Content</label>
        <textarea id="content" name="content" required></textarea>
    </div>
    
    <button type="submit">
        Create Post
    </button>
    
    <div id="submit-progress" class="htmx-indicator">
        <div class="spinner"></div>
        Creating post...
    </div>
</form>
```

## Real-time Updates

### Server-Sent Events with HTMX
```html
<!-- Live feed component -->
<div id="live-feed"
     hx-get="/api/live/feed"
     hx-trigger="load, every 5s"
     hx-swap="innerHTML">
    Loading live updates...
</div>

<!-- Notifications -->
<div id="notifications"
     hx-get="/api/notifications"
     hx-trigger="load, every 10s"
     hx-swap="innerHTML">
</div>
```

```go
// Live feed endpoint with Server-Sent Events
e.Router.GET("/api/live/feed", func(c echo.Context) error {
    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    
    var records []*core.Record
    err := app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.HashExp{"published": true}).
        OrderBy("created DESC").
        Limit(10).
        All(&records)
    if err != nil {
        return err
    }

    if isHTMXRequest(c) {
        // Return HTML fragment
        var html strings.Builder
        for _, record := range records {
            component := views.PostCard(record)
            component.Render(c.Request().Context(), &html)
        }
        return c.HTML(http.StatusOK, html.String())
    }

    // Server-Sent Events for real-time updates
    c.Response().Header().Set("Content-Type", "text/event-stream")
    c.Response().Header().Set("Cache-Control", "no-cache")
    c.Response().Header().Set("Connection", "keep-alive")

    // Initial data
    fmt.Fprintf(c.Response(), "data: %s\n\n", generateFeedJSON(records))
    
    // Keep connection alive
    for {
        time.Sleep(30 * time.Second)
        fmt.Fprintf(c.Response(), "data: {\"ping\": true}\n\n")
        c.Response().Flush()
    }
})
```

### WebSocket Integration for Real-time Chat
```html
<!-- Chat interface -->
<div id="chat-container">
    <div id="messages"
         hx-ws="connect:/ws/chat"
         hx-ws-send="submit-message">
        <!-- Messages will appear here -->
    </div>
    
    <form id="chat-form"
          hx-ws="send:/ws/chat"
          hx-target="#messages"
          hx-swap="beforeend">
        <input type="text" 
               id="message-input" 
               name="message" 
               placeholder="Type a message..."
               required>
        <button type="submit">Send</button>
    </form>
</div>
```

## Advanced HTMX Patterns

### Modal Windows
```html
<!-- Modal for post editing -->
<div id="edit-modal" class="modal">
    <div class="modal-content">
        <span class="close" 
              hx-get="/api/modal/close"
              hx-target="#edit-modal"
              hx-swap="outerHTML">&times;</span>
        
        <div id="modal-body">
            <!-- Content loaded here -->
        </div>
    </div>
</div>

<!-- Trigger edit modal -->
<button hx-get="/api/posts/{id}/edit"
        hx-target="#modal-body"
        hx-target="#edit-modal">
    Edit Post
</button>
```

```go
// Modal content endpoint
e.Router.GET("/api/posts/:id/edit", func(c echo.Context) error {
    postId := c.PathParam("id")
    
    post, err := app.Dao().FindRecordById("posts", postId)
    if err != nil {
        return apis.NewNotFoundError("Post not found", err)
    }

    // Check permissions
    authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*core.Record)
    if authRecord == nil || post.GetString("author") != authRecord.Id {
        return apis.NewForbiddenError("Access denied", nil)
    }

    return views.EditPostForm(post).Render(c.Request().Context(), c.Response().Writer)
})

// Modal close endpoint
e.Router.GET("/api/modal/close", func(c echo.Context) error {
    return c.String(http.StatusOK, "")
})
```

### Search with Auto-complete
```html
<!-- Search with live results -->
<div class="search-container">
    <input type="search" 
           id="search-input"
           name="q"
           placeholder="Search posts..."
           hx-get="/api/search"
           hx-target="#search-results"
           hx-trigger="keyup changed delay:300ms"
           hx-swap="innerHTML">
    
    <div id="search-results" class="search-dropdown">
        <!-- Search results appear here -->
    </div>
</div>
```

```go
// Search endpoint with debouncing
e.Router.GET("/api/search", func(c echo.Context) error {
    query := c.QueryParam("q")
    if len(query) < 2 {
        return c.String(http.StatusOK, "")
    }

    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    
    var records []*core.Record
    err := app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.NewExp("title LIKE {:query} OR content LIKE {:query}", 
            dbx.Params{"query": "%" + query + "%"})).
        AndWhere(dbx.HashExp{"published": true}).
        OrderBy("created DESC").
        Limit(10).
        All(&records)
    if err != nil {
        return err
    }

    return views.SearchResults(records, query).Render(c.Request().Context(), c.Response().Writer)
})
```

## Error Handling and UX

### Error Display Patterns
```html
<!-- Error display component -->
<div id="error-container" class="error-toast">
    <div class="error-content">
        <h3>Error</h3>
        <p id="error-message"></p>
        <button hx-get="/api/errors/clear"
                hx-target="#error-container"
                hx-swap="outerHTML">
            Close
        </button>
    </div>
</div>
```

```go
// Global error handler for HTMX
app.OnServe().BindFunc(func(e *core.ServeEvent) error {
    e.Router.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            err := next(c)
            if err != nil {
                if isHTMXRequest(c) {
                    errorHtml := fmt.Sprintf(`
                        <div class="error-message">
                            <h3>Error</h3>
                            <p>%s</p>
                            <button hx-get="/api/errors/clear" 
                                    hx-target="#error-container" 
                                    hx-swap="outerHTML">
                                Close
                            </button>
                        </div>`, err.Error())
                    
                    c.Response().Header().Set("HX-Retarget", "#error-container")
                    return c.HTML(http.StatusBadRequest, errorHtml)
                }
                return err
            }
            return nil
        }
    })
    
    return e.Next()
})
```

### Loading States and Progress
```html
<!-- Progress indicators -->
<div class="loading-overlay" id="loading-overlay">
    <div class="spinner"></div>
    <p>Loading...</p>
</div>

<!-- HTMX configuration for global loading -->
<body hx-indicator="#loading-overlay">
    <!-- Content -->
</body>
```

```css
/* Loading animations */
.htmx-indicator {
    display: none;
}

.htmx-request .htmx-indicator {
    display: block;
}

.spinner {
    border: 4px solid #f3f3f3;
    border-top: 4px solid #3498db;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    animation: spin 1s linear infinite;
}

@keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
}
```

## Performance Optimization

### Lazy Loading and Caching
```html
<!-- Lazy load images -->
<img src="/placeholder.jpg"
     data-src="/images/post/{id}.jpg"
     class="lazy-image"
     hx-get="/images/post/{id}"
     hx-trigger="revealed"
     hx-target="self"
     hx-swap="outerHTML">

<!-- Cache configuration -->
<script>
    document.addEventListener('htmx:afterRequest', function(event) {
        // Cache successful GET responses
        if (event.detail.xhr.status === 200 && event.detail.requestMethod === 'GET') {
            const url = event.detail.requestConfig.path;
            const response = event.detail.xhr.responseText;
            localStorage.setItem(url, response);
        }
    });
</script>
```

### Optimized Database Queries
```go
// Efficient query with HTMX pagination
e.Router.GET("/api/posts", func(c echo.Context) error {
    page, _ := strconv.Atoi(c.QueryParam("page"))
    if page < 1 {
        page = 1
    }
    
    postsCollection, _ := app.Dao().FindCollectionByNameOrId("posts")
    
    // Use query builder for efficient pagination
    query := app.Dao().RecordQuery(postsCollection).
        AndWhere(dbx.HashExp{"published": true}).
        OrderBy("created DESC").
        Limit(20)
    
    if page > 1 {
        query = query.Offset((page - 1) * 20)
    }
    
    var records []*core.Record
    if err := query.All(&records); err != nil {
        return err
    }
    
    // Return minimal data for HTMX
    if isHTMXRequest(c) {
        return views.PostsList(records).Render(c.Request().Context(), c.Response().Writer)
    }
    
    // Full page with metadata for initial load
    data := views.PostsPageData{
        Posts: records,
        Page: page,
        // Add pagination metadata
    }
    return views.PostsPage(data).Render(c.Request().Context(), c.Response().Writer)
})
```

## Testing HTMX Integration

### Unit Testing HTMX Responses
```go
func TestHTMXPostCreation(t *testing.T) {
    app := pocketbase.New()
    
    // Create test request with HTMX headers
    req := httptest.NewRequest("POST", "/api/posts", strings.NewReader("title=Test&content=Test content"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("HX-Request", "true")
    
    rec := httptest.NewRecorder()
    c := echo.New().NewContext(req, rec)
    
    // Simulate authenticated user
    user := createTestUser(t, app)
    c.Set(apis.ContextAuthRecordKey, user)
    
    // Call handler
    err := postHandler(app)(c)
    
    assert.NoError(t, err)
    assert.Equal(t, 200, rec.Code)
    assert.Contains(t, rec.Body.String(), "Test") // HTML fragment
}

func TestNonHTMXPostCreation(t *testing.T) {
    app := pocketbase.New()
    
    // Create test request without HTMX headers
    req := httptest.NewRequest("POST", "/api/posts", strings.NewReader("title=Test&content=Test content"))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    
    rec := httptest.NewRecorder()
    c := echo.New().NewContext(req, rec)
    
    // Simulate authenticated user
    user := createTestUser(t, app)
    c.Set(apis.ContextAuthRecordKey, user)
    
    // Call handler
    err := postHandler(app)(c)
    
    assert.NoError(t, err)
    assert.Equal(t, 302, rec.Code) // Redirect for non-HTMX
}
```

## Best Practices

### Security Considerations
1. **Validate HTMX headers** before processing requests
2. **CSRF protection** for state-changing requests
3. **Rate limiting** on AJAX endpoints
4. **Input sanitization** before rendering responses
5. **Access control** checks in all handlers

### Performance Guidelines
1. **Minimize response size** for HTMX requests
2. **Use partial templates** for dynamic content
3. **Implement caching** for frequently requested data
4. **Optimize database queries** with proper indexing
5. **Compress responses** for better performance

### UX Enhancement
1. **Provide loading indicators** for all async operations
2. **Use transitions** for smooth content updates
3. **Implement error handling** with clear user feedback
4. **Maintain browser history** with hx-push-url
5. **Support accessibility** with proper ARIA attributes

---

This skill provides comprehensive patterns for building modern, responsive web applications with PocketBase and HTMX, enabling rich user interactions without complex JavaScript frameworks.