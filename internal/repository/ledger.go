package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AccountTotalsRow is one account's posted debit and credit totals.
type AccountTotalsRow struct {
	AccountID       string
	Code            string
	Name            string
	Type            string
	IsContra        bool
	ParentAccountID *string
	TotalDebit      int64
	TotalCredit     int64
}

// LedgerLineRow is one posted line on an account. RunningNet is the
// cumulative debit-credit over the requested range up to and including it.
type LedgerLineRow struct {
	LineID          string
	JournalEntryID  string
	Reference       string
	TransactionDate time.Time
	Notes           *string
	Debit           int64
	Credit          int64
	RunningNet      int64
}

// LedgerRepository reads balances and lines from posted journal entries only.
type LedgerRepository interface {
	TrialBalance(ctx context.Context, clientID string, asOf time.Time) ([]AccountTotalsRow, error)
	NetBefore(ctx context.Context, clientID string, accountID string, before time.Time) (int64, error)
	NetBetween(ctx context.Context, clientID string, accountID string, start *time.Time, end time.Time) (int64, error)
	Lines(
		ctx context.Context,
		clientID string,
		accountID string,
		start *time.Time,
		end time.Time,
		page int,
		pageSize int,
	) ([]LedgerLineRow, int64, error)
}

type ledgerRepository struct {
	DB *gorm.DB
}

func NewLedgerRepository(DB *gorm.DB) LedgerRepository {
	return &ledgerRepository{DB}
}

// postedLines selects the live lines of the client's posted entries.
const postedLines = `
FROM journal_entry_lines l
JOIN journal_entries e ON e.id = l.journal_entry_id
WHERE e.client_id = @client
  AND e.status = 'POSTED'
  AND e.deleted_at IS NULL
  AND l.deleted_at IS NULL`

func (r *ledgerRepository) TrialBalance(
	ctx context.Context,
	clientID string,
	asOf time.Time,
) ([]AccountTotalsRow, error) {
	rows := make([]AccountTotalsRow, 0)
	result := r.DB.WithContext(ctx).Raw(`
SELECT a.id AS account_id, a.code, a.name, a.type, a.is_contra, a.parent_account_id,
       COALESCE(t.total_debit, 0) AS total_debit, COALESCE(t.total_credit, 0) AS total_credit
FROM accounts a
LEFT JOIN (
  SELECT l.account_id, SUM(l.debit) AS total_debit, SUM(l.credit) AS total_credit
  `+postedLines+`
    AND e.transaction_date < @as_of
  GROUP BY l.account_id
) t ON t.account_id = a.id
WHERE a.client_id = @client AND a.is_group = false AND a.deleted_at IS NULL
ORDER BY a.code`,
		map[string]any{"client": clientID, "as_of": asOf},
	).Scan(&rows)

	return rows, result.Error
}

func (r *ledgerRepository) NetBefore(
	ctx context.Context,
	clientID string,
	accountID string,
	before time.Time,
) (int64, error) {
	var net int64
	result := r.DB.WithContext(ctx).Raw(`
SELECT COALESCE(SUM(l.debit - l.credit), 0)
`+postedLines+`
  AND l.account_id = @account
  AND e.transaction_date < @before`,
		map[string]any{"client": clientID, "account": accountID, "before": before},
	).Scan(&net)

	return net, result.Error
}

func (r *ledgerRepository) NetBetween(
	ctx context.Context,
	clientID string,
	accountID string,
	start *time.Time,
	end time.Time,
) (int64, error) {
	var net int64
	result := r.DB.WithContext(ctx).Raw(`
SELECT COALESCE(SUM(l.debit - l.credit), 0)
`+postedLines+`
  AND l.account_id = @account
  AND (CAST(@start AS timestamptz) IS NULL OR e.transaction_date >= @start)
  AND e.transaction_date < @end`,
		map[string]any{"client": clientID, "account": accountID, "start": start, "end": end},
	).Scan(&net)

	return net, result.Error
}

func (r *ledgerRepository) Lines(
	ctx context.Context,
	clientID string,
	accountID string,
	start *time.Time,
	end time.Time,
	page int,
	pageSize int,
) ([]LedgerLineRow, int64, error) {
	params := map[string]any{
		"client":  clientID,
		"account": accountID,
		"start":   start,
		"end":     end,
		"limit":   pageSize,
		"offset":  (page - 1) * pageSize,
	}

	inRange := postedLines + `
  AND l.account_id = @account
  AND (CAST(@start AS timestamptz) IS NULL OR e.transaction_date >= @start)
  AND e.transaction_date < @end`

	var total int64
	if err := r.DB.WithContext(ctx).Raw("SELECT COUNT(*) "+inRange, params).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// The running sum is taken over the whole range before paginating, so
	// later pages continue from where earlier pages ended.
	rows := make([]LedgerLineRow, 0)
	result := r.DB.WithContext(ctx).Raw(`
SELECT * FROM (
  SELECT l.id AS line_id, e.id AS journal_entry_id, e.reference, e.transaction_date, l.notes,
         l.debit, l.credit,
         SUM(l.debit - l.credit) OVER (ORDER BY e.transaction_date, e.created_at, l.id) AS running_net,
         e.created_at AS entry_created_at
  `+inRange+`
) ranged
ORDER BY transaction_date, entry_created_at, line_id
LIMIT @limit OFFSET @offset`, params).Scan(&rows)

	return rows, total, result.Error
}
