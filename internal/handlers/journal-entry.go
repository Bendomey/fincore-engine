package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/repository"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/Bendomey/fincore-engine/internal/transformations"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type JournalEntryHandler struct {
	service  services.JournalEntryService
	validate *validator.Validate
}

func NewJournalEntryHandler(service services.JournalEntryService, validate *validator.Validate) JournalEntryHandler {
	return JournalEntryHandler{service, validate}
}

type CreateJournalEntryLineInput struct {
	AccountID string  `json:"account_id" validate:"required,uuid4"`
	Notes     *string `json:"notes"      validate:"omitempty,max=1024"`
	Debit     int64   `json:"debit"      validate:"required,number,min=0"`
	Credit    int64   `json:"credit"     validate:"required,number,min=0"`
}

type CreateJournalEntryRequest struct {
	Status          string                        `json:"status"           validate:"required,oneof=DRAFT POSTED"`
	Reference       string                        `json:"reference"        validate:"required,min=3,max=255"`
	TransactionDate *string                       `json:"transaction_date" validate:"omitempty,datetime"`
	Metadata        *map[string]interface{}       `json:"metadata"         validate:"omitempty,json"`
	Lines           []CreateJournalEntryLineInput `json:"lines"            validate:"required,min=2,dive"`
}

func (h *JournalEntryHandler) CreateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var body CreateJournalEntryRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	isPassedValidation := lib.ValidateRequest(h.validate, body, w)

	if !isPassedValidation {
		return
	}

	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lines := make([]services.CreateJournalEntryLineInput, 0)
	for _, line := range body.Lines {
		lines = append(lines, services.CreateJournalEntryLineInput{
			AccountID: line.AccountID,
			Notes:     line.Notes,
			Debit:     line.Debit,
			Credit:    line.Credit,
		})
	}

	journalEntry, err := h.service.CreateJournalEntry(r.Context(), services.CreateJournalEntryInput{
		Status:          body.Status,
		Reference:       body.Reference,
		TransactionDate: body.TransactionDate,
		Metadata:        body.Metadata,
		Lines:           lines,

		ClientID: client.ID.String(),
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBJournalEntryToRestJournalEntry(journalEntry, nil),
	})
}

type UpdateJournalEntryLineInput struct {
	ID        *string `json:"id"         validate:"omitempty,uuid4"`
	AccountID *string `json:"account_id" validate:"omitempty,uuid4"`
	Notes     *string `json:"notes"      validate:"omitempty,max=1024"`
	Debit     *int64  `json:"debit"      validate:"omitempty,number,min=0"`
	Credit    *int64  `json:"credit"     validate:"omitempty,number,min=0"`
}

type UpdateJournalEntryRequest struct {
	Reference       *string                        `json:"reference"        validate:"omitempty,min=3,max=255"`
	TransactionDate *string                        `json:"transaction_date" validate:"omitempty,datetime"`
	Metadata        *map[string]interface{}        `json:"metadata"         validate:"omitempty,json"`
	Lines           *[]UpdateJournalEntryLineInput `json:"lines"            validate:"omitempty,min=2,dive"`
}

func (h *JournalEntryHandler) UpdateJournalEntry(w http.ResponseWriter, r *http.Request) {
	var body UpdateJournalEntryRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	isPassedValidation := lib.ValidateRequest(h.validate, body, w)

	if !isPassedValidation {
		return
	}

	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lines := make([]services.UpdateJournalEntryLineInput, 0)
	if body.Lines != nil {
		for _, line := range *body.Lines {
			lines = append(lines, services.UpdateJournalEntryLineInput{
				ID:        line.ID,
				AccountID: line.AccountID,
				Notes:     line.Notes,
				Debit:     line.Debit,
				Credit:    line.Credit,
			})
		}
	}

	journalEntry, err := h.service.UpdateJournalEntry(
		r.Context(),
		chi.URLParam(r, "journal_entry_id"),
		services.UpdateJournalEntryInput{
			ID:              chi.URLParam(r, "journal_entry_id"),
			Reference:       body.Reference,
			TransactionDate: body.TransactionDate,
			Metadata:        body.Metadata,
			Lines:           &lines,

			ClientID: client.ID.String(),
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBJournalEntryToRestJournalEntry(journalEntry, nil),
	})
}

