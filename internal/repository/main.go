package repository

import "gorm.io/gorm"

type Repository struct {
	ClientRepository           ClientRepository
	AccountRepository          AccountRepository
	JournalEntryRepository     JournalEntryRepository
	JournalEntryLineRepository JournalEntryLineRepository
	LedgerRepository           LedgerRepository
}

func NewRepository(db *gorm.DB) Repository {
	clientRepository := NewClientRepository(db)
	accountRepository := NewAccountRepository(db)
	journalEntryRepository := NewJournalEntryRepository(db)
	journalEntryLineRepository := NewJournalEntryLineRepository(db)
	ledgerRepository := NewLedgerRepository(db)

	return Repository{
		ClientRepository:           clientRepository,
		AccountRepository:          accountRepository,
		JournalEntryRepository:     journalEntryRepository,
		JournalEntryLineRepository: journalEntryLineRepository,
		LedgerRepository:           ledgerRepository,
	}
}
