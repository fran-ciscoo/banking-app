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
func (c *ChatService) SendMessage(ctx context.Context, userID string, userMessage string) (string, error) {
	mcpTools, err := c.MCPClient.ListTools(ctx)
	if err != nil {
		return "", fmt.Errorf("error obteniendo tools: %w", err)
	}

	orTools := mcpToolsToOpenRouterTools(mcpTools)

	systemPrompt := fmt.Sprintf(`Eres un asistente bancario amable y profesional. Ayudas al usuario con sus operaciones bancarias: consultar saldo, ver historial, depositar, retirar y transferir dinero.

IMPORTANTE: El ID del usuario actual ya lo conoces, es: %s
NUNCA le preguntes al usuario su ID, ya lo tienes. Úsalo directamente en la tool get_balance cuando lo necesites.

Cuando necesites el ID de una cuenta específica (account_id), primero usa get_balance con este user_id para obtenerlo de la lista de cuentas devuelta.

Antes de ejecutar depósitos, retiros o transferencias, confirma los detalles con el usuario en tu respuesta si no están claros.
Responde siempre en español, de forma clara y breve.`, userID)

	messages := []ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

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