package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

var (
	limiters = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getLimiter(key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	l, exists := limiters[key]
	if !exists {
		// 50 requests per second with burst of 100
		// TODO: figure out free/premium
		l = rate.NewLimiter(50, 100)
		limiters[key] = l
	}
	return l
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := r.Header.Get("X-FinCore-Client-Id")

		if client != "" {
			limiter := getLimiter(client)

			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
