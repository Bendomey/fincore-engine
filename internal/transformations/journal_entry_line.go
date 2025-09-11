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
		"account_id":       i.AccountID,
		"debit":            i.Debit,
		"credit":           i.Credit,
		"notes":            i.Notes,
		"created_at":       i.CreatedAt,
		"updated_at":       i.UpdatedAt,
	}

	if populate != nil {
		for _, field := range *populate {
			switch field {
			case "Account":
				data["account"] = DBAccountToRestAccount(&i.Account, populate)
			case "JournalEntry":
				data["journal_entry"] = DBJournalEntryToRestJournalEntry(&i.JournalEntry, populate)
			}
		}
	}

	return data
}
