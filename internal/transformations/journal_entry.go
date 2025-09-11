package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBJournalEntryToRestJournalEntry transforms journal_entry db input to rest type
func DBJournalEntryToRestJournalEntry(i *models.JournalEntry, populate *[]string) interface{} {
	if i == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":               i.ID.String(),
		"reference":        i.Reference,
		"status":           i.Status,
		"posted_at":        i.PostedAt,
		"transaction_date": i.TransactionDate,
		"metadata":         i.Metadata,
		"created_at":       i.CreatedAt,
		"updated_at":       i.UpdatedAt,
	}

	if len(i.JournalEntryLines) > 0 {
		lines := make([]interface{}, 0)
		for _, line := range i.JournalEntryLines {
			lines = append(lines, DBJournalEntryLineToRestJournalEntryLine(&line, populate))
		}
		data["lines"] = lines
	}

	return data
}
