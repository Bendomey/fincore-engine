package middleware

import (
	"net/http"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/go-chi/httprate"
)

const clientIDHeader = "X-FinCore-Client-Id"

// KeyByClientIP keys rate limits by the caller's IP. Behind Fly's proxy every
// connection comes from the proxy, so the caller's address is taken from
// Fly-Client-IP, which the proxy sets; elsewhere it's the connection address.
func KeyByClientIP(r *http.Request) (string, error) {
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return ip, nil
	}

	return httprate.KeyByIP(r)
}

// RateLimitMiddleware limits requests that name a client to 50 per second per
// client id. It runs before auth so bogus secrets can't flood bcrypt for a
// client; counters expire, so made-up ids can't grow memory without bound.
func RateLimitMiddleware(next http.Handler) http.Handler {
	limited := httprate.Limit(50, time.Second, httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
		return r.Header.Get(clientIDHeader), nil
	}))(next)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(clientIDHeader) == "" {
			next.ServeHTTP(w, r)
			return
		}

		limited.ServeHTTP(w, r)
	})
}

// AnonymousIPRateLimitMiddleware limits requests by IP unless they
// authenticated. It runs after auth, so sending a client id header without
// valid credentials doesn't skip it, while one server calling on behalf of
// many tenants isn't throttled by its single IP.
func AnonymousIPRateLimitMiddleware(requests int, window time.Duration) func(http.Handler) http.Handler {
	byIP := httprate.Limit(requests, window, httprate.WithKeyFuncs(KeyByClientIP))

	return func(next http.Handler) http.Handler {
		limited := byIP(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if client, ok := lib.ClientFromContext(r.Context()); ok && client != nil {
				next.ServeHTTP(w, r)
				return
			}

			limited.ServeHTTP(w, r)
		})
	}
}
