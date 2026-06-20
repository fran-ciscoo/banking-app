package internal

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(databaseURL string) (*Repository, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error haciendo ping: %w", err)
	}
	return &Repository{DB: db}, nil
}

type Account struct {
	ID       string  `db:"id" json:"id"`
	UserID   string  `db:"user_id" json:"user_id"`
	Type     string  `db:"type" json:"type"`
	Nickname *string `db:"nickname" json:"nickname"`
	Balance  float64 `db:"balance" json:"balance"`
	Currency string  `db:"currency" json:"currency"`
}

func (r *Repository) GetAccountsByUserID(userID string) ([]Account, error) {
	var accounts []Account
	query := `SELECT id, user_id, type, nickname, balance, currency FROM accounts WHERE user_id = $1`
	err := r.DB.Select(&accounts, query, userID)
	return accounts, err
}

func (r *Repository) GetAccountByID(accountID string) (*Account, error) {
	var account Account
	query := `SELECT id, user_id, type, nickname, balance, currency FROM accounts WHERE id = $1`
	err := r.DB.Get(&account, query, accountID)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) UpdateBalance(accountID string, amount float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE id = $2`
	_, err := r.DB.Exec(query, amount, accountID)
	return err
}

func (r *Repository) CreateTransaction(fromAccount, toAccount string, amount float64, txType, description string) error {
	query := `
	INSERT INTO transactions (from_account, to_account, amount, type, description)
	VALUES ($1, $2, $3, $4, $5)`
	_, err := r.DB.Exec(query, fromAccount, toAccount, amount, txType, description)
	return err
}

type Transaction struct {
	ID          string  `db:"id" json:"id"`
	FromAccount string  `db:"from_account" json:"from_account"`
	ToAccount   string  `db:"to_account" json:"to_account"`
	Amount      float64 `db:"amount" json:"amount"`
	Type        string  `db:"type" json:"type"`
	Description string  `db:"description" json:"description"`
	Timestamp   string  `db:"timestamp" json:"timestamp"`
}

func (r *Repository) GetTransactionsByAccount(accountID string, limit int) ([]Transaction, error) {
	var txs []Transaction
	query := `
	SELECT id, from_account, to_account, amount, type, description, timestamp
	FROM transactions
	WHERE from_account = $1 OR to_account = $1
	ORDER BY timestamp DESC
	LIMIT $2`
	err := r.DB.Select(&txs, query, accountID, limit)
	return txs, err
}