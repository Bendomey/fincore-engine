package repository

import (
	"context"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"gorm.io/gorm"
)

type JournalEntryRepository interface {
	Create(context context.Context, journalEntry *models.JournalEntry) error
	Update(context context.Context, journalEntry *models.JournalEntry) error
	Delete(context context.Context, journalEntry *models.JournalEntry) error
	FindAndDelete(context context.Context, id string) error
	GetByIDAndClientID(
		context context.Context,
		id string,
		clientID string,
		populate *[]string,
	) (*models.JournalEntry, error)
	GetByID(context context.Context, id string, populate *[]string) (*models.JournalEntry, error)
	List(
		context context.Context,
		filterQuery lib.FilterQuery,
		filters ListJournalEntriesFilter,
	) (*[]models.JournalEntry, error)
	Count(context context.Context, filterQuery lib.FilterQuery, filters ListJournalEntriesFilter) (int64, error)
}

type journalEntryRepository struct {
	DB *gorm.DB
}

func NewJournalEntryRepository(DB *gorm.DB) JournalEntryRepository {
	return &journalEntryRepository{DB}
}

func (r *journalEntryRepository) Create(ctx context.Context, journalEntry *models.JournalEntry) error {
	return r.DB.WithContext(ctx).Create(journalEntry).Error
}

func (r *journalEntryRepository) Update(ctx context.Context, journalEntry *models.JournalEntry) error {
	journalEntry.UpdatedAt = time.Now()
	return r.DB.WithContext(ctx).Save(journalEntry).Error
}

func (r *journalEntryRepository) Delete(ctx context.Context, journalEntry *models.JournalEntry) error {
	return r.DB.WithContext(ctx).Delete(journalEntry).Error
}

func (r *journalEntryRepository) FindAndDelete(ctx context.Context, id string) error {
	var journalEntry models.JournalEntry
	if err := r.DB.WithContext(ctx).Where("id = ?", id).First(&journalEntry).Error; err != nil {
		return err
	}

	return r.DB.WithContext(ctx).Delete(&journalEntry).Error
}

func (r *journalEntryRepository) GetByIDAndClientID(
	context context.Context,
	id string,
	clientID string,
	populate *[]string,
) (*models.JournalEntry, error) {
	var journalEntry models.JournalEntry
	db := r.DB.WithContext(context)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ? AND client_id = ?", id, clientID).First(&journalEntry)

	if result.Error != nil {
		return nil, result.Error
	}

	return &journalEntry, nil
}

func (r *journalEntryRepository) GetByID(
	ctx context.Context,
	id string,
	populate *[]string,
) (*models.JournalEntry, error) {
	var journalEntry models.JournalEntry
	db := r.DB.WithContext(ctx)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ?", id).First(&journalEntry)

	if result.Error != nil {
		return nil, result.Error
	}

	return &journalEntry, nil
}

type ListJournalEntriesFilter struct {
	ClientId string
	Status   *string
}

func (r *journalEntryRepository) List(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters ListJournalEntriesFilter,
) (*[]models.JournalEntry, error) {
	var journalEntries []models.JournalEntry

	db := r.DB.WithContext(ctx).
		Scopes(
			DateRangeScope("journal_entries", filterQuery.DateRange),
			ClientFilterScope("journal_entries", filters.ClientId),
			StatusFilterScope(filters.Status),
			SearchScope("journal_entries", filterQuery.Search),

			PaginationScope(filterQuery.Page, filterQuery.PageSize),
			OrderScope("journal_entries", filterQuery.OrderBy, filterQuery.Order),
		)

	if filterQuery.Populate != nil {
		for _, field := range *filterQuery.Populate {
			db = db.Preload(field)
		}
	}

	results := db.Find(&journalEntries)

	if results.Error != nil {
		return nil, results.Error
	}

	return &journalEntries, nil
}

func (r *journalEntryRepository) Count(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters ListJournalEntriesFilter,
) (int64, error) {
	var count int64

	result := r.DB.
		WithContext(ctx).
		Model(&models.JournalEntry{}).
		Scopes(
			DateRangeScope("journal_entries", filterQuery.DateRange),
			ClientFilterScope("journal_entries", filters.ClientId),
			StatusFilterScope(filters.Status),
			SearchScope("journal_entries", filterQuery.Search),
		).
		Count(&count)

	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}

func StatusFilterScope(status *string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if status == nil || *status == "" {
			return db
		}

		return db.Where("journal_entries.status = ?", *status)
	}
}
