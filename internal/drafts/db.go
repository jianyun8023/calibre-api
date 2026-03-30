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

	// Connect to SQLite with busy_timeout parameter
	db, err := sql.Open("sqlite3", dbPath+"?_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %v", err)
	}

	// Set WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Warnf("Failed to enable WAL mode for drafts db: %v", err)
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

	draftsIndex := `CREATE INDEX IF NOT EXISTS idx_book_drafts_status ON book_drafts(status, created_at);`
	draftsLookupIndex := `CREATE INDEX IF NOT EXISTS idx_book_drafts_lookup ON book_drafts(book_id, action, status);`
	historyIndex := `CREATE INDEX IF NOT EXISTS idx_book_draft_history_time ON book_draft_history(processed_at DESC);`

	if _, err := db.Exec(draftsTable); err != nil {
		return err
	}
	if _, err := db.Exec(historyTable); err != nil {
		return err
	}
	if _, err := db.Exec(draftsIndex); err != nil {
		log.Warnf("Failed to create index idx_book_drafts_status: %v", err)
	}
	if _, err := db.Exec(draftsLookupIndex); err != nil {
		log.Warnf("Failed to create index idx_book_drafts_lookup: %v", err)
	}
	if _, err := db.Exec(historyIndex); err != nil {
		log.Warnf("Failed to create index idx_book_draft_history_time: %v", err)
	}

	return nil
}
