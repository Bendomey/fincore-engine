package services

import (
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type Services struct {
	ClientService       ClientService
	AccountService      AccountService
	JournalEntryService JournalEntryService
}

func NewServices(repository repository.Repository) Services {

	clientService := NewClientService(repository.ClientRepository)
	accountService := NewAccountService(repository.AccountRepository)
	journalEntryService := NewJournalEntryService(repository.JournalEntryRepository, repository.AccountRepository, repository.JournalEntryLineRepository)

	return Services{ClientService: clientService, AccountService: accountService, JournalEntryService: journalEntryService}
}
