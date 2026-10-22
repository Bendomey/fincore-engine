package services

import (
	"context"
	"errors"
	"time"

	"github.com/Bendomey/fincore-engine/internal/models"
	"github.com/Bendomey/fincore-engine/internal/repository"
)

// ErrLedgerAccountNotFound means the account is missing, another client's, or a group.
var ErrLedgerAccountNotFound = errors.New("account not found")

// NormalSide returns the side an account's balance grows on: assets and
// expenses grow with debits, everything else with credits; contra flips it.
func NormalSide(accountType string, isContra bool) string {
	debitNormal := accountType == "ASSET" || accountType == "EXPENSE"
	if isContra {
		debitNormal = !debitNormal
	}

	if debitNormal {
		return "DEBIT"
	}

	return "CREDIT"
}

// SideBalance converts a debit-minus-credit net into a balance on the given side.
func SideBalance(net int64, side string) int64 {
	if side == "DEBIT" {
		return net
	}

	return -net
}

type TrialBalanceAccount struct {
	repository.AccountTotalsRow
	NormalBalance string
	Balance       int64
}

type TrialBalance struct {
	AsOf        time.Time
	Accounts    []TrialBalanceAccount
	TotalDebit  int64
	TotalCredit int64
}

type LedgerInput struct {
	ClientID  string
	AccountID string
	Start     *time.Time
	End       time.Time
	Page      int
	PageSize  int
}

type LedgerLine struct {
	repository.LedgerLineRow
	RunningBalance int64
}

type AccountLedger struct {
	Account        *models.Account
	Start          *time.Time
	End            time.Time
	NormalBalance  string
	OpeningBalance int64
	ClosingBalance int64
	Lines          []LedgerLine
	Total          int64
}

type ReportService interface {
	TrialBalance(ctx context.Context, clientID string, asOf time.Time) (*TrialBalance, error)
	AccountLedger(ctx context.Context, input LedgerInput) (*AccountLedger, error)
}

type reportService struct {
	ledger  repository.LedgerRepository
	account repository.AccountRepository
}

func NewReportService(ledger repository.LedgerRepository, account repository.AccountRepository) ReportService {
	return &reportService{ledger, account}
}

func (s *reportService) TrialBalance(ctx context.Context, clientID string, asOf time.Time) (*TrialBalance, error) {
	rows, err := s.ledger.TrialBalance(ctx, clientID, asOf)
	if err != nil {
		return nil, err
	}

	report := &TrialBalance{AsOf: asOf, Accounts: make([]TrialBalanceAccount, 0, len(rows))}
	for _, row := range rows {
		side := NormalSide(row.Type, row.IsContra)
		report.Accounts = append(report.Accounts, TrialBalanceAccount{
			AccountTotalsRow: row,
			NormalBalance:    side,
			Balance:          SideBalance(row.TotalDebit-row.TotalCredit, side),
		})
		report.TotalDebit += row.TotalDebit
		report.TotalCredit += row.TotalCredit
	}

	return report, nil
}

func (s *reportService) AccountLedger(ctx context.Context, input LedgerInput) (*AccountLedger, error) {
	account, err := s.account.GetByIDAndClientID(ctx, input.AccountID, input.ClientID, nil)
	if err != nil || account.IsGroup {
		return nil, ErrLedgerAccountNotFound
	}

	side := NormalSide(account.Type, account.IsContra)

	opening := int64(0)
	if input.Start != nil {
		netBefore, netErr := s.ledger.NetBefore(ctx, input.ClientID, input.AccountID, *input.Start)
		if netErr != nil {
			return nil, netErr
		}
		opening = SideBalance(netBefore, side)
	}

	netInRange, err := s.ledger.NetBetween(ctx, input.ClientID, input.AccountID, input.Start, input.End)
	if err != nil {
		return nil, err
	}

	rows, total, err := s.ledger.Lines(
		ctx, input.ClientID, input.AccountID, input.Start, input.End, input.Page, input.PageSize,
	)
	if err != nil {
		return nil, err
	}

	lines := make([]LedgerLine, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, LedgerLine{
			LedgerLineRow:  row,
			RunningBalance: opening + SideBalance(row.RunningNet, side),
		})
	}

	return &AccountLedger{
		Account:        account,
		Start:          input.Start,
		End:            input.End,
		NormalBalance:  side,
		OpeningBalance: opening,
		ClosingBalance: opening + SideBalance(netInRange, side),
		Lines:          lines,
		Total:          total,
	}, nil
}
