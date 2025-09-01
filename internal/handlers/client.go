package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/Bendomey/fincore-engine/internal/transformations"
	"github.com/go-playground/validator/v10"
)

type ClientHandler struct {
	service  services.ClientService
	validate *validator.Validate
}

func NewClientHandler(service services.ClientService, validate *validator.Validate) ClientHandler {
	return ClientHandler{service, validate}
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3,max=50"`
	Email string `json:"email" validate:"required,email"`
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&body); decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}

	isPassedValidation := lib.ValidateRequest(h.validate, body, w)

	if !isPassedValidation {
		return
	}

	client, err := h.service.CreateClient(r.Context(), services.CreateUserInput{
		Name:  body.Name,
		Email: body.Email,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message": "Client created successfully",
		"data":    transformations.DBClientToRestClient(&client.Client, &client.Secret),
	})

}

func (h *ClientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	client, err := h.service.GetClient(r.Context(), id)

	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(client)
}
