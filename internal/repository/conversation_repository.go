package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/metalpoch/local-synapse/internal/entity"
)

type ConversationRepository interface {
	GetOrCreateActiveConversation(userID string) (*entity.Conversation, error)
	GetConversationMessages(conversationID string, limit int) ([]entity.Message, error)
	SaveMessage(message *entity.Message) error
	CreateConversation(userID string) (*entity.Conversation, error)
}

type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

// GetOrCreateActiveConversation returns the user's most recent conversation or starts a new one.
func (r *conversationRepository) GetOrCreateActiveConversation(userID string) (*entity.Conversation, error) {
	query := `SELECT id, user_id, title, created_at, updated_at 
	          FROM chat_conversations 
	          WHERE user_id = ? 
	          ORDER BY updated_at DESC 
	          LIMIT 1`

	var conv entity.Conversation
	err := r.db.QueryRow(query, userID).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return r.CreateConversation(userID)
	}

	if err != nil {
		return nil, fmt.Errorf("error getting active conversation: %w", err)
	}

	return &conv, nil
}

// GetConversationMessages retrieves the last N messages from a specific conversation.
func (r *conversationRepository) GetConversationMessages(conversationID string, limit int) ([]entity.Message, error) {
	query := `SELECT id, conversation_id, role, content, tool_calls, created_at 
	          FROM (
	              SELECT id, conversation_id, role, content, tool_calls, created_at 
	              FROM chat_messages 
	              WHERE conversation_id = ? 
	              ORDER BY created_at DESC 
	              LIMIT ?
	          ) ORDER BY created_at ASC`

	rows, err := r.db.Query(query, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying messages: %w", err)
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&msg.ToolCalls,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// SaveMessage persists a new message and updates the conversation timestamp.
func (r *conversationRepository) SaveMessage(message *entity.Message) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	query := `INSERT INTO chat_messages (id, conversation_id, role, content, tool_calls) 
	          VALUES (?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		message.ID,
		message.ConversationID,
		message.Role,
		message.Content,
		message.ToolCalls,
	)

	if err != nil {
		return fmt.Errorf("error saving message: %w", err)
	}

	updateQuery := `UPDATE chat_conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = r.db.Exec(updateQuery, message.ConversationID)
	if err != nil {
		return fmt.Errorf("error updating conversation timestamp: %w", err)
	}

	return nil
}

// CreateConversation initializes a new empty conversation for a user.
func (r *conversationRepository) CreateConversation(userID string) (*entity.Conversation, error) {
	conv := &entity.Conversation{
		ID:     uuid.New().String(),
		UserID: userID,
	}

	query := `INSERT INTO chat_conversations (id, user_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, conv.ID, conv.UserID)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation: %w", err)
	}

	return r.getConversationByID(conv.ID)
}

// getConversationByID fetches a single conversation record by its ID.
func (r *conversationRepository) getConversationByID(id string) (*entity.Conversation, error) {
	query := `SELECT id, user_id, title, created_at, updated_at 
	          FROM chat_conversations 
	          WHERE id = ?`

	var conv entity.Conversation
	err := r.db.QueryRow(query, id).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Title,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting conversation: %w", err)
	}

	return &conv, nil
}
