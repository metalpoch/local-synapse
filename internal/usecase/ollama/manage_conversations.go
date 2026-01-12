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

// ---

type CreateConversation struct {
	conversationRepo repository.ConversationRepository
}

func NewCreateConversation(cr repository.ConversationRepository) *CreateConversation {
	return &CreateConversation{cr}
}

func (uc *CreateConversation) Execute(ctx context.Context, userID string) (*entity.Conversation, error) {
	return uc.conversationRepo.CreateConversation(userID)
}

// ---

type DeleteConversation struct {
	conversationRepo repository.ConversationRepository
}

func NewDeleteConversation(cr repository.ConversationRepository) *DeleteConversation {
	return &DeleteConversation{cr}
}

func (uc *DeleteConversation) Execute(ctx context.Context, id string, userID string) error {
	return uc.conversationRepo.DeleteConversation(id, userID)
}
