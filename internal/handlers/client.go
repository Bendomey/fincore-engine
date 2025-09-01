package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/go-chi/chi/v5"
)

type ClientHandler struct {
	service services.ClientService
}

func NewClientHandler(service services.ClientService) ClientHandler {
	return ClientHandler{service: service}
}

func (h *ClientHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var user models.Client
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.service.CreateClient(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *ClientHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	client, err := h.service.GetClient(uint(id))
	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(client)
}
