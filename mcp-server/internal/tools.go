package internal

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tools struct {
	Repo *Repository
	TbDB *TigerBeetleDB
}

func NewTools(repo *Repository, tbDB *TigerBeetleDB) *Tools {
	return &Tools{Repo: repo, TbDB: tbDB}
}

// ---- get_balance ----

type GetBalanceInput struct {
	UserID string `json:"user_id" jsonschema:"el ID del usuario"`
}

type GetBalanceOutput struct {
	Accounts []Account `json:"accounts"`
}

func (t *Tools) GetBalance(ctx context.Context, req *mcp.CallToolRequest, input GetBalanceInput) (*mcp.CallToolResult, GetBalanceOutput, error) {
	accounts, err := t.Repo.GetAccountsByUserID(input.UserID)
	if err != nil {
		return nil, GetBalanceOutput{}, fmt.Errorf("error obteniendo cuentas: %w", err)
	}
	return nil, GetBalanceOutput{Accounts: accounts}, nil
}

// ---- get_history ----

type GetHistoryInput struct {
	AccountID string `json:"account_id" jsonschema:"el ID de la cuenta"`
	Limit     int    `json:"limit" jsonschema:"cantidad de transacciones a mostrar"`
}

type GetHistoryOutput struct {
	Transactions []Transaction `json:"transactions"`
}

func (t *Tools) GetHistory(ctx context.Context, req *mcp.CallToolRequest, input GetHistoryInput) (*mcp.CallToolResult, GetHistoryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	txs, err := t.Repo.GetTransactionsByAccount(input.AccountID, limit)
	if err != nil {
		return nil, GetHistoryOutput{}, fmt.Errorf("error obteniendo historial: %w", err)
	}
	return nil, GetHistoryOutput{Transactions: txs}, nil
}

// ---- deposit ----

type DepositInput struct {
	AccountID   string  `json:"account_id" jsonschema:"el ID de la cuenta donde depositar"`
	Amount      float64 `json:"amount" jsonschema:"el monto a depositar"`
	Description string  `json:"description" jsonschema:"descripcion opcional del deposito"`
}

type DepositOutput struct {
	Message string `json:"message"`
}

func (t *Tools) Deposit(ctx context.Context, req *mcp.CallToolRequest, input DepositInput) (*mcp.CallToolResult, DepositOutput, error) {
	if input.Amount <= 0 {
		return nil, DepositOutput{}, fmt.Errorf("el monto debe ser mayor a 0")
	}

	amountCents := uint64(input.Amount * 100)
	tbAccountID := AccountIDFromString(input.AccountID)
	transferID := AccountIDFromString(input.AccountID + "-" + uuid.New().String())

	if err := t.TbDB.Deposit(tbAccountID, amountCents, transferID); err != nil {
		return nil, DepositOutput{}, fmt.Errorf("error registrando deposito contable: %w", err)
	}

	if err := t.Repo.UpdateBalance(input.AccountID, input.Amount); err != nil {
		return nil, DepositOutput{}, fmt.Errorf("error actualizando balance: %w", err)
	}

	if err := t.Repo.CreateTransaction("EXTERNAL", input.AccountID, input.Amount, "deposit", input.Description); err != nil {
		return nil, DepositOutput{}, fmt.Errorf("error registrando transaccion: %w", err)
	}

	return nil, DepositOutput{Message: fmt.Sprintf("Deposito de $%.2f realizado correctamente", input.Amount)}, nil
}

// ---- withdraw ----

type WithdrawInput struct {
	AccountID   string  `json:"account_id" jsonschema:"el ID de la cuenta de donde retirar"`
	Amount      float64 `json:"amount" jsonschema:"el monto a retirar"`
	Description string  `json:"description" jsonschema:"descripcion opcional del retiro"`
}

type WithdrawOutput struct {
	Message string `json:"message"`
}

