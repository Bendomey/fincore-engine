package router

import (
	"net/http"
	"time"

	appMiddleware "github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/httprate"
)

func New(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)

	// TODO: figure out how to get the origins from db and then set the cors.
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"User-Agent", "Content-Type", "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Connection", "DNT", "Host", "Origin", "Pragma", "Referer", "X-FinCore-Client-Id", "X-FinCore-Client-Secret"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Rate limit: max 100 requests per minute per IP.
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	r.Use(appMiddleware.VerifyAuthMiddleware(appCtx))

	// rate limit for authed routes
	r.Use(appMiddleware.RateLimitMiddleware)

	r.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// health check
	r.Use(middleware.Heartbeat("/"))

	// serve openapi.yaml + docs
	r.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/openapi.yaml")
	})

	if appCtx.Config.Env != "production" {
		r.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "api/swagger-ui/index.html")
		})
	}

	r.Route("/api", func(r chi.Router) {
		r.Mount("/", NewClientRouter(appCtx)) // clients
	})

	return r
}
