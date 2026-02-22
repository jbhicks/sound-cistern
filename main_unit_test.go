//go:build unit

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMockActivities(t *testing.T) {
	mock := getMockActivities()

	assert.Contains(t, mock, "collection")

	collection, ok := mock["collection"]
	require.True(t, ok, "collection should exist")

	collectionSlice, ok := collection.([]map[string]interface{})
	require.True(t, ok, "collection should be a slice of maps")
	assert.Greater(t, len(collectionSlice), 0, "should have at least one track")

	firstItem := collectionSlice[0]
	assert.Contains(t, firstItem, "origin")

	origin := firstItem["origin"].(map[string]interface{})
	assert.Contains(t, origin, "id")
	assert.Contains(t, origin, "title")
	assert.Contains(t, origin, "user")

	user := origin["user"].(map[string]interface{})
	assert.Contains(t, user, "username")
}

func TestMockSoundcloudActivitiesResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/me/activities", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := mockSoundcloudActivitiesResponse(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "collection")
	collection := response["collection"].([]interface{})
	assert.Len(t, collection, 5)

	firstTrack := collection[0].(map[string]interface{})
	origin := firstTrack["origin"].(map[string]interface{})
	assert.Equal(t, "DJ PFUNK Mix \"Frequencies\"", origin["title"])
}

func TestSyncTestMode(t *testing.T) {
	t.Skip("Requires PocketBase database setup - run E2E test instead")
}
