package governance

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var ErrConcurrentModification = errors.New("concurrent modification detected")
var ErrDraftNotFound = errors.New("draft not found")
var ErrChangelogNotFound = errors.New("changelog not found")

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := runMigrations(conn); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func runMigrations(conn *sql.DB) error {
	migrationFiles, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range migrationFiles {
		if file.IsDir() {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		if _, err := conn.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file.Name(), err)
		}
	}

	return nil
}

func (db *DB) CreateDraft(draft *MetadataDraft) error {
	result, err := db.conn.Exec(`
		INSERT INTO metadata_drafts 
			(book_id, book_title, field, old_value, new_value, source, confidence, 
			 confidence_breakdown, flags, status, suggested_action, session_id, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
	`,
		draft.BookID, draft.BookTitle, draft.Field, draft.OldValue, draft.NewValue,
		draft.Source, draft.Confidence, ConfidenceBreakdownToJSON(draft.ConfidenceBreakdown),
		FlagsToJSON(draft.Flags), draft.Status, draft.SuggestedAction, draft.SessionID,
	)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	draft.ID = id
	draft.Version = 1
	draft.CreatedAt = time.Now()

	return nil
}

func (db *DB) GetDraft(id int64) (*MetadataDraft, error) {
	var draft MetadataDraft
	var confidenceBreakdown, flags sql.NullString
	var reviewedAt, appliedAt sql.NullTime
	var reviewedBy sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, book_id, book_title, field, old_value, new_value, source, 
			   confidence, confidence_breakdown, flags, status, suggested_action,
			   session_id, version, created_at, reviewed_at, reviewed_by, applied_at
		FROM metadata_drafts WHERE id = ?
	`, id).Scan(
		&draft.ID, &draft.BookID, &draft.BookTitle, &draft.Field, &draft.OldValue,
		&draft.NewValue, &draft.Source, &draft.Confidence, &confidenceBreakdown,
		&flags, &draft.Status, &draft.SuggestedAction, &draft.SessionID,
		&draft.Version, &draft.CreatedAt, &reviewedAt, &reviewedBy, &appliedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrDraftNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get draft: %w", err)
	}

	if confidenceBreakdown.Valid {
		draft.ConfidenceBreakdown = JSONToConfidenceBreakdown(confidenceBreakdown.String)
	}
	if flags.Valid {
		draft.Flags = JSONToFlags(flags.String)
	}
	if reviewedAt.Valid {
		draft.ReviewedAt = &reviewedAt.Time
	}
	if reviewedBy.Valid {
		draft.ReviewedBy = reviewedBy.String
	}
	if appliedAt.Valid {
		draft.AppliedAt = &appliedAt.Time
	}

	return &draft, nil
}

func (db *DB) ListDrafts(filter DraftFilter) ([]*MetadataDraft, int, error) {
	query := `SELECT id, book_id, book_title, field, old_value, new_value, source, 
			  confidence, confidence_breakdown, flags, status, suggested_action,
			  session_id, version, created_at, reviewed_at, reviewed_by, applied_at
			  FROM metadata_drafts WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM metadata_drafts WHERE 1=1`
	args := []interface{}{}

	if filter.Status != nil {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, *filter.Status)
	}
	if filter.ConfidenceMin != nil {
		query += " AND confidence >= ?"
		countQuery += " AND confidence >= ?"
		args = append(args, *filter.ConfidenceMin)
	}
	if filter.ConfidenceMax != nil {
		query += " AND confidence < ?"
		countQuery += " AND confidence < ?"
		args = append(args, *filter.ConfidenceMax)
	}
	if filter.HasFlags != nil && *filter.HasFlags {
		query += " AND flags != '[]' AND flags IS NOT NULL AND flags != ''"
		countQuery += " AND flags != '[]' AND flags IS NOT NULL AND flags != ''"
	}
	if filter.SessionID != "" {
		query += " AND session_id = ?"
		countQuery += " AND session_id = ?"
		args = append(args, filter.SessionID)
	}
	if filter.BookID != nil {
		query += " AND book_id = ?"
		countQuery += " AND book_id = ?"
		args = append(args, *filter.BookID)
	}
	if filter.Field != nil {
		query += " AND field = ?"
		countQuery += " AND field = ?"
		args = append(args, *filter.Field)
	}

	var total int
	if err := db.conn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count drafts: %w", err)
	}

	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list drafts: %w", err)
	}
	defer rows.Close()

	var drafts []*MetadataDraft
	for rows.Next() {
		var draft MetadataDraft
		var confidenceBreakdown, flags sql.NullString
		var reviewedAt, appliedAt sql.NullTime
		var reviewedBy sql.NullString

		if err := rows.Scan(
			&draft.ID, &draft.BookID, &draft.BookTitle, &draft.Field, &draft.OldValue,
			&draft.NewValue, &draft.Source, &draft.Confidence, &confidenceBreakdown,
			&flags, &draft.Status, &draft.SuggestedAction, &draft.SessionID,
			&draft.Version, &draft.CreatedAt, &reviewedAt, &reviewedBy, &appliedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan draft: %w", err)
		}

		if confidenceBreakdown.Valid {
			draft.ConfidenceBreakdown = JSONToConfidenceBreakdown(confidenceBreakdown.String)
		}
		if flags.Valid {
			draft.Flags = JSONToFlags(flags.String)
		}
		if reviewedAt.Valid {
			draft.ReviewedAt = &reviewedAt.Time
		}
		if reviewedBy.Valid {
			draft.ReviewedBy = reviewedBy.String
		}
		if appliedAt.Valid {
			draft.AppliedAt = &appliedAt.Time
		}

		drafts = append(drafts, &draft)
	}

	return drafts, total, nil
}

func (db *DB) UpdateDraftStatus(id int64, status DraftStatus, version int, reviewedBy string) error {
	now := time.Now()
	result, err := db.conn.Exec(`
		UPDATE metadata_drafts 
		SET status = ?, reviewed_at = ?, reviewed_by = ?, version = version + 1
		WHERE id = ? AND version = ? AND status = 'pending'
	`, status, now, reviewedBy, id, version)

	if err != nil {
		return fmt.Errorf("failed to update draft status: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrConcurrentModification
	}
	return nil
}

func (db *DB) MarkDraftApplied(id int64) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE metadata_drafts 
		SET status = 'applied', applied_at = ?, version = version + 1
		WHERE id = ? AND status = 'approved'
	`, now, id)
	return err
}

