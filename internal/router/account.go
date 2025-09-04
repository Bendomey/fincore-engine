package router

import (
	"github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewAccountRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	// ensure auth is present
	r.Use(middleware.CheckForAuthPresenceMiddleware)

	r.Post("/", appCtx.Handlers.AccountHandler.CreateAccount)
	r.Get("/", appCtx.Handlers.AccountHandler.ListAccounts)

	r.Get("/{account_id}", appCtx.Handlers.AccountHandler.GetAccount)
	r.Patch("/{account_id}", appCtx.Handlers.AccountHandler.UpdateAccount)
	r.Delete("/{account_id}", appCtx.Handlers.AccountHandler.DeleteAccount)

	return r
}
