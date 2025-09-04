package handlers

import (
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/go-playground/validator/v10"
)

type Handlers struct {
	ClientHandler  ClientHandler
	AccountHandler AccountHandler
}

func NewHandlers(services services.Services, validate *validator.Validate) Handlers {

	clientHandler := NewClientHandler(services.ClientService, validate)
	accountHandler := NewAccountHandler(services.AccountService, validate)

	return Handlers{ClientHandler: clientHandler, AccountHandler: accountHandler}
}
