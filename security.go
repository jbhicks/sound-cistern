package main

// Security helpers: stream-route credential selection, outbound URL guards,
// log redaction, security headers and a small in-memory rate limiter.
//
// These are deliberately free of PocketBase app dependencies (beyond
// *models.Record) so they can be unit-tested without booting PocketBase.

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"

	"github.com/jbhicks/sound-cistern/services"
)

// authRecordFromContext returns the authenticated PocketBase record, or nil.
// Unlike a bare type assertion it never panics when the request is anonymous.
func authRecordFromContext(c echo.Context) *models.Record {
	rec, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	return rec
}

// soundcloudUserLookup returns the soundcloud_users record owned by the given
// PocketBase user id (or an error when none exists).
type soundcloudUserLookup func(userID string) (*models.Record, error)

// streamCredentialsForRequest selects the SoundCloud credentials used by the
// stream proxy. It ONLY ever returns the soundcloud_users record that belongs
// to the authenticated user; there is no debug / "first available token" path.
// On failure it returns a non-zero HTTP status and an error message.
func streamCredentialsForRequest(c echo.Context, lookup soundcloudUserLookup) (*models.Record, int, string) {
	authRecord := authRecordFromContext(c)
	if authRecord == nil || authRecord.Id == "" {
		return nil, http.StatusUnauthorized, "Authentication required"
	}
	if lookup == nil {
		return nil, http.StatusInternalServerError, "Database error"
	}
	scUser, err := lookup(authRecord.Id)
	if err != nil || scUser == nil {
		return nil, http.StatusNotFound, "No SoundCloud account connected"
	}
	// Defence in depth: never use a record that is not linked to this user.
	if owner := scUser.GetString("user_id"); owner != authRecord.Id {
		return nil, http.StatusNotFound, "No SoundCloud account connected"
	}
	return scUser, 0, ""
}

// isValidSoundCloudTrackID accepts only numeric SoundCloud track ids. The id is
// interpolated into upstream API URLs, so anything else is rejected.
func isValidSoundCloudTrackID(id string) bool {
	if id == "" || len(id) > 20 {
		return false
	}
	for _, r := range id {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isSoundCloudHost reports whether host belongs to SoundCloud (API or CDN).
// Only these hosts ever receive the user's bearer token.
func isSoundCloudHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	for _, d := range []string{"soundcloud.com", "sndcdn.com", "soundcloud.cloud"} {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// isSoundCloudHTTPSURL reports whether raw is an https URL on a SoundCloud host.
func isSoundCloudHTTPSURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.User == nil && isSoundCloudHost(u.Hostname())
}

// isAllowedOutboundURL is the SSRF guard for URLs the stream proxy fetches
// (resolved from SoundCloud or read from the soundcloud_tracks.stream_url
// column). It requires https and rejects localhost and literal private,
// loopback, link-local and unspecified IP addresses.
func isAllowedOutboundURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" || host == "localhost" || strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".internal") || strings.HasSuffix(host, ".local") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
			return false
		}
	}
	return true
}

// redactURL drops the query string and fragment (signed CDN URLs carry
// credentials there) so a URL can be logged safely.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "<unparseable url>"
	}
	u.RawQuery = ""
	u.Fragment = ""
	u.User = nil
	return u.String()
}

// trackStreamDeps holds everything the stream proxy needs, so tests can swap
// out the database and SoundCloud.
type trackStreamDeps struct {
	lookupSoundcloudUser soundcloudUserLookup
	saveRecord           func(*models.Record) error
	storedStreamURL      func(trackID string) string
	resolveStreamURL     func(accessToken, trackID, quality, userAgent string) string
	// validateToken returns (valid, checked). checked=false means the check
	// could not be performed and the token should be used as-is.
	validateToken  func(accessToken string) (bool, bool)
	refreshToken   func(refreshToken string) (*services.TokenRefreshResult, error)
	allowURL       func(string) bool
	httpClient     *http.Client
	skipValidation bool // test mode: SoundCloud is mocked
}

