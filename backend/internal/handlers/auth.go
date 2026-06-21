package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/fran-ciscoo/banking-app/internal/models"
	"github.com/fran-ciscoo/banking-app/internal/repository"
	"github.com/fran-ciscoo/banking-app/pkg/config"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		respondError(w, http.StatusBadRequest, "Todos los campos son requeridos")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error procesando contraseña")
		return
	}

	userID := uuid.New().String()

	if err := h.DB.CreateUser(userID, req.Email, string(hashedPassword), req.FullName); err != nil {
		respondError(w, http.StatusConflict, "El email ya está registrado")
		return
	}

	// Crear cuenta bancaria automáticamente
	accountUUID := uuid.New().String()
	accountID := fmt.Sprintf("4001-%s-%s-%s",
		accountUUID[0:4],
		accountUUID[4:8],
		accountUUID[9:13],
	)

	if err := h.DB.CreateAccount(accountID, userID, "savings"); err != nil {
		respondError(w, http.StatusInternalServerError, "Error creando cuenta bancaria")
		return
	}

	tbAccountID := repository.AccountIDFromString(accountID)
	fmt.Printf("Creando cuenta TigerBeetle: accountID=%s -> tbAccountID=%d\n", accountID, tbAccountID)
	if err := h.TbDB.CreateAccount(tbAccountID); err != nil {
		fmt.Printf("ERROR TigerBeetle: %v\n", err)
		respondError(w, http.StatusInternalServerError, "Error creando cuenta contable: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Usuario registrado correctamente",
		"user_id":    userID,
		"account_id": accountID,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	user, err := h.DB.GetUserByEmail(req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Email o contraseña incorrectos")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "Email o contraseña incorrectos")
		return
	}

	// Si el usuario tiene 2FA activado, no entregamos el token todavía:
	// pedimos el código TOTP en un segundo paso.
	if user.TotpEnabled {
		if req.Code == "" {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"requires_2fa": true,
				"message":      "Ingresa el código de tu app de autenticación",
			})
			return
		}

		if user.TotpSecret == nil || !totp.Validate(req.Code, *user.TotpSecret) {
			respondError(w, http.StatusUnauthorized, "Código de verificación incorrecto")
			return
		}
	}

	cfg := config.Load()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error generando token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token":         tokenString,
		"user_id":       user.ID,
		"full_name":     user.FullName,
		"email":         user.Email,
		"totp_enabled":  user.TotpEnabled,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Sesión cerrada correctamente",
	})
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Obtener el token del header Authorization: Bearer <token>
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "Token requerido")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "Formato de token inválido")
			return
		}

		tokenString := parts[1]
		cfg := config.Load()

		// Validar el token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			respondError(w, http.StatusUnauthorized, "Token inválido o expirado")
			return
		}

		// Pasar el request al siguiente handler
		next.ServeHTTP(w, r)
	})
}