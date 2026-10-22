package middleware

import (
	"net/http"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/httprate"
)

// Failed authentications are counted per IP and client id, so guessing one
// client's secret is throttled without locking out other clients served from
// the same IP (one caller often acts for many tenants), and per IP overall,
// so rotating client ids doesn't escape the limit. Once either is over its
// limit, credentialed requests are refused before any lookup or bcrypt.
const (
	failedAuthPerClientLimit = 20
	failedAuthPerIPLimit     = 300
)

var (
	failedAuthPerClient = httprate.NewRateLimiter(failedAuthPerClientLimit, time.Minute)
	failedAuthPerIP     = httprate.NewRateLimiter(failedAuthPerIPLimit, time.Minute)
)

// atLimit reports whether key has used up limit, without counting a request.
// OnLimit stops counting once a key reaches its limit, so the check is >=.
func atLimit(limiter *httprate.RateLimiter, key string, limit int) bool {
	_, rate, err := limiter.Status(key)
	return err == nil && rate >= float64(limit)
}

func VerifyAuthMiddleware(appCtx pkg.AppContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientId := r.Header.Get("X-FinCore-Client-Id")
			clientSecret := r.Header.Get("X-FinCore-Client-Secret")

			if clientId != "" && clientSecret != "" {
				ipKey, _ := KeyByClientIP(r)
				clientKey := ipKey + "|" + clientId
				if atLimit(failedAuthPerClient, clientKey, failedAuthPerClientLimit) ||
					atLimit(failedAuthPerIP, ipKey, failedAuthPerIPLimit) {
					http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
					return
				}

				client, err := appCtx.Services.ClientService.AuthenticateClient(r.Context(), clientId, clientSecret)
				if err != nil || client == nil {
					limitedClient := failedAuthPerClient.OnLimit(w, r, clientKey)
					limitedIP := failedAuthPerIP.OnLimit(w, r, ipKey)
					if limitedClient || limitedIP {
						http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
						return
					}

					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Attach client to context
				ctx := lib.WithClient(r.Context(), client)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CheckForAuthPresenceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client, ok := lib.ClientFromContext(r.Context())
		if !ok || client == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
