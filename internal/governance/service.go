package governance

import (
	"fmt"
	"strconv"

	"github.com/jianyun8023/calibre-api/pkg/content"
)

type Config struct {
	DBPath string `yaml:"db_path" mapstructure:"db_path"`

	Confidence struct {
		AutoApplyThreshold float64 `yaml:"auto_apply_threshold" mapstructure:"auto_apply_threshold"`
		ReviewThreshold    float64 `yaml:"review_threshold" mapstructure:"review_threshold"`
	} `yaml:"confidence" mapstructure:"confidence"`

	Batch struct {
		DefaultLimit int `yaml:"default_limit" mapstructure:"default_limit"`
		MaxLimit     int `yaml:"max_limit" mapstructure:"max_limit"`
		WorkerCount  int `yaml:"worker_count" mapstructure:"worker_count"`
	} `yaml:"batch" mapstructure:"batch"`
}

func DefaultConfig() Config {
	cfg := Config{
		DBPath: "data/governance.db",
	}
	cfg.Confidence.AutoApplyThreshold = 0.80
	cfg.Confidence.ReviewThreshold = 0.50
	cfg.Batch.DefaultLimit = 100
	cfg.Batch.MaxLimit = 1000
	cfg.Batch.WorkerCount = 5
	return cfg
}

type Service struct {
	db         *DB
	config     Config
	contentApi *content.Api
}

func NewService(db *DB, config Config, contentApi *content.Api) *Service {
	return &Service{
		db:         db,
		config:     config,
		contentApi: contentApi,
	}
}

func (s *Service) GetDB() *DB {
	return s.db
}

func (s *Service) GetConfig() Config {
	return s.config
}

func (s *Service) CreateDraft(draft *MetadataDraft) error {
	return s.db.CreateDraft(draft)
}

func (s *Service) GetDraft(id int64) (*MetadataDraft, error) {
	return s.db.GetDraft(id)
}

func (s *Service) ListDrafts(filter DraftFilter) ([]*MetadataDraft, int, error) {
	return s.db.ListDrafts(filter)
}

func (s *Service) ApproveDraft(id int64, version int, reviewedBy string) error {
	return s.db.UpdateDraftStatus(id, DraftStatusApproved, version, reviewedBy)
}

func (s *Service) RejectDraft(id int64, version int, reviewedBy string) error {
	return s.db.UpdateDraftStatus(id, DraftStatusRejected, version, reviewedBy)
}

func (s *Service) SkipDraft(id int64, version int, reviewedBy string) error {
	return s.db.UpdateDraftStatus(id, DraftStatusSkipped, version, reviewedBy)
}

func (s *Service) UpdateDraftValue(id int64, newValue string, version int) error {
	return s.db.UpdateDraftValue(id, newValue, version)
}

func (s *Service) BatchApprove(items []BatchItem, reviewedBy string) *BatchResult {
	result := &BatchResult{
		Success:   []int64{},
		Conflicts: []int64{},
		Errors:    []int64{},
	}

	for _, item := range items {
		err := s.ApproveDraft(item.ID, item.Version, reviewedBy)
		if err == ErrConcurrentModification {
			result.Conflicts = append(result.Conflicts, item.ID)
		} else if err != nil {
			result.Errors = append(result.Errors, item.ID)
		} else {
			result.Success = append(result.Success, item.ID)
		}
	}

	return result
}

func (s *Service) BatchReject(items []BatchItem, reviewedBy string) *BatchResult {
	result := &BatchResult{
		Success:   []int64{},
		Conflicts: []int64{},
		Errors:    []int64{},
	}

	for _, item := range items {
		err := s.RejectDraft(item.ID, item.Version, reviewedBy)
		if err == ErrConcurrentModification {
			result.Conflicts = append(result.Conflicts, item.ID)
		} else if err != nil {
			result.Errors = append(result.Errors, item.ID)
		} else {
			result.Success = append(result.Success, item.ID)
		}
	}

	return result
}

