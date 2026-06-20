package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ChatService struct {
	MCPClient    *MCPClient
	OpenRouter   *OpenRouterClient
}

func NewChatService(mcpClient *MCPClient, openRouter *OpenRouterClient) *ChatService {
	return &ChatService{
		MCPClient:  mcpClient,
		OpenRouter: openRouter,
	}
}

// convierte las tools de MCP al formato que espera OpenRouter
func mcpToolsToOpenRouterTools(mcpTools []*mcp.Tool) []ToolDefinition {
	var tools []ToolDefinition
	for _, t := range mcpTools {
		var schema map[string]any
		if t.InputSchema != nil {
			schemaBytes, _ := json.Marshal(t.InputSchema)
			json.Unmarshal(schemaBytes, &schema)
		}
		if schema == nil {
			schema = map[string]any{"type": "object", "properties": map[string]any{}}
		}

		tools = append(tools, ToolDefinition{
			Type: "function",
			Function: ToolFunctionSchema{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  schema,
			},
		})
	}
	return tools
}

// SendMessage procesa un mensaje del usuario y devuelve la respuesta final
func (c *ChatService) SendMessage(ctx context.Context, userID string, userMessage string, history []ChatMessage) (string, error) {
	mcpTools, err := c.MCPClient.ListTools(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo tools: %w", err)
	}

	orTools := mcpToolsToOpenRouterTools(mcpTools)

	systemPrompt := fmt.Sprintf(`Eres un asistente bancario amable y profesional. Ayudas al usuario con sus operaciones bancarias: consultar saldo, ver historial, depositar, retirar y transferir dinero.

	IMPORTANTE: El ID del usuario actual ya lo conoces, es: %s
	NUNCA le preguntes al usuario su ID, ya lo tienes. Úsalo directamente en la tool get_balance cuando lo necesites.

	Cuando necesites el ID de una cuenta específica (account_id), primero usa get_balance con este user_id para obtenerlo de la lista de cuentas devuelta.

	REGLA CRÍTICA DE CONFIRMACIÓN:
	Antes de ejecutar CUALQUIER depósito, retiro o transferencia (las tools deposit, withdraw, transfer), DEBES primero responder al usuario describiendo exactamente la operación que vas a realizar (monto, cuenta origen, cuenta destino si aplica) y preguntar "¿Confirmas esta operación?". NO ejecutes la tool en ese mismo turno.

	Solo ejecuta la tool de depósito, retiro o transferencia en el turno SIGUIENTE, después de que el usuario responda afirmativamente (por ejemplo "sí", "confirmo", "adelante", "dale").

	Si el usuario responde negativamente o cambia de opinión, no ejecutes la operación y confírmalo.

	Las consultas de saldo (get_balance) e historial (get_history) NO requieren confirmación, ejecútalas directamente.

	POLÍTICA DE CIERRE DE CUENTAS:
	El sistema NO permite cerrar o eliminar una cuenta que tenga saldo distinto a $0. Si el usuario pregunta por qué no puede cerrar su cuenta, o pide cerrarla, primero usa get_balance para revisar el saldo de esa cuenta:
	- Si el saldo es $0, infórmale que sí puede cerrarla desde el botón de eliminar en el dashboard.
	- Si el saldo es distinto de $0, explícale que por política del banco no se pueden cerrar cuentas con saldo disponible, y que debe retirar o transferir todo el dinero primero, o acercarse físicamente a una sucursal del banco para gestionar el cierre.

	LIMITACIÓN DE TRANSFERENCIAS:
	Por seguridad, SOLO puedes ejecutar transferencias entre cuentas que pertenecen al mismo usuario (transferencias propias, por ejemplo de su cuenta corriente a su cuenta de ahorros). NO puedes transferir dinero a la cuenta de otra persona.
	Si el usuario te pide transferir dinero a un tercero, explícale amablemente que el asistente solo puede mover dinero entre sus propias cuentas, y que para transferencias a terceros debe usar la sección de Transacciones en el dashboard.

	Responde siempre en español, de forma clara y breve.`, userID)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, history...)
	messages = append(messages, ChatMessage{Role: "user", Content: userMessage})

	// Loop de hasta 5 iteraciones para permitir múltiples tool calls
	for i := 0; i < 5; i++ {
		response, err := c.OpenRouter.Chat(ctx, messages, orTools)
		if err != nil {
			return "", fmt.Errorf("error en chat: %w", err)
		}

		// Si no hay tool calls, esta es la respuesta final
		if len(response.ToolCalls) == 0 {
			return response.Content, nil
		}

		// Agregar la respuesta del asistente (con tool calls) al historial
		messages = append(messages, *response)

		// Ejecutar cada tool call y agregar el resultado
		for _, tc := range response.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tc.Function.Arguments), &args)

			result, err := c.MCPClient.CallTool(ctx, tc.Function.Name, args)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			messages = append(messages, ChatMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return "Lo siento, no pude completar la solicitud después de varios intentos.", nil
}