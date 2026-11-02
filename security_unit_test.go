//go:build unit

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jbhicks/sound-cistern/services"
)

func scUsersCollection() *models.Collection {
	col := &models.Collection{Name: "soundcloud_users", Type: models.CollectionTypeBase}
	col.Schema = schema.NewSchema(
		&schema.SchemaField{Name: "user_id", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "access_token", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "refresh_token", Type: schema.FieldTypeText},
		&schema.SchemaField{Name: "expires_at", Type: schema.FieldTypeText},
	)
	return col
}

func newSCUser(id, userID, access, refresh string) *models.Record {
	r := models.NewRecord(scUsersCollection())
	r.Id = id
	r.Set("user_id", userID)
	r.Set("access_token", access)
	r.Set("refresh_token", refresh)
	return r
}

func newAuthUser(id string) *models.Record {
	col := &models.Collection{Name: "users", Type: models.CollectionTypeAuth}
	r := models.NewRecord(col)
	r.Id = id
	return r
}

// fakeStore mimics the soundcloud_users collection keyed by owner user id.
type fakeStore struct {
	byOwner     map[string]*models.Record
	lookups     []string
	saved       []*models.Record
	resolveSeen []string
}

func (f *fakeStore) lookup(userID string) (*models.Record, error) {
	f.lookups = append(f.lookups, userID)
	if r, ok := f.byOwner[userID]; ok {
		return r, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeStore) deps(upstreamURL string) trackStreamDeps {
	return trackStreamDeps{
		lookupSoundcloudUser: f.lookup,
		saveRecord:           func(r *models.Record) error { f.saved = append(f.saved, r); return nil },
		resolveStreamURL: func(tok, trackID, q, ua string) string {
			f.resolveSeen = append(f.resolveSeen, tok)
			return upstreamURL
		},
		allowURL: func(u string) bool { return strings.HasPrefix(u, upstreamURL) && upstreamURL != "" },
	}
}

// newStreamTestServer mounts the handler exactly like main.go does.
func newStreamTestServer(d trackStreamDeps, as *models.Record) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = func(c echo.Context, err error) {
		var apiErr *apis.ApiError
		if errors.As(err, &apiErr) {
			_ = c.JSON(apiErr.Code, apiErr)
			return
		}
		_ = c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if as != nil {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set(apis.ContextAuthRecordKey, as)
				return next(c)
			}
		})
	}
	e.GET("/api/track/:id/stream", newTrackStreamHandler(d), apis.RequireRecordAuth())
	return e
}

func doGet(e *echo.Echo, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestStreamRoute_UnauthenticatedReturns401(t *testing.T) {
	store := &fakeStore{byOwner: map[string]*models.Record{
		"owner": newSCUser("sc1", "owner", "OWNER_TOKEN", "OWNER_REFRESH"),
	}}
	e := newStreamTestServer(store.deps("https://upstream.example/"), nil)

	rec := doGet(e, "/api/track/123/stream")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, store.lookups, "no token lookup for anonymous requests")
	assert.Empty(t, store.resolveSeen)
}

func TestStreamRoute_DebugParamUnauthenticatedStill401(t *testing.T) {
	store := &fakeStore{byOwner: map[string]*models.Record{
		"owner": newSCUser("sc1", "owner", "OWNER_TOKEN", "OWNER_REFRESH"),
	}}
	e := newStreamTestServer(store.deps("https://upstream.example/"), nil)

	for _, p := range []string{"/api/track/123/stream?debug=1", "/api/track/123/stream?debug=true&test_mode=true"} {
		rec := doGet(e, p)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, p)
	}
	assert.Empty(t, store.lookups, "debug param must never trigger a token lookup")
	assert.Empty(t, store.resolveSeen, "no SoundCloud call may be made with anyone's token")
}

func TestStreamHandler_DirectCallWithoutAuthReturns401(t *testing.T) {
	// Even if mounted without middleware by mistake, the handler refuses.
	store := &fakeStore{byOwner: map[string]*models.Record{
		"owner": newSCUser("sc1", "owner", "OWNER_TOKEN", ""),
	}}
	e := echo.New()
	e.GET("/api/track/:id/stream", newTrackStreamHandler(store.deps("https://upstream.example/")))
	rec := doGet(e, "/api/track/123/stream?debug=1")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, store.lookups)
}

func TestStreamRoute_AuthedUserWithoutSoundCloudGets404NotOthersToken(t *testing.T) {
	store := &fakeStore{byOwner: map[string]*models.Record{
		"owner": newSCUser("sc1", "owner", "OWNER_TOKEN", "OWNER_REFRESH"),
	}}
	e := newStreamTestServer(store.deps("https://upstream.example/"), newAuthUser("intruder"))

	rec := doGet(e, "/api/track/123/stream?debug=1")
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, []string{"intruder"}, store.lookups, "lookup must be scoped to the caller")
	assert.Empty(t, store.resolveSeen, "owner's token must never be used")
}

