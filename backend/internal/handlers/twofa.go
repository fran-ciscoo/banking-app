package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Setup2FA genera un secreto TOTP nuevo y devuelve el código QR en base64
func (h *Handler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	user, err := h.DB.GetUserByID(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "BankingApp",
		AccountName: user.Email,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error generando secreto 2FA")
		return
	}

	// Guardar el secreto (aún no activado hasta que el usuario confirme con un código válido)
	if err := h.DB.SetTotpSecret(userID, key.Secret()); err != nil {
		respondError(w, http.StatusInternalServerError, "Error guardando secreto")
		return
	}

	// Generar la imagen QR
	img, err := key.Image(256, 256)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error generando QR")
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		respondError(w, http.StatusInternalServerError, "Error codificando QR")
		return
	}

	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"secret":   key.Secret(),
		"qr_image": "data:image/png;base64," + qrBase64,
	})
}

// Confirm2FA verifica el primer código y activa 2FA definitivamente
func (h *Handler) Confirm2FA(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	user, err := h.DB.GetUserByID(userID)
	if err != nil || user.TotpSecret == nil {
		respondError(w, http.StatusBadRequest, "Primero debes iniciar la configuración de 2FA")
		return
	}

	valid := totp.Validate(req.Code, *user.TotpSecret)
	if !valid {
		respondError(w, http.StatusBadRequest, "Código incorrecto")
		return
	}

	if err := h.DB.EnableTotp(userID, true); err != nil {
		respondError(w, http.StatusInternalServerError, "Error activando 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Autenticación de dos factores activada correctamente",
	})
}

// Disable2FA desactiva 2FA para el usuario
func (h *Handler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	if err := h.DB.EnableTotp(userID, false); err != nil {
		respondError(w, http.StatusInternalServerError, "Error desactivando 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Autenticación de dos factores desactivada",
	})
}

var _ = otp.Algorithm(0)