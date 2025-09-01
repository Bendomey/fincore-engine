package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Bendomey/fincore-engine/pkg"
)

func New(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	// health check
	r.Use(middleware.Heartbeat("/healthz"))

	r.Route("/api", func(r chi.Router) {
		r.Mount("/", NewClientRouter(appCtx)) // clients
	})

	// serve openapi.yaml + docs
	r.Get("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "api/openapi.yaml")
	})

	if appCtx.Config.Env != "production" {
		r.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "api/swagger-ui/index.html")
		})
	}

	return r
}
