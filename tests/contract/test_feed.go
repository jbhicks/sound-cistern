package contract

import (
	"net/http/httptest"
	"testing"
)

func Test_Feed(t *testing.T) {
	app := App()
	req := httptest.NewRequest("GET", "/feed", nil)
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect list of tracks
}
