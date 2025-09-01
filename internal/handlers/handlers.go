package handlers

import (
	"github.com/Bendomey/fincore-engine/internal/services"
)

type Handlers struct {
	ClientHandler ClientHandler
}

func NewHandlers(services services.Services) Handlers {

	clientHandler := NewClientHandler(services.ClientService)
	return Handlers{ClientHandler: clientHandler}
}
