package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/valkey-io/valkey-go"
)

const (
	conversationKeyPrefix = "chat:conversation:"
	conversationTTL       = 30 * time.Minute
)

type ConversationCache interface {
	GetConversationFromCache(ctx context.Context, userID string) ([]dto.OllamaChatMessage, error)
	SaveConversationToCache(ctx context.Context, userID string, messages []dto.OllamaChatMessage) error
	InvalidateConversation(ctx context.Context, userID string) error
}

type conversationCache struct {
	client valkey.Client
}

func NewConversationCache(client valkey.Client) ConversationCache {
	return &conversationCache{client: client}
}

// GetConversationFromCache retrieves the conversation history from the Valkey storage.
func (c *conversationCache) GetConversationFromCache(ctx context.Context, userID string) ([]dto.OllamaChatMessage, error) {
	if c.client == nil {
		return nil, fmt.Errorf("valkey client is nil")
	}

	key := conversationKeyPrefix + userID
	cmd := c.client.B().Get().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	data, err := result.ToString()
	if err != nil {
		return nil, fmt.Errorf("conversation not found in cache: %w", err)
	}

	var messages []dto.OllamaChatMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return nil, fmt.Errorf("error unmarshaling cached conversation: %w", err)
	}

	return messages, nil
}

// SaveConversationToCache stores the conversation history with a TTL.
func (c *conversationCache) SaveConversationToCache(ctx context.Context, userID string, messages []dto.OllamaChatMessage) error {
	if c.client == nil {
		return fmt.Errorf("valkey client is nil")
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("error marshaling conversation: %w", err)
	}

	key := conversationKeyPrefix + userID
	cmd := c.client.B().Set().Key(key).Value(string(data)).Ex(conversationTTL).Build()
	result := c.client.Do(ctx, cmd)

	if err := result.Error(); err != nil {
		return fmt.Errorf("error saving conversation to cache: %w", err)
	}

	return nil
}

// InvalidateConversation removes the conversation record from the cache.
func (c *conversationCache) InvalidateConversation(ctx context.Context, userID string) error {
	if c.client == nil {
		return fmt.Errorf("valkey client is nil")
	}

	key := conversationKeyPrefix + userID
	cmd := c.client.B().Del().Key(key).Build()
	result := c.client.Do(ctx, cmd)

	if err := result.Error(); err != nil {
		return fmt.Errorf("error invalidating conversation: %w", err)
	}

	return nil
}
