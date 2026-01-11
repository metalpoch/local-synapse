package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/metalpoch/local-synapse/internal/domain"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/entity"
	"github.com/metalpoch/local-synapse/internal/infrastructure/cache"
	"github.com/metalpoch/local-synapse/internal/repository"
)

// StreamChatUsecase orchestrates the chat streaming flow with Ollama
type StreamChatUsecase struct {
	ollamaClient         *OllamaClient
	toolExecutor         *ToolExecutor
	mcpClient            domain.MCPClient
	model                string
	systemPrompt         string
	conversationRepo     repository.ConversationRepository
	conversationCache    cache.ConversationCache
}

// NewStreamChatUsecase creates a new stream chat usecase
func NewStreamChatUsecase(
	ollamaURL string,
	model string,
	systemPrompt string,
	mcpClient domain.MCPClient,
	conversationRepo repository.ConversationRepository,
	conversationCache cache.ConversationCache,
) *StreamChatUsecase {
	return &StreamChatUsecase{
		ollamaClient:      NewOllamaClient(ollamaURL),
		toolExecutor:      NewToolExecutor(mcpClient),
		mcpClient:         mcpClient,
		model:             model,
		systemPrompt:      systemPrompt,
		conversationRepo:  conversationRepo,
		conversationCache: conversationCache,
	}
}

// StreamChat executes the complete chat flow with tool calling support
// onChunk is called for each response chunk from Ollama
func (uc *StreamChatUsecase) StreamChat(ctx context.Context, user *dto.UserResponse, userPrompt string, onChunk func(dto.OllamaChatResponse) error) error {
	tools := uc.getAvailableTools(ctx)

	// 1. Obtener o crear conversación activa
	conversation, err := uc.conversationRepo.GetOrCreateActiveConversation(user.ID)
	if err != nil {
		log.Printf("[Context] Error getting/creating conversation: %v", err)
		return fmt.Errorf("failed to get conversation: %w", err)
	}

	// 2. Intentar cargar contexto del cache
	var messages []dto.OllamaChatMessage
	cachedMessages, err := uc.conversationCache.GetConversationFromCache(ctx, user.ID)
	
	if err != nil || len(cachedMessages) == 0 {
		// 3. Si no está en cache, cargar desde DB
		log.Printf("[Context] Cache miss, loading from DB for user %s", user.ID)
		dbMessages, err := uc.conversationRepo.GetConversationMessages(conversation.ID, 20)
		if err != nil {
			log.Printf("[Context] Error loading messages from DB: %v", err)
		} else {
			messages = convertMessagesToDTO(dbMessages)
		}
	} else {
		log.Printf("[Context] Cache hit, loaded %d messages for user %s", len(cachedMessages), user.ID)
		messages = cachedMessages
	}

	// 4. Inyectar información del usuario en el system prompt
	personalizedPrompt := fmt.Sprintf("%s\n\nUsuario actual: %s (Email: %s)", 
		uc.systemPrompt, user.Name, user.Email)

	// 5. Si no hay mensajes previos, agregar el system prompt
	if len(messages) == 0 {
		messages = []dto.OllamaChatMessage{
			{Role: "system", Content: personalizedPrompt},
		}
	} else {
		// Actualizar el system prompt si ya existe
		if messages[0].Role == "system" {
			messages[0].Content = personalizedPrompt
		} else {
			// Insertar al inicio si no existe
			messages = append([]dto.OllamaChatMessage{
				{Role: "system", Content: personalizedPrompt},
			}, messages...)
		}
	}

	// 6. Agregar el nuevo mensaje del usuario
	messages = append(messages, dto.OllamaChatMessage{
		Role:    "user",
		Content: userPrompt,
	})

	log.Printf("[MCP] Sending initial request to Ollama (Streaming mode)")

	var fullContent string
	var allToolCalls []dto.ToolCall

	request := dto.OllamaChatRequest{
		Model:    uc.model,
		Messages: messages,
		Stream:   true,
		Tools:    tools,
	}

	// 7. Stream first response and accumulate content/tool calls
	err = uc.ollamaClient.StreamChatRequest(ctx, request, func(chunk dto.OllamaChatResponse) error {
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

	// 8. Check if tools were called
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

		// 9. Final request with tool results
		log.Printf("[MCP] Sending final request with tool results")
		
		// Reset fullContent for final response
		fullContent = ""
		allToolCalls = nil
		
		finalRequest := dto.OllamaChatRequest{
			Model:    uc.model,
			Messages: messages,
			Stream:   true,
			Tools:    tools, // Keep tools enabled for multi-turn
		}

		err = uc.ollamaClient.StreamChatRequest(ctx, finalRequest, func(chunk dto.OllamaChatResponse) error {
			// Accumulate final content
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

	// 10. Guardar mensajes en la base de datos
	go uc.saveConversationContext(context.Background(), conversation.ID, user.ID, userPrompt, fullContent, allToolCalls, messages)

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

// convertMessagesToDTO converts entity messages to DTO messages
func convertMessagesToDTO(dbMessages []entity.Message) []dto.OllamaChatMessage {
	messages := make([]dto.OllamaChatMessage, 0, len(dbMessages))
	
	for _, msg := range dbMessages {
		dtoMsg := dto.OllamaChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
		
		// Deserializar tool calls si existen
		if msg.ToolCalls != nil && *msg.ToolCalls != "" {
			var toolCalls []dto.ToolCall
			if err := json.Unmarshal([]byte(*msg.ToolCalls), &toolCalls); err == nil {
				dtoMsg.ToolCalls = toolCalls
			}
		}
		
		messages = append(messages, dtoMsg)
	}
	
	return messages
}

// saveConversationContext guarda el contexto de la conversación en DB y cache
func (uc *StreamChatUsecase) saveConversationContext(
	ctx context.Context,
	conversationID string,
	userID string,
	userPrompt string,
	assistantContent string,
	toolCalls []dto.ToolCall,
	allMessages []dto.OllamaChatMessage,
) {
	// Guardar mensaje del usuario
	userMsg := &entity.Message{
		ConversationID: conversationID,
		Role:           "user",
		Content:        userPrompt,
	}
	
	if err := uc.conversationRepo.SaveMessage(userMsg); err != nil {
		log.Printf("[Context] Error saving user message: %v", err)
	}
	
	// Guardar respuesta del asistente
	assistantMsg := &entity.Message{
		ConversationID: conversationID,
		Role:           "assistant",
		Content:        assistantContent,
	}
	
	// Serializar tool calls si existen
	if len(toolCalls) > 0 {
		toolCallsJSON, err := json.Marshal(toolCalls)
		if err == nil {
			toolCallsStr := string(toolCallsJSON)
			assistantMsg.ToolCalls = &toolCallsStr
		}
	}
	
	if err := uc.conversationRepo.SaveMessage(assistantMsg); err != nil {
		log.Printf("[Context] Error saving assistant message: %v", err)
	}
	
	// Actualizar cache con todos los mensajes
	if err := uc.conversationCache.SaveConversationToCache(ctx, userID, allMessages); err != nil {
		log.Printf("[Context] Error updating cache: %v", err)
	} else {
		log.Printf("[Context] Cache updated for user %s", userID)
	}
}