func (db *DB) UpdateDraftValue(id int64, newValue string, version int) error {
	result, err := db.conn.Exec(`
		UPDATE metadata_drafts 
		SET new_value = ?, version = version + 1
		WHERE id = ? AND version = ? AND status = 'pending'
	`, newValue, id, version)

	if err != nil {
		return fmt.Errorf("failed to update draft value: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrConcurrentModification
	}
	return nil
}

func (db *DB) CreateChangelog(log *MetadataChangelog) error {
	result, err := db.conn.Exec(`
		INSERT INTO metadata_changelog 
			(book_id, book_title, field, old_value, new_value, source, draft_id, applied_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		log.BookID, log.BookTitle, log.Field, log.OldValue, log.NewValue,
		log.Source, log.DraftID, log.AppliedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to create changelog: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	log.ID = id
	log.AppliedAt = time.Now()

	return nil
}

func (db *DB) GetChangelog(id int64) (*MetadataChangelog, error) {
	var log MetadataChangelog
	var draftID sql.NullInt64
	var revertedAt sql.NullTime
	var revertedBy, revertReason, appliedBy sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, book_id, book_title, field, old_value, new_value, source, draft_id,
			   applied_at, applied_by, reverted_at, reverted_by, revert_reason
		FROM metadata_changelog WHERE id = ?
	`, id).Scan(
		&log.ID, &log.BookID, &log.BookTitle, &log.Field, &log.OldValue,
		&log.NewValue, &log.Source, &draftID, &log.AppliedAt, &appliedBy,
		&revertedAt, &revertedBy, &revertReason,
	)

	if err == sql.ErrNoRows {
		return nil, ErrChangelogNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get changelog: %w", err)
	}

	if draftID.Valid {
		log.DraftID = &draftID.Int64
	}
	if appliedBy.Valid {
		log.AppliedBy = appliedBy.String
	}
	if revertedAt.Valid {
		log.RevertedAt = &revertedAt.Time
	}
	if revertedBy.Valid {
		log.RevertedBy = revertedBy.String
	}
	if revertReason.Valid {
		log.RevertReason = revertReason.String
	}

	return &log, nil
}

func (db *DB) ListChangelogs(filter ChangelogFilter) ([]*MetadataChangelog, int, error) {
	query := `SELECT id, book_id, book_title, field, old_value, new_value, source, draft_id,
			  applied_at, applied_by, reverted_at, reverted_by, revert_reason
			  FROM metadata_changelog WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM metadata_changelog WHERE 1=1`
	args := []interface{}{}

	if filter.BookID != nil {
		query += " AND book_id = ?"
		countQuery += " AND book_id = ?"
		args = append(args, *filter.BookID)
	}
	if filter.Field != nil {
		query += " AND field = ?"
		countQuery += " AND field = ?"
		args = append(args, *filter.Field)
	}
	if filter.FromDate != nil {
		query += " AND applied_at >= ?"
		countQuery += " AND applied_at >= ?"
		args = append(args, *filter.FromDate)
	}
	if filter.ToDate != nil {
		query += " AND applied_at <= ?"
		countQuery += " AND applied_at <= ?"
		args = append(args, *filter.ToDate)
	}
	if filter.Reverted != nil {
		if *filter.Reverted {
			query += " AND reverted_at IS NOT NULL"
			countQuery += " AND reverted_at IS NOT NULL"
		} else {
			query += " AND reverted_at IS NULL"
			countQuery += " AND reverted_at IS NULL"
		}
	}

	var total int
	if err := db.conn.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count changelogs: %w", err)
	}

	query += " ORDER BY applied_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list changelogs: %w", err)
	}
	defer rows.Close()

	var logs []*MetadataChangelog
	for rows.Next() {
		var log MetadataChangelog
		var draftID sql.NullInt64
		var revertedAt sql.NullTime
		var revertedBy, revertReason, appliedBy sql.NullString

		if err := rows.Scan(
			&log.ID, &log.BookID, &log.BookTitle, &log.Field, &log.OldValue,
			&log.NewValue, &log.Source, &draftID, &log.AppliedAt, &appliedBy,
			&revertedAt, &revertedBy, &revertReason,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan changelog: %w", err)
		}

		if draftID.Valid {
			log.DraftID = &draftID.Int64
		}
		if appliedBy.Valid {
			log.AppliedBy = appliedBy.String
		}
		if revertedAt.Valid {
			log.RevertedAt = &revertedAt.Time
		}
		if revertedBy.Valid {
			log.RevertedBy = revertedBy.String
		}
		if revertReason.Valid {
			log.RevertReason = revertReason.String
		}

		logs = append(logs, &log)
	}

	return logs, total, nil
}

