package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "github.com/jbhicks/sound-cistern/pb_migrations"
	"github.com/jbhicks/sound-cistern/views"
)

func main() {
	app := pocketbase.New()

	var publicDir string = "./pb_public"

	isGoRun := false
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	jsvm.MustRegister(app, jsvm.Config{})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// Security headers middleware
		e.Router.Use(middleware.SecurityWithConfig(middleware.SecurityConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			HSTSMaxAge:            3600,
			HSTSPreload:           true,
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'; media-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none';",
		}))

		// CSRF protection
		e.Router.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLength: 32,
			TokenLookup: "form:csrf_token",
		}))

		// Rate limiting
		e.Router.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store:    middleware.NewRateLimiterMemoryStore(20),
			Max:      20,
			Duration: 1 * time.Minute,
		}))

		// Static file serving
		e.Router.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   publicDir,
			Browse: false,
			HTML5:  true,
		}))

		// Health check endpoint
		e.Router.GET("/health", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
		})

		// Home page
		e.Router.GET("/", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Home",
				Description: "Your Soundcloud feed aggregator",
				CurrentPath: "/",
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.Home(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Sign in page
		e.Router.GET("/signin", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Sign In",
				Description: "Sign in to your account",
				CurrentPath: "/signin",
			}

			return views.SignIn(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Sign up page
		e.Router.GET("/signup", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Sign Up",
				Description: "Create your account",
				CurrentPath: "/signup",
			}

			return views.SignUp(data).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Blog index page (without enhanced features)
		e.Router.GET("/blog", func(c echo.Context) error {
			data := views.PageData{
				Title:       "Blog",
				Description: "Latest posts and articles",
				CurrentPath: "/blog",
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
			if err != nil {
				return err
			}

			records, err := app.Dao().FindRecordsByFilter(
				postsCollection.Id,
				"published = true",
				"-created",
				100,
				0,
			)
			if err != nil {
				return err
			}

			posts := make([]views.Post, 0, len(records))
			for _, record := range records {
				posts = append(posts, views.Post{
					ID:         record.Id,
					Title:      record.GetString("title"),
					Slug:       record.GetString("slug"),
					Content:    record.GetString("content"),
					Excerpt:    record.GetString("excerpt"),
					Image:      record.GetString("image"),
					ImageAlt:   record.GetString("image_alt"),
					CreatedAt:  record.Created.Time(),
					AuthorID:   record.GetString("author"),
					AuthorName: "",
				})
			}

			return views.BlogIndex(data, posts).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Blog post show page (without enhanced features)
		e.Router.GET("/blog/:slug", func(c echo.Context) error {
			slug := c.PathParam("slug")

			postsCollection, err := app.Dao().FindCollectionByNameOrId("posts")
			if err != nil {
				return err
			}

			record, err := app.Dao().FindFirstRecordByFilter(
				postsCollection.Id,
				"slug = {:slug} && published = true",
				map[string]any{"slug": slug},
			)
			if err != nil {
				return apis.NewNotFoundError("Post not found", err)
			}

			authorName := ""
			authorID := record.GetString("author")
			if authorID != "" {
				authorRecord, err := app.Dao().FindRecordById("users", authorID)
				if err == nil {
					firstName := authorRecord.GetString("first_name")
					lastName := authorRecord.GetString("last_name")
					authorName = firstName + " " + lastName
				}
			}

			post := views.Post{
				ID:         record.Id,
				Title:      record.GetString("title"),
				Slug:       record.GetString("slug"),
				Content:    record.GetString("content"),
				Excerpt:    record.GetString("excerpt"),
				Image:      record.GetString("image"),
				ImageAlt:   record.GetString("image_alt"),
				CreatedAt:  record.Created.Time(),
				AuthorID:   authorID,
				AuthorName: authorName,
			}

			data := views.PageData{
				Title:       post.Title,
				Description: post.Excerpt,
				CurrentPath: "/blog/" + slug,
			}

			authRecord, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			if authRecord != nil {
				data.User = authRecord
			}

			return views.BlogShow(data, post).Render(c.Request().Context(), c.Response().Writer)
		}, apis.ActivityLogger(app))

		// Public assets
		e.Router.GET("/public/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), false))

		// API routes (protected)
		api := e.Group("/api", apis.RequireRecordAuth())
		api.GET("/user", func(c echo.Context) error {
			authRecord := c.Get(apis.ContextAuthRecordKey).(*models.Record)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"id":    authRecord.Id,
				"email": authRecord.GetString("email"),
				"name":  authRecord.GetString("first_name") + " " + authRecord.GetString("last_name"),
			})
		})

		// Error handling
		e.HTTPErrorHandler = func(err error, c echo.Context) {
			code := http.StatusInternalServerError
			if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
			}

			// Log the error
			log.Printf("HTTP Error: %d - %v", code, err)

			// Return JSON for API errors
			if c.Request().Header.Get("Accept") == "application/json" || c.Request().URL.Path[:5] == "/api/" {
				c.JSON(code, map[string]interface{}{
					"error": err.Error(),
				})
				return
			}

			// For HTML responses
			if code == http.StatusNotFound {
				c.HTML(code, `<html><head><title>Not Found</title></head><body><h1>404 Not Found</h1><p>The requested page was not found.</p></body></html>`)
				return
			}

			c.HTML(code, `<html><head><title>Server Error</title></head><body><h1>500 Internal Server Error</h1><p>Something went wrong.</p></body></html>`)
		}

		return nil
	})

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Handle OS signals for graceful shutdown
		ch := make(chan os.Signal, 1)
		// In a real implementation, you would listen for SIGINT/SIGTERM here
		<-ch
		log.Println("Shutting down server...")
		cancel()
	}()

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
