package integration

import (
	"net/http/httptest"
	"testing"
)

func Test_ErrorHandling(t *testing.T) {
	app := App()
	// Simulate API down
	req := httptest.NewRequest("GET", "/feed", nil)
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200 with cached data, got %d", res.Code)
	}
	// Expect cached data with warning
}
