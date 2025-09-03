package router

import (
	"net/http"
	"time"

	appMiddleware "github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"

	"github.com/Bendomey/fincore-engine/pkg"
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

	r.Route("/api", func(r chi.Router) {
		r.Mount("/", NewClientRouter(appCtx)) // clients
	})

	// serve openapi.yaml + docs
	r.Handle("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./api/service-specs"))))

	if appCtx.Config.Env != "production" {
		r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
<!DOCTYPE html>
<html>
  <head>
    <title>FinCore Docs</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function() {
        SwaggerUIBundle({
          url: '/swagger/index.yaml',
          dom_id: '#swagger-ui'
        });
      };
    </script>
  </body>
</html>
	`))
		})
	}

	return r
}
