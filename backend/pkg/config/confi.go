package config

import (
	"os"
	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	JWTSecret       string
	TigerBeetleAddr string
	OpenRouterKey   string
	MCPServerURL    string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", "cambia-esto-en-produccion"),
		TigerBeetleAddr: getEnv("TIGERBEETLE_ADDR", "localhost:3000"),
		OpenRouterKey:   getEnv("OPENROUTER_API_KEY", ""),
		MCPServerURL:    getEnv("MCP_SERVER_URL", "http://localhost:9090"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}