func (h *JournalEntryHandler) PostJournalEntry(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	journalEntry, err := h.service.PostJournalEntry(r.Context(), services.GetJournalEntryInput{
		ClientID: client.ID.String(),
		ID:       chi.URLParam(r, "journal_entry_id"),
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBJournalEntryToRestJournalEntry(journalEntry, nil),
	})
}

func (h *JournalEntryHandler) DeleteJournalEntry(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.service.DeleteJournalEntry(r.Context(), services.GetJournalEntryInput{
		ClientID: client.ID.String(),
		ID:       chi.URLParam(r, "journal_entry_id"),
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
	json.NewEncoder(w).Encode(map[string]any{})
}

type GetJournalEntryRequest struct {
	ClientID string    `json:"client_id" validate:"required,uuid4"`
	ID       string    `json:"id"        validate:"required,uuid4"`
	Populate *[]string `json:"populate"  validate:"omitempty,dive,oneof=JournalEntryLines, Account"`
}

func (h *JournalEntryHandler) GetJournalEntry(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	input := GetJournalEntryRequest{
		ClientID: client.ID.String(),
		ID:       chi.URLParam(r, "journal_entry_id"),
		Populate: getPopulateFields(r),
	}

	isPassedValidation := lib.ValidateRequest(h.validate, input, w)

	if !isPassedValidation {
		return
	}

	journalEntry, err := h.service.GetJournalEntry(r.Context(), services.GetJournalEntryInput{
		ClientID: input.ClientID,
		ID:       input.ID,
		Populate: input.Populate,
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": err.Error(),
			},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": transformations.DBJournalEntryToRestJournalEntry(journalEntry, input.Populate),
	})
}

type ListJournalEntriesFilterRequest struct {
	ClientID string  `json:"client_id" validate:"required,uuid4"`
	Status   *string `json:"status"    validate:"omitempty,oneof=DRAFT POSTED"`
}

func (h *JournalEntryHandler) ListJournalEntries(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filters := ListJournalEntriesFilterRequest{
		ClientID: client.ID.String(),
		Status:   lib.NullOrString(r.URL.Query().Get("status")),
	}

	isFiltersPassedValidation := lib.ValidateRequest(h.validate, filters, w)
	if !isFiltersPassedValidation {
		return
	}

	filterQuery, filterErr := lib.GenerateQuery(r.URL.Query())
	if filterErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": filterErr.Error(),
			},
		})
		return
	}

	isFilterQueryPassedValidation := lib.ValidateRequest(h.validate, filterQuery, w)
	if !isFilterQueryPassedValidation {
		return
	}

	journalEntries, journalEntriesErr := h.service.ListJournalEntries(
		r.Context(),
		*filterQuery,
		repository.ListJournalEntriesFilter{
			ClientId: filters.ClientID,
			Status:   filters.Status,
		},
	)

	if journalEntriesErr != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": journalEntriesErr.Error(),
			},
		})
		return
	}

	count, countsErr := h.service.CountJournalEntries(r.Context(), *filterQuery, repository.ListJournalEntriesFilter{
		ClientId: filters.ClientID,
		Status:   filters.Status,
	})

	if countsErr != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": countsErr.Error(),
			},
		})
		return
	}

	journalEntriesTransformed := make([]interface{}, 0)
	for _, journalEntry := range journalEntries {
		journalEntriesTransformed = append(
			journalEntriesTransformed,
			transformations.DBJournalEntryToRestJournalEntry(&journalEntry, filterQuery.Populate),
		)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": journalEntriesTransformed,
		"meta": map[string]any{
			"page":              filterQuery.Page,
			"page_size":         filterQuery.PageSize,
			"order":             filterQuery.Order,
			"order_by":          filterQuery.OrderBy,
			"total":             count,
			"has_next_page":     (filterQuery.Page * filterQuery.PageSize) < int(count),
			"has_previous_page": filterQuery.Page > 1,
		},
	})
}