// persistRefreshedToken stores a refreshed token pair on the user's own record.
func (d trackStreamDeps) persistRefreshedToken(scUser *models.Record, res *services.TokenRefreshResult) {
	scUser.Set("access_token", res.AccessToken)
	if res.RefreshToken != "" {
		scUser.Set("refresh_token", res.RefreshToken)
	}
	expiresIn := res.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600
	}
	scUser.Set("expires_at", time.Now().Add(time.Duration(expiresIn)*time.Second).Format(time.RFC3339))
	if d.saveRecord != nil {
		if err := d.saveRecord(scUser); err != nil {
			log.Printf("[Stream] Warning: failed to persist refreshed SoundCloud token for record %s", scUser.Id)
		}
	}
}

// newTrackStreamHandler builds GET /api/track/:id/stream. It must be mounted
// with apis.RequireRecordAuth(); it also re-checks auth itself.
func newTrackStreamHandler(d trackStreamDeps) echo.HandlerFunc {
	allow := d.allowURL
	if allow == nil {
		allow = isAllowedOutboundURL
	}
	return func(c echo.Context) error {
		scUser, status, msg := streamCredentialsForRequest(c, d.lookupSoundcloudUser)
		if status != 0 {
			return c.JSON(status, map[string]string{"error": msg})
		}

		trackID := c.PathParam("id")
		if !isValidSoundCloudTrackID(trackID) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid track id"})
		}

		accessToken := scUser.GetString("access_token")
		refreshToken := scUser.GetString("refresh_token")

		doRefresh := func() {
			if refreshToken == "" || d.refreshToken == nil {
				return
			}
			res, err := d.refreshToken(refreshToken)
			if err != nil || res == nil || res.AccessToken == "" {
				log.Printf("[Stream] SoundCloud token refresh failed for record %s", scUser.Id)
				return
			}
			accessToken = res.AccessToken
			if res.RefreshToken != "" {
				refreshToken = res.RefreshToken
			}
			d.persistRefreshedToken(scUser, res)
			log.Printf("[Stream] SoundCloud token refreshed for record %s", scUser.Id)
		}

		// No access token stored: try a refresh first.
		if accessToken == "" {
			doRefresh()
		} else if !d.skipValidation && refreshToken != "" && d.validateToken != nil {
			if valid, checked := d.validateToken(accessToken); checked && !valid {
				doRefresh()
			}
		}

		if accessToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "No access token available"})
		}

		quality := c.QueryParam("quality")
		streamURL := ""
		if d.resolveStreamURL != nil {
			streamURL = d.resolveStreamURL(accessToken, trackID, quality, c.Request().Header.Get("User-Agent"))
		}
		if streamURL == "" && d.storedStreamURL != nil {
			streamURL = d.storedStreamURL(trackID)
		}
		if streamURL == "" {
			log.Printf("[Stream] No stream URL for track %s", trackID)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "No stream URL available for this track. Try syncing tracks again."})
		}
		if !allow(streamURL) {
			log.Printf("[Stream] Refusing to fetch disallowed stream URL %s", redactURL(streamURL))
			return c.JSON(http.StatusBadGateway, map[string]string{"error": "Stream not available"})
		}

		client := d.httpClient
		if client == nil {
			client = &http.Client{}
		}
		// Re-check every redirect hop with the SSRF guard. Go already strips the
		// Authorization header when a redirect leaves the original domain.
		guarded := *client
		guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			if !allow(req.URL.String()) {
				return errors.New("redirect to disallowed URL")
			}
			if !isSoundCloudHost(req.URL.Hostname()) {
				req.Header.Del("Authorization")
			}
			return nil
		}

		req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, streamURL, nil)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]string{"error": "Stream not available"})
		}
		if u, _ := url.Parse(streamURL); u != nil && isSoundCloudHost(u.Hostname()) {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		}
		if rangeHeader := c.Request().Header.Get("Range"); rangeHeader != "" {
			req.Header.Set("Range", rangeHeader)
		}

		proxyResp, err := guarded.Do(req)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch stream"})
		}
		defer proxyResp.Body.Close()

		if proxyResp.StatusCode != http.StatusOK && proxyResp.StatusCode != http.StatusPartialContent {
			log.Printf("[Stream] upstream status %d for %s", proxyResp.StatusCode, redactURL(streamURL))
			return c.JSON(http.StatusBadGateway, map[string]string{"error": "Stream not available"})
		}

		h := c.Response().Header()
		h.Set("Content-Type", proxyResp.Header.Get("Content-Type"))
		h.Set("Accept-Ranges", "bytes")
		// Same-origin only: the player is served from this origin. No wildcard CORS.
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Cache-Control", "private, no-store")
		if v := proxyResp.Header.Get("Content-Range"); v != "" {
			h.Set("Content-Range", v)
		}
		if v := proxyResp.Header.Get("Content-Length"); v != "" {
			h.Set("Content-Length", v)
		}
		c.Response().WriteHeader(proxyResp.StatusCode)
		_, _ = io.Copy(c.Response(), proxyResp.Body)
		return nil
	}
}

