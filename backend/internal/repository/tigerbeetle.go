package repository

import (
	"fmt"
	"math/big"

	tb "github.com/tigerbeetle/tigerbeetle-go"
)

type TigerBeetleDB struct {
	Client tb.Client
}

func NewTigerBeetleDB(address string) (*TigerBeetleDB, error) {
	client, err := tb.NewClient(tb.ToUint128(0), []string{address})
	if err != nil {
		return nil, fmt.Errorf("error conectando a TigerBeetle: %w", err)
	}

	fmt.Println("Conectado a TigerBeetle correctamente")
	return &TigerBeetleDB{Client: client}, nil
}

func (t *TigerBeetleDB) Close() {
	t.Client.Close()
}

func (t *TigerBeetleDB) CreateAccount(id uint64) error {
	accounts := []tb.Account{
		{
			ID:     tb.ToUint128(id),
			Ledger: 1,
			Code:   1,
		},
	}

	results, err := t.Client.CreateAccounts(accounts)
	if err != nil {
		return fmt.Errorf("error creando cuenta: %w", err)
	}

	if len(results) > 0 {
		return fmt.Errorf("error creando cuenta: %v", results[0].Status)
	}

	return nil
}

func (t *TigerBeetleDB) GetBalance(id uint64) (uint64, error) {
	ids := []tb.Uint128{tb.ToUint128(id)}

	accounts, err := t.Client.LookupAccounts(ids)
	if err != nil {
		return 0, fmt.Errorf("error obteniendo cuenta: %w", err)
	}

	if len(accounts) == 0 {
		return 0, fmt.Errorf("cuenta no encontrada")
	}

	creditsBytes := accounts[0].CreditsPosted.Bytes()
	debitsBytes  := accounts[0].DebitsPosted.Bytes()

	credits := new(big.Int).SetBytes(creditsBytes[:])
	debits  := new(big.Int).SetBytes(debitsBytes[:])
	balance := new(big.Int).Sub(credits, debits)

	return balance.Uint64(), nil
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

	if len(results) > 0 {
		return fmt.Errorf("error en deposito: %v", results[0].Status)
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

	if len(results) > 0 {
		return fmt.Errorf("error en retiro: %v", results[0].Status)
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

	if len(results) > 0 {
		return fmt.Errorf("error en transferencia: %v", results[0].Status)
	}

	return nil
}