func TestStreamCredentials_RejectsRecordOwnedBySomeoneElse(t *testing.T) {
	// A buggy lookup that returns another user's record must not be honoured.
	e := echo.New()
	c := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder())
	c.Set(apis.ContextAuthRecordKey, newAuthUser("intruder"))
	rec, status, _ := streamCredentialsForRequest(c, func(string) (*models.Record, error) {
		return newSCUser("sc1", "owner", "OWNER_TOKEN", ""), nil
	})
	assert.Nil(t, rec)
	assert.Equal(t, http.StatusNotFound, status)
}

func TestStreamRoute_UsesOnlyCallersOwnToken(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("audio"))
	}))
	defer upstream.Close()

	store := &fakeStore{byOwner: map[string]*models.Record{
		"owner": newSCUser("sc1", "owner", "OWNER_TOKEN", ""),
		"alice": newSCUser("sc2", "alice", "ALICE_TOKEN", ""),
	}}
	d := store.deps(upstream.URL + "/audio.mp3")
	d.httpClient = upstream.Client()
	e := newStreamTestServer(d, newAuthUser("alice"))

	rec := doGet(e, "/api/track/123/stream")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, "audio", rec.Body.String())
	assert.Equal(t, []string{"ALICE_TOKEN"}, store.resolveSeen)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"), "no wildcard CORS on the stream proxy")
	assert.Equal(t, "same-origin", rec.Header().Get("Cross-Origin-Resource-Policy"))
}

func TestStreamRoute_PersistsRefreshedToken(t *testing.T) {
	store := &fakeStore{byOwner: map[string]*models.Record{
		"alice": newSCUser("sc2", "alice", "EXPIRED", "OLD_REFRESH"),
	}}
	d := store.deps("") // no upstream; we only care about refresh persistence
	d.validateToken = func(string) (bool, bool) { return false, true }
	d.refreshToken = func(rt string) (*services.TokenRefreshResult, error) {
		assert.Equal(t, "OLD_REFRESH", rt)
		return &services.TokenRefreshResult{AccessToken: "NEW_ACCESS", RefreshToken: "NEW_REFRESH", ExpiresIn: 3600}, nil
	}
	e := newStreamTestServer(d, newAuthUser("alice"))

	rec := doGet(e, "/api/track/123/stream")
	assert.Equal(t, http.StatusNotFound, rec.Code) // no stream URL available
	require.Len(t, store.saved, 1)
	assert.Equal(t, "NEW_ACCESS", store.saved[0].GetString("access_token"))
	assert.Equal(t, "NEW_REFRESH", store.saved[0].GetString("refresh_token"))
	assert.Equal(t, []string{"NEW_ACCESS"}, store.resolveSeen)
}

func TestStreamRoute_RejectsNonNumericTrackID(t *testing.T) {
	store := &fakeStore{byOwner: map[string]*models.Record{
		"alice": newSCUser("sc2", "alice", "ALICE_TOKEN", ""),
	}}
	e := newStreamTestServer(store.deps("https://upstream.example/"), newAuthUser("alice"))
	rec := doGet(e, "/api/track/..%2Fme/stream")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Empty(t, store.resolveSeen)
}

func TestStreamRoute_BlocksRedirectToDisallowedURL(t *testing.T) {
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer upstream.Close()
	store := &fakeStore{byOwner: map[string]*models.Record{
		"alice": newSCUser("sc2", "alice", "ALICE_TOKEN", ""),
	}}
	d := store.deps(upstream.URL + "/x")
	d.httpClient = upstream.Client()
	e := newStreamTestServer(d, newAuthUser("alice"))
	rec := doGet(e, "/api/track/123/stream")
	assert.Equal(t, http.StatusBadGateway, rec.Code)
}

func TestIsAllowedOutboundURL(t *testing.T) {
	ok := []string{
		"https://cf-media.sndcdn.com/abc.mp3?Policy=x",
		"https://api.soundcloud.com/tracks/1/stream",
		"https://playback.media-streaming.soundcloud.cloud/x/playlist.m3u8",
	}
	bad := []string{
		"http://api.soundcloud.com/tracks/1/stream",
		"file:///etc/passwd",
		"https://localhost/x", "https://127.0.0.1/x", "https://10.0.0.5/x",
		"https://169.254.169.254/latest/meta-data/", "https://[::1]/x",
		"https://192.168.1.49:8090/x", "https://user:pass@api.soundcloud.com/x",
		"gopher://x", "",
	}
	for _, u := range ok {
		assert.True(t, isAllowedOutboundURL(u), u)
	}
	for _, u := range bad {
		assert.False(t, isAllowedOutboundURL(u), u)
	}
}

