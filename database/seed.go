// seed.go — Script de carga de datos de prueba.
//
// Carga una muestra del JSON de datos de prueba directamente en
// PostgreSQL (usuarios, metadatos de cuentas) y TigerBeetle
// (balances iniciales y transferencias), usando la misma lógica
// de IDs determinísticos que usa el backend principal.
//
// Uso:
//   cd backend
//   go run ../database/seed.go
//
// Requiere que docker compose esté corriendo (Postgres y TigerBeetle
// accesibles en localhost:5432 y localhost:3000).
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type SeedUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FullName  string `json:"full_name"`
	CreatedAt string `json:"created_at"`
}

type SeedAccount struct {
	AccountNumber  string  `json:"account_number"`
	UserID         string  `json:"user_id"`
	InitialBalance float64 `json:"initial_balance"`
	Currency       string  `json:"currency"`
	AccountType    string  `json:"account_type"`
}

type SeedTransaction struct {
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Timestamp   string  `json:"timestamp"`
	Status      string  `json:"status"`
}

type SeedData struct {
	Users        []SeedUser        `json:"users"`
	Accounts     []SeedAccount     `json:"accounts"`
	Transactions []SeedTransaction `json:"transactions"`
}

func accountIDFromString(accountID string) uint64 {
	hash := sha256.Sum256([]byte(accountID))
	return binary.BigEndian.Uint64(hash[:8])
}

// normalizeAccountType mapea tipos del dataset a los soportados por el sistema.
// "investment" no es un tipo soportado en el dashboard actual, se trata como "savings".
func normalizeAccountType(t string) string {
	if t == "checking" || t == "savings" {
		return t
	}
	return "savings"
}

