package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPClient struct {
	session *mcp.ClientSession
}

func NewMCPClient(ctx context.Context, serverURL string) (*MCPClient, error) {
	client := mcp.NewClient(&mcp.Implementation{Name: "banking-backend", Version: "v1.0.0"}, nil)

	transport := &mcp.StreamableClientTransport{
		Endpoint:   serverURL,
		HTTPClient: &http.Client{},
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("error conectando al servidor MCP: %w", err)
	}

	return &MCPClient{session: session}, nil
}

func (m *MCPClient) Close() {
	m.session.Close()
}

// ListTools devuelve las tools disponibles en el servidor MCP
func (m *MCPClient) ListTools(ctx context.Context) ([]*mcp.Tool, error) {
	result, err := m.session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error listando tools: %w", err)
	}
	return result.Tools, nil
}

// CallTool ejecuta una tool específica con los argumentos dados
func (m *MCPClient) CallTool(ctx context.Context, name string, args map[string]any) (string, error) {
	params := &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	}

	res, err := m.session.CallTool(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error llamando tool %s: %w", name, err)
	}

	if res.IsError {
		errMsg := "error desconocido"
		if len(res.Content) > 0 {
			if tc, ok := res.Content[0].(*mcp.TextContent); ok {
				errMsg = tc.Text
			}
		}
		return "", fmt.Errorf("la tool %s falló: %s", name, errMsg)
	}

	var output string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			output += tc.Text
		}
	}

	return output, nil
}