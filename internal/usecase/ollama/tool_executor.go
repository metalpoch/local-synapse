package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/dto"
)

// ToolExecutor handles execution of MCP tool calls
type ToolExecutor struct {
	mcpClient domain.MCPClient
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(mcpClient domain.MCPClient) *ToolExecutor {
	return &ToolExecutor{mcpClient: mcpClient}
}

// ExecuteToolCalls executes all tool calls and returns messages to append to chat history
func (e *ToolExecutor) ExecuteToolCalls(ctx context.Context, toolCalls []dto.ToolCall) ([]dto.OllamaChatMessage, error) {
	if e.mcpClient == nil {
		return nil, fmt.Errorf("MCP client not available")
	}

	var messages []dto.OllamaChatMessage

	for _, tc := range toolCalls {
		log.Printf("[MCP] Executing tool: %s", tc.Function.Name)

		result, err := e.mcpClient.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
		content := ""

		if err != nil {
			log.Printf("[MCP] Tool execution failed: %v", err)
			content = fmt.Sprintf("Error executing tool: %v", err)
		} else {
			log.Printf("[MCP] Tool execution successful")
			content = e.formatToolResult(result)
		}

		// Append tool result message
		messages = append(messages, dto.OllamaChatMessage{
			Role:    "tool",
			Content: content,
		})
	}

	return messages, nil
}

// formatToolResult converts MCP result to string format
func (e *ToolExecutor) formatToolResult(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}

	var content string
	for _, c := range result.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			content += tc.Text + " "
		} else {
			b, _ := json.Marshal(c)
			content += string(b) + " "
		}
	}

	return content
}
