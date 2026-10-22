package services

import (
	"github.com/Bendomey/fincore-engine/internal/repository"
)

type Services struct {
	ClientService       ClientService
	AccountService      AccountService
	JournalEntryService JournalEntryService
	ReportService       ReportService
}

func NewServices(repository repository.Repository) Services {
	clientService := NewClientService(repository.ClientRepository)
	accountService := NewAccountService(repository.AccountRepository)
	journalEntryService := NewJournalEntryService(
		repository.JournalEntryRepository,
		repository.AccountRepository,
		repository.JournalEntryLineRepository,
	)

	reportService := NewReportService(repository.LedgerRepository, repository.AccountRepository)

	return Services{
		ClientService:       clientService,
		AccountService:      accountService,
		JournalEntryService: journalEntryService,
		ReportService:       reportService,
	}
}
