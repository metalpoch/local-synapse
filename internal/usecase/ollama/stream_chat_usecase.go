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

// StreamChat handles the full chat flow with Ollama, including tool calling and persistence.
func (uc *StreamChatUsecase) StreamChat(ctx context.Context, user *dto.UserResponse, conversationID, userPrompt string, onChunk func(dto.OllamaChatResponse) error) error {
	tools := uc.getAvailableTools(ctx)

	var conversation *entity.Conversation
	var err error

	if conversationID != "" {
		conversation, err = uc.conversationRepo.GetConversationByID(conversationID, user.ID)
		if err != nil {
			return fmt.Errorf("failed to get conversation: %w", err)
		}
		if conversation == nil {
			return fmt.Errorf("conversation not found or not owned by user")
		}
	} else {
		// Fetch or start a new conversation for the user
		conversation, err = uc.conversationRepo.GetOrCreateActiveConversation(user.ID)
		if err != nil {
			log.Printf("[Context] Error getting/creating conversation: %v", err)
			return fmt.Errorf("failed to get conversation: %w", err)
		}
	}

	// Try loading context from cache first
	var messages []dto.OllamaChatMessage
	cachedMessages, err := uc.conversationCache.GetConversationFromCache(ctx, conversation.ID)
	
	if err != nil || len(cachedMessages) == 0 {
		log.Printf("[Context] Cache miss, loading from DB for conversation %s", conversation.ID)
		dbMessages, err := uc.conversationRepo.GetConversationMessages(conversation.ID, 50)
		if err != nil {
			log.Printf("[Context] Error loading messages from DB: %v", err)
		} else {
			messages = convertMessagesToDTO(dbMessages)
		}
	} else {
		log.Printf("[Context] Cache hit, loaded %d messages for conversation %s", len(cachedMessages), conversation.ID)
		messages = cachedMessages
	}

	// Add user details to the system prompt to personalize the interaction
	personalizedPrompt := fmt.Sprintf("%s\n\nActive user: %s (Email: %s)", 
		uc.systemPrompt, user.Name, user.Email)

	// Ensure the system prompt is always at the start
	if len(messages) == 0 {
		messages = []dto.OllamaChatMessage{
			{Role: "system", Content: personalizedPrompt},
		}
	} else {
		if messages[0].Role == "system" {
			messages[0].Content = personalizedPrompt
		} else {
			messages = append([]dto.OllamaChatMessage{
				{Role: "system", Content: personalizedPrompt},
			}, messages...)
		}
	}

	// Add the new user message to the stack
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

	// Stream the first response and gather chunks
	err = uc.ollamaClient.StreamChatRequest(ctx, request, func(chunk dto.OllamaChatResponse) error {
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

	// Persist the conversation context asynchronously
	go uc.saveConversationContext(context.Background(), conversation, user.ID, userPrompt, fullContent, allToolCalls, messages)

	return nil
}

// ... (getAvailableTools and convertMessagesToDTO omitted for brevity but remain unchanged)

// saveConversationContext persists the chat session to both the database and cache
func (uc *StreamChatUsecase) saveConversationContext(
	ctx context.Context,
	conversation *entity.Conversation,
	userID string,
	userPrompt string,
	assistantContent string,
	toolCalls []dto.ToolCall,
	allMessages []dto.OllamaChatMessage,
) {
	// Generate title if it doesn't exist
	if conversation.Title == nil || *conversation.Title == "" {
		title, err := uc.ollamaClient.GenerateTitle(ctx, userPrompt)
		if err == nil && title != "" {
			if err := uc.conversationRepo.UpdateConversation(conversation.ID, userID, title); err != nil {
				log.Printf("[Context] Error updating conversation title: %v", err)
			} else {
				log.Printf("[Context] Generated title for conversation %s: %s", conversation.ID, title)
			}
		}
	}

	// Store user input
	userMsg := &entity.Message{
		ConversationID: conversation.ID,
		Role:           "user",
		Content:        userPrompt,
	}
	
	if err := uc.conversationRepo.SaveMessage(userMsg); err != nil {
		log.Printf("[Context] Error saving user message: %v", err)
	}
	
	// Store assistant response
	assistantMsg := &entity.Message{
		ConversationID: conversation.ID,
		Role:           "assistant",
		Content:        assistantContent,
	}
	
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
	
	// Update cache with the latest state
	if err := uc.conversationCache.SaveConversationToCache(ctx, conversation.ID, allMessages); err != nil {
		log.Printf("[Context] Error updating cache: %v", err)
	} else {
		log.Printf("[Context] Cache updated for conversation %s", conversation.ID)
	}
}