func TestIsSoundCloudHost(t *testing.T) {
	assert.True(t, isSoundCloudHost("api.soundcloud.com"))
	assert.True(t, isSoundCloudHost("cf-hls-media.sndcdn.com"))
	assert.False(t, isSoundCloudHost("soundcloud.com.evil.example"))
	assert.False(t, isSoundCloudHost("evilsoundcloud.com"))
}

func TestIsValidSoundCloudTrackID(t *testing.T) {
	assert.True(t, isValidSoundCloudTrackID("2267275367"))
	for _, id := range []string{"", "abc", "1/../me", "123?x=1", "soundcloud:tracks:1", strings.Repeat("9", 21)} {
		assert.False(t, isValidSoundCloudTrackID(id), id)
	}
}

func TestRedactURL(t *testing.T) {
	assert.Equal(t, "https://cf-media.sndcdn.com/a.mp3", redactURL("https://cf-media.sndcdn.com/a.mp3?Policy=SECRET&Signature=SIG"))
	assert.NotContains(t, redactURL("https://u:p@h/x?y=1#f"), "p@")
}

func signSession(t *testing.T, secret string, claims jwt.MapClaims) string {
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return s
}

func TestParseSessionUserID(t *testing.T) {
	const secret = "test-secret"
	valid := signSession(t, secret, jwt.MapClaims{"id": "u1", "type": "authRecord", "exp": time.Now().Add(time.Hour).Unix()})
	expired := signSession(t, secret, jwt.MapClaims{"id": "u1", "type": "authRecord", "exp": time.Now().Add(-time.Hour).Unix()})
	forged := signSession(t, "attacker-secret", jwt.MapClaims{"id": "victim", "type": "authRecord", "exp": time.Now().Add(time.Hour).Unix()})
	wrongType := signSession(t, secret, jwt.MapClaims{"id": "u1", "type": "admin"})
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"id": "victim", "type": "authRecord"}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	id, ok := parseSessionUserID(valid, secret)
	assert.True(t, ok)
	assert.Equal(t, "u1", id)

	id, ok = parseSessionUserID(expired, secret)
	assert.True(t, ok, "expired-but-genuine sessions may be refreshed")
	assert.Equal(t, "u1", id)

	for name, tok := range map[string]string{"forged": forged, "wrongType": wrongType, "unsigned": unsigned, "garbage": "a.b.c", "empty": ""} {
		_, ok := parseSessionUserID(tok, secret)
		assert.False(t, ok, name)
	}
}

func TestRateLimiter(t *testing.T) {
	now := time.Unix(0, 0)
	r := newRateLimiter(2, time.Minute)
	r.now = func() time.Time { return now }
	assert.True(t, r.allow("a"))
	assert.True(t, r.allow("a"))
	assert.False(t, r.allow("a"))
	assert.True(t, r.allow("b"))
	now = now.Add(61 * time.Second)
	assert.True(t, r.allow("a"))
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	e := echo.New()
	e.Use(securityHeadersMiddleware())
	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, rec.Header().Get("Permissions-Policy"))
	assert.Contains(t, rec.Header().Get("Strict-Transport-Security"), "max-age=")
	csp := rec.Header().Get("Content-Security-Policy-Report-Only")
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "https://*.sndcdn.com")
	assert.Empty(t, rec.Header().Get("Content-Security-Policy"), "CSP must not be enforced yet")
}

func TestCSPReportHandler(t *testing.T) {
	e := echo.New()
	e.POST("/api/csp-report", cspReportHandler)
	body := `{"csp-report":{"document-uri":"https://soundcistern.com/stream?x=1","violated-directive":"img-src","blocked-uri":"https://evil.example/a.png?t=secret"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/csp-report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/csp-report")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestSafeLocalRedirectPath(t *testing.T) {
	assert.Equal(t, "/stream", safeLocalRedirectPath("/stream"))
	assert.Equal(t, "/", safeLocalRedirectPath(""))
	assert.Equal(t, "/example.org/x", safeLocalRedirectPath("//example.org/x"))
	assert.Equal(t, "/example.org", safeLocalRedirectPath("///example.org"))
	assert.Equal(t, "/", safeLocalRedirectPath("/\\example.org"))
	assert.Equal(t, "/", safeLocalRedirectPath("https://example.org"))
}
