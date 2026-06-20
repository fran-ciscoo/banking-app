package internal

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type TigerBeetleDB struct {
	Client tb.Client
}

func NewTigerBeetleDB(address string) (*TigerBeetleDB, error) {
	var client tb.Client
	var err error

	for i := 0; i < 10; i++ {
		client, err = tb.NewClient(tb.ToUint128(0), []string{address})
		if err == nil {
			fmt.Println("MCP Server conectado a TigerBeetle correctamente")
			return &TigerBeetleDB{Client: client}, nil
		}
		fmt.Printf("Intento %d/10 conectando a TigerBeetle... reintentando en 2s\n", i+1)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("error conectando a TigerBeetle tras varios intentos: %w", err)
}

func (t *TigerBeetleDB) Deposit(toAccountID uint64, amount uint64, transferID uint64) error {
	transfers := []tb.Transfer{
		{
			ID:              tb.ToUint128(transferID),
			DebitAccountID:  tb.ToUint128(1),
			CreditAccountID: tb.ToUint128(toAccountID),
			Amount:          tb.ToUint128(amount),
			Ledger:          1,
			Code:            1,
		},
	}

	results, err := t.Client.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("error en deposito: %w", err)
	}

	for _, r := range results {
		if r.Status != tb.TransferCreated {
			return fmt.Errorf("error en deposito: %v", r.Status)
		}
	}

	return nil
}

func (t *TigerBeetleDB) Withdraw(fromAccountID uint64, amount uint64, transferID uint64) error {
	transfers := []tb.Transfer{
		{
			ID:              tb.ToUint128(transferID),
			DebitAccountID:  tb.ToUint128(fromAccountID),
			CreditAccountID: tb.ToUint128(1),
			Amount:          tb.ToUint128(amount),
			Ledger:          1,
			Code:            1,
		},
	}

	results, err := t.Client.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("error en retiro: %w", err)
	}

	for _, r := range results {
		if r.Status != tb.TransferCreated {
			return fmt.Errorf("error en retiro: %v", r.Status)
		}
	}

	return nil
}

func (t *TigerBeetleDB) Transfer(fromAccountID, toAccountID, amount, transferID uint64) error {
	transfers := []tb.Transfer{
		{
			ID:              tb.ToUint128(transferID),
			DebitAccountID:  tb.ToUint128(fromAccountID),
			CreditAccountID: tb.ToUint128(toAccountID),
			Amount:          tb.ToUint128(amount),
			Ledger:          1,
			Code:            1,
		},
	}

	results, err := t.Client.CreateTransfers(transfers)
	if err != nil {
		return fmt.Errorf("error en transferencia: %w", err)
	}

	for _, r := range results {
		if r.Status != tb.TransferCreated {
			return fmt.Errorf("error en transferencia: %v", r.Status)
		}
	}

	return nil
}

// AccountIDFromString convierte el ID de cuenta (string) en un uint64 determinístico,
// igual que en el backend, para que ambos generen el mismo ID de cuenta en TigerBeetle.
func AccountIDFromString(accountID string) uint64 {
	hash := sha256.Sum256([]byte(accountID))
	return binary.BigEndian.Uint64(hash[:8])
}