package ollama

import (
	"context"
	"fmt"
	"log"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/entity"
	"github.com/metalpoch/local-synapse/internal/repository"
)

type GetChatHistory struct {
	conversationRepo repository.ConversationRepository
}

func NewGetChatHistory(cr repository.ConversationRepository) *GetChatHistory {
	return &GetChatHistory{conversationRepo: cr}
}

func (uc *GetChatHistory) Execute(ctx context.Context, userID, conversationID string) ([]dto.OllamaChatMessage, error) {
	var conversation *entity.Conversation
	var err error

	if conversationID != "" {
		conversation, err = uc.conversationRepo.GetConversationByID(conversationID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
		if conversation == nil {
			return nil, fmt.Errorf("conversation not found or not owned by user")
		}
	} else {
		conversation, err = uc.conversationRepo.GetOrCreateActiveConversation(userID)
		if err != nil {
			log.Printf("[Context] Error getting conversation: %v", err)
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
	}

	dbMessages, err := uc.conversationRepo.GetFullConversationHistory(conversation.ID)
	if err != nil {
		log.Printf("[Context] Error loading full history from DB: %v", err)
		return nil, fmt.Errorf("failed to load history: %w", err)
	}

	return convertMessagesToDTO(dbMessages), nil
}
