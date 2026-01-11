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

// GetOrCreateActiveConversation obtiene la conversación más reciente del usuario o crea una nueva
func (r *conversationRepository) GetOrCreateActiveConversation(userID string) (*entity.Conversation, error) {
	// Intentar obtener la conversación más reciente
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
		// No existe conversación, crear una nueva
		return r.CreateConversation(userID)
	}

	if err != nil {
		return nil, fmt.Errorf("error getting active conversation: %w", err)
	}

	return &conv, nil
}

// GetConversationMessages obtiene los últimos N mensajes de una conversación
func (r *conversationRepository) GetConversationMessages(conversationID string, limit int) ([]entity.Message, error) {
	query := `SELECT id, conversation_id, role, content, tool_calls, created_at 
	          FROM chat_messages 
	          WHERE conversation_id = ? 
	          ORDER BY created_at ASC 
	          LIMIT ?`

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

// SaveMessage guarda un nuevo mensaje en la base de datos
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

	// Actualizar el updated_at de la conversación
	updateQuery := `UPDATE chat_conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = r.db.Exec(updateQuery, message.ConversationID)
	if err != nil {
		return fmt.Errorf("error updating conversation timestamp: %w", err)
	}

	return nil
}

// CreateConversation crea una nueva conversación para un usuario
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

	// Obtener la conversación completa con timestamps
	return r.getConversationByID(conv.ID)
}

// getConversationByID obtiene una conversación por su ID
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
