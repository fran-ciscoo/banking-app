package handlers

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/fran-ciscoo/banking-app/pkg/config"
)

func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	// Obtener user_id del token JWT
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	// Obtener cuentas del usuario
	accounts, err := h.DB.GetAccountsByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error obteniendo cuentas")
		return
	}

	// Obtener info del usuario
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