package ollama

import (
	"context"

	"github.com/metalpoch/local-synapse/internal/entity"
	"github.com/metalpoch/local-synapse/internal/repository"
)

type ListConversations struct {
	conversationRepo repository.ConversationRepository
}

func NewListConversations(cr repository.ConversationRepository) *ListConversations {
	return &ListConversations{cr}
}

func (uc *ListConversations) Execute(ctx context.Context, userID string) ([]entity.Conversation, error) {
	return uc.conversationRepo.GetConversations(userID)
}
