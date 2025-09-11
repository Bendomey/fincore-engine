package models

type JournalEntryLine struct {
	BaseModelSoftDelete
	JournalEntryID string       `json:"journal_entry_id" gorm:"not null;index;"`
	JournalEntry   JournalEntry `json:"journal_entry" gorm:"foreignKey:JournalEntryID;references:ID"`

	AccountID string  `json:"account_id" gorm:"not null;index;"`
	Account   Account `json:"account" gorm:"foreignKey:AccountID;references:ID"`

	Notes  *string `json:"notes"`
	Debit  int64   `json:"debit" gorm:"not null; default: 0"`
	Credit int64   `json:"credit" gorm:"not null; default: 0"`
}
