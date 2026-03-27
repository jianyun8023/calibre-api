package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jianyun8023/calibre-api/internal/repository"
)

// DraftService defines the logic for handling book drafts
type DraftService interface {
	ReceiveDeletes(ctx context.Context, ids []string) error
	ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error
	GetPendingDrafts(ctx context.Context) ([]repository.BookDraft, error)
	ApplyDrafts(ctx context.Context, ids []int64) error
	RejectDrafts(ctx context.Context, ids []int64) error
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

func (s *draftService) ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error {
	for _, u := range updates {
		dataBytes, err := json.Marshal(u.Data)
		if err != nil {
			return fmt.Errorf("failed to serialize update data for book %s: %v", u.ID, err)
		}

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
	return nil
}

func (s *draftService) GetPendingDrafts(ctx context.Context) ([]repository.BookDraft, error) {
	return s.draftRepo.GetPendingDrafts(ctx)
}

func (s *draftService) ApplyDrafts(ctx context.Context, ids []int64) error {
	for _, draftID := range ids {
		draft, err := s.draftRepo.GetDraftByID(ctx, draftID)
		if err != nil {
			return fmt.Errorf("failed to get draft %d: %v", draftID, err)
		}
		if draft == nil || draft.Status != repository.DraftStatusPending {
			continue // Skip non-existent or already processed drafts
		}

		// Apply the action
		if draft.Action == repository.DraftActionDelete {
			err = s.bookService.DeleteBook(draft.BookID)
		} else if draft.Action == repository.DraftActionUpdate {
			var updateData BookUpdate
			if errUnmarshal := json.Unmarshal([]byte(draft.Data), &updateData); errUnmarshal != nil {
				return fmt.Errorf("failed to unmarshal update data for draft %d: %v", draftID, errUnmarshal)
			}
			err = s.bookService.UpdateMetadata(draft.BookID, &updateData)
		}

		if err != nil {
			return fmt.Errorf("failed to apply draft %d for book %s: %v", draftID, draft.BookID, err)
		}

		// If successfully applied, update status and record history
		if err := s.draftRepo.UpdateDraftStatus(ctx, draftID, repository.DraftStatusApplied); err != nil {
			return fmt.Errorf("failed to update draft status for %d: %v", draftID, err)
		}

		history := &repository.BookDraftHistory{
			DraftID: draftID,
			BookID:  draft.BookID,
			Action:  draft.Action,
			Data:    draft.Data,
			Status:  repository.DraftStatusApplied,
		}
		if _, err := s.draftRepo.CreateHistory(ctx, history); err != nil {
			return fmt.Errorf("failed to create history for draft %d: %v", draftID, err)
		}
	}
	return nil
}

func (s *draftService) RejectDrafts(ctx context.Context, ids []int64) error {
	for _, draftID := range ids {
		draft, err := s.draftRepo.GetDraftByID(ctx, draftID)
		if err != nil {
			return fmt.Errorf("failed to get draft %d: %v", draftID, err)
		}
		if draft == nil || draft.Status != repository.DraftStatusPending {
			continue
		}

		if err := s.draftRepo.UpdateDraftStatus(ctx, draftID, repository.DraftStatusRejected); err != nil {
			return fmt.Errorf("failed to update draft status for %d: %v", draftID, err)
		}

		history := &repository.BookDraftHistory{
			DraftID: draftID,
			BookID:  draft.BookID,
			Action:  draft.Action,
			Data:    draft.Data,
			Status:  repository.DraftStatusRejected,
		}
		if _, err := s.draftRepo.CreateHistory(ctx, history); err != nil {
			return fmt.Errorf("failed to create history for draft %d: %v", draftID, err)
		}
	}
	return nil
}
