package handlers

import (
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/go-playground/validator/v10"
)

type Handlers struct {
	ClientHandler       ClientHandler
	AccountHandler      AccountHandler
	JournalEntryHandler JournalEntryHandler
}

func NewHandlers(services services.Services, validate *validator.Validate) Handlers {

	clientHandler := NewClientHandler(services.ClientService, validate)
	accountHandler := NewAccountHandler(services.AccountService, validate)
	journalEntryHandler := NewJournalEntryHandler(services.JournalEntryService, validate)

	return Handlers{ClientHandler: clientHandler, AccountHandler: accountHandler, JournalEntryHandler: journalEntryHandler}
}
