package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBJournalEntryLineToRestJournalEntryLine transforms journal_entry_line db input to rest type
func DBJournalEntryLineToRestJournalEntryLine(i *models.JournalEntryLine, populate *[]string) interface{} {
	if i == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":               i.ID.String(),
		"journal_entry_id": i.JournalEntryID,
		"journal_entry":    DBJournalEntryToRestJournalEntry(&i.JournalEntry),
		"account_id":       i.AccountID,
		"account":          DBAccountToRestAccount(&i.Account, populate),
		"debit":            i.Debit,
		"credit":           i.Credit,
		"notes":            i.Notes,
		"created_at":       i.CreatedAt,
		"updated_at":       i.UpdatedAt,
	}

	return data
}
