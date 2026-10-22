package repository

import (
	"context"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"gorm.io/gorm"
)

type AccountRepository interface {
	Create(context context.Context, account *models.Account) error
	Update(context context.Context, account *models.Account) error
	Delete(context context.Context, account *models.Account) error
	FindAndDelete(context context.Context, id string) error
	GetByID(context context.Context, id string, populate *[]string) (*models.Account, error)
	GetByIDAndClientID(ctx context.Context, id string, clientID string, populate *[]string) (*models.Account, error)
	List(context context.Context, filterQuery lib.FilterQuery, filters ListAccountsFilter) (*[]models.Account, error)
	Count(context context.Context, filterQuery lib.FilterQuery, filters ListAccountsFilter) (int64, error)
	MaxCode(ctx context.Context, clientID string, accountType string) (*int64, error)
	FirstFreeCode(ctx context.Context, clientID string, from int64) (*int64, error)
}

type accountRepository struct {
	DB *gorm.DB
}

func NewAccountRepository(DB *gorm.DB) AccountRepository {
	return &accountRepository{DB}
}

func (r *accountRepository) Create(ctx context.Context, account *models.Account) error {
	return r.DB.WithContext(ctx).Create(account).Error
}

func (r *accountRepository) Update(ctx context.Context, account *models.Account) error {
	account.UpdatedAt = time.Now()
	return r.DB.WithContext(ctx).Save(account).Error
}

func (r *accountRepository) Delete(ctx context.Context, account *models.Account) error {
	return r.DB.WithContext(ctx).Delete(account).Error
}

func (r *accountRepository) FindAndDelete(ctx context.Context, id string) error {
	var account models.Account
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&account).Error; err != nil {
		return err
	}

	return r.DB.WithContext(ctx).Delete(&account).Error
}

func (r *accountRepository) GetByCode(ctx context.Context, code string) (*models.Account, error) {
	var account models.Account
	result := r.DB.WithContext(ctx).Where("code = ?", code).First(&account)
	if result.Error != nil {
		return nil, result.Error
	}
	return &account, nil
}

func (r *accountRepository) GetByID(ctx context.Context, id string, populate *[]string) (*models.Account, error) {
	var account models.Account
	db := r.DB.WithContext(ctx)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ?", id).First(&account)

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil
}

func (r *accountRepository) GetByIDAndClientID(
	ctx context.Context,
	id string,
	clientID string,
	populate *[]string,
) (*models.Account, error) {
	var account models.Account
	db := r.DB.WithContext(ctx)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ? AND client_id = ?", id, clientID).First(&account)

	if result.Error != nil {
		return nil, result.Error
	}

	return &account, nil
}

type ListAccountsFilter struct {
	ClientId        string
	ParentAccountId *string
	AccountType     *string
	IsContra        *bool
	IsGroup         *bool
}

func (r *accountRepository) List(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters ListAccountsFilter,
) (*[]models.Account, error) {
	var accounts []models.Account

	db := r.DB.WithContext(ctx).
		Scopes(
			DateRangeScope("accounts", filterQuery.DateRange),
			ClientFilterScope("accounts", filters.ClientId),
			ParentAccountFilterScope(filters.ParentAccountId),
			AccountTypeFilterScope(filters.AccountType),
			IsContraFilterScope(filters.IsContra),
			IsGroupFilterScope(filters.IsGroup),
			SearchScope("accounts", filterQuery.Search),

			PaginationScope(filterQuery.Page, filterQuery.PageSize),
			OrderScope("accounts", filterQuery.OrderBy, filterQuery.Order),
		)

	if filterQuery.Populate != nil {
		for _, field := range *filterQuery.Populate {
			db = db.Preload(field)
		}
	}

	results := db.Find(&accounts)

	if results.Error != nil {
		return nil, results.Error
	}

	return &accounts, nil
}

func (r *accountRepository) Count(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters ListAccountsFilter,
) (int64, error) {
	var count int64

	result := r.DB.
		WithContext(ctx).
		Model(&models.Account{}).
		Scopes(
			DateRangeScope("accounts", filterQuery.DateRange),
			ClientFilterScope("accounts", filters.ClientId),
			ParentAccountFilterScope(filters.ParentAccountId),
			AccountTypeFilterScope(filters.AccountType),
			IsContraFilterScope(filters.IsContra),
			IsGroupFilterScope(filters.IsGroup),
			SearchScope("accounts", filterQuery.Search),
		).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

func IsGroupFilterScope(isGroup *bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isGroup == nil {
			return db
		}

		return db.Where("accounts.is_group = ?", *isGroup)
	}
}

func IsContraFilterScope(isContra *bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if isContra == nil {
			return db
		}

		return db.Where("accounts.is_contra = ?", *isContra)
	}
}

func AccountTypeFilterScope(accountType *string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if accountType == nil || *accountType == "" {
			return db
		}

		return db.Where("accounts.type = ?", *accountType)
	}
}

func ParentAccountFilterScope(parentAccountId *string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if parentAccountId == nil || *parentAccountId == "" {
			return db
		}

		return db.Where("accounts.parent_account_id = ?", *parentAccountId)
	}
}

// MaxCode returns the client's highest numeric code for an account type,
// including soft-deleted accounts (their codes stay reserved), or nil if none.
func (r *accountRepository) MaxCode(ctx context.Context, clientID string, accountType string) (*int64, error) {
	var maxCode *int64
	result := r.DB.WithContext(ctx).
		Unscoped().
		Model(&models.Account{}).
		Select("MAX(code::bigint)").
		Where("client_id = ? AND type = ? AND code ~ '^[0-9]+$'", clientID, accountType).
		Scan(&maxCode)
	if result.Error != nil {
		return nil, result.Error
	}

	return maxCode, nil
}

// FirstFreeCode returns the lowest code at or above from that no account of
// the client uses (any type, including soft-deleted), or nil if the next 1000
// codes are all taken.
func (r *accountRepository) FirstFreeCode(ctx context.Context, clientID string, from int64) (*int64, error) {
	var free *int64
	result := r.DB.WithContext(ctx).Raw(`
SELECT MIN(c) FROM generate_series(CAST(@from AS bigint), CAST(@from AS bigint) + 999) AS c
WHERE NOT EXISTS (SELECT 1 FROM accounts a WHERE a.client_id = @client AND a.code = c::text)`,
		map[string]any{"from": from, "client": clientID},
	).Scan(&free)
	if result.Error != nil {
		return nil, result.Error
	}

	return free, nil
}
