package integration

import (
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Performance(t *testing.T) {
	as := &TestSuite{Action: suite.NewAction(actions.App())}
	start := time.Now()
	req := as.HTMLRequest("GET", "/feed", nil)
	res := as.HTML(app).Request(req)
	duration := time.Since(start)
	if duration > 2*time.Second {
		t.Errorf("Feed load took %v, expected <2s", duration)
	}
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
}
