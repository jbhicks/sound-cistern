package contract

import (
	"net/http/httptest"
	"testing"
)

func Test_AuthSoundcloud(t *testing.T) {
	app := App()
	req := httptest.NewRequest("GET", "/auth/soundcloud", nil)
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect redirect to Soundcloud
}
