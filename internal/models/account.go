package models

type Account struct {
	BaseModelSoftDelete
	ClientID string `json:"client_id" gorm:"not null;index;"`
	Client   Client `json:"client" gorm:"foreignKey:ClientID;references:ID"`

	Code        string  `json:"code" gorm:"not null;uniqueIndex;"`
	Name        string  `json:"name" gorm:"not null;"`
	Description *string `json:"description"`
	Type        string  `json:"type" gorm:"not null; index;"` // EXPENSE | LIABILITY | EQUITY | ASSET | INCOME
	IsContra    bool    `json:"is_contra" gorm:"not null;"`

	ParentAccount   *Account `json:"parent_account" gorm:"foreignKey:ParentAccountID;references:ID"`
	ParentAccountID *string  `json:"parent_account_id"`
}
