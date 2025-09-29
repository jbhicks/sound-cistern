package actions

import (
	"net/http"
	"strings"
	"sync"

	"github.com/jbhicks/sound-cistern/locales"
	"github.com/jbhicks/sound-cistern/models"
	"github.com/jbhicks/sound-cistern/pkg/logging"
	"github.com/jbhicks/sound-cistern/public"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/middleware/csrf"
	"github.com/gobuffalo/middleware/forcessl"
	"github.com/gobuffalo/middleware/i18n"
	"github.com/gobuffalo/middleware/paramlogger"
	"github.com/unrolled/secure"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
	T       *i18n.Translator
)

// isStaticAsset checks if the path is for a static asset that should not be cached in development
func isStaticAsset(path string) bool {
	return strings.HasSuffix(path, ".css") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".ico") ||
		strings.HasSuffix(path, ".png") ||
		strings.HasSuffix(path, ".jpg") ||
		strings.HasSuffix(path, ".svg") ||
		strings.Contains(path, "/public/") ||
		strings.Contains(path, "/css/") ||
		strings.Contains(path, "/js/")
}

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	appOnce.Do(func() {
		// Initialize logging service
		if err := logging.Init(nil); err != nil {
			// Fallback to standard library logging if our logger fails
			println("Warning: Failed to initialize logging service:", err.Error())
		}

		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_sound_cistern_session",
			Host:        "0.0.0.0:3000",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		if ENV == "production" {
			app.Use(csrf.New)
		}

		// Wraps each request in a transaction.
		//   c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))
		// Setup and use translations:
		app.Use(translations())

		// NOTE: this block should go before any resources
		// that need to be protected by buffalo-auth!
		//AuthMiddlewares
		app.Use(SetCurrentUser)
		app.Use(Authorize)

		// Skip Authorize middleware for public routes following official buffalo-auth pattern
		app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthLanding, AuthNew, AuthCreate, BlogIndex, BlogShow)

		// Public routes
		app.GET("/", HomeHandler)
		app.GET("/health", HealthCheck)

		// Blog routes
		app.GET("/blog", BlogIndex)
		app.GET("/blog/{slug}", BlogShow)

		//Routes for Auth
		app.GET("/auth/", AuthLanding)
		app.GET("/auth/new", AuthNew)
		app.POST("/auth/", AuthCreate)
		app.DELETE("/auth/", AuthDestroy)
		app.GET("/signout", AuthDestroy)  // Add GET route for signout links
		app.POST("/signout", AuthDestroy) // Add POST route for HTMX signout

		//Routes for User registration
		app.GET("/users/new", UsersNew)
		app.POST("/users/", UsersCreate)

		// Protected routes - these will use the global Authorize middleware
		app.GET("/dashboard", DashboardHandler)
		app.GET("/profile", ProfileSettings)
		app.POST("/profile", ProfileUpdate)
		app.GET("/account", AccountSettings)
		app.POST("/account", AccountUpdate)

		// Admin-only routes
		adminGroup := app.Group("/admin")
		adminGroup.Use(AdminRequired)
		adminGroup.GET("/", AdminDashboard)
		adminGroup.GET("/dashboard", AdminDashboard)
		adminGroup.GET("/users", AdminUsers)
		adminGroup.GET("/users/{user_id}", AdminUserShow)
		adminGroup.POST("/users/{user_id}", AdminUserUpdate)
		adminGroup.DELETE("/users/{user_id}", AdminUserDelete)

		// Admin blog post routes
		adminGroup.GET("/posts", AdminPostsIndex)
		adminGroup.POST("/posts/bulk", AdminPostsBulk)
		adminGroup.GET("/posts/new", AdminPostsNew)
		adminGroup.POST("/posts", AdminPostsCreate)
		adminGroup.GET("/posts/{post_id}", AdminPostsShow)
		adminGroup.GET("/posts/{post_id}/edit", AdminPostsEdit)
		adminGroup.POST("/posts/{post_id}", AdminPostsUpdate)
		adminGroup.DELETE("/posts/{post_id}", AdminPostsDelete)

		// Sound Cistern routes
		// Skip Authorize middleware for Soundcloud auth routes
		app.Middleware.Skip(Authorize, SoundcloudAuth, SoundcloudCallback)

		// Auth routes for Soundcloud
		app.GET("/auth/soundcloud", SoundcloudAuth)
		app.GET("/auth/callback", SoundcloudCallback)

		// Protected Sound Cistern routes
		app.GET("/feed", FeedIndex)
		app.POST("/filter", FeedFilter)

		// Add no-cache headers for static files in development
		if ENV == "development" {
			app.Use(func(next buffalo.Handler) buffalo.Handler {
				return func(c buffalo.Context) error {
					// Add no-cache headers for CSS, JS, and other static assets
					if req := c.Request(); req != nil {
						path := req.URL.Path
						if isStaticAsset(path) {
							c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
							c.Response().Header().Set("Pragma", "no-cache")
							c.Response().Header().Set("Expires", "0")
						}
					}
					return next(c)
				}
			})
		}

		// Serve static files
		app.ServeFiles("/", http.FS(public.FS()))
	})

	return app
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(locales.FS(), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
