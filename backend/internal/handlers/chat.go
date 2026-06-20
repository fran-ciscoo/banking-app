package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fran-ciscoo/banking-app/internal/services"
)

type ChatHandler struct {
	ChatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{ChatService: chatService}
}

type ChatHistoryItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Message string            `json:"message"`
	History []ChatHistoryItem `json:"history"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func (h *ChatHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Datos inválidos")
		return
	}

	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "El mensaje no puede estar vacío")
		return
	}

	var history []services.ChatMessage
	for _, h := range req.History {
		history = append(history, services.ChatMessage{Role: h.Role, Content: h.Content})
	}

	reply, err := h.ChatService.SendMessage(r.Context(), userID, req.Message, history)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error procesando el mensaje: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ChatResponse{Reply: reply})
}