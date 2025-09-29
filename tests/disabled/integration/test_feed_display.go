package integration

import (
	"net/http/httptest"
	"testing"
)

func Test_FeedDisplay(t *testing.T) {
	as := &TestSuite{Action: suite.NewAction(actions.App())}
	req := as.HTMLRequest("GET", "/feed", nil)
	res := as.HTML(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect feed from database
}
