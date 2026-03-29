package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// DraftService defines the logic for handling book drafts
type DraftService interface {
	ReceiveDeletes(ctx context.Context, ids []string) error
	ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error
	GetPendingDrafts(ctx context.Context, limit, offset int) ([]repository.BookDraft, int64, error)
	ApplyDrafts(ctx context.Context, ids []int64) []error
	RejectDrafts(ctx context.Context, ids []int64) []error
	CancelDrafts(ctx context.Context, bookIDs []string) (int, int, []error)
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
	if len(ids) == 0 {
		return nil
	}

	// Bulk check for existing pending deletes
	existingDrafts, err := s.draftRepo.GetPendingDraftsByBookIDsAndAction(ctx, ids, repository.DraftActionDelete)
	if err != nil {
		return fmt.Errorf("failed to check existing drafts: %v", err)
	}

	existingMap := make(map[string]bool)
	for _, d := range existingDrafts {
		existingMap[d.BookID] = true
	}

	for _, id := range ids {
		if existingMap[id] {
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
	// 1. Reset stuck processing drafts (e.g. from app crashes)
	stuckThreshold := time.Now().Add(-1 * time.Hour) // If processing for > 1 hour, assume crashed
	if resetCount, err := s.draftRepo.ResetStuckProcessingDrafts(ctx, stuckThreshold); err != nil {
		log.Warnf("Failed to reset stuck processing drafts: %v", err)
	} else if resetCount > 0 {
		log.Infof("Reset %d stuck processing drafts back to pending", resetCount)
	}

	// 2. Expire old pending drafts
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
		// Expire atomically (status + history)
		if err := s.draftRepo.ExpireDraftAtomically(ctx, &draft); err != nil {
			log.Warnf("Failed to atomically expire draft %d: %v", draft.ID, err)
			continue
		}
		count++
	}

	return count, nil
}

func (s *draftService) ReceiveUpdates(ctx context.Context, updates []BookDraftUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]string, len(updates))
	for i, u := range updates {
		ids[i] = u.ID
	}

	// Bulk check for existing pending updates
	existingDrafts, err := s.draftRepo.GetPendingDraftsByBookIDsAndAction(ctx, ids, repository.DraftActionUpdate)
	if err != nil {
		return fmt.Errorf("failed to check existing drafts: %v", err)
	}

	existingMap := make(map[string]*repository.BookDraft)
	for i := range existingDrafts {
		d := &existingDrafts[i]
		existingMap[d.BookID] = d
	}

	for _, u := range updates {
		var finalData *BookUpdate

		if existing := existingMap[u.ID]; existing != nil {
			// Merge with existing draft data
			var existingData BookUpdate
			if err := json.Unmarshal([]byte(existing.Data), &existingData); err != nil {
				return fmt.Errorf("failed to parse existing draft data for book %s: %v", u.ID, err)
			}

			// Merge: new fields override old fields, but keep old fields if not provided in new data
			finalData = mergeDraftUpdates(&existingData, u.Data)
			
			dataBytes, err := json.Marshal(finalData)
			if err != nil {
				return fmt.Errorf("failed to serialize merged data for book %s: %v", u.ID, err)
			}

			if err := s.draftRepo.UpdateDraftData(ctx, existing.ID, string(dataBytes)); err != nil {
				return fmt.Errorf("failed to update existing draft data for book %s: %v", u.ID, err)
			}
			log.Infof("Merged and updated existing pending update draft for book %s", u.ID)
		} else {
			// Create new draft
			finalData = u.Data
			dataBytes, err := json.Marshal(finalData)
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
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Concurrency control: limit to 5 concurrent external API calls
	sem := make(chan struct{}, 5)

	for _, draftID := range ids {
		wg.Add(1)
		sem <- struct{}{} // Block if limit reached

		go func(id int64) {
			defer wg.Done()
			defer func() { <-sem }() // Release token

			draft, err := s.draftRepo.GetDraftByID(ctx, id)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to get draft %d: %v", id, err))
				mu.Unlock()
				return
			}
			if draft == nil || draft.Status != repository.DraftStatusPending {
				return // Skip non-existent or already processed drafts
			}

			// 1. Mark as Processing (Atomic)
			updated, err := s.draftRepo.UpdateDraftStatusIfPending(ctx, id, repository.DraftStatusProcessing)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to mark draft %d as processing: %v", id, err))
				mu.Unlock()
				return
			}
			if !updated {
				return // Another process got it
			}

		// 2. Perform External Action
		var actionErr error
		if draft.Action == repository.DraftActionDelete {
			log.Infof("Applying draft #%d: DELETE book %s (created: %s)", id, draft.BookID, draft.CreatedAt.Format("2006-01-02 15:04:05"))
			actionErr = s.bookService.DeleteBook(draft.BookID)
		} else if draft.Action == repository.DraftActionUpdate {
			var updateData BookUpdate
			if errUnmarshal := json.Unmarshal([]byte(draft.Data), &updateData); errUnmarshal != nil {
				actionErr = fmt.Errorf("failed to unmarshal update data for draft %d: %v", id, errUnmarshal)
			} else {
				// 构建更新字段摘要
				var fields []string
				if updateData.Title != nil && *updateData.Title != "" {
					fields = append(fields, fmt.Sprintf("title=%s", *updateData.Title))
				}
				if updateData.Authors != nil && len(*updateData.Authors) > 0 {
					fields = append(fields, fmt.Sprintf("authors=%v", *updateData.Authors))
				}
				if updateData.Publisher != nil && *updateData.Publisher != "" {
					fields = append(fields, fmt.Sprintf("publisher=%s", *updateData.Publisher))
				}
				if updateData.Tags != nil {
					if len(*updateData.Tags) > 0 {
						fields = append(fields, fmt.Sprintf("tags=%v", *updateData.Tags))
					} else {
						fields = append(fields, "tags=[] (clear)")
					}
				}
				if updateData.Isbn != nil && *updateData.Isbn != "" {
					fields = append(fields, fmt.Sprintf("isbn=%s", *updateData.Isbn))
				}
				if updateData.PubDate != nil && !updateData.PubDate.IsZero() {
					fields = append(fields, fmt.Sprintf("pubdate=%s", updateData.PubDate.Format("2006-01-02")))
				}
				if updateData.Rating != 0 {
					fields = append(fields, fmt.Sprintf("rating=%.1f", updateData.Rating))
				}
				if updateData.Comments != nil && *updateData.Comments != "" {
					commentPreview := *updateData.Comments
					if len(commentPreview) > 50 {
						commentPreview = commentPreview[:50] + "..."
					}
					fields = append(fields, fmt.Sprintf("comments=%s", commentPreview))
				}

				// 记录被拒绝的空值字段
				var rejectedFields []string
				if updateData.Title != nil && *updateData.Title == "" {
					rejectedFields = append(rejectedFields, "title")
				}
				if updateData.Publisher != nil && *updateData.Publisher == "" {
					rejectedFields = append(rejectedFields, "publisher")
				}
				if updateData.Comments != nil && *updateData.Comments == "" {
					rejectedFields = append(rejectedFields, "comments")
				}
				if updateData.Isbn != nil && *updateData.Isbn == "" {
					rejectedFields = append(rejectedFields, "isbn")
				}
				if updateData.Authors != nil && len(*updateData.Authors) == 0 {
					rejectedFields = append(rejectedFields, "authors")
				}

				if len(fields) > 0 {
					log.Infof("Applying draft #%d: UPDATE book %s, fields: %s", id, draft.BookID, strings.Join(fields, ", "))
				}
				if len(rejectedFields) > 0 {
					log.Warnf("Draft #%d: Rejected empty fields (not allowed): %s", id, strings.Join(rejectedFields, ", "))
				}

				actionErr = s.bookService.UpdateMetadata(draft.BookID, &updateData)
			}
		}

			// 3. Finalize
			if actionErr != nil {
				// Revert to pending
				if revertErr := s.draftRepo.UpdateDraftStatus(ctx, id, repository.DraftStatusPending); revertErr != nil {
					log.Warnf("Failed to revert draft %d back to pending after failure: %v", id, revertErr)
				}
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to apply external action for draft %d: %v", id, actionErr))
				mu.Unlock()
			} else {
				// Success: update to applied and create history (transactionally)
				if finalizeErr := s.draftRepo.ApplyDraftSuccess(ctx, draft, repository.DraftStatusApplied); finalizeErr != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("failed to finalize draft %d: %v", id, finalizeErr))
					mu.Unlock()
				}
			}
		}(draftID)
	}

	wg.Wait()
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

