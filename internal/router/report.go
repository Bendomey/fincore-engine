package router

import (
	"github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewReportRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	// ensure auth is present
	r.Use(middleware.CheckForAuthPresenceMiddleware)

	r.Get("/trial-balance", appCtx.Handlers.ReportHandler.TrialBalance)

	return r
}
