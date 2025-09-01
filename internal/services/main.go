package services

import (
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type Services struct {
	ClientService ClientService
}

func NewServices(repository repository.Repository) Services {

	clientService := NewClientService(repository.ClientRepository)
	return Services{ClientService: clientService}
}