// mergeDraftUpdates merges new draft data into existing draft data
// New fields override old fields, but old fields are preserved if not provided in new data
func mergeDraftUpdates(old, new *BookUpdate) *BookUpdate {
	merged := &BookUpdate{}

	// Title: new overrides old
	if new.Title != nil {
		merged.Title = new.Title
	} else if old.Title != nil {
		merged.Title = old.Title
	}

	// Authors: new overrides old
	if new.Authors != nil {
		merged.Authors = new.Authors
	} else if old.Authors != nil {
		merged.Authors = old.Authors
	}

	// Publisher: new overrides old
	if new.Publisher != nil {
		merged.Publisher = new.Publisher
	} else if old.Publisher != nil {
		merged.Publisher = old.Publisher
	}

	// Tags: new overrides old
	if new.Tags != nil {
		merged.Tags = new.Tags
	} else if old.Tags != nil {
		merged.Tags = old.Tags
	}

	// Isbn: new overrides old
	if new.Isbn != nil {
		merged.Isbn = new.Isbn
	} else if old.Isbn != nil {
		merged.Isbn = old.Isbn
	}

	// Comments: new overrides old
	if new.Comments != nil {
		merged.Comments = new.Comments
	} else if old.Comments != nil {
		merged.Comments = old.Comments
	}

	// PubDate: new overrides old
	if new.PubDate != nil && !new.PubDate.IsZero() {
		merged.PubDate = new.PubDate
	} else if old.PubDate != nil && !old.PubDate.IsZero() {
		merged.PubDate = old.PubDate
	}

	// Rating: new overrides old (0 means not provided)
	if new.Rating != 0 {
		merged.Rating = new.Rating
	} else if old.Rating != 0 {
		merged.Rating = old.Rating
	}

	return merged
}

