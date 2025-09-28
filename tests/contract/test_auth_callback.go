package contract

import (
	"net/http/httptest"
	"testing"
)

func Test_AuthCallback(t *testing.T) {
	app := App()
	req := httptest.NewRequest("GET", "/auth/callback", nil)
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect user authenticated
}
