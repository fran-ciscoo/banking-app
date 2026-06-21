package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/fran-ciscoo/banking-app/internal/repository"
	"github.com/fran-ciscoo/banking-app/pkg/config"
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error obteniendo cuentas")
		return
	}

	// Sobrescribir el balance con el valor real de TigerBeetle (fuente de verdad financiera)
	for i := range accounts {
		tbAccountID := repository.AccountIDFromString(accounts[i].ID)
		balanceCents, err := h.TbDB.GetBalance(tbAccountID)
		if err == nil {
			accounts[i].Balance = float64(balanceCents) / 100
		}
	}

	user, err := h.DB.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error obteniendo usuario")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user":     user,
		"accounts": accounts,
	})
}

// Helper para extraer el user_id del JWT
func getUserIDFromToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	parts := strings.Split(authHeader, " ")
	tokenString := parts[1]

	cfg := config.Load()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return "", err
	}

	claims := token.Claims.(jwt.MapClaims)
	return claims["user_id"].(string), nil
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Type != "checking" && req.Type != "savings" {
		respondError(w, http.StatusBadRequest, "Tipo de cuenta inválido")
		return
	}

	accountUUID := uuid.New().String()
	accountID := fmt.Sprintf("4001-%s-%s-%s",
		accountUUID[0:4],
		accountUUID[4:8],
		accountUUID[9:13],
	)

	// Crear la cuenta en PostgreSQL (metadatos: usuario, tipo, nickname)
	if err := h.DB.CreateAccount(accountID, userID, req.Type); err != nil {
		respondError(w, http.StatusInternalServerError, "Error creando cuenta")
		return
	}

	// Crear la cuenta contable en TigerBeetle (balances y transferencias)
	tbAccountID := repository.AccountIDFromString(accountID)
	if err := h.TbDB.CreateAccount(tbAccountID); err != nil {
		respondError(w, http.StatusInternalServerError, "Error creando cuenta contable: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Cuenta creada correctamente",
		"account_id": accountID,
		"type":       req.Type,
	})
}

func (h *Handler) UpdateAccountNickname(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	accountID := chi.URLParam(r, "id")

	var req struct {
		Nickname string `json:"nickname"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Nickname == "" {
		respondError(w, http.StatusBadRequest, "El nombre no puede estar vacío")
		return
	}

	if err := h.DB.UpdateAccountNickname(accountID, userID, req.Nickname); err != nil {
		respondError(w, http.StatusBadRequest, "Error actualizando el nombre")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Nombre actualizado correctamente",
	})
}

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	accountID := chi.URLParam(r, "id")

	if err := h.DB.DeleteAccount(accountID, userID); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Cuenta eliminada correctamente",
	})
}