func (s *draftService) CancelDrafts(ctx context.Context, bookIDs []string) (int, int, []error) {
	var errs []error
	cancelledBooks := 0
	totalDeleted := 0

	for _, bookID := range bookIDs {
		drafts := s.getPendingDraftsByBookID(ctx, bookID)

		if len(drafts) == 0 {
			errs = append(errs, fmt.Errorf("no pending drafts found for book %s", bookID))
			continue
		}

		deletedCount := 0
		for _, draft := range drafts {
			if err := s.draftRepo.DeleteDraft(ctx, draft.ID); err != nil {
				errs = append(errs, fmt.Errorf("failed to delete draft %d for book %s: %v", draft.ID, bookID, err))
				continue
			}
			deletedCount++
			totalDeleted++
		}

		if deletedCount > 0 {
			cancelledBooks++
			log.Infof("Cancelled %d draft(s) for book %s", deletedCount, bookID)
		}
	}

	return cancelledBooks, totalDeleted, errs
}

func (s *draftService) getPendingDraftsByBookID(ctx context.Context, bookID string) []repository.BookDraft {
	var allDrafts []repository.BookDraft

	if updateDraft, err := s.draftRepo.GetPendingDraftByBookIDAndAction(ctx, bookID, repository.DraftActionUpdate); err == nil && updateDraft != nil {
		allDrafts = append(allDrafts, *updateDraft)
	}

	if deleteDraft, err := s.draftRepo.GetPendingDraftByBookIDAndAction(ctx, bookID, repository.DraftActionDelete); err == nil && deleteDraft != nil {
		allDrafts = append(allDrafts, *deleteDraft)
	}

	return allDrafts
}
