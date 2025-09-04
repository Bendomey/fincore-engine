package services

import (
	"context"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type AccountService interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (*models.Account, error)
	UpdateAccount(ctx context.Context, accountId string, input UpdateAccountInput) (*models.Account, error)
	DeleteAccount(ctx context.Context, accountId string) error
	GetAccount(ctx context.Context, accountId string, populate *[]string) (*models.Account, error)
	ListAccounts(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListAccountsFilter) ([]models.Account, error)
	CountAccounts(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListAccountsFilter) (int64, error)
}

type accountService struct {
	repo repository.AccountRepository
}

func NewAccountService(repo repository.AccountRepository) AccountService {
	return &accountService{repo}
}

type CreateAccountInput struct {
	Name        string
	AccountType string
	IsContra    bool
	ClientID    string

	ParentAccountID *string
	Description     *string
}

func (s *accountService) CreateAccount(ctx context.Context, input CreateAccountInput) (*models.Account, error) {
	// TODO: create code.

	account := &models.Account{
		Name:            input.Name,
		Type:            input.AccountType,
		IsContra:        input.IsContra,
		ParentAccountID: input.ParentAccountID,
		ClientID:        input.ClientID,
		Description:     input.Description,
	}

	err := s.repo.Create(ctx, account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

type UpdateAccountInput struct {
	Name            *string
	IsContra        *bool
	ParentAccountID *string
	Description     *string
}

func (s *accountService) UpdateAccount(ctx context.Context, accountId string, input UpdateAccountInput) (*models.Account, error) {
	account, err := s.repo.GetByID(ctx, accountId, nil)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		account.Name = *input.Name
	}

	if input.IsContra != nil {
		account.IsContra = *input.IsContra
	}

	account.ParentAccountID = input.ParentAccountID
	account.Description = input.Description

	err = s.repo.Update(ctx, account)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, accountId string) error {
	_, err := s.repo.GetByID(ctx, accountId, nil)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, accountId)
}

func (s *accountService) GetAccount(ctx context.Context, accountId string, populate *[]string) (*models.Account, error) {
	return s.repo.GetByID(ctx, accountId, populate)
}

func (s *accountService) ListAccounts(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListAccountsFilter) ([]models.Account, error) {
	accounts, err := s.repo.List(ctx, filterQuery, filters)
	if err != nil {
		return nil, err
	}

	return *accounts, nil
}

func (s *accountService) CountAccounts(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListAccountsFilter) (int64, error) {
	return s.repo.Count(ctx, filterQuery, filters)
}
