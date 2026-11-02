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
