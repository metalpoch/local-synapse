package sqlite

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

	for _, query := range [2]string{tableUsers, triggerUsers} {
		if _, err := db.Exec(query); err != nil {
			panic(err)
		}
	}

}
