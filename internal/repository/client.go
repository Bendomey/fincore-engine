package repository

import (
	"context"

	"github.com/Bendomey/fincore-engine/internal/models"
	"gorm.io/gorm"
)

type ClientRepository interface {
	GetByClientEmail(context context.Context, email string) (*models.Client, error)
	GetByClientID(context context.Context, clientId string) (*models.Client, error)
	GetByID(context context.Context, id string) (*models.Client, error)
	Create(context context.Context, client *models.Client) error
}

type clientRepository struct {
	DB *gorm.DB
}

func NewClientRepository(DB *gorm.DB) ClientRepository {
	return &clientRepository{DB}
}

func (r *clientRepository) GetByClientID(ctx context.Context, clientId string) (*models.Client, error) {
	var client models.Client
	result := r.DB.WithContext(ctx).Where("client_id = ?", clientId).First(&client)
	if result.Error != nil {
		return nil, result.Error
	}
	return &client, nil
}

func (r *clientRepository) GetByClientEmail(ctx context.Context, email string) (*models.Client, error) {
	var client models.Client
	result := r.DB.WithContext(ctx).Where("email = ?", email).First(&client)
	if result.Error != nil {
		return nil, result.Error
	}

	return &client, nil
}

func (r *clientRepository) GetByID(ctx context.Context, id string) (*models.Client, error) {
	var client models.Client
	result := r.DB.WithContext(ctx).Where("id = ?", id).First(&client)
	if result.Error != nil {
		return nil, result.Error
	}
	return &client, nil
}

func (r *clientRepository) Create(ctx context.Context, client *models.Client) error {
	return r.DB.WithContext(ctx).Create(client).Error
}
