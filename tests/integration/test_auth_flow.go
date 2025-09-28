package integration

import (
	"net/http/httptest"
	"testing"
)

func Test_AuthFlow(t *testing.T) {
	app := App()
	// Simulate login
	req := httptest.NewRequest("GET", "/auth/soundcloud", nil)
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Check redirect
}
