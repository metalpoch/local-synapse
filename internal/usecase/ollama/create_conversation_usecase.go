package ollama

import (
	"context"

	"github.com/metalpoch/local-synapse/internal/entity"
	"github.com/metalpoch/local-synapse/internal/repository"
)

type CreateConversation struct {
	conversationRepo repository.ConversationRepository
}

func NewCreateConversation(cr repository.ConversationRepository) *CreateConversation {
	return &CreateConversation{cr}
}

func (uc *CreateConversation) Execute(ctx context.Context, userID string) (*entity.Conversation, error) {
	return uc.conversationRepo.CreateConversation(userID)
}
