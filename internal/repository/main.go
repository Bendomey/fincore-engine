package repository

import "gorm.io/gorm"

type Repository struct {
	ClientRepository  ClientRepository
	AccountRepository AccountRepository
}

func NewRepository(db *gorm.DB) Repository {

	clientRepository := NewClientRepository(db)
	accountRepository := NewAccountRepository(db)

	return Repository{ClientRepository: clientRepository, AccountRepository: accountRepository}
}
