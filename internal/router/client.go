package router

import (
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewClientRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	r.Route("/clients", func(r chi.Router) {
		r.Post("/", appCtx.Handlers.ClientHandler.CreateClient)
		r.Get("/{id}", appCtx.Handlers.ClientHandler.GetClient)
	})

	return r
}
