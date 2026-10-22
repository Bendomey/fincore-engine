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
	UpdateAccount(ctx context.Context, input UpdateAccountInput) (*models.Account, error)
	DeleteAccount(ctx context.Context, input DeleteAccountInput) error
	GetAccount(ctx context.Context, input GetAccountInput) (*models.Account, error)
	ListAccounts(
		ctx context.Context,
		filterQuery lib.FilterQuery,
		filters repository.ListAccountsFilter,
	) ([]models.Account, error)
	CountAccounts(
		ctx context.Context,
		filterQuery lib.FilterQuery,
		filters repository.ListAccountsFilter,
	) (int64, error)
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

		parentAccount, parentAccountErr := s.repo.GetByIDAndClientID(ctx, *input.ParentAccountID, input.ClientID, nil)
		if parentAccountErr != nil {
			return nil, parentAccountErr
		}

		if !parentAccount.IsGroup {
			return nil, errors.New("Parent account must be a group account")
		}

	}

	account := &models.Account{
		Name:            input.Name,
		Type:            input.AccountType,
		IsContra:        input.IsContra,
		IsGroup:         input.IsGroup,
		ParentAccountID: input.ParentAccountID,
		ClientID:        input.ClientID,
		Description:     input.Description,
	}

	// Two creates for the same client and type can pick the same code; the
	// unique index rejects the loser, which retries with a fresh code.
	for attempt := 0; ; attempt++ {
		code, codeErr := GenerateAccountCode(ctx, s.repo, input.ClientID, input.AccountType)
		if codeErr != nil {
			return nil, codeErr
		}
		account.Code = *code

		err := s.repo.Create(ctx, account)
		if err == nil {
			break
		}

		if !lib.IsUniqueViolation(err) {
			return nil, err
		}

		if attempt == 2 {
			return nil, errors.New("could not allocate an account code, try again")
		}
	}

	return account, nil
}

// GenerateAccountCode returns the next code for the client and account type:
// the first code, from one above the client's highest code of that type (or
// the type's base code), that no other account of the client uses. Codes are
// unique per client across types, so a type that has grown into the next
// type's range skips over the codes already taken there.
func GenerateAccountCode(
	ctx context.Context,
	repo repository.AccountRepository,
	clientID string,
	accountType string,
) (*string, error) {
	var base int64
	switch accountType {
	case "ASSET":
		base = 1000
	case "LIABILITY":
		base = 2000
	case "EQUITY":
		base = 3000
	case "INCOME":
		base = 4000
	case "EXPENSE":
		base = 5000
	default:
		return nil, errors.New("invalid account type")
	}

	maxCode, err := repo.MaxCode(ctx, clientID, accountType)
	if err != nil {
		return nil, err
	}

	from := base
	if maxCode != nil {
		from = *maxCode + 1
	}

	next, err := repo.FirstFreeCode(ctx, clientID, from)
	if err != nil {
		return nil, err
	}

	if next == nil {
		return nil, fmt.Errorf("no free account code after %d", from)
	}

	code := fmt.Sprintf("%d", *next)
	return &code, nil
}

type UpdateAccountInput struct {
	ClientID    string
	ID          string
	Name        *string
	Description *string
}

func (s *accountService) UpdateAccount(
	ctx context.Context,
	input UpdateAccountInput,
) (*models.Account, error) {
	account, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, nil)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		account.Name = *input.Name
	}

	if input.Description != nil {
		account.Description = input.Description
	}

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

func (s *accountService) ListAccounts(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters repository.ListAccountsFilter,
) ([]models.Account, error) {
	accounts, err := s.repo.List(ctx, filterQuery, filters)
	if err != nil {
		return nil, err
	}

	return *accounts, nil
}

func (s *accountService) CountAccounts(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters repository.ListAccountsFilter,
) (int64, error) {
	return s.repo.Count(ctx, filterQuery, filters)
}
