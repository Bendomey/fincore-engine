package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type AccountService interface {
	CreateAccount(ctx context.Context, input CreateAccountInput) (*models.Account, error)
	UpdateAccount(ctx context.Context, accountId string, input UpdateAccountInput) (*models.Account, error)
	DeleteAccount(ctx context.Context, input DeleteAccountInput) error
	GetAccount(ctx context.Context, input GetAccountInput) (*models.Account, error)
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
	IsGroup     bool
	ClientID    string

	ParentAccountID *string
	Description     *string
}

func (s *accountService) CreateAccount(ctx context.Context, input CreateAccountInput) (*models.Account, error) {

	if input.IsGroup {
		if input.ParentAccountID != nil {
			return nil, errors.New("Cannot set parent account for a group account")
		}
	}

	if input.ParentAccountID != nil {

		parentAccount, parentAccountErr := s.repo.GetByID(ctx, *input.ParentAccountID, nil)
		if parentAccountErr != nil {
			return nil, parentAccountErr
		}

		if !parentAccount.IsGroup {
			return nil, errors.New("Parent account must be a group account")
		}

	}

	code, codeErr := GenerateAccountCode(s.repo, input.AccountType)
	if codeErr != nil {
		return nil, codeErr
	}

	account := &models.Account{
		Code:            *code,
		Name:            input.Name,
		Type:            input.AccountType,
		IsContra:        input.IsContra,
		IsGroup:         input.IsGroup,
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

func GenerateAccountCode(repo repository.AccountRepository, accountType string) (*string, error) {
	generatedCode := "1000"
	switch accountType {
	case "ASSET":
		generatedCode = "1000"
	case "LIABILITY":
		generatedCode = "2000"
	case "EQUITY":
		generatedCode = "3000"
	case "INCOME":
		generatedCode = "4000"
	case "EXPENSE":
		generatedCode = "5000"
	default:
		return nil, errors.New("invalid account type")
	}

	accounts, err := repo.List(context.Background(), lib.FilterQuery{
		Page:     1,
		PageSize: 1,
		Order:    "desc",
		OrderBy:  "created_at",
	}, repository.ListAccountsFilter{
		AccountType: &accountType,
	})

	if err != nil {
		return nil, err
	}

	if len(*accounts) > 0 {
		lastAccount := (*accounts)[0]

		// Increment the last account code by 1
		// This assumes that the account codes are numeric strings
		var lastCodeInt int
		_, scanErr := fmt.Sscanf(lastAccount.Code, "%d", &lastCodeInt)
		if scanErr != nil {
			return nil, errors.New("failed to parse last account code")
		}

		generatedCode = fmt.Sprintf("%d", lastCodeInt+1)
	}

	return &generatedCode, nil
}

type UpdateAccountInput struct {
	Name        *string
	Description *string
}

func (s *accountService) UpdateAccount(ctx context.Context, accountId string, input UpdateAccountInput) (*models.Account, error) {
	account, err := s.repo.GetByID(ctx, accountId, nil)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		account.Name = *input.Name
	}

	account.Description = input.Description

	err = s.repo.Update(ctx, account)

	if err != nil {
		return nil, err
	}

	return account, nil
}

type DeleteAccountInput struct {
	ClientID string
	ID       string
}

func (s *accountService) DeleteAccount(ctx context.Context, input DeleteAccountInput) error {
	account, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, nil)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, account)
}

type GetAccountInput struct {
	ClientID string
	ID       string
	Populate *[]string
}

func (s *accountService) GetAccount(ctx context.Context, input GetAccountInput) (*models.Account, error) {
	account, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, input.Populate)
	if err != nil {
		return nil, err
	}

	return account, nil
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
