package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fran-ciscoo/banking-app/internal/models"
)

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

	// Obtener cuenta del usuario
	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}

	account := accounts[0]

	// Actualizar balance en PostgreSQL
	if err := h.DB.UpdateBalance(account.ID, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance")
		return
	}

	// Registrar transacción
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

	// Obtener cuenta
	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}

	account := accounts[0]

	// Verificar saldo suficiente
	if account.Balance < req.Amount {
		respondError(w, http.StatusBadRequest, "Saldo insuficiente")
		return
	}

	// Actualizar balance (negativo para restar)
	if err := h.DB.UpdateBalance(account.ID, -req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance")
		return
	}

	// Registrar transacción
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

	// Obtener cuenta origen
	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil || len(accounts) == 0 {
		respondError(w, http.StatusNotFound, "Cuenta no encontrada")
		return
	}
	fromAccount := accounts[0]

	// Verificar que no se transfiere a sí mismo
	if fromAccount.ID == req.ToAccountID {
		respondError(w, http.StatusBadRequest, "No puedes transferirte a ti mismo")
		return
	}

	// Verificar saldo
	if fromAccount.Balance < req.Amount {
		respondError(w, http.StatusBadRequest, "Saldo insuficiente")
		return
	}

	// Verificar que la cuenta destino existe
	toAccount, err := h.DB.GetAccountByID(req.ToAccountID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Cuenta destino no encontrada")
		return
	}

	// Actualizar balances
	if err := h.DB.UpdateBalance(fromAccount.ID, -req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance origen")
		return
	}

	if err := h.DB.UpdateBalance(toAccount.ID, req.Amount); err != nil {
		respondError(w, http.StatusInternalServerError, "Error actualizando balance destino")
		return
	}

	// Registrar transacción
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

	transactions, err := h.DB.GetTransactionsByUserID(userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error obteniendo historial")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"transactions": transactions,
		"timestamp":    time.Now(),
	})
}