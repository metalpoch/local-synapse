package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewSqliteClient(addr string) *sql.DB {
	db, err := sql.Open("sqlite3", addr)
	if err != nil {
		panic(err)
	}

	initDB(db)

	return db
}

func initDB(db *sql.DB) {
	tableUsers := `CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    auth_provider TEXT NOT NULL,
    code_provider INTEGER,
    image_url TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	triggerUsers := `CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
	AFTER UPDATE ON users 
	FOR EACH ROW 
	BEGIN
  	UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	END;`

	tableConversations := `CREATE TABLE IF NOT EXISTS chat_conversations (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	triggerConversations := `CREATE TRIGGER IF NOT EXISTS update_conversations_updated_at 
	AFTER UPDATE ON chat_conversations 
	FOR EACH ROW 
	BEGIN
  	UPDATE chat_conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	END;`

	tableMessages := `CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('system', 'user', 'assistant', 'tool')),
    content TEXT NOT NULL,
    tool_calls TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (conversation_id) REFERENCES chat_conversations(id) ON DELETE CASCADE
	);`

	indexMessages := `CREATE INDEX IF NOT EXISTS idx_messages_conversation 
	ON chat_messages(conversation_id, created_at DESC);`

	queries := []string{
		tableUsers,
		triggerUsers,
		tableConversations,
		triggerConversations,
		tableMessages,
		indexMessages,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			panic(err)
		}
	}

}
