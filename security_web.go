package main

// Web-layer security helpers (request filtering, CORS, CSRF/origin checks,
// cookies, CSP). Kept free of PocketBase app state so they are unit-testable.

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

// rejectEncodedPathSeparators blocks request paths that contain an encoded
// slash or backslash (%2F / %5C). The pinned echo v5 pre-release that
// PocketBase v0.22 depends on is affected by GO-2026-6293 (an encoded slash
// can bypass route-level middleware and reach the static file handler) and
// there is no fixed echo version compatible with PocketBase v0.22, so the
// pattern is refused up front. No legitimate route of this app uses them.
func rejectEncodedPathSeparators() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if hasEncodedPathSeparator(c.Request().URL.RawPath) || hasEncodedPathSeparator(c.Request().RequestURI) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid path"})
			}
			return next(c)
		}
	}
}

func hasEncodedPathSeparator(p string) bool {
	if i := strings.IndexAny(p, "?#"); i >= 0 {
		p = p[:i]
	}
	lp := strings.ToLower(p)
	return strings.Contains(lp, "%2f") || strings.Contains(lp, "%5c")
}

// userProtectedFields are users fields a user must never change on their own
// record through the REST API (users.updateRule lets them update themselves).
// role is not used for authorization today, but it is the obvious field to
// grow privileges on, so it is admin-only. (email/verified/password are
// already protected by PocketBase's auth record form for non-admins.)
var userProtectedFields = []string{"role"}

// isAdminRequest reports whether the request is authenticated as a
// PocketBase admin.
func isAdminRequest(c echo.Context) bool {
	if c == nil {
		return false
	}
	admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
	return admin != nil
}

// guardUserProtectedFields is an OnRecordBeforeUpdateRequest("users") hook
// that rejects non-admin changes to userProtectedFields.
func guardUserProtectedFields(e *core.RecordUpdateEvent) error {
	if e == nil || e.Record == nil || isAdminRequest(e.HttpContext) {
		return nil
	}
	orig := e.Record.OriginalCopy()
	for _, f := range userProtectedFields {
		if orig.GetString(f) != e.Record.GetString(f) {
			return apis.NewForbiddenError("You are not allowed to change "+f+".", nil)
		}
	}
	return nil
}

// productionOrigin is the only browser origin allowed in production.
const productionOrigin = "https://soundcistern.com"

// devOrigins are additionally allowed outside production (local dev / tests).
var devOrigins = []string{
	"http://localhost:8090", "http://127.0.0.1:8090",
	"http://localhost:5173", "http://127.0.0.1:5173",
}

// isProductionEnv reports whether this process runs in production: anything
// that is not explicitly test mode or a development GO_ENV.
func isProductionEnv() bool {
	if os.Getenv("TEST_MODE") == "true" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV"))) {
	case "development", "dev", "local", "test":
		return false
	}
	return true
}

// allowedCORSOrigins returns the browser origins allowed to call the API.
// CORS_ALLOWED_ORIGINS (comma separated) overrides the defaults.
func allowedCORSOrigins() []string {
	if v := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); v != "" {
		var out []string
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" && o != "*" {
				out = append(out, strings.TrimSuffix(o, "/"))
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	origins := []string{productionOrigin}
	if !isProductionEnv() {
		origins = append(origins, devOrigins...)
	}
	return origins
}

// withDefaultCORSOrigins appends --origins=<origins> to a "serve" command line
// that does not already set --origins, so PocketBase does not fall back to
// its default "*".
func withDefaultCORSOrigins(args []string, origins []string) []string {
	if len(args) < 2 || args[1] != "serve" || len(origins) == 0 {
		return args
	}
	for _, a := range args[2:] {
		if a == "--origins" || strings.HasPrefix(a, "--origins=") {
			return args
		}
	}
	out := make([]string, 0, len(args)+1)
	out = append(out, args...)
	return append(out, "--origins="+strings.Join(origins, ","))
}

// isAllowedOrigin reports whether origin is one of the allowed origins.
func isAllowedOrigin(origin string, allowed []string) bool {
	origin = strings.TrimSuffix(origin, "/")
	for _, a := range allowed {
		if strings.EqualFold(origin, a) {
			return true
		}
	}
	return false
}

// originHost returns the host[:port] of an Origin/Referer header value.
func originHost(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}
	return strings.ToLower(u.Host)
}
