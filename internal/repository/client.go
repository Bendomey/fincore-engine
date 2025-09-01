package repository

import (
	"github.com/Bendomey/fincore-engine/internal/models"
	"gorm.io/gorm"
)

type ClientRepository interface {
	GetByID(id uint) (*models.Client, error)
	Create(client *models.Client) error
}

type clientRepository struct {
	DB *gorm.DB
}

func NewClientRepository(DB *gorm.DB) ClientRepository {
	return &clientRepository{DB}
}

func (r *clientRepository) GetByID(id uint) (*models.Client, error) {
	var client models.Client
	result := r.DB.First(&client, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &client, nil
}

func (r *clientRepository) Create(client *models.Client) error {
	return r.DB.Create(client).Error
}
