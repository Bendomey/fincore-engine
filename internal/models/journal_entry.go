package models

import (
	"time"

	"gorm.io/datatypes"
)

type JournalEntry struct {
	BaseModelSoftDelete
	ClientID string `json:"client_id" gorm:"not null;index;"`
	Client   Client `json:"client" gorm:"foreignKey:ClientID;references:ID"`

	Reference       string    `json:"reference" gorm:"not null;"`
	TransactionID   *string   `json:"transaction_id"`
	TransactionDate time.Time `json:"transaction_date" gorm:"not null;"`

	TotalDebit  int64 `json:"total_debit" gorm:"not null; default: 0"`
	TotalCredit int64 `json:"total_credit" gorm:"not null; default: 0"`

	Metadata *datatypes.JSON `json:"metadata"` // save any client related data.

	JournalEntryLines []JournalEntryLine `json:"journal_entry_lines" gorm:"foreignKey:JournalEntryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
