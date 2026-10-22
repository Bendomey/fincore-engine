package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/Bendomey/fincore-engine/internal/transformations"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

const (
	defaultLedgerPageSize = 50
	maxLedgerPageSize     = 200
)

type ReportHandler struct {
	service  services.ReportService
	validate *validator.Validate
}

func NewReportHandler(service services.ReportService, validate *validator.Validate) ReportHandler {
	return ReportHandler{service, validate}
}

// dateParam parses an optional date query param; ok is false after writing a 400.
func dateParam(w http.ResponseWriter, r *http.Request, name string, endOfDay bool) (*time.Time, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, true
	}

	t, err := lib.ParseDateParam(value, endOfDay)
	if err != nil {
		writeError(w, http.StatusBadRequest, name+": "+err.Error())
		return nil, false
	}

	return &t, true
}

func (h *ReportHandler) TrialBalance(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())
	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	asOf, ok := dateParam(w, r, "as_of", true)
	if !ok {
		return
	}
	if asOf == nil {
		now := time.Now()
		asOf = &now
	}

	report, err := h.service.TrialBalance(r.Context(), client.ID.String(), *asOf)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBTrialBalanceToRestTrialBalance(report),
	})
}

func (h *ReportHandler) AccountLedger(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())
	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	start, ok := dateParam(w, r, "start_date", false)
	if !ok {
		return
	}

	end, ok := dateParam(w, r, "end_date", true)
	if !ok {
		return
	}
	if end == nil {
		now := time.Now()
		end = &now
	}

	page, pageSize := 1, defaultLedgerPageSize
	if v := r.URL.Query().Get("page"); v != "" {
		p, err := lib.ConvertStringToInt(v)
		if err != nil || p < 1 {
			writeError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = p
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		ps, err := lib.ConvertStringToInt(v)
		if err != nil || ps < 1 || ps > maxLedgerPageSize {
			writeError(w, http.StatusBadRequest, "page_size must be between 1 and 200")
			return
		}
		pageSize = ps
	}

	ledger, err := h.service.AccountLedger(r.Context(), services.LedgerInput{
		ClientID:  client.ID.String(),
		AccountID: chi.URLParam(r, "account_id"),
		Start:     start,
		End:       *end,
		Page:      page,
		PageSize:  pageSize,
	})
	if errors.Is(err, services.ErrLedgerAccountNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBAccountLedgerToRestAccountLedger(ledger),
		"meta": map[string]any{
			"page":              page,
			"page_size":         pageSize,
			"total":             ledger.Total,
			"has_next_page":     int64(page*pageSize) < ledger.Total,
			"has_previous_page": page > 1,
		},
	})
}