func (t *Tools) Withdraw(ctx context.Context, req *mcp.CallToolRequest, input WithdrawInput) (*mcp.CallToolResult, WithdrawOutput, error) {
	if input.Amount <= 0 {
		return nil, WithdrawOutput{}, fmt.Errorf("el monto debe ser mayor a 0")
	}

	account, err := t.Repo.GetAccountByID(input.AccountID)
	if err != nil {
		return nil, WithdrawOutput{}, fmt.Errorf("cuenta no encontrada: %w", err)
	}

	if account.Balance < input.Amount {
		return nil, WithdrawOutput{}, fmt.Errorf("saldo insuficiente")
	}

	amountCents := uint64(input.Amount * 100)
	tbAccountID := AccountIDFromString(input.AccountID)
	transferID := AccountIDFromString(input.AccountID + "-" + uuid.New().String())

	if err := t.TbDB.Withdraw(tbAccountID, amountCents, transferID); err != nil {
		return nil, WithdrawOutput{}, fmt.Errorf("error registrando retiro contable: %w", err)
	}

	if err := t.Repo.UpdateBalance(input.AccountID, -input.Amount); err != nil {
		return nil, WithdrawOutput{}, fmt.Errorf("error actualizando balance: %w", err)
	}

	if err := t.Repo.CreateTransaction(input.AccountID, "EXTERNAL", input.Amount, "withdrawal", input.Description); err != nil {
		return nil, WithdrawOutput{}, fmt.Errorf("error registrando transaccion: %w", err)
	}

	return nil, WithdrawOutput{Message: fmt.Sprintf("Retiro de $%.2f realizado correctamente", input.Amount)}, nil
}

// ---- transfer ----

type TransferInput struct {
	FromAccountID string  `json:"from_account_id" jsonschema:"el ID de la cuenta origen"`
	ToAccountID   string  `json:"to_account_id" jsonschema:"el ID de la cuenta destino"`
	Amount        float64 `json:"amount" jsonschema:"el monto a transferir"`
	Description   string  `json:"description" jsonschema:"descripcion opcional de la transferencia"`
}

type TransferOutput struct {
	Message string `json:"message"`
}

func (t *Tools) Transfer(ctx context.Context, req *mcp.CallToolRequest, input TransferInput) (*mcp.CallToolResult, TransferOutput, error) {
	if input.Amount <= 0 {
		return nil, TransferOutput{}, fmt.Errorf("el monto debe ser mayor a 0")
	}

	fromAccount, err := t.Repo.GetAccountByID(input.FromAccountID)
	if err != nil {
		return nil, TransferOutput{}, fmt.Errorf("cuenta origen no encontrada: %w", err)
	}

	if fromAccount.Balance < input.Amount {
		return nil, TransferOutput{}, fmt.Errorf("saldo insuficiente")
	}

	toAccount, err := t.Repo.GetAccountByID(input.ToAccountID)
	if err != nil {
		return nil, TransferOutput{}, fmt.Errorf("cuenta destino no encontrada: %w", err)
	}

	// SEGURIDAD: el chat con IA solo puede transferir entre cuentas del mismo propietario
	if toAccount.UserID != fromAccount.UserID {
		return nil, TransferOutput{}, fmt.Errorf("por seguridad, el asistente solo puede transferir dinero entre tus propias cuentas. Para transferir a un tercero, usa la sección de Transacciones en el dashboard")
	}

	amountCents := uint64(input.Amount * 100)
	fromTbID := AccountIDFromString(input.FromAccountID)
	toTbID := AccountIDFromString(input.ToAccountID)
	transferID := AccountIDFromString(input.FromAccountID + "-" + input.ToAccountID + "-" + uuid.New().String())

	if err := t.TbDB.Transfer(fromTbID, toTbID, amountCents, transferID); err != nil {
		return nil, TransferOutput{}, fmt.Errorf("error registrando transferencia contable: %w", err)
	}

	if err := t.Repo.UpdateBalance(input.FromAccountID, -input.Amount); err != nil {
		return nil, TransferOutput{}, fmt.Errorf("error actualizando balance origen: %w", err)
	}

	if err := t.Repo.UpdateBalance(input.ToAccountID, input.Amount); err != nil {
		return nil, TransferOutput{}, fmt.Errorf("error actualizando balance destino: %w", err)
	}

	if err := t.Repo.CreateTransaction(input.FromAccountID, input.ToAccountID, input.Amount, "transfer", input.Description); err != nil {
		return nil, TransferOutput{}, fmt.Errorf("error registrando transaccion: %w", err)
	}

	return nil, TransferOutput{Message: fmt.Sprintf("Transferencia de $%.2f realizada correctamente", input.Amount)}, nil
}