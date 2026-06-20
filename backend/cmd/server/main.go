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

	fmt.Printf("Intentando conectar a TigerBeetle con dirección: [%s]\n", cfg.TigerBeetleAddr)
	tbDB, err := repository.NewTigerBeetleDB(cfg.TigerBeetleAddr)
	if err != nil {
		log.Fatalf("Error conectando a TigerBeetle: %v", err)
	}

	// Crear la cuenta especial EXTERNAL (ID 1) que representa dinero entrando/saliendo del banco
	if err := tbDB.CreateAccount(1); err != nil {
		log.Printf("Advertencia: no se pudo crear cuenta EXTERNAL (puede que ya exista): %v", err)
	}

	if err := db.CreateTables(); err != nil {
		log.Fatalf("Error creando tablas: %v", err)
	}

	h := handlers.NewHandler(db, tbDB)

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