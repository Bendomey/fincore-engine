package repository

import (
	"context"
	"time"

	"github.com/Bendomey/fincore-engine/internal/models"
	"gorm.io/gorm"
)

type JournalEntryLineRepository interface {
	GetByIDAndEntryID(
		context context.Context,
		id string,
		entryID string,
		populate *[]string,
	) (*models.JournalEntryLine, error)
	GetByID(context context.Context, id string, populate *[]string) (*models.JournalEntryLine, error)
	Update(ctx context.Context, journalEntryLine *models.JournalEntryLine) error
}

type journalEntryLineRepository struct {
	DB *gorm.DB
}

func (r *journalEntryLineRepository) Update(ctx context.Context, journalEntryLine *models.JournalEntryLine) error {
	journalEntryLine.UpdatedAt = time.Now()
	return r.DB.WithContext(ctx).Save(journalEntryLine).Error
}

func NewJournalEntryLineRepository(DB *gorm.DB) JournalEntryLineRepository {
	return &journalEntryLineRepository{DB}
}

func (r *journalEntryLineRepository) GetByIDAndEntryID(
	context context.Context,
	id string,
	entryID string,
	populate *[]string,
) (*models.JournalEntryLine, error) {
	var journalEntryLine models.JournalEntryLine
	db := r.DB.WithContext(context)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ? AND journal_entry_id = ?", id, entryID).First(&journalEntryLine)

	if result.Error != nil {
		return nil, result.Error
	}

	return &journalEntryLine, nil
}

func (r *journalEntryLineRepository) GetByID(
	ctx context.Context,
	id string,
	populate *[]string,
) (*models.JournalEntryLine, error) {
	var journalEntryLine models.JournalEntryLine
	db := r.DB.WithContext(ctx)

	if populate != nil {
		for _, field := range *populate {
			db = db.Preload(field)
		}
	}

	result := db.Where("id = ?", id).First(&journalEntryLine)

	if result.Error != nil {
		return nil, result.Error
	}

	return &journalEntryLine, nil
}
