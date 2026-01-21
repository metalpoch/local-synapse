package ollama

import (
	"context"

	"github.com/metalpoch/local-synapse/internal/repository"
)

type DeleteConversation struct {
	conversationRepo repository.ConversationRepository
}

func NewDeleteConversation(cr repository.ConversationRepository) *DeleteConversation {
	return &DeleteConversation{cr}
}

func (uc *DeleteConversation) Execute(ctx context.Context, id string, userID string) error {
	return uc.conversationRepo.DeleteConversation(id, userID)
}
