package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/fran-ciscoo/banking-app/internal/models"
	"github.com/fran-ciscoo/banking-app/internal/repository"
)

func uuidNow() string {
	return uuid.New().String()
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "El monto debe ser mayor a 0")
		return
	}

	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}

	account := accounts[0]

	// Convertir el monto a centavos (TigerBeetle trabaja con enteros, sin decimales)
	amountCents := uint64(req.Amount * 100)
	tbAccountID := repository.AccountIDFromString(account.ID)
	transferID := repository.AccountIDFromString(account.ID + "-" + uuidNow())

	// Registrar el depósito en TigerBeetle (motor contable real)
	if err := h.TbDB.Deposit(tbAccountID, amountCents, transferID); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando depósito contable: "+err.Error())
		return
	}

	// Actualizar el balance espejo en PostgreSQL (lectura rápida para el dashboard)
	if err := h.DB.UpdateBalance(account.ID, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance")
		return
	}

	if err := h.DB.CreateTransaction("EXTERNAL", account.ID, req.Amount, "deposit", req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando transacción")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Depósito realizado correctamente",
		"amount":  req.Amount,
	})
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req models.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "El monto debe ser mayor a 0")
		return
	}

	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}

	account := accounts[0]

	if account.Balance < req.Amount {
		respondError(w, http.StatusBadRequest, "Saldo insuficiente")
		return
	}

	amountCents := uint64(req.Amount * 100)
	tbAccountID := repository.AccountIDFromString(account.ID)
	transferID := repository.AccountIDFromString(account.ID + "-" + uuidNow())

	// TigerBeetle valida automáticamente que haya saldo suficiente (cuentas no pueden ir negativas por defecto)
	if err := h.TbDB.Withdraw(tbAccountID, amountCents, transferID); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando retiro contable: "+err.Error())
		return
	}

	if err := h.DB.UpdateBalance(account.ID, -req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance")
		return
	}

	if err := h.DB.CreateTransaction(account.ID, "EXTERNAL", req.Amount, "withdrawal", req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando transacción")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Retiro realizado correctamente",
		"amount":  req.Amount,
	})
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "El monto debe ser mayor a 0")
		return
	}

	if req.ToAccountID == "" {
		respondError(w, http.StatusBadRequest, "Cuenta destino requerida")
		return
	}

	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}
	fromAccount := accounts[0]

	if fromAccount.ID == req.ToAccountID {
		respondError(w, http.StatusBadRequest, "No puedes transferirte a ti mismo")
		return
	}

	if fromAccount.Balance < req.Amount {
		respondError(w, http.StatusBadRequest, "Saldo insuficiente")
		return
	}

	toAccount, err := h.DB.GetAccountByID(req.ToAccountID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Cuenta destino no encontrada")
		return
	}

	amountCents := uint64(req.Amount * 100)
	fromTbID := repository.AccountIDFromString(fromAccount.ID)
	toTbID := repository.AccountIDFromString(toAccount.ID)
	transferID := repository.AccountIDFromString(fromAccount.ID + "-" + toAccount.ID + "-" + uuidNow())

	if err := h.TbDB.Transfer(fromTbID, toTbID, amountCents, transferID); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando transferencia contable: "+err.Error())
		return
	}

	if err := h.DB.UpdateBalance(fromAccount.ID, -req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance origen")
		return
	}

	if err := h.DB.UpdateBalance(toAccount.ID, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance destino")
		return
	}

	if err := h.DB.CreateTransaction(fromAccount.ID, toAccount.ID, req.Amount, "transfer", req.Description); err != nil {
		respondError(w, http.StatusInternalServerError, "Error registrando transacción")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Transferencia realizada correctamente",
		"amount":  req.Amount,
		"to":      toAccount.ID,
	})
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	accountID := r.URL.Query().Get("account")

	var transactions []repository.TransactionRecord

	if accountID != "" {
		// Verificar que la cuenta pertenece al usuario antes de mostrar su historial
		account, err := h.DB.GetAccountByID(accountID)
		if err != nil || account.UserID != userID {
			respondError(w, http.StatusForbidden, "No tienes acceso a esta cuenta")
			return
		}
		transactions, err = h.DB.GetTransactionsByAccountID(accountID, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error obteniendo historial")
			return
		}
	} else {
		transactions, err = h.DB.GetTransactionsByUserID(userID, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Error obteniendo historial")
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"timestamp":    time.Now(),
	})
}