package models

type Client struct {
	BaseModelSoftDelete
	Name             string `json:"name"`
	Email            string `json:"email"`
	ClientID         string `json:"client_id"`
	ClientSecretHash string `json:"client_secret_hash"`
}
