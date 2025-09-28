package integration

import (
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Performance(t *testing.T) {
	app := App()
	start := time.Now()
	req := httptest.NewRequest("GET", "/feed", nil)
	res := httptest.New(app).Request(req)
	duration := time.Since(start)
	if duration > 2*time.Second {
		t.Errorf("Feed load took %v, expected <2s", duration)
	}
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
}
