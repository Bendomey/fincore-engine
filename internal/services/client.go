package services

import (
	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type ClientService interface {
	GetClient(id uint) (*models.Client, error)
	CreateClient(client *models.Client) error
}

type clientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) ClientService {
	return &clientService{repo}
}

func (s *clientService) GetClient(id uint) (*models.Client, error) {
	return s.repo.GetByID(id)
}

func (s *clientService) CreateClient(client *models.Client) error {
	return s.repo.Create(client)
}