func (s *Service) ApplyDrafts(draftIDs []int64, appliedBy string) (*BatchResult, error) {
	result := &BatchResult{
		Success:   []int64{},
		Conflicts: []int64{},
		Errors:    []int64{},
	}

	for _, draftID := range draftIDs {
		draft, err := s.db.GetDraft(draftID)
		if err != nil {
			result.Errors = append(result.Errors, draftID)
			continue
		}

		if draft.Status != DraftStatusApproved {
			result.Conflicts = append(result.Conflicts, draftID)
			continue
		}

		if err := s.applyToCalibur(draft); err != nil {
			result.Errors = append(result.Errors, draftID)
			continue
		}

		changelog := &MetadataChangelog{
			BookID:    draft.BookID,
			BookTitle: draft.BookTitle,
			Field:     draft.Field,
			OldValue:  draft.OldValue,
			NewValue:  draft.NewValue,
			Source:    draft.Source,
			DraftID:   &draft.ID,
			AppliedBy: appliedBy,
		}
		if err := s.db.CreateChangelog(changelog); err != nil {
			result.Errors = append(result.Errors, draftID)
			continue
		}

		if err := s.db.MarkDraftApplied(draftID); err != nil {
			result.Errors = append(result.Errors, draftID)
			continue
		}

		result.Success = append(result.Success, draftID)
	}

	return result, nil
}

func (s *Service) ApplyAllApproved(appliedBy string) (*BatchResult, error) {
	status := DraftStatusApproved
	drafts, _, err := s.db.ListDrafts(DraftFilter{Status: &status, Limit: s.config.Batch.MaxLimit})
	if err != nil {
		return nil, err
	}

	var draftIDs []int64
	for _, d := range drafts {
		draftIDs = append(draftIDs, d.ID)
	}

	return s.ApplyDrafts(draftIDs, appliedBy)
}

func (s *Service) applyToCalibur(draft *MetadataDraft) error {
	if s.contentApi == nil {
		return fmt.Errorf("content API not available")
	}

	bookIDStr := strconv.FormatInt(draft.BookID, 10)

	var metadata map[string]interface{}

	switch draft.Field {
	case FieldISBN:
		metadata = map[string]interface{}{
			"identifiers": map[string]string{
				"isbn": draft.NewValue,
			},
		}
	case FieldTitle:
		metadata = map[string]interface{}{
			"title": draft.NewValue,
		}
	case FieldAuthors:
		metadata = map[string]interface{}{
			"authors": draft.NewValue,
		}
	case FieldPublisher:
		metadata = map[string]interface{}{
			"publisher": draft.NewValue,
		}
	case FieldPublishDate:
		metadata = map[string]interface{}{
			"pubdate": draft.NewValue,
		}
	default:
		return fmt.Errorf("unknown field: %s", draft.Field)
	}

	_, err := s.contentApi.UpdateMetaData(bookIDStr, metadata, "library")
	return err
}

func (s *Service) GetChangelog(id int64) (*MetadataChangelog, error) {
	return s.db.GetChangelog(id)
}

func (s *Service) ListChangelogs(filter ChangelogFilter) ([]*MetadataChangelog, int, error) {
	return s.db.ListChangelogs(filter)
}

func (s *Service) RevertChangelog(id int64, reason, revertedBy string) error {
	changelog, err := s.db.GetChangelog(id)
	if err != nil {
		return err
	}

	if changelog.RevertedAt != nil {
		return fmt.Errorf("changelog already reverted")
	}

	revertDraft := &MetadataDraft{
		BookID:          changelog.BookID,
		BookTitle:       changelog.BookTitle,
		Field:           changelog.Field,
		OldValue:        changelog.NewValue,
		NewValue:        changelog.OldValue,
		Source:          SourceManual,
		Confidence:      1.0,
		Status:          DraftStatusApproved,
		SuggestedAction: ActionAutoApply,
	}

	if err := s.applyToCalibur(revertDraft); err != nil {
		return fmt.Errorf("failed to revert in Calibre: %w", err)
	}

	return s.db.RevertChangelog(id, reason, revertedBy)
}

func (s *Service) GetStats() (*GovernanceStats, error) {
	return s.db.GetStats()
}

func (s *Service) CreateSession(taskType, mode string, totalBooks int) (*ExtractionSession, error) {
	session := &ExtractionSession{
		TaskType:   taskType,
		Mode:       mode,
		TotalBooks: totalBooks,
	}
	if err := s.db.CreateSession(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) UpdateSessionStats(sessionID string, processed, success, failed, skipped, autoApproved, pendingReview int) error {
	return s.db.UpdateSessionStats(sessionID, processed, success, failed, skipped, autoApproved, pendingReview)
}

func (s *Service) CompleteSession(sessionID string, state SessionState, errorMsg string) error {
	return s.db.CompleteSession(sessionID, state, errorMsg)
}

func (s *Service) GetSession(id string) (*ExtractionSession, error) {
	return s.db.GetSession(id)
}
