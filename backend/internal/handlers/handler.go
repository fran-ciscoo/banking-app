package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fran-ciscoo/banking-app/internal/repository"
)

type Handler struct {
	DB *repository.PostgresDB
}

func NewHandler(db *repository.PostgresDB) *Handler {
	return &Handler{DB: db}
}

// Helper para responder JSON fácilmente
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper para responder errores
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}