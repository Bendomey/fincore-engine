package models

import (
	"errors"

	"gorm.io/gorm"
)

type Account struct {
	BaseModelSoftDelete
	ClientID string `json:"client_id" gorm:"not null;index;"`
	Client   Client `json:"client" gorm:"foreignKey:ClientID;references:ID"`

	Code        string  `json:"code" gorm:"not null;uniqueIndex;"`
	Name        string  `json:"name" gorm:"not null;"`
	Description *string `json:"description"`
	Type        string  `json:"type" gorm:"not null; index;"` // EXPENSE | LIABILITY | EQUITY | ASSET | INCOME
	IsContra    bool    `json:"is_contra" gorm:"not null;"`
	IsGroup     bool    `json:"is_group" gorm:"not null;default:false;index;"`

	ParentAccount   *Account `json:"parent_account" gorm:"foreignKey:ParentAccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ParentAccountID *string  `json:"parent_account_id"`

	JournalEntries []JournalEntry `json:"journal_entries" gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (acc *Account) BeforeDelete(tx *gorm.DB) (err error) {
	// Prevent deletion if the account is a group account and has child accounts
	if acc.IsGroup {
		var count int64
		tx.Model(&Account{}).Where("parent_account_id = ?", acc.ID).Count(&count)
		if count > 0 {
			return errors.New("cannot delete a group account that has child accounts")
		}
	}

	// prevent deletion if the account has associated journal entries
	var journalEntriesCount int64
	tx.Model(&JournalEntry{}).Where("account_id = ?", acc.ID).Count(&journalEntriesCount)

	if journalEntriesCount > 0 {
		return errors.New("cannot delete an account that has associated journal entries")
	}

	return
}
