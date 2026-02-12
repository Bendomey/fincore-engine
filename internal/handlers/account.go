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

type AccountHandler struct {
	service  services.AccountService
	validate *validator.Validate
}

func NewAccountHandler(service services.AccountService, validate *validator.Validate) AccountHandler {
	return AccountHandler{service, validate}
}

type CreateAccountRequest struct {
	Name            string  `json:"name"              validate:"required,min=3,max=255"`
	Type            string  `json:"type"              validate:"required,oneof=EXPENSE LIABILITY EQUITY ASSET INCOME"`
	IsContra        bool    `json:"is_contra"         validate:"boolean"`
	IsGroup         bool    `json:"is_group"          validate:"boolean"`
	ParentAccountID *string `json:"parent_account_id" validate:"omitempty,uuid4"`
	Description     *string `json:"description"       validate:"omitempty,max=1024"`
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var body CreateAccountRequest
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

	account, err := h.service.CreateAccount(r.Context(), services.CreateAccountInput{
		Name:            body.Name,
		AccountType:     body.Type,
		IsContra:        body.IsContra,
		IsGroup:         body.IsGroup,
		ParentAccountID: body.ParentAccountID,
		Description:     body.Description,
		ClientID:        client.ID.String(),
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
		"data": transformations.DBAccountToRestAccount(account, nil),
	})
}

type UpdateAccountRequest struct {
	Name        *string `json:"name"        validate:"omitempty,min=3,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1024"`
}

func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	var body UpdateAccountRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	isPassedValidation := lib.ValidateRequest(h.validate, body, w)

	if !isPassedValidation {
		return
	}

	account, err := h.service.UpdateAccount(r.Context(), chi.URLParam(r, "account_id"), services.UpdateAccountInput{
		Name:        body.Name,
		Description: body.Description,
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
		"data": transformations.DBAccountToRestAccount(account, nil),
	})
}

func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := h.service.DeleteAccount(r.Context(), services.DeleteAccountInput{
		ClientID: client.ID.String(),
		ID:       chi.URLParam(r, "account_id"),
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

type GetAccountRequest struct {
	ClientID string    `json:"client_id" validate:"required,uuid4"`
	ID       string    `json:"name"      validate:"required,uuid4"`
	Populate *[]string `json:"populate"  validate:"omitempty,dive,oneof=ParentAccount"`
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	input := GetAccountRequest{
		ClientID: client.ID.String(),
		ID:       chi.URLParam(r, "account_id"),
		Populate: getPopulateFields(r),
	}

	isPassedValidation := lib.ValidateRequest(h.validate, input, w)

	if !isPassedValidation {
		return
	}

	account, err := h.service.GetAccount(r.Context(), services.GetAccountInput{
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
		"data": transformations.DBAccountToRestAccount(account, input.Populate),
	})
}

type ListAccountsFilterRequest struct {
	ClientID        string  `json:"client_id"         validate:"required,uuid4"`
	ParentAccountID *string `json:"parent_account_id" validate:"omitempty,uuid4"`
	AccountType     *string `json:"account_type"      validate:"omitempty,oneof=EXPENSE LIABILITY EQUITY ASSET INCOME"`
	IsContra        *string `json:"is_contra"         validate:"omitempty,boolean"`
	IsGroup         *string `json:"is_group"          validate:"omitempty,boolean"`
}

func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	client, clientOk := lib.ClientFromContext(r.Context())

	if !clientOk {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	filters := ListAccountsFilterRequest{
		ClientID:        client.ID.String(),
		ParentAccountID: lib.NullOrString(r.URL.Query().Get("parent_account_id")),
		AccountType:     lib.NullOrString(r.URL.Query().Get("account_type")),
		IsContra:        lib.NullOrString(r.URL.Query().Get("is_contra")),
		IsGroup:         lib.NullOrString(r.URL.Query().Get("is_group")),
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

	accounts, accountsErr := h.service.ListAccounts(r.Context(), *filterQuery, repository.ListAccountsFilter{
		ClientId:        filters.ClientID,
		ParentAccountId: filters.ParentAccountID,
		AccountType:     filters.AccountType,
		IsContra:        lib.ConvertStringPointerToBoolPointer(filters.IsContra),
		IsGroup:         lib.ConvertStringPointerToBoolPointer(filters.IsGroup),
	})

	if accountsErr != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]any{
			"errors": map[string]string{
				"message": accountsErr.Error(),
			},
		})
		return
	}

	count, countsErr := h.service.CountAccounts(r.Context(), *filterQuery, repository.ListAccountsFilter{
		ClientId:        filters.ClientID,
		ParentAccountId: filters.ParentAccountID,
		AccountType:     filters.AccountType,
		IsContra:        lib.ConvertStringPointerToBoolPointer(filters.IsContra),
		IsGroup:         lib.ConvertStringPointerToBoolPointer(filters.IsGroup),
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

	accountsTransformed := make([]interface{}, 0)
	for _, account := range accounts {
		accountsTransformed = append(
			accountsTransformed,
			transformations.DBAccountToRestAccount(&account, filterQuery.Populate),
		)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"data": accountsTransformed,
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
