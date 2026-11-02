//go:build unit

package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRejectEncodedPathSeparators(t *testing.T) {
	e := echo.New()
	e.Pre(rejectEncodedPathSeparators())
	e.GET("/*", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	cases := map[string]int{
		"/api/health":                http.StatusOK,
		"/blog/my-post?next=%2Fhome": http.StatusOK,
		"/public%2F..%2Fpb_data":     http.StatusBadRequest,
		"/app%2fassets":              http.StatusBadRequest,
		"/x%5C..%5Cy":                http.StatusBadRequest,
	}
	for path, want := range cases {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, want, rec.Code, path)
	}
}

func usersRecord(role, firstName string) *models.Record {
	col := &models.Collection{Name: "users", Type: models.CollectionTypeAuth}
	col.Schema = schema.NewSchema(
		&schema.SchemaField{Name: "role", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "first_name", Type: schema.FieldTypeText},
	)
	// Simulate a record loaded from the DB (sets the original copy).
	data := dbx.NullStringMap{}
	for k, v := range map[string]string{"id": "u1", "role": role, "first_name": firstName} {
		data[k] = sql.NullString{String: v, Valid: true}
	}
	return models.NewRecordFromNullStringMap(col, data)
}

func updateEvent(rec *models.Record, admin bool) *core.RecordUpdateEvent {
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodPatch, "/api/collections/users/records/u1", nil), httptest.NewRecorder())
	if admin {
		c.Set(apis.ContextAdminKey, &models.Admin{})
	} else {
		c.Set(apis.ContextAuthRecordKey, rec)
	}
	return &core.RecordUpdateEvent{HttpContext: c, Record: rec}
}

func TestGuardUserProtectedFields_UserCannotChangeOwnRole(t *testing.T) {
	rec := usersRecord("user", "Al")
	rec.Set("role", "admin")
	err := guardUserProtectedFields(updateEvent(rec, false))
	require.Error(t, err)
	var apiErr *apis.ApiError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.Code)
}

func TestGuardUserProtectedFields_UserCanChangeOtherFields(t *testing.T) {
	rec := usersRecord("user", "Al")
	rec.Set("first_name", "Alice")
	rec.Set("role", "user") // unchanged value is fine
	assert.NoError(t, guardUserProtectedFields(updateEvent(rec, false)))
}

func TestGuardUserProtectedFields_AdminCanChangeRole(t *testing.T) {
	rec := usersRecord("user", "Al")
	rec.Set("role", "admin")
	assert.NoError(t, guardUserProtectedFields(updateEvent(rec, true)))
}

func TestWithDefaultCORSOrigins(t *testing.T) {
	o := []string{"https://soundcistern.com"}
	got := withDefaultCORSOrigins([]string{"app", "serve", "--http=0.0.0.0:8090"}, o)
	assert.Equal(t, []string{"app", "serve", "--http=0.0.0.0:8090", "--origins=https://soundcistern.com"}, got)

	explicit := []string{"app", "serve", "--origins=https://x.example"}
	assert.Equal(t, explicit, withDefaultCORSOrigins(explicit, o))
	explicit2 := []string{"app", "serve", "--origins", "https://x.example"}
	assert.Equal(t, explicit2, withDefaultCORSOrigins(explicit2, o))

	migrate := []string{"app", "migrate", "up"}
	assert.Equal(t, migrate, withDefaultCORSOrigins(migrate, o))
}

func TestAllowedCORSOrigins(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	t.Setenv("TEST_MODE", "")
	t.Setenv("GO_ENV", "production")
	assert.Equal(t, []string{"https://soundcistern.com"}, allowedCORSOrigins())

	t.Setenv("GO_ENV", "")
	assert.Equal(t, []string{"https://soundcistern.com"}, allowedCORSOrigins())

	t.Setenv("GO_ENV", "development")
	got := allowedCORSOrigins()
	assert.Contains(t, got, "https://soundcistern.com")
	assert.Contains(t, got, "http://localhost:5173")

	t.Setenv("CORS_ALLOWED_ORIGINS", "https://a.example/, *, https://b.example")
	assert.Equal(t, []string{"https://a.example", "https://b.example"}, allowedCORSOrigins())
}

func TestSameOriginWriteGuard(t *testing.T) {
	e := echo.New()
	e.Pre(sameOriginWriteGuard([]string{"https://soundcistern.com"}))
	e.Any("/*", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })

	type tc struct {
		method, host, origin, referer string
		want                          int
	}
	cases := []tc{
		{http.MethodGet, "soundcistern.com", "https://evil.example", "", http.StatusOK},
		{http.MethodPost, "soundcistern.com", "https://soundcistern.com", "", http.StatusOK},
		{http.MethodPost, "localhost:8090", "http://localhost:8090", "", http.StatusOK},
		{http.MethodPost, "soundcistern.com", "", "https://soundcistern.com/stream", http.StatusOK},
		{http.MethodPost, "soundcistern.com", "", "", http.StatusOK}, // non-browser client
		{http.MethodPost, "soundcistern.com", "https://evil.example", "", http.StatusForbidden},
		{http.MethodDelete, "soundcistern.com", "null", "", http.StatusForbidden},
		{http.MethodPatch, "soundcistern.com", "", "https://evil.example/x", http.StatusForbidden},
		{http.MethodPost, "soundcistern.com", "https://soundcistern.com.evil.example", "", http.StatusForbidden},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, "/api/favorites/toggle", nil)
		req.Host = c.host
		if c.origin != "" {
			req.Header.Set("Origin", c.origin)
		}
		if c.referer != "" {
			req.Header.Set("Referer", c.referer)
		}
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, c.want, rec.Code, "%+v", c)
	}
}

func TestClearSessionCookieAndDomain(t *testing.T) {
	assert.Equal(t, ".soundcistern.com", cookieDomainFor("soundcistern.com"))
	assert.Equal(t, ".soundcistern.com", cookieDomainFor("soundcistern.com:443"))
	assert.Equal(t, "", cookieDomainFor("localhost:8090"))
	assert.Equal(t, "", cookieDomainFor("soundcistern.com.evil.example"))

	e := echo.New()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/signout", nil)
	req.Host = "soundcistern.com"
	clearSessionCookie(e.NewContext(req, rec))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 2)
	for _, ck := range cookies {
		assert.Equal(t, "pb_auth", ck.Name)
		assert.True(t, ck.MaxAge < 0)
		assert.True(t, ck.HttpOnly)
		assert.True(t, ck.Secure)
	}
}
