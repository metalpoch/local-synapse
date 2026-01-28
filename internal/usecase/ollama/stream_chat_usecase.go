package ollama

import (
	"context"
	"log"

	"github.com/metalpoch/local-synapse/internal/dto"
	mcpclient "github.com/metalpoch/local-synapse/internal/infrastructure/mcp_client"
	"github.com/metalpoch/local-synapse/internal/infrastructure/ollama"
)

// StreamChatUsecase orchestrates the chat streaming flow with Ollama
type StreamChatUsecase struct {
	ollamaClient *ollama_infra.OllamaClient
	toolExecutor *ToolExecutor
	mcpClient    mcpclient.MCPClient
	model        string
	systemPrompt string
}

// NewStreamChatUsecase creates a new stream chat usecase
func NewStreamChatUsecase(
	ollamaURL string,
	model string,
	systemPrompt string,
	mcpClient mcpclient.MCPClient,
) *StreamChatUsecase {
	return &StreamChatUsecase{
		ollamaClient: ollama_infra.NewOllamaClient(ollamaURL),
		toolExecutor: NewToolExecutor(mcpClient),
		mcpClient:    mcpClient,
		model:        model,
		systemPrompt: systemPrompt,
	}
}

// Execute handles the full chat flow with Ollama, including tool calling and persistence.
func (uc *StreamChatUsecase) Execute(ctx context.Context, userPrompt string, onChunk func(dto.OllamaChatResponse) error) error {
	tools := uc.getAvailableTools(ctx)

	var messages []dto.OllamaChatMessage = []dto.OllamaChatMessage{
		{Role: "system", Content: uc.systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	log.Printf("[MCP] Sending initial request to Ollama (Streaming mode)")

	var fullContent string
	var allToolCalls []dto.ToolCall

	request := dto.OllamaChatRequest{
		Model:    uc.model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
	}

	// Stream the first response and gather chunks
	err := uc.ollamaClient.StreamChatRequest(ctx, request, func(chunk dto.OllamaChatResponse) error {
		fullContent += chunk.Message.Content
		if len(chunk.Message.ToolCalls) > 0 {
			allToolCalls = append(allToolCalls, chunk.Message.ToolCalls...)
		}
		return onChunk(chunk)
	})

	if err != nil {
		return err
	}

	// Handle tool execution if requested by the model
	if len(allToolCalls) > 0 {
		log.Printf("[MCP] Ollama requested %d tools", len(allToolCalls))

		messages = append(messages, dto.OllamaChatMessage{
			Role:      "assistant",
			Content:   fullContent,
			ToolCalls: allToolCalls,
		})

		toolMessages, err := uc.toolExecutor.ExecuteToolCalls(ctx, allToolCalls)
		if err != nil {
			log.Printf("[MCP] Tool execution error: %v", err)
		}

		messages = append(messages, toolMessages...)

		log.Printf("[MCP] Sending final request with tool results")

		fullContent = ""
		allToolCalls = nil

		finalRequest := dto.OllamaChatRequest{
			Model:    uc.model,
			Messages: messages,
			Stream:   true,
			Tools:    tools,
		}

		err = uc.ollamaClient.StreamChatRequest(ctx, finalRequest, func(chunk dto.OllamaChatResponse) error {
			fullContent += chunk.Message.Content
			if len(chunk.Message.ToolCalls) > 0 {
				allToolCalls = append(allToolCalls, chunk.Message.ToolCalls...)
			}
			return onChunk(chunk)
		})
		if err != nil {
			return err
		}
	}

	// Add the final assistant response to the messages for caching
	finalAssistantMsg := dto.OllamaChatMessage{
		Role:    "assistant",
		Content: fullContent,
	}
	if len(allToolCalls) > 0 {
		finalAssistantMsg.ToolCalls = allToolCalls
	}
	messages = append(messages, finalAssistantMsg)

	return nil
}

// getAvailableTools fetches tools from the MCP client
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
