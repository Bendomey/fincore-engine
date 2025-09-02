package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
	"github.com/getsentry/raven-go"
	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ClientService interface {
	AuthenticateClient(ctx context.Context, clientId string, clientSecret string) (*models.Client, error)
	GetClient(ctx context.Context, clientId string) (*models.Client, error)
	CreateClient(ctx context.Context, input CreateUserInput) (*CreateUserResponse, error)
}

type clientService struct {
	repo repository.ClientRepository
}

func NewClientService(repo repository.ClientRepository) ClientService {
	return &clientService{repo}
}

func (s *clientService) AuthenticateClient(ctx context.Context, clientId string, clientSecret string) (*models.Client, error) {
	client, err := s.repo.GetByClientID(ctx, clientId)
	if err != nil {
		return nil, err
	}

	if client == nil {
		return nil, errors.New("invalid client credentials")
	}

	isVerified := VerifyClientSecret(client.ClientSecretHash, clientSecret)
	if !isVerified {
		return nil, errors.New("invalid client credentials")
	}

	return client, nil
}

func (s *clientService) GetClient(ctx context.Context, clientId string) (*models.Client, error) {
	return s.repo.GetByID(ctx, clientId)
}

type CreateUserInput struct {
	Name  string
	Email string
}

type CreateUserResponse struct {
	Client models.Client
	Secret string
}

func (s *clientService) CreateClient(ctx context.Context, input CreateUserInput) (*CreateUserResponse, error) {

	// does email exists?
	clientByEmail, clientByEmailErr := s.repo.GetByClientEmail(ctx, input.Email)

	if clientByEmailErr != nil {
		if !errors.Is(clientByEmailErr, gorm.ErrRecordNotFound) {
			raven.CaptureError(clientByEmailErr, nil)
			return nil, clientByEmailErr
		}
	}

	if clientByEmail != nil {
		return nil, errors.New("email already in use")
	}

	uuidV4, err := uuid.NewV4()
	if err != nil {
		raven.CaptureError(err, map[string]string{
			"function": "CreateClient",
			"action":   "generating client id",
		})

		return nil, err
	}

	clientID := "c_" + uuidV4.String()

	// Generate 32 random bytes for secret
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		raven.CaptureError(err, map[string]string{
			"function": "CreateClient",
			"action":   "generating random bytes",
		})
		return nil, err
	}

	clientSecret := base64.RawURLEncoding.EncodeToString(secretBytes) // URL-safe

	hash, err := bcrypt.GenerateFromPassword([]byte(clientSecret), bcrypt.DefaultCost)
	if err != nil {
		raven.CaptureError(err, map[string]string{
			"function": "CreateClient",
			"action":   "hashing random bytes",
		})
		return nil, err
	}

	client := &models.Client{
		Name:             input.Name,
		Email:            input.Email,
		ClientID:         clientID,
		ClientSecretHash: string(hash),
	}

	if err := s.repo.Create(ctx, client); err != nil {
		return nil, err
	}

	return &CreateUserResponse{
		Client: *client,
		Secret: clientSecret,
	}, nil
}

// VerifyClientSecret checks plaintext secret against hashed version
func VerifyClientSecret(hash, secret string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret)) == nil
}
