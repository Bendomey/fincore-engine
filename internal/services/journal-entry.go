package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bendomey/fincore-engine/internal/lib"
	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type JournalEntryService interface {
	CreateJournalEntry(ctx context.Context, input CreateJournalEntryInput) (*models.JournalEntry, error)
	UpdateJournalEntry(
		ctx context.Context,
		journalEntryId string,
		input UpdateJournalEntryInput,
	) (*models.JournalEntry, error)
	PostJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error)
	DeleteJournalEntry(ctx context.Context, input GetJournalEntryInput) error
	GetJournalEntry(ctx context.Context, input GetJournalEntryInput) (*models.JournalEntry, error)
	ListJournalEntries(
		ctx context.Context,
		filterQuery lib.FilterQuery,
		filters repository.ListJournalEntriesFilter,
	) ([]models.JournalEntry, error)
	CountJournalEntries(
		ctx context.Context,
		filterQuery lib.FilterQuery,
		filters repository.ListJournalEntriesFilter,
	) (int64, error)
}

// DuplicateReferenceError means the client already has a live entry with the
// requested reference. Callers retrying a create can treat it as success.
type DuplicateReferenceError struct {
	ExistingID string
}

func (e *DuplicateReferenceError) Error() string {
	return "a journal entry with this reference already exists"
}

// duplicateReference returns a DuplicateReferenceError if another live entry
// of the client uses reference.
func (s *journalEntryService) duplicateReference(
	ctx context.Context,
	clientID string,
	reference string,
	exceptID string,
) error {
	existing, err := s.repo.GetByReference(ctx, clientID, reference)
	if err != nil {
		return err
	}

	if existing != nil && existing.ID.String() != exceptID {
		return &DuplicateReferenceError{ExistingID: existing.ID.String()}
	}

	return nil
}

type journalEntryService struct {
	repo      repository.JournalEntryRepository
	account   repository.AccountRepository
	entryLine repository.JournalEntryLineRepository
}

func NewJournalEntryService(
	repo repository.JournalEntryRepository,
	account repository.AccountRepository,
	entryLine repository.JournalEntryLineRepository,
) JournalEntryService {
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

func (s *journalEntryService) CreateJournalEntry(
	ctx context.Context,
	input CreateJournalEntryInput,
) (*models.JournalEntry, error) {
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

	validateLinesErr := validateLines(ctx, s.account, input.ClientID, lines)
	if validateLinesErr != nil {
		return nil, validateLinesErr
	}

	if dupErr := s.duplicateReference(ctx, input.ClientID, input.Reference, ""); dupErr != nil {
		return nil, dupErr
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

	journalEntry.TransactionDate = time.Now()
	if input.TransactionDate != nil {
		t, err := lib.ParseDateParam(*input.TransactionDate, false)
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
	if lib.IsUniqueViolation(err) {
		// A concurrent create with the same reference won the race.
		if dupErr := s.duplicateReference(ctx, input.ClientID, input.Reference, ""); dupErr != nil {
			return nil, dupErr
		}
	}
	if err != nil {
		return nil, err
	}

	return &journalEntry, nil
}

func validateLines(
	ctx context.Context,
	accountRepo repository.AccountRepository,
	clientID string,
	lines []models.JournalEntryLine,
) error {
	if len(lines) < 2 {
		return errors.New("a journal entry needs at least two lines")
	}

	debitTotal := int64(0)
	creditTotal := int64(0)

	for _, line := range lines {
		if line.Debit < 0 || line.Credit < 0 {
			return errors.New("debit and credit amounts cannot be negative")
		}

		if (line.Debit > 0) == (line.Credit > 0) {
			return errors.New("each line must have either a debit or a credit")
		}

		debitTotal += line.Debit
		creditTotal += line.Credit
	}

	if debitTotal == 0 {
		return errors.New("journal entry total must be greater than zero")
	}

	if debitTotal != creditTotal {
		return errors.New("debit and credit totals must be equal")
	}

	// accounts must belong to the client and accept postings
	for _, line := range lines {
		account, err := accountRepo.GetByIDAndClientID(ctx, line.AccountID, clientID, nil)
		if err != nil {
			return fmt.Errorf("account %s not found", line.AccountID)
		}

		if account.IsGroup {
			return fmt.Errorf("cannot post to group account %s", account.Code)
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

func (s *journalEntryService) UpdateJournalEntry(
	ctx context.Context,
	journalEntryId string,
	input UpdateJournalEntryInput,
) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, nil)
	if err != nil {
		return nil, err
	}

	if entry.Status == "POSTED" {
		return nil, errors.New("journal entry is already posted")
	}

	if input.Reference != nil && *input.Reference != entry.Reference {
		if dupErr := s.duplicateReference(ctx, input.ClientID, *input.Reference, entry.ID.String()); dupErr != nil {
			return nil, dupErr
		}

		entry.Reference = *input.Reference
	}

	if input.TransactionDate != nil {
		t, err := lib.ParseDateParam(*input.TransactionDate, false)
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
		validateLinesErr := validateLines(ctx, s.account, input.ClientID, lines)
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

func (s *journalEntryService) PostJournalEntry(
	ctx context.Context,
	input GetJournalEntryInput,
) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, &[]string{"JournalEntryLines"})
	if err != nil {
		return nil, err
	}

	if entry.Status == "POSTED" {
		return nil, errors.New("journal entry is already posted")
	}

	// Drafts may predate the line rules, and PATCH only checks the lines it
	// sends, so the whole entry is checked before it reaches the books.
	if validateErr := validateLines(ctx, s.account, input.ClientID, entry.JournalEntryLines); validateErr != nil {
		return nil, validateErr
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

func (s *journalEntryService) GetJournalEntry(
	ctx context.Context,
	input GetJournalEntryInput,
) (*models.JournalEntry, error) {
	entry, err := s.repo.GetByIDAndClientID(ctx, input.ID, input.ClientID, input.Populate)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *journalEntryService) ListJournalEntries(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters repository.ListJournalEntriesFilter,
) ([]models.JournalEntry, error) {
	entries, err := s.repo.List(ctx, filterQuery, filters)
	if err != nil {
		return nil, err
	}

	return *entries, nil
}

func (s *journalEntryService) CountJournalEntries(
	ctx context.Context,
	filterQuery lib.FilterQuery,
	filters repository.ListJournalEntriesFilter,
) (int64, error) {
	return s.repo.Count(ctx, filterQuery, filters)
}
