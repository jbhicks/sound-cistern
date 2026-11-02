//go:build unit

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
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