func (db *DB) RevertChangelog(id int64, reason, revertedBy string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE metadata_changelog 
		SET reverted_at = ?, reverted_by = ?, revert_reason = ?
		WHERE id = ? AND reverted_at IS NULL
	`, now, revertedBy, reason, id)
	return err
}

func (db *DB) CreateSession(session *ExtractionSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.StartedAt = time.Now()
	session.State = SessionStateRunning

	_, err := db.conn.Exec(`
		INSERT INTO extraction_sessions 
			(id, task_type, mode, total_books, state)
		VALUES (?, ?, ?, ?, ?)
	`, session.ID, session.TaskType, session.Mode, session.TotalBooks, session.State)

	return err
}

func (db *DB) UpdateSessionStats(id string, processed, success, failed, skipped, autoApproved, pendingReview int) error {
	_, err := db.conn.Exec(`
		UPDATE extraction_sessions 
		SET processed = ?, success = ?, failed = ?, skipped = ?,
			auto_approved = ?, pending_review = ?
		WHERE id = ?
	`, processed, success, failed, skipped, autoApproved, pendingReview, id)
	return err
}

func (db *DB) CompleteSession(id string, state SessionState, errorMsg string) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE extraction_sessions 
		SET state = ?, error_message = ?, completed_at = ?
		WHERE id = ?
	`, state, errorMsg, now, id)
	return err
}

func (db *DB) GetSession(id string) (*ExtractionSession, error) {
	var session ExtractionSession
	var completedAt sql.NullTime
	var errorMsg sql.NullString

	err := db.conn.QueryRow(`
		SELECT id, task_type, mode, total_books, processed, success, failed, skipped,
			   auto_approved, pending_review, state, error_message, started_at, completed_at
		FROM extraction_sessions WHERE id = ?
	`, id).Scan(
		&session.ID, &session.TaskType, &session.Mode, &session.TotalBooks,
		&session.Processed, &session.Success, &session.Failed, &session.Skipped,
		&session.AutoApproved, &session.PendingReview, &session.State, &errorMsg,
		&session.StartedAt, &completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}
	if errorMsg.Valid {
		session.ErrorMessage = errorMsg.String
	}

	return &session, nil
}

func (db *DB) GetStats() (*GovernanceStats, error) {
	stats := &GovernanceStats{
		BySource:   make(map[MetadataSource]int),
		FlagsCount: make(map[DraftFlag]int),
	}

	err := db.conn.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'applied' THEN 1 ELSE 0 END), 0)
		FROM metadata_drafts
	`).Scan(&stats.Drafts.Pending, &stats.Drafts.Approved, &stats.Drafts.Rejected, &stats.Drafts.Applied)
	if err != nil {
		return nil, fmt.Errorf("failed to get draft stats: %w", err)
	}

	err = db.conn.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN confidence >= 0.8 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN confidence >= 0.5 AND confidence < 0.8 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN confidence < 0.5 THEN 1 ELSE 0 END), 0)
		FROM metadata_drafts WHERE status = 'pending'
	`).Scan(&stats.ConfidenceDistribution.High, &stats.ConfidenceDistribution.Medium, &stats.ConfidenceDistribution.Low)
	if err != nil {
		return nil, fmt.Errorf("failed to get confidence distribution: %w", err)
	}

	rows, err := db.conn.Query(`
		SELECT source, COUNT(*) FROM metadata_drafts GROUP BY source
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get source stats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var source MetadataSource
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			continue
		}
		stats.BySource[source] = count
	}

	flagRows, err := db.conn.Query(`
		SELECT flags FROM metadata_drafts WHERE flags IS NOT NULL AND flags != '' AND flags != '[]'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get flags stats: %w", err)
	}
	defer flagRows.Close()

	for flagRows.Next() {
		var flagsJSON string
		if err := flagRows.Scan(&flagsJSON); err != nil {
			continue
		}
		for _, flag := range JSONToFlags(flagsJSON) {
			stats.FlagsCount[flag]++
		}
	}

	return stats, nil
}
