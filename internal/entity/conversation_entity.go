package entity

import "time"

type Conversation struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Title     *string   `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Message struct {
	ID             string    `db:"id"`
	ConversationID string    `db:"conversation_id"`
	Role           string    `db:"role"`
	Content        string    `db:"content"`
	ToolCalls      *string   `db:"tool_calls"` // JSON serializado
	CreatedAt      time.Time `db:"created_at"`
}
