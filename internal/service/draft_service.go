package service

import (
	"context"
	"encoding/json"
	"fmt"
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
		dataBytes, err := json.Marshal(u.Data)
		if err != nil {
			return fmt.Errorf("failed to serialize update data for book %s: %v", u.ID, err)
		}

		if existing := existingMap[u.ID]; existing != nil {
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

			// 2. Perform External Action (DRY RUN MODE - 仅日志，不执行)
			var actionErr error
			if draft.Action == repository.DraftActionDelete {
				log.Infof("========================================")
				log.Infof("📋 Draft ID: %d", id)
				log.Infof("📚 Book ID: %s", draft.BookID)
				log.Infof("🔥 Action: DELETE")
				log.Infof("⏰ Created At: %s", draft.CreatedAt)
				log.Infof("🎯 Will Execute: s.bookService.DeleteBook(\"%s\")", draft.BookID)
				log.Infof("📌 Description: 将调用 Calibre Content Server API 删除此书籍")
				log.Infof("⚠️  Note: DRY RUN MODE - 实际删除操作已被注释，不会执行真实删除")
				log.Infof("========================================")

				// DRY RUN: 注释掉实际执行代码
				// actionErr = s.bookService.DeleteBook(draft.BookID)

			} else if draft.Action == repository.DraftActionUpdate {
				var updateData BookUpdate
				if errUnmarshal := json.Unmarshal([]byte(draft.Data), &updateData); errUnmarshal != nil {
					actionErr = fmt.Errorf("failed to unmarshal update data for draft %d: %v", id, errUnmarshal)
				} else {
					log.Infof("========================================")
					log.Infof("📋 Draft ID: %d", id)
					log.Infof("📚 Book ID: %s", draft.BookID)
					log.Infof("✏️  Action: UPDATE")
					log.Infof("⏰ Created At: %s", draft.CreatedAt)
					log.Infof("📝 Update Data:")

					// 详细输出每个更新字段
					if updateData.Title != nil && *updateData.Title != "" {
						log.Infof("  ✓ Title: %s", *updateData.Title)
					}
					if updateData.Authors != nil && len(*updateData.Authors) > 0 {
						log.Infof("  ✓ Authors: %v", *updateData.Authors)
					}
					if updateData.Publisher != nil && *updateData.Publisher != "" {
						log.Infof("  ✓ Publisher: %s", *updateData.Publisher)
					}
					if updateData.Tags != nil {
						if len(*updateData.Tags) > 0 {
							log.Infof("  ✓ Tags: %v", *updateData.Tags)
						} else {
							log.Infof("  ✓ Tags: [] (清空标签)")
						}
					}
					if updateData.Isbn != nil && *updateData.Isbn != "" {
						log.Infof("  ✓ ISBN: %s", *updateData.Isbn)
					}
					if updateData.PubDate != nil && !updateData.PubDate.IsZero() {
						log.Infof("  ✓ PubDate: %s", updateData.PubDate.Format("2006-01-02"))
					}
					if updateData.Rating != 0 {
						log.Infof("  ✓ Rating: %.1f", updateData.Rating)
					}
					if updateData.Comments != nil && *updateData.Comments != "" {
						comments := *updateData.Comments
						if len(comments) > 100 {
							comments = comments[:100] + "..."
						}
						log.Infof("  ✓ Comments: %s", comments)
					}

					// 显示被拒绝的字段（空值）
					if updateData.Title != nil && *updateData.Title == "" {
						log.Warnf("  ✗ Title: \"\" (拒绝：不允许清空)")
					}
					if updateData.Publisher != nil && *updateData.Publisher == "" {
						log.Warnf("  ✗ Publisher: \"\" (拒绝：不允许清空)")
					}
					if updateData.Comments != nil && *updateData.Comments == "" {
						log.Warnf("  ✗ Comments: \"\" (拒绝：不允许清空)")
					}
					if updateData.Isbn != nil && *updateData.Isbn == "" {
						log.Warnf("  ✗ ISBN: \"\" (拒绝：不允许清空)")
					}
					if updateData.Authors != nil && len(*updateData.Authors) == 0 {
						log.Warnf("  ✗ Authors: [] (拒绝：不允许清空)")
					}

					log.Infof("🎯 Will Execute: s.bookService.UpdateMetadata(\"%s\", updateData)", draft.BookID)
					log.Infof("📌 Description: 将调用 Calibre Content Server API 更新书籍元数据")
					log.Infof("⚠️  Note: DRY RUN MODE - 实际更新操作已被注释，不会执行真实更新")
					log.Infof("========================================")

					// DRY RUN: 注释掉实际执行代码
					// actionErr = s.bookService.UpdateMetadata(draft.BookID, &updateData)
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
