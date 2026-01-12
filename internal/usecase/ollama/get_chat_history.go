package ollama

import (
	"context"
	"fmt"
	"log"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/repository"
)

type GetChatHistory struct {
	conversationRepo repository.ConversationRepository
}

func NewGetChatHistory(cr repository.ConversationRepository) *GetChatHistory {
	return &GetChatHistory{conversationRepo: cr}
}

func (uc *GetChatHistory) Execute(ctx context.Context, userID string) ([]dto.OllamaChatMessage, error) {
	conversation, err := uc.conversationRepo.GetOrCreateActiveConversation(userID)
	if err != nil {
		log.Printf("[Context] Error getting conversation: %v", err)
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	dbMessages, err := uc.conversationRepo.GetFullConversationHistory(conversation.ID)
	if err != nil {
		log.Printf("[Context] Error loading full history from DB: %v", err)
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	return convertMessagesToDTO(dbMessages), nil
}
