package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fran-ciscoo/banking-app/mcp-server/internal"
)

func main() {
	godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/banking?sslmode=disable"
	}

	repo, err := internal.NewRepository(databaseURL)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	tigerbeetleAddr := os.Getenv("TIGERBEETLE_ADDR")
	if tigerbeetleAddr == "" {
		tigerbeetleAddr = "localhost:3000"
	}

	tbDB, err := internal.NewTigerBeetleDB(tigerbeetleAddr)
	if err != nil {
		log.Fatalf("Error conectando a TigerBeetle: %v", err)
	}

	tools := internal.NewTools(repo, tbDB)

	server := mcp.NewServer(&mcp.Implementation{Name: "banking-mcp-server", Version: "v1.0.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_balance",
		Description: "Obtiene el saldo y las cuentas bancarias de un usuario",
	}, tools.GetBalance)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_history",
		Description: "Obtiene el historial de transacciones de una cuenta",
	}, tools.GetHistory)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deposit",
		Description: "Deposita dinero en una cuenta bancaria",
	}, tools.Deposit)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "withdraw",
		Description: "Retira dinero de una cuenta bancaria",
	}, tools.Withdraw)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "transfer",
		Description: "Transfiere dinero entre dos cuentas bancarias",
	}, tools.Transfer)

	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, nil)

	log.Println("Servidor MCP corriendo en http://localhost:9090")
	log.Fatal(http.ListenAndServe(":9090", handler))
}