package integration

import (
	"net/http/httptest"
	"testing"
)

func Test_AuthFlow(t *testing.T) {
	as := &TestSuite{Action: suite.NewAction(actions.App())}
	// Simulate login
	req := as.HTMLRequest("GET", "/auth/soundcloud", nil)
	res := as.HTML(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Check redirect
}
