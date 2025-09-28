package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func Test_Filtering(t *testing.T) {
	app := App()
	filter := map[string]interface{}{
		"min_length": 3600,
	}
	body, _ := json.Marshal(filter)
	req := httptest.NewRequest("POST", "/filter", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.New(app).Request(req)
	if res.Code != 200 {
		t.Errorf("Expected 200, got %d", res.Code)
	}
	// Expect filtered results
}
