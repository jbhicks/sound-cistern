package main

// Web-layer security helpers (request filtering, CORS, CSRF/origin checks,
// cookies, CSP). Kept free of PocketBase app state so they are unit-testable.

import (
	"net/http"
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
