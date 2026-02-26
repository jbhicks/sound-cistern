package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
)

const RequestIDKey = "request_id"

func GenerateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)[:8]
}

func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get("X-Request-ID")
			if id == "" {
				id = GenerateRequestID()
			}
			c.Response().Header().Set("X-Request-ID", id)
			return next(c)
		}
	}
}

func RequestLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			reqID := c.Response().Header().Get("X-Request-ID")
			if reqID == "" {
				reqID = "-"
			}

			log.Printf("[%s] %s %s %d %v %s",
				reqID,
				c.Request().Method,
				c.Request().URL.Path,
				c.Response().Status,
				duration,
				c.RealIP(),
			)

			return err
		}
	}
}

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, times := range rl.requests {
			var valid []time.Time
			for _, t := range times {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			if len(valid) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = valid
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	var valid []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.limit {
		rl.requests[ip] = valid
		return false
	}

	rl.requests[ip] = append(valid, now)
	return true
}

func RateLimitMiddleware(limiter *RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if !limiter.Allow(ip) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}

func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func ValidateQueryParams(params ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, p := range params {
				if val := c.QueryParam(p); val != "" {
					if len(val) > 1000 {
						return echo.NewHTTPError(http.StatusBadRequest, "parameter too long: "+p)
					}
				}
			}
			return next(c)
		}
	}
}
