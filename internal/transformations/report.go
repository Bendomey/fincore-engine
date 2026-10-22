package transformations

import (
	"github.com/Bendomey/fincore-engine/internal/services"
)

// DBTrialBalanceToRestTrialBalance transforms a trial balance to rest type
func DBTrialBalanceToRestTrialBalance(i *services.TrialBalance) interface{} {
	accounts := make([]interface{}, 0, len(i.Accounts))
	for _, a := range i.Accounts {
		accounts = append(accounts, map[string]interface{}{
			"account_id":        a.AccountID,
			"code":              a.Code,
			"name":              a.Name,
			"type":              a.Type,
			"is_contra":         a.IsContra,
			"parent_account_id": a.ParentAccountID,
			"normal_balance":    a.NormalBalance,
			"total_debit":       a.TotalDebit,
			"total_credit":      a.TotalCredit,
			"balance":           a.Balance,
		})
	}

	return map[string]interface{}{
		"as_of":        i.AsOf,
		"accounts":     accounts,
		"total_debit":  i.TotalDebit,
		"total_credit": i.TotalCredit,
	}
}

// DBAccountLedgerToRestAccountLedger transforms an account ledger to rest type
func DBAccountLedgerToRestAccountLedger(i *services.AccountLedger) interface{} {
	lines := make([]interface{}, 0, len(i.Lines))
	for _, l := range i.Lines {
		lines = append(lines, map[string]interface{}{
			"line_id":          l.LineID,
			"journal_entry_id": l.JournalEntryID,
			"reference":        l.Reference,
			"transaction_date": l.TransactionDate,
			"notes":            l.Notes,
			"debit":            l.Debit,
			"credit":           l.Credit,
			"running_balance":  l.RunningBalance,
		})
	}

	return map[string]interface{}{
		"account":         DBAccountToRestAccount(i.Account, nil),
		"start_date":      i.Start,
		"end_date":        i.End,
		"normal_balance":  i.NormalBalance,
		"opening_balance": i.OpeningBalance,
		"closing_balance": i.ClosingBalance,
		"lines":           lines,
	}
}
