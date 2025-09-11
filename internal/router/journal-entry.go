package router

import (
	"github.com/Bendomey/fincore-engine/internal/middleware"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/go-chi/chi/v5"
)

func NewJournalEntryRouter(appCtx pkg.AppContext) *chi.Mux {
	r := chi.NewRouter()

	// ensure auth is present
	r.Use(middleware.CheckForAuthPresenceMiddleware)

	r.Post("/", appCtx.Handlers.JournalEntryHandler.CreateJournalEntry)
	r.Get("/", appCtx.Handlers.JournalEntryHandler.ListJournalEntries)

	r.Get("/{journal_entry_id}", appCtx.Handlers.JournalEntryHandler.GetJournalEntry)
	r.Patch("/{account_id}", appCtx.Handlers.JournalEntryHandler.UpdateJournalEntry)
	r.Patch("/{journal_entry_id}/post", appCtx.Handlers.JournalEntryHandler.PostJournalEntry)
	r.Delete("/{account_id}", appCtx.Handlers.JournalEntryHandler.DeleteJournalEntry)

	return r
}
