package ollama

import (
	"context"
	"log"

	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/dto"
)

// StreamChatUsecase orchestrates the chat streaming flow with Ollama
type StreamChatUsecase struct {
	ollamaClient  *OllamaClient
	toolExecutor  *ToolExecutor
	mcpClient     domain.MCPClient
	model         string
	systemPrompt  string
}

// NewStreamChatUsecase creates a new stream chat usecase
func NewStreamChatUsecase(
	ollamaURL string,
	model string,
	systemPrompt string,
	mcpClient domain.MCPClient,
) *StreamChatUsecase {
	return &StreamChatUsecase{
		ollamaClient:  NewOllamaClient(ollamaURL),
		toolExecutor:  NewToolExecutor(mcpClient),
		mcpClient:     mcpClient,
		model:         model,
		systemPrompt:  systemPrompt,
	}
}

// StreamChat executes the complete chat flow with tool calling support
// onChunk is called for each response chunk from Ollama
func (uc *StreamChatUsecase) StreamChat(
	ctx context.Context,
	userPrompt string,
	onChunk func(dto.OllamaChatResponse) error,
) error {
	// 1. Get available MCP tools
	tools := uc.getAvailableTools(ctx)

	// 2. Build initial messages
	messages := []dto.OllamaChatMessage{
		{Role: "system", Content: uc.systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	// 3. First request to Ollama (streaming)
	log.Printf("[MCP] Sending initial request to Ollama (Streaming mode)")

	var fullContent string
	var allToolCalls []dto.ToolCall

	request := dto.OllamaChatRequest{
		Model:    uc.model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
	}

	// Stream first response and accumulate content/tool calls
	err := uc.ollamaClient.StreamChatRequest(ctx, request, func(chunk dto.OllamaChatResponse) error {
		// Accumulate content and tool calls
		fullContent += chunk.Message.Content
		if len(chunk.Message.ToolCalls) > 0 {
			allToolCalls = append(allToolCalls, chunk.Message.ToolCalls...)
		}

		// Forward chunk to caller
		return onChunk(chunk)
	})

	if err != nil {
		return err
	}

	// 4. Check if tools were called
	if len(allToolCalls) > 0 {
		log.Printf("[MCP] Ollama requested %d tools", len(allToolCalls))

		// Append assistant's message with tool calls
		messages = append(messages, dto.OllamaChatMessage{
			Role:      "assistant",
			Content:   fullContent,
			ToolCalls: allToolCalls,
		})

		// Execute tools
		toolMessages, err := uc.toolExecutor.ExecuteToolCalls(ctx, allToolCalls)
		if err != nil {
			log.Printf("[MCP] Tool execution error: %v", err)
			// Continue anyway with error messages
		}

		// Append tool results to messages
		messages = append(messages, toolMessages...)

		// 5. Final request with tool results
		log.Printf("[MCP] Sending final request with tool results")
		finalRequest := dto.OllamaChatRequest{
			Model:    uc.model,
			Messages: messages,
			Stream:   true,
			Tools:    tools, // Keep tools enabled for multi-turn
		}

		return uc.ollamaClient.StreamChatRequest(ctx, finalRequest, onChunk)
	}

	return nil
}

// getAvailableTools fetches available tools from MCP client
func (uc *StreamChatUsecase) getAvailableTools(ctx context.Context) []dto.Tool {
	var tools []dto.Tool

	if uc.mcpClient == nil {
		return tools
	}

	mcpTools, err := uc.mcpClient.ListTools(ctx)
	if err != nil {
		log.Printf("[MCP] Error listing tools: %v", err)
		return tools
	}

	if len(mcpTools) == 0 {
		return tools
	}

	log.Printf("[MCP] Discovered %d tools", len(mcpTools))

	for _, t := range mcpTools {
		tools = append(tools, dto.Tool{
			Type: "function",
			Function: dto.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	return tools
}
