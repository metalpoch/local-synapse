package ollama

import (
	"context"

	"github.com/metalpoch/local-synapse/internal/repository"
)

type RenameConversation struct {
	conversationRepo repository.ConversationRepository
}

func NewRenameConversation(cr repository.ConversationRepository) *RenameConversation {
	return &RenameConversation{cr}
}

func (uc *RenameConversation) Execute(ctx context.Context, id string, userID string, title string) error {
	return uc.conversationRepo.UpdateConversation(id, userID, title)
}