// securityHeadersMiddleware adds baseline security headers that
// echo's middleware.Secure() does not set. It intentionally does not set a
// Content-Security-Policy (the Templ pages rely on inline scripts and CDNs;
// roll that out separately in Report-Only mode first).
func securityHeadersMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
			h.Set("X-Content-Type-Options", "nosniff")
			if h.Get("X-Frame-Options") == "" {
				h.Set("X-Frame-Options", "SAMEORIGIN")
			}
			if c.Request().TLS != nil || strings.EqualFold(c.Request().Header.Get("X-Forwarded-Proto"), "https") {
				h.Set("Strict-Transport-Security", "max-age=15552000")
			}
			return next(c)
		}
	}
}

// rateLimiter is a tiny fixed-window, in-memory limiter keyed by string.
type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*rateBucket
	now     func() time.Time
	lastGC  time.Time
}

type rateBucket struct {
	start time.Time
	count int
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, buckets: map[string]*rateBucket{}, now: time.Now}
}

// allow records a hit for key and reports whether it is within the limit.
func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	if now.Sub(r.lastGC) > r.window {
		for k, b := range r.buckets {
			if now.Sub(b.start) >= r.window {
				delete(r.buckets, k)
			}
		}
		r.lastGC = now
	}
	b, ok := r.buckets[key]
	if !ok || now.Sub(b.start) >= r.window {
		r.buckets[key] = &rateBucket{start: now, count: 1}
		return true
	}
	if b.count >= r.limit {
		return false
	}
	b.count++
	return true
}

// rateLimitKey prefers the authenticated user id, then Cloudflare's client IP,
// then echo's RealIP.
func rateLimitKey(c echo.Context) string {
	if rec := authRecordFromContext(c); rec != nil && rec.Id != "" {
		return "u:" + rec.Id
	}
	if ip := c.Request().Header.Get("CF-Connecting-IP"); ip != "" {
		return "ip:" + ip
	}
	return "ip:" + c.RealIP()
}

// rateLimitMiddleware returns 429 once a key exceeds the limiter's budget.
func rateLimitMiddleware(r *rateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !r.allow(rateLimitKey(c)) {
				c.Response().Header().Set("Retry-After", "60")
				return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "Too many requests"})
			}
			return next(c)
		}
	}
}

// parseSessionUserID verifies a pb_auth session JWT minted by this app and
// returns its user id. The HMAC signature is ALWAYS verified; only the expiry
// claim is ignored, so an expired-but-genuine session can be refreshed.
// A token with a bad signature, a non-HMAC alg, or the wrong type is rejected.
func parseSessionUserID(tokenString, secret string) (string, bool) {
	if tokenString == "" || secret == "" {
		return "", false
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithoutClaimsValidation(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || token == nil || !token.Valid {
		return "", false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false
	}
	if typ, _ := claims["type"].(string); typ != "authRecord" {
		return "", false
	}
	id, _ := claims["id"].(string)
	if id == "" {
		return "", false
	}
	return id, true
}

// truncateForLog bounds attacker-controlled strings before logging them.
func truncateForLog(s string, n int) string {
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
