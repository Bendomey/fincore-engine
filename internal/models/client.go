package models

type Client struct {
	BaseModelSoftDelete
	Name             string `json:"name"               gorm:"not null;index"`
	Email            string `json:"email"              gorm:"not null;uniqueIndex"`
	ClientId         string `json:"client_id"          gorm:"not null;uniqueIndex;"`
	ClientSecretHash string `json:"client_secret_hash"`

	Accounts []Account
}
