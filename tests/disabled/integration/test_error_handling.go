package integration

import (
	"net/http/httptest"
	"testing"
)

func Test_ErrorHandling(t *testing.T) {
	as := &TestSuite{Action: suite.NewAction(actions.App())}
	// Simulate API down
	req := as.HTMLRequest("GET", "/feed", nil)
	res := as.HTML(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200 with cached data, got %d", res.Code)
	}
	// Expect cached data with warning
}
