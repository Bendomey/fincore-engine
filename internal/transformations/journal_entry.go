package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/models"
)

// DBJournalEntryToRestJournalEntry transforms journal_entry db input to rest type
func DBJournalEntryToRestJournalEntry(i *models.JournalEntry) interface{} {
	if i == nil {
		return nil
	}

	data := map[string]interface{}{
		"id":               i.ID.String(),
		"reference":        i.Reference,
		"transaction_id":   i.TransactionID,
		"transaction_date": i.TransactionDate,
		"total_debit":      i.TotalDebit,
		"total_credit":     i.TotalCredit,
		"metadata":         i.Metadata,
		"created_at":       i.CreatedAt,
		"updated_at":       i.UpdatedAt,
	}

	return data
}
