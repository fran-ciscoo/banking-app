package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fran-ciscoo/banking-app/internal/repository"
)

type Handler struct {
	DB   *repository.PostgresDB
	TbDB *repository.TigerBeetleDB
}

func NewHandler(db *repository.PostgresDB, tbDB *repository.TigerBeetleDB) *Handler {
	return &Handler{DB: db, TbDB: tbDB}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}