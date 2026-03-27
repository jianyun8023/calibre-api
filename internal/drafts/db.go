package drafts

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

func InitDB(dbPath string) (*sql.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %v", err)
	}

	// Create tables if not exist
	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	log.Infof("Initialized SQLite database for drafts at %s", dbPath)
	return db, nil
}

func createTables(db *sql.DB) error {
	draftsTable := `
	CREATE TABLE IF NOT EXISTS book_drafts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		book_id TEXT NOT NULL,
		action TEXT NOT NULL,
		data TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	historyTable := `
	CREATE TABLE IF NOT EXISTS book_draft_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		draft_id INTEGER NOT NULL,
		book_id TEXT NOT NULL,
		action TEXT NOT NULL,
		data TEXT,
		status TEXT NOT NULL,
		processed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := db.Exec(draftsTable); err != nil {
		return err
	}
	if _, err := db.Exec(historyTable); err != nil {
		return err
	}

	return nil
}
