package models

import (
	"time"

	"gorm.io/datatypes"
)

type JournalEntry struct {
	BaseModelSoftDelete
	ClientID string `json:"client_id" gorm:"not null;index;"`
	Client   Client `json:"client" gorm:"foreignKey:ClientID;references:ID"`

	Status          string     `json:"status" gorm:"not null; index; default: POSTED;"` // DRAFT, POSTED
	PostedAt        *time.Time `json:"posted_at"`
	Reference       string     `json:"reference" gorm:"not null;"`
	TransactionDate time.Time  `json:"transaction_date" gorm:"not null;"`

	Metadata *datatypes.JSON `json:"metadata"` // save any client related data.

	JournalEntryLines []JournalEntryLine `json:"journal_entry_lines" gorm:"foreignKey:JournalEntryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
