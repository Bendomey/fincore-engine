package services

import (
	"context"
	"errors"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type JournalEntryService interface {
	CreateJournalEntry(ctx context.Context, input CreateJournalEntryInput) (*models.JournalEntry, error)
	UpdateJournalEntry(ctx context.Context, journalEntryId string, input UpdateJournalEntryInput) (*models.JournalEntry, error)
	PostJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, input GetJournalEntryInput) error
	GetJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error)
	ListJournalEntries(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListJournalEntriesFilter) ([]models.JournalEntry, error)
	CountJournalEntries(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListJournalEntriesFilter) (int64, error)
}

type journalEntryService struct {
	repo      repository.JournalEntryRepository
	account   repository.AccountRepository
	entryLine repository.JournalEntryLineRepository
}

func NewJournalEntryService(repo repository.JournalEntryRepository, account repository.AccountRepository, entryLine repository.JournalEntryLineRepository) JournalEntryService {
	return &journalEntryService{repo, account, entryLine}
}

type CreateJournalEntryLineInput struct {
	AccountID string
	Notes     *string
	Debit     int64
	Credit    int64
}

type CreateJournalEntryInput struct {
	ClientID string

	Status          string
	Reference       string
	TransactionDate *string
	Metadata        *map[string]interface{}
	Lines           []CreateJournalEntryLineInput
}

func (s *journalEntryService) CreateJournalEntry(ctx context.Context, input CreateJournalEntryInput) (*models.JournalEntry, error) {
	// create journal entry and lines in a transaction
	lines := make([]models.JournalEntryLine, 0)
	for _, line := range input.Lines {
		lines = append(lines, models.JournalEntryLine{
			AccountID: line.AccountID,
			Notes:     line.Notes,
			Debit:     line.Debit,
			Credit:    line.Credit,
		})
	}

	validateLinesErr := validateLines(s.account, ctx, lines)
	if validateLinesErr != nil {
		return nil, validateLinesErr
	}

	journalEntry := models.JournalEntry{
		ClientID:          input.ClientID,
		Status:            input.Status,
		Reference:         input.Reference,
		JournalEntryLines: lines,
	}

	if input.Status == "POSTED" {
		now := time.Now()
		journalEntry.PostedAt = &now
	}

	if input.TransactionDate != nil {
		t, err := time.Parse(time.RFC3339, *input.TransactionDate)
		if err != nil {
			return nil, errors.New("invalid transaction date format")
		}

		journalEntry.TransactionDate = t
	}

	if input.Metadata != nil {
		metadata, err := lib.InterfaceToJSON(*input.Metadata)
		if err != nil {
			return nil, errors.New("invalid metadata format")
		}

		journalEntry.Metadata = metadata
	}

	err := s.repo.Create(ctx, &journalEntry)
	if err != nil {
		return nil, err
	}

	return &journalEntry, nil
}

func validateLines(accountRepo repository.AccountRepository, ctx context.Context, lines []models.JournalEntryLine) error {
	// make sure debits equal credits
	debitTotal := int64(0)
	creditTotal := int64(0)

	for _, line := range lines {
		debitTotal += line.Debit
		creditTotal += line.Credit
	}

	if debitTotal != creditTotal {
		return errors.New("debit and credit totals must be equal")
	}

	// make sure accounts exist and belong to the client
	for _, line := range lines {
		_, err := accountRepo.GetByID(ctx, line.AccountID, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

type UpdateJournalEntryLineInput struct {
	ID        *string
	AccountID *string
	Notes     *string
	Debit     *int64
	Credit    *int64
}

type UpdateJournalEntryInput struct {
	ClientID string

	ID              string
	Reference       *string
	TransactionDate *string
	Metadata        *map[string]interface{}
	Lines           *[]UpdateJournalEntryLineInput
}

func (s *journalEntryService) UpdateJournalEntry(ctx context.Context, journalEntryId string, input UpdateJournalEntryInput) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, nil)
	if err != nil {
		return nil, err
	}

	if entry.Status == "POSTED" {
		return nil, errors.New("journal entry is already posted")
	}

	if input.Reference != nil {
		entry.Reference = *input.Reference
	}

	if input.TransactionDate != nil {
		t, err := time.Parse(time.RFC3339, *input.TransactionDate)
		if err != nil {
			return nil, errors.New("invalid transaction date format")
		}

		entry.TransactionDate = t
	}

	if input.Metadata != nil {
		metadata, err := lib.InterfaceToJSON(*input.Metadata)
		if err != nil {
			return nil, errors.New("invalid metadata format")
		}

		entry.Metadata = metadata
	}

	if input.Lines != nil && len(*input.Lines) > 0 {
		// fetch existing lines and create new lines(without ID)
		lines := make([]models.JournalEntryLine, 0)
		for _, line := range *input.Lines {
			if line.ID != nil {
				lineEntry, err := s.entryLine.GetByIDAndEntryID(ctx, *line.ID, input.ID, nil)
				if err != nil {
					return nil, err
				}

				if line.AccountID != nil {
					lineEntry.AccountID = *line.AccountID
				}

				if line.Debit != nil {
					lineEntry.Debit = *line.Debit
				}

				if line.Credit != nil {
					lineEntry.Credit = *line.Credit
				}

				lineEntry.Notes = line.Notes

				lines = append(lines, *lineEntry)
			} else {
				if line.AccountID == nil || line.Credit == nil || line.Debit == nil {
					return nil, errors.New("account_id, debit and credit are required for new lines")
				}

				newLine := models.JournalEntryLine{
					AccountID:      *line.AccountID,
					Notes:          line.Notes,
					Debit:          *line.Debit,
					Credit:         *line.Credit,
					JournalEntryID: input.ID,
				}

				lines = append(lines, newLine)
			}
		}

		// validate lines
		validateLinesErr := validateLines(s.account, ctx, lines)
		if validateLinesErr != nil {
			return nil, validateLinesErr
		}

		// save lines
		for _, line := range lines {
			err = s.entryLine.Update(ctx, &line)
			if err != nil {
				return nil, err
			}
		}
	}

	// save entry.
	err = s.repo.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

type GetJournalEntryInput struct {
	ClientID string
	ID       string
	Populate *[]string
}

func (s *journalEntryService) PostJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, input.Populate)
	if err != nil {
		return nil, err
	}

	if entry.Status == "POSTED" {
		return nil, errors.New("journal entry is already posted")
	}

	now := time.Now()
	entry.Status = "POSTED"
	entry.PostedAt = &now

	err = s.repo.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *journalEntryService) DeleteJournalEntry(ctx context.Context, input GetJournalEntryInput) error {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, nil)
	if err != nil {
		return err
	}

	if entry.Status == "POSTED" {
		return errors.New("journal entry is already posted")
	}

	return s.repo.Delete(ctx, entry)
}

func (s *journalEntryService) GetJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, input.Populate)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *journalEntryService) ListJournalEntries(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListJournalEntriesFilter) ([]models.JournalEntry, error) {
	entries, err := s.repo.List(ctx, filterQuery, filters)
	if err != nil {
		return nil, err
	}

	return *entries, nil
}

func (s *journalEntryService) CountJournalEntries(ctx context.Context, filterQuery lib.FilterQuery, filters repository.ListJournalEntriesFilter) (int64, error) {
	return s.repo.Count(ctx, filterQuery, filters)
}
