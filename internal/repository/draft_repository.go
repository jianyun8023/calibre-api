package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	DraftStatusPending   DraftStatus = "pending"
	DraftStatusProcessing DraftStatus = "processing"
	DraftStatusApplied   DraftStatus = "applied"
	DraftStatusRejected  DraftStatus = "rejected"
	DraftStatusExpired   DraftStatus = "expired"
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
	GetPendingDrafts(ctx context.Context, limit, offset int) ([]BookDraft, int64, error)
	GetDraftByID(ctx context.Context, id int64) (*BookDraft, error)
	GetPendingDraftByBookIDAndAction(ctx context.Context, bookID string, action DraftType) (*BookDraft, error)
	GetPendingDraftsByBookIDsAndAction(ctx context.Context, bookIDs []string, action DraftType) ([]BookDraft, error)
	GetPendingDraftsBefore(ctx context.Context, before time.Time) ([]BookDraft, error)
	ResetStuckProcessingDrafts(ctx context.Context, threshold time.Time) (int64, error)
	UpdateDraftStatus(ctx context.Context, id int64, status DraftStatus) error
	UpdateDraftData(ctx context.Context, id int64, data string) error
	DeleteDraft(ctx context.Context, id int64) error

	// Atomic Operations
	UpdateDraftStatusIfPending(ctx context.Context, id int64, newStatus DraftStatus) (bool, error)
	ApplyDraftSuccess(ctx context.Context, draft *BookDraft, newStatus DraftStatus) error
	ExpireDraftAtomically(ctx context.Context, draft *BookDraft) error

	// History
	CreateHistory(ctx context.Context, history *BookDraftHistory) (int64, error)
	GetHistory(ctx context.Context, limit, offset int) ([]BookDraftHistory, int64, error)
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

func (r *sqliteDraftRepository) GetPendingDrafts(ctx context.Context, limit, offset int) ([]BookDraft, int64, error) {
	// First get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM book_drafts WHERE status = ?`
	if err := r.db.QueryRowContext(ctx, countQuery, DraftStatusPending).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Then get paginated data
	query := `SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, DraftStatusPending, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var drafts []BookDraft
	for rows.Next() {
		var draft BookDraft
		if err := rows.Scan(&draft.ID, &draft.BookID, &draft.Action, &draft.Data, &draft.Status, &draft.CreatedAt); err != nil {
			return nil, 0, err
		}
		drafts = append(drafts, draft)
	}
	return drafts, total, nil
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

func (r *sqliteDraftRepository) GetPendingDraftByBookIDAndAction(ctx context.Context, bookID string, action DraftType) (*BookDraft, error) {
	query := `SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE book_id = ? AND action = ? AND status = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, bookID, action, DraftStatusPending)

	var draft BookDraft
	err := row.Scan(&draft.ID, &draft.BookID, &draft.Action, &draft.Data, &draft.Status, &draft.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, this is fine
		}
		return nil, err
	}
	return &draft, nil
}

func (r *sqliteDraftRepository) GetPendingDraftsByBookIDsAndAction(ctx context.Context, bookIDs []string, action DraftType) ([]BookDraft, error) {
	if len(bookIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(bookIDs))
	args := make([]interface{}, 0, len(bookIDs)+2)
	for i, id := range bookIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, action, DraftStatusPending)

	query := fmt.Sprintf(`SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE book_id IN (%s) AND action = ? AND status = ?`, strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, query, args...)
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

func (r *sqliteDraftRepository) GetPendingDraftsBefore(ctx context.Context, before time.Time) ([]BookDraft, error) {
	query := `SELECT id, book_id, action, data, status, created_at FROM book_drafts WHERE status = ? AND created_at < ?`
	rows, err := r.db.QueryContext(ctx, query, DraftStatusPending, before)
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

func (r *sqliteDraftRepository) ResetStuckProcessingDrafts(ctx context.Context, threshold time.Time) (int64, error) {
	// A draft in "processing" state that was created before the threshold is considered stuck.
	// Since we don't have an 'updated_at' column, and processing should be very fast,
	// we just use the 'created_at' column and assume a long enough threshold.
	query := `UPDATE book_drafts SET status = ? WHERE status = ? AND created_at < ?`
	res, err := r.db.ExecContext(ctx, query, DraftStatusPending, DraftStatusProcessing, threshold)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *sqliteDraftRepository) UpdateDraftStatus(ctx context.Context, id int64, status DraftStatus) error {
	query := `UPDATE book_drafts SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *sqliteDraftRepository) UpdateDraftData(ctx context.Context, id int64, data string) error {
	query := `UPDATE book_drafts SET data = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, data, id)
	return err
}

func (r *sqliteDraftRepository) DeleteDraft(ctx context.Context, id int64) error {
	query := `DELETE FROM book_drafts WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqliteDraftRepository) UpdateDraftStatusIfPending(ctx context.Context, id int64, newStatus DraftStatus) (bool, error) {
	query := `UPDATE book_drafts SET status = ? WHERE id = ? AND status = ?`
	res, err := r.db.ExecContext(ctx, query, newStatus, id, DraftStatusPending)
	if err != nil {
		return false, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *sqliteDraftRepository) ApplyDraftSuccess(ctx context.Context, draft *BookDraft, newStatus DraftStatus) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update Draft Status (from processing to applied/rejected) with defensive check
	updateQuery := `UPDATE book_drafts SET status = ? WHERE id = ? AND status = ?`
	res, err := tx.ExecContext(ctx, updateQuery, newStatus, draft.ID, DraftStatusProcessing)
	if err != nil {
		return err // Rollback
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		// If 0 rows affected, it was not in processing state (defensive check failed)
		return fmt.Errorf("draft %d was not in processing state, affected %d rows", draft.ID, affected)
	}

	// 2. Record History
	historyQuery := `INSERT INTO book_draft_history (draft_id, book_id, action, data, status) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, historyQuery, draft.ID, draft.BookID, draft.Action, draft.Data, newStatus); err != nil {
		return err // Rollback
	}

	return tx.Commit()
}

func (r *sqliteDraftRepository) ExpireDraftAtomically(ctx context.Context, draft *BookDraft) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Update Draft Status (from pending to expired)
	updateQuery := `UPDATE book_drafts SET status = ? WHERE id = ? AND status = ?`
	res, err := tx.ExecContext(ctx, updateQuery, DraftStatusExpired, draft.ID, DraftStatusPending)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("draft %d was not in pending state, affected %d rows", draft.ID, affected)
	}

	// 2. Record History
	historyQuery := `INSERT INTO book_draft_history (draft_id, book_id, action, data, status) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.ExecContext(ctx, historyQuery, draft.ID, draft.BookID, draft.Action, draft.Data, DraftStatusExpired); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *sqliteDraftRepository) CreateHistory(ctx context.Context, history *BookDraftHistory) (int64, error) {
	query := `INSERT INTO book_draft_history (draft_id, book_id, action, data, status) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, history.DraftID, history.BookID, history.Action, history.Data, history.Status)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqliteDraftRepository) GetHistory(ctx context.Context, limit, offset int) ([]BookDraftHistory, int64, error) {
	// First get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM book_draft_history`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, draft_id, book_id, action, data, status, processed_at FROM book_draft_history ORDER BY processed_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []BookDraftHistory
	for rows.Next() {
		var h BookDraftHistory
		if err := rows.Scan(&h.ID, &h.DraftID, &h.BookID, &h.Action, &h.Data, &h.Status, &h.ProcessedAt); err != nil {
			return nil, 0, err
		}
		histories = append(histories, h)
	}
	return histories, total, nil
}
