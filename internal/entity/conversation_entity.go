package entity

import "time"

type Conversation struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"userId"`
	Title     *string   `db:"title" json:"title"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type Message struct {
	ID             string    `db:"id" json:"id"`
	ConversationID string    `db:"conversation_id" json:"conversationId"`
	Role           string    `db:"role" json:"role"`
	Content        string    `db:"content" json:"content"`
	ToolCalls      *string   `db:"tool_calls" json:"toolCalls,omitempty"` // JSON serializado
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
}
