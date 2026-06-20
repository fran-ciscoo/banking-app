package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/fran-ciscoo/banking-app/internal/handlers"
	"github.com/fran-ciscoo/banking-app/internal/repository"
	"github.com/fran-ciscoo/banking-app/internal/services"
	"github.com/fran-ciscoo/banking-app/pkg/config"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	if err := db.CreateTables(); err != nil {
		log.Fatalf("Error creando tablas: %v", err)
	}

	h := handlers.NewHandler(db)

	// Conectar al servidor MCP
	mcpServerURL := cfg.MCPServerURL
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:9090"
	}
	mcpClient, err := services.NewMCPClient(ctx, mcpServerURL)

	var chatHandler *handlers.ChatHandler
	if mcpClient != nil {
		openRouter := services.NewOpenRouterClient(cfg.OpenRouterKey)
		chatService := services.NewChatService(mcpClient, openRouter)
		chatHandler = handlers.NewChatHandler(chatService)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Post("/api/auth/register", h.Register)
	r.Post("/api/auth/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Post("/api/auth/logout", h.Logout)
		r.Get("/api/account", h.GetAccount)
		r.Post("/api/account/create", h.CreateAccount)
		r.Put("/api/account/{id}/nickname", h.UpdateAccountNickname)
		r.Delete("/api/account/{id}", h.DeleteAccount)
		r.Post("/api/transactions/deposit", h.Deposit)
		r.Post("/api/transactions/withdraw", h.Withdraw)
		r.Post("/api/transactions/transfer", h.Transfer)
		r.Get("/api/transactions/history", h.GetHistory)

		if chatHandler != nil {
			r.Post("/api/chat", chatHandler.Chat)
		}
	})

	fmt.Printf("Servidor corriendo en http://localhost:%s\n", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}