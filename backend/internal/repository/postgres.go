package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sqlx.DB
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error conectando a PostgreSQL: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error haciendo ping a PostgreSQL: %w", err)
	}

	fmt.Println("Conectado a PostgreSQL correctamente")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) CreateTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id         UUID PRIMARY KEY,
		email      VARCHAR(255) UNIQUE NOT NULL,
		password   VARCHAR(255) NOT NULL,
		full_name  VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS accounts (
		id         VARCHAR(20) PRIMARY KEY,
		user_id    UUID NOT NULL REFERENCES users(id),
		type       VARCHAR(20) NOT NULL,
		nickname   VARCHAR(100),
		balance    DECIMAL(15,2) DEFAULT 0,
		currency   VARCHAR(10) DEFAULT 'USD',
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		from_account VARCHAR(20),
		to_account   VARCHAR(20),
		amount       DECIMAL(15,2) NOT NULL,
		type         VARCHAR(30) NOT NULL,
		description  VARCHAR(255),
		status       VARCHAR(20) DEFAULT 'completed',
		timestamp    TIMESTAMP DEFAULT NOW()
	);`

	_, err := p.DB.Exec(query)
	if err != nil {
		return fmt.Errorf("error creando tablas: %w", err)
	}

	// Migración: agregar columna nickname si no existe
	alterQuery := `ALTER TABLE accounts ADD COLUMN IF NOT EXISTS nickname VARCHAR(100);`
	p.DB.Exec(alterQuery)

	// Migración: actualizar moneda a USD
	updateCurrencyQuery := `UPDATE accounts SET currency = 'USD' WHERE currency = 'HNL';`
	p.DB.Exec(updateCurrencyQuery)

	fmt.Println("Tablas creadas correctamente")
	return nil
}

// ---- Usuarios ----

func (p *PostgresDB) CreateUser(id, email, password, fullName string) error {
	query := `INSERT INTO users (id, email, password, full_name) VALUES ($1, $2, $3, $4)`
	_, err := p.DB.Exec(query, id, email, password, fullName)
	if err != nil {
		return fmt.Errorf("error creando usuario: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetUserByEmail(email string) (*UserRecord, error) {
	var user UserRecord
	query := `SELECT id, email, password, full_name, created_at FROM users WHERE email = $1`
	err := p.DB.Get(&user, query, email)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %w", err)
	}
	return &user, nil
}

func (p *PostgresDB) GetUserByID(id string) (*UserRecord, error) {
	var user UserRecord
	query := `SELECT id, email, password, full_name, created_at FROM users WHERE id = $1`
	err := p.DB.Get(&user, query, id)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %w", err)
	}
	return &user, nil
}

// ---- Cuentas ----

func (p *PostgresDB) CreateAccount(id, userID, accountType string) error {
	query := `INSERT INTO accounts (id, user_id, type) VALUES ($1, $2, $3)`
	_, err := p.DB.Exec(query, id, userID, accountType)
	if err != nil {
		return fmt.Errorf("error creando cuenta: %w", err)
	}
	return nil
}

func (p *PostgresDB) UpdateBalance(accountID string, amount float64) error {
	query := `UPDATE accounts SET balance = balance + $1 WHERE id = $2`
	_, err := p.DB.Exec(query, amount, accountID)
	if err != nil {
		return fmt.Errorf("error actualizando balance: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetAccountsByUserID(userID string) ([]AccountRecord, error) {
	var accounts []AccountRecord
	query := `SELECT id, user_id, type, nickname, balance, currency, created_at FROM accounts WHERE user_id = $1`
	err := p.DB.Select(&accounts, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo cuentas: %w", err)
	}
	return accounts, nil
}

func (p *PostgresDB) GetAccountByID(accountID string) (*AccountRecord, error) {
	var account AccountRecord
	query := `SELECT id, user_id, type, nickname, balance, currency, created_at FROM accounts WHERE id = $1`
	err := p.DB.Get(&account, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("cuenta no encontrada: %w", err)
	}
	return &account, nil
}

func (p *PostgresDB) DeleteAccount(accountID, userID string) error {
	// Verificar que el balance sea 0
	var balance float64
	query := `SELECT balance FROM accounts WHERE id = $1 AND user_id = $2`
	err := p.DB.Get(&balance, query, accountID, userID)
	if err != nil {
		return fmt.Errorf("cuenta no encontrada")
	}

	if balance != 0 {
		return fmt.Errorf("la cuenta debe tener saldo en cero para poder eliminarla")
	}

	deleteQuery := `DELETE FROM accounts WHERE id = $1 AND user_id = $2`
	result, err := p.DB.Exec(deleteQuery, accountID, userID)
	if err != nil {
		return fmt.Errorf("error eliminando cuenta: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("cuenta no encontrada")
	}

	return nil
}

func (p *PostgresDB) UpdateAccountNickname(accountID, userID, nickname string) error {
	query := `UPDATE accounts SET nickname = $1 WHERE id = $2 AND user_id = $3`
	result, err := p.DB.Exec(query, nickname, accountID, userID)
	if err != nil {
		return fmt.Errorf("error actualizando nombre: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("cuenta no encontrada o no pertenece al usuario")
	}
	return nil
}

// ---- Transacciones ----

func (p *PostgresDB) CreateTransaction(fromAccount, toAccount string, amount float64, txType, description string) error {
	query := `
	INSERT INTO transactions (from_account, to_account, amount, type, description)
	VALUES ($1, $2, $3, $4, $5)`
	_, err := p.DB.Exec(query, fromAccount, toAccount, amount, txType, description)
	if err != nil {
		return fmt.Errorf("error creando transacción: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetTransactionsByAccount(accountID string, limit int) ([]TransactionRecord, error) {
	var txs []TransactionRecord
	query := `
	SELECT id, from_account, to_account, amount, type, description, status, timestamp
	FROM transactions
	WHERE from_account = $1 OR to_account = $1
	ORDER BY timestamp DESC
	LIMIT $2`
	err := p.DB.Select(&txs, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo transacciones: %w", err)
	}
	return txs, nil
}

func (p *PostgresDB) GetTransactionsByUserID(userID string, limit int) ([]TransactionRecord, error) {
	var txs []TransactionRecord
	query := `
	SELECT t.id, t.from_account, t.to_account, t.amount, t.type, t.description, t.status, t.timestamp
	FROM transactions t
	WHERE t.from_account IN (SELECT id FROM accounts WHERE user_id = $1)
	   OR t.to_account IN (SELECT id FROM accounts WHERE user_id = $1)
	ORDER BY t.timestamp DESC
	LIMIT $2`
	err := p.DB.Select(&txs, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo transacciones: %w", err)
	}
	return txs, nil
}

func (p *PostgresDB) GetTransactionsByAccountID(accountID string, limit int) ([]TransactionRecord, error) {
	var txs []TransactionRecord
	query := `
	SELECT id, from_account, to_account, amount, type, description, status, timestamp
	FROM transactions
	WHERE from_account = $1 OR to_account = $1
	ORDER BY timestamp DESC
	LIMIT $2`
	err := p.DB.Select(&txs, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo transacciones: %w", err)
	}
	return txs, nil
}

// ---- Structs internos para leer de la DB ----

type UserRecord struct {
	ID        string `db:"id" json:"id"`
	Email     string `db:"email" json:"email"`
	Password  string `db:"password" json:"-"`
	FullName  string `db:"full_name" json:"full_name"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

type AccountRecord struct {
	ID        string  `db:"id" json:"id"`
	UserID    string  `db:"user_id" json:"user_id"`
	Type      string  `db:"type" json:"type"`
	Nickname  *string `db:"nickname" json:"nickname"`
	Balance   float64 `db:"balance" json:"balance"`
	Currency  string  `db:"currency" json:"currency"`
	CreatedAt string  `db:"created_at" json:"created_at"`
}

type TransactionRecord struct {
	ID          string  `db:"id" json:"id"`
	FromAccount string  `db:"from_account" json:"from_account"`
	ToAccount   string  `db:"to_account" json:"to_account"`
	Amount      float64 `db:"amount" json:"amount"`
	Type        string  `db:"type" json:"type"`
	Description string  `db:"description" json:"description"`
	Status      string  `db:"status" json:"status"`
	Timestamp   string  `db:"timestamp" json:"timestamp"`
}