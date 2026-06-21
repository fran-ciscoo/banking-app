package models

import "time"

type Account struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Type      string    `db:"type" json:"type"` // "checking" o "savings"
	Balance   float64   `db:"balance" json:"balance"`
	Currency  string    `db:"currency" json:"currency"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Transaction struct {
	ID          string    `db:"id" json:"id"`
	FromAccount string    `db:"from_account" json:"from_account"`
	ToAccount   string    `db:"to_account" json:"to_account"`
	Amount      float64   `db:"amount" json:"amount"`
	Type        string    `db:"type" json:"type"`
	Description string    `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
}

type TransactionRequest struct {
	AccountID   string  `json:"account_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
}