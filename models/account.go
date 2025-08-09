package models

type Account struct {
	AccountID      int     `json:"account_id"`
	CurrentBalance float64 `json:"balance"`
}