func main() {
	dataPath := "database/sample-data.json"
	if len(os.Args) > 1 {
		dataPath = os.Args[1]
	}

	raw, err := os.ReadFile(dataPath)
	if err != nil {
		log.Fatalf("Error leyendo %s: %v", dataPath, err)
	}

	var data SeedData
	if err := json.Unmarshal(raw, &data); err != nil {
		log.Fatalf("Error parseando JSON: %v", err)
	}

	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/banking?sslmode=disable")
	tigerbeetleAddr := getEnv("TIGERBEETLE_ADDR", "localhost:3000")

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()

	tbClient, err := tb.NewClient(tb.ToUint128(0), []string{tigerbeetleAddr})
	if err != nil {
		log.Fatalf("Error conectando a TigerBeetle: %v", err)
	}
	defer tbClient.Close()

	fmt.Printf("Cargando %d usuarios, %d cuentas, %d transacciones...\n",
		len(data.Users), len(data.Accounts), len(data.Transactions))

	// ---- 1. Usuarios ----
	for _, u := range data.Users {
		hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("ADVERTENCIA: error hasheando password de %s: %v", u.Email, err)
			continue
		}

		_, err = db.Exec(
			`INSERT INTO users (id, email, password, full_name, created_at)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO NOTHING`,
			u.ID, u.Email, string(hashed), u.FullName, u.CreatedAt,
		)
		if err != nil {
			log.Printf("ADVERTENCIA: error insertando usuario %s: %v", u.Email, err)
			continue
		}
		fmt.Printf("  Usuario creado: %s (%s)\n", u.FullName, u.Email)
	}

	// ---- 2. Cuentas (Postgres + TigerBeetle) ----
	var tbAccounts []tb.Account
	accountBalances := map[string]float64{}

	for _, a := range data.Accounts {
		accType := normalizeAccountType(a.AccountType)

		_, err := db.Exec(
			`INSERT INTO accounts (id, user_id, type, balance, currency)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO NOTHING`,
			a.AccountNumber, a.UserID, accType, a.InitialBalance, a.Currency,
		)
		if err != nil {
			log.Printf("ADVERTENCIA: error insertando cuenta %s: %v", a.AccountNumber, err)
			continue
		}

		tbID := accountIDFromString(a.AccountNumber)
		tbAccounts = append(tbAccounts, tb.Account{
			ID:     tb.ToUint128(tbID),
			Ledger: 1,
			Code:   1,
		})
		accountBalances[a.AccountNumber] = a.InitialBalance

		fmt.Printf("  Cuenta creada: %s (%s) saldo inicial $%.2f\n", a.AccountNumber, accType, a.InitialBalance)
	}

	if len(tbAccounts) > 0 {
		results, err := tbClient.CreateAccounts(tbAccounts)
		if err != nil {
			log.Fatalf("Error creando cuentas en TigerBeetle: %v", err)
		}
		for _, r := range results {
			if r.Status != tb.AccountCreated && r.Status != tb.AccountExists {
				log.Printf("ADVERTENCIA: cuenta TigerBeetle con status inesperado: %v", r.Status)
			}
		}
	}

	// ---- 3. Establecer saldo inicial vía depósito desde EXTERNAL (cuenta 1) ----
	var initialDeposits []tb.Transfer
	depositCounter := uint64(900000000)
	for accountNumber, balance := range accountBalances {
		if balance <= 0 {
			continue
		}
		tbID := accountIDFromString(accountNumber)
		depositCounter++
		initialDeposits = append(initialDeposits, tb.Transfer{
			ID:              tb.ToUint128(depositCounter),
			DebitAccountID:  tb.ToUint128(1),
			CreditAccountID: tb.ToUint128(tbID),
			Amount:          tb.ToUint128(uint64(balance * 100)),
			Ledger:          1,
			Code:            1,
		})
	}

	if len(initialDeposits) > 0 {
		results, err := tbClient.CreateTransfers(initialDeposits)
		if err != nil {
			log.Fatalf("Error estableciendo saldos iniciales: %v", err)
		}
		for _, r := range results {
			if r.Status != tb.TransferCreated {
				log.Printf("ADVERTENCIA: saldo inicial con status inesperado: %v", r.Status)
			}
		}
		fmt.Printf("  Saldos iniciales establecidos para %d cuentas\n", len(initialDeposits))
	}

	// ---- 4. Transacciones de ejemplo (transferencias entre cuentas) ----
	var tbTransfers []tb.Transfer
	transferCounter := uint64(800000000)

	for _, t := range data.Transactions {
		fromTbID := accountIDFromString(t.FromAccount)
		toTbID := accountIDFromString(t.ToAccount)
		transferCounter++

		tbTransfers = append(tbTransfers, tb.Transfer{
			ID:              tb.ToUint128(transferCounter),
			DebitAccountID:  tb.ToUint128(fromTbID),
			CreditAccountID: tb.ToUint128(toTbID),
			Amount:          tb.ToUint128(uint64(t.Amount * 100)),
			Ledger:          1,
			Code:            1,
		})

		_, err := db.Exec(
			`INSERT INTO transactions (from_account, to_account, amount, type, description, status, timestamp)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			t.FromAccount, t.ToAccount, t.Amount, "transfer", t.Description, t.Status, t.Timestamp,
		)
		if err != nil {
			log.Printf("ADVERTENCIA: error insertando transacción: %v", err)
		}
	}

	if len(tbTransfers) > 0 {
		results, err := tbClient.CreateTransfers(tbTransfers)
		if err != nil {
			log.Fatalf("Error registrando transacciones en TigerBeetle: %v", err)
		}
		for _, r := range results {
			if r.Status != tb.TransferCreated {
				log.Printf("ADVERTENCIA: transferencia con status inesperado: %v", r.Status)
			}
		}
		fmt.Printf("  %d transacciones registradas en TigerBeetle\n", len(tbTransfers))
	}

	// ---- 5. Actualizar balance espejo en Postgres ----
	for _, t := range data.Transactions {
		db.Exec(`UPDATE accounts SET balance = balance - $1 WHERE id = $2`, t.Amount, t.FromAccount)
		db.Exec(`UPDATE accounts SET balance = balance + $1 WHERE id = $2`, t.Amount, t.ToAccount)
	}

	fmt.Println("\nCarga de datos de prueba completada.")
	fmt.Println("Puedes iniciar sesión con cualquiera de los usuarios cargados, por ejemplo:")
	if len(data.Users) > 0 {
		fmt.Printf("  Email: %s\n  Password: %s\n", data.Users[0].Email, data.Users[0].Password)
	}

	_ = time.Now
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}