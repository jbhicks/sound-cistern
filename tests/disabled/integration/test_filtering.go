package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func Test_Filtering(t *testing.T) {
	as := &TestSuite{Action: suite.NewAction(actions.App())}
	filter := map[string]interface{}{
		"min_length": 3600,
	}
	body, _ := json.Marshal(filter)
	req := as.HTMLRequest("POST", "/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := as.HTML(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect filtered results
}
