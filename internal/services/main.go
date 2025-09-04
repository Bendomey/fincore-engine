package services

import (
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type Services struct {
	ClientService  ClientService
	AccountService AccountService
}

func NewServices(repository repository.Repository) Services {

	clientService := NewClientService(repository.ClientRepository)
	accountService := NewAccountService(repository.AccountRepository)

	return Services{ClientService: clientService, AccountService: accountService}
}
