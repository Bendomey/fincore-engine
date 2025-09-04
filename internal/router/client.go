package router

import (
	"github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewClientRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/", appCtx.Handlers.ClientHandler.CreateClient)
	})

	// protected client routes ...
	r.Group(func(r chi.Router) {
		// ensure auth is present
		r.Use(middleware.CheckForAuthPresenceMiddleware)

		r.Get("/me", appCtx.Handlers.ClientHandler.GetClient)
	})

	return r
}
