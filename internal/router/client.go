package router

import (
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewClientRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/v1/clients", func(r chi.Router) {
		r.Post("/", appCtx.Handlers.ClientHandler.CreateClient)
	})

	// r.Route("/v1/clients", func(r chi.Router) {
	// 	// ensure auth is present
	// 	r.Use(appMiddleware.CheckForAuthPresenceMiddleware)

	// 	// register all routes....
	// 	r.Get("/me", appCtx.Handlers.ClientHandler.GetClient)
	// })

	return r
}
