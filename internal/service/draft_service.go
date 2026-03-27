package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// DraftService defines the logic for handling book drafts
type DraftService interface {
	ReceiveDeletes(ctx context.Context, ids []string) error
	ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error
	GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error)
	ApplyDrafts(ctx context.Context, ids []int64) ([]error)
	RejectDrafts(ctx context.Context, ids []int64) ([]error)
	CleanupExpiredDrafts(ctx context.Context, days int) (int, error)
	GetHistory(ctx context.Context, limit, offset int) ([]repository.BookDraftHistory, int64, error)
}

type BookDraftUpdate struct {
	ID   string      `json:"id"`
	Data *BookUpdate `json:"data"`
}

type draftService struct {
	draftRepo   repository.DraftRepository
	bookService BookService
}

// NewDraftService creates a new DraftService
func NewDraftService(draftRepo repository.DraftRepository, bookService BookService) DraftService {
	return &draftService{
		draftRepo:   draftRepo,
		bookService: bookService,
	}
}

func (s *draftService) ReceiveDeletes(ctx context.Context, ids []string) error {
	for _, id := range ids {
		// Check for existing pending delete to ensure idempotency
		existing, err := s.draftRepo.GetPendingDraftByBookIDAndAction(ctx, id, repository.DraftActionDelete)
		if err != nil {
			return fmt.Errorf("failed to check existing draft for book %s: %v", id, err)
		}
		if existing != nil {
			log.Infof("Pending delete draft already exists for book %s, ignoring duplicate", id)
			continue
		}

		draft := &repository.BookDraft{
			BookID: id,
			Action: repository.DraftActionDelete,
			Data:   "{}", // No extra data needed for delete
			Status: repository.DraftStatusPending,
		}
		if _, err := s.draftRepo.CreateDraft(ctx, draft); err != nil {
			return fmt.Errorf("failed to create delete draft for book %s: %v", id, err)
		}
	}
	return nil
}

func (s *draftService) GetHistory(ctx context.Context, limit, offset int) ([]repository.BookDraftHistory, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.draftRepo.GetHistory(ctx, limit, offset)
}

func (s *draftService) CleanupExpiredDrafts(ctx context.Context, days int) (int, error) {
	if days <= 0 {
		return 0, fmt.Errorf("expiration days must be greater than 0")
	}

	before := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	drafts, err := s.draftRepo.GetPendingDraftsBefore(ctx, before)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired drafts: %v", err)
	}

	count := 0
	for _, draft := range drafts {
		// Mark as expired
		if err := s.draftRepo.UpdateDraftStatus(ctx, draft.ID, repository.DraftStatusExpired); err != nil {
			log.Warnf("Failed to update status for expired draft %d: %v", draft.ID, err)
			continue
		}

		// Save to history
		history := &repository.BookDraftHistory{
			DraftID: draft.ID,
			BookID:  draft.BookID,
			Action:  draft.Action,
			Data:    draft.Data,
			Status:  repository.DraftStatusExpired,
		}
		if _, err := s.draftRepo.CreateHistory(ctx, history); err != nil {
			log.Warnf("Failed to create history for expired draft %d: %v", draft.ID, err)
		} else {
			count++
		}
	}

	return count, nil
}

func (s *draftService) ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error {
	for _, u := range updates {
		dataBytes, err := json.Marshal(u.Data)
		if err != nil {
			return fmt.Errorf("failed to serialize update data for book %s: %v", u.ID, err)
		}

		// Check for existing pending update to ensure idempotency
		existing, err := s.draftRepo.GetPendingDraftByBookIDAndAction(ctx, u.ID, repository.DraftActionUpdate)
		if err != nil {
			return fmt.Errorf("failed to check existing draft for book %s: %v", u.ID, err)
		}

		if existing != nil {
			// Update the existing draft's data
			if err := s.draftRepo.UpdateDraftData(ctx, existing.ID, string(dataBytes)); err != nil {
				return fmt.Errorf("failed to update existing draft data for book %s: %v", u.ID, err)
			}
			log.Infof("Updated existing pending update draft for book %s", u.ID)
		} else {
			draft := &repository.BookDraft{
				BookID: u.ID,
				Action: repository.DraftActionUpdate,
				Data:   string(dataBytes),
				Status: repository.DraftStatusPending,
			}
			if _, err := s.draftRepo.CreateDraft(ctx, draft); err != nil {
				return fmt.Errorf("failed to create update draft for book %s: %v", u.ID, err)
			}
		}
	}
	return nil
}

func (s *draftService) GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.draftRepo.GetPendingDrafts(ctx, limit, offset)
}

func (s *draftService) ApplyDrafts(ctx context.Context, ids []int64) []error {
	var errs []error
	for _, draftID := range ids {
		// Fetch draft first to know the action needed
		draft, err := s.draftRepo.GetDraftByID(ctx, draftID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get draft %d: %v", draftID, err))
			continue
		}
		if draft == nil || draft.Status != repository.DraftStatusPending {
			continue // Skip non-existent or already processed drafts
		}

		// 1. Mark as Processing (Atomic)
		updated, err := s.draftRepo.UpdateDraftStatusIfPending(ctx, draftID, repository.DraftStatusProcessing)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to mark draft %d as processing: %v", draftID, err))
			continue
		}
		if !updated {
			continue // Another process got it
		}

		// 2. Perform External Action
		var actionErr error
		if draft.Action == repository.DraftActionDelete {
			actionErr = s.bookService.DeleteBook(draft.BookID)
		} else if draft.Action == repository.DraftActionUpdate {
			var updateData BookUpdate
			if errUnmarshal := json.Unmarshal([]byte(draft.Data), &updateData); errUnmarshal != nil {
				actionErr = fmt.Errorf("failed to unmarshal update data for draft %d: %v", draftID, errUnmarshal)
			} else {
				actionErr = s.bookService.UpdateMetadata(draft.BookID, &updateData)
			}
		}

		// 3. Finalize
		if actionErr != nil {
			// Revert to pending
			if revertErr := s.draftRepo.UpdateDraftStatus(ctx, draftID, repository.DraftStatusPending); revertErr != nil {
				log.Warnf("Failed to revert draft %d back to pending after failure: %v", draftID, revertErr)
			}
			errs = append(errs, fmt.Errorf("failed to apply external action for draft %d: %v", draftID, actionErr))
		} else {
			// Success: update to applied and create history (transactionally)
			if finalizeErr := s.draftRepo.ApplyDraftSuccess(ctx, draft, repository.DraftStatusApplied); finalizeErr != nil {
				errs = append(errs, fmt.Errorf("failed to finalize draft %d: %v", draftID, finalizeErr))
			}
		}
	}

	return errs
}

func (s *draftService) RejectDrafts(ctx context.Context, ids []int64) []error {
	var errs []error
	for _, draftID := range ids {
		draft, err := s.draftRepo.GetDraftByID(ctx, draftID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get draft %d: %v", draftID, err))
			continue
		}
		if draft == nil || draft.Status != repository.DraftStatusPending {
			continue
		}

		updated, err := s.draftRepo.UpdateDraftStatusIfPending(ctx, draftID, repository.DraftStatusProcessing)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to mark draft %d as processing: %v", draftID, err))
			continue
		}
		if !updated {
			continue
		}

		if err := s.draftRepo.ApplyDraftSuccess(ctx, draft, repository.DraftStatusRejected); err != nil {
			errs = append(errs, fmt.Errorf("failed to reject draft %d atomically: %v", draftID, err))
		}
	}

	return errs
}
