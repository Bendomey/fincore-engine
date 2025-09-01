package middleware

import (
	"net/http"

	"github.com/Bendomey/fincore-engine/pkg"
)

func VerifyAuthMiddleware(appCtx pkg.AppContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientId := r.Header.Get("X-FinCore-Client-Id")
			clientSecret := r.Header.Get("X-FinCore-Client-Secret")

			if clientId != "" && clientSecret != "" {
				client, err := appCtx.Services.ClientService.AuthenticateClient(r.Context(), clientId, clientSecret)
				if err != nil || client == nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Attach client to context
				ctx := WithClient(r.Context(), client)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CheckForAuthPresenceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client, ok := ClientFromContext(r.Context())
		if !ok || client == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
