package repository

import (
	"context"
	"database/sql"
	"time"
)

// DraftType defines the type of a draft action
type DraftType string

const (
	DraftActionDelete DraftType = "delete"
	DraftActionUpdate DraftType = "update"
)

// DraftStatus defines the status of a draft
type DraftStatus string

const (
	DraftStatusPending  DraftStatus = "pending"
	DraftStatusApplied  DraftStatus = "applied"
	DraftStatusRejected DraftStatus = "rejected"
)

// BookDraft represents a pending draft change for a book
type BookDraft struct {
	ID        int64       `json:"id"`
	BookID    string      `json:"book_id"`
	Action    DraftType   `json:"action"`
	Data      string      `json:"data"`
	Status    DraftStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// BookDraftHistory represents the history of a draft change
type BookDraftHistory struct {
	ID          int64       `json:"id"`
	DraftID     int64       `json:"draft_id"`
	BookID      string      `json:"book_id"`
	Action      DraftType   `json:"action"`
	Data        string      `json:"data"`
	Status      DraftStatus `json:"status"`
	ProcessedAt time.Time   `json:"processed_at"`
}

// DraftRepository handles data access for drafts
type DraftRepository interface {
	// Drafts
	CreateDraft(ctx context.Context, draft *BookDraft) (int64, error)
	GetPendingDrafts(ctx context.Context) ([]BookDraft, error)
	GetDraftByID(ctx context.Context, id int64) (*BookDraft, error)
	UpdateDraftStatus(ctx context.Context, id int64, status DraftStatus) error
	DeleteDraft(ctx context.Context, id int64) error

	// History
	CreateHistory(ctx context.Context, history *BookDraftHistory) (int64, error)
	GetHistory(ctx context.Context, limit, offset int) ([]BookDraftHistory, error)
}

type sqliteDraftRepository struct {
	db *sql.DB
}

// NewSqliteDraftRepository creates a new SQLite-based draft repository
func NewSqliteDraftRepository(db *sql.DB) DraftRepository {
	return &sqliteDraftRepository{
		db: db,
	}
}

func (r *sqliteDraftRepository) CreateDraft(ctx context.Context, draft *BookDraft) (int64, error) {
	query := `INSERT INTO book_drafts (book_id, action, data, status) VALUES (?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, draft.BookID, draft.Action, draft.Data, draft.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqliteDraftRepository) GetPendingDrafts(ctx context.Context) ([]BookDraft, error) {
	query := `SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE status = ? ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, DraftStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []BookDraft
	for rows.Next() {
		var draft BookDraft
		if err := rows.Scan(&draft.ID, &draft.BookID, &draft.Action, &draft.Data, &draft.Status, &draft.CreatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}
	return drafts, nil
}

func (r *sqliteDraftRepository) GetDraftByID(ctx context.Context, id int64) (*BookDraft, error) {
	query := `SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var draft BookDraft
	err := row.Scan(&draft.ID, &draft.BookID, &draft.Action, &draft.Data, &draft.Status, &draft.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // or define a specific error
		}
		return nil, err
	}
	return &draft, nil
}

func (r *sqliteDraftRepository) UpdateDraftStatus(ctx context.Context, id int64, status DraftStatus) error {
	query := `UPDATE book_drafts SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *sqliteDraftRepository) DeleteDraft(ctx context.Context, id int64) error {
	query := `DELETE FROM book_drafts WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqliteDraftRepository) CreateHistory(ctx context.Context, history *BookDraftHistory) (int64, error) {
	query := `INSERT INTO book_draft_history (draft_id, book_id, action, data, status) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, history.DraftID, history.BookID, history.Action, history.Data, history.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqliteDraftRepository) GetHistory(ctx context.Context, limit, offset int) ([]BookDraftHistory, error) {
	query := `SELECT id, draft_id, book_id, action, data, status, processed_at FROM book_draft_history ORDER BY processed_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []BookDraftHistory
	for rows.Next() {
		var h BookDraftHistory
		if err := rows.Scan(&h.ID, &h.DraftID, &h.BookID, &h.Action, &h.Data, &h.Status, &h.ProcessedAt); err != nil {
			return nil, err
		}
		histories = append(histories, h)
	}
	return histories, nil
}
