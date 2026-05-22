package tasks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

// CleanupOrphansTask cleans up orphaned books in the search index
// (books that exist in MeiliSearch but have been deleted from Calibre)
type CleanupOrphansTask struct {
	id         string
	contentApi *content.Api
	searcher   semantic.Searcher
	status     TaskStatus
	mu         sync.RWMutex
	cancel     context.CancelFunc
}

// NewCleanupOrphansTask creates a new cleanup orphans task
func NewCleanupOrphansTask(id string, contentApi *content.Api, searcher semantic.Searcher) *CleanupOrphansTask {
	return &CleanupOrphansTask{
		id:         id,
		contentApi: contentApi,
		searcher:   searcher,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeCleanupOrphans,
			Mode:      TaskModeFull,
			State:     "pending",
			Progress:  0,
			Message:   "Pending orphan cleanup",
			StartTime: time.Now(),
		},
	}
}

func (t *CleanupOrphansTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Fetching all book IDs from Calibre..."
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 1. Fetch all IDs from Calibre
	calibreIDs, err := t.contentApi.GetAllBooksIds("")
	if err != nil {
		return fmt.Errorf("failed to get calibre IDs: %w", err)
	}

	calibreMap := make(map[int64]bool, len(calibreIDs))
	for _, id := range calibreIDs {
		calibreMap[id] = true
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books in Calibre. Fetching search index IDs...", len(calibreIDs))
	t.status.Progress = 10
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 2. Fetch all IDs from MeiliSearch and find orphans
	var orphanIDs []int64
	cursor := ""
	limit := 1000

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		books, _, nextCursor, err := t.searcher.GetAllWithCursor(limit, cursor)
		if err != nil {
			return fmt.Errorf("failed to fetch from search engine: %w", err)
		}

		for _, b := range books {
			if !calibreMap[b.ID] {
				orphanIDs = append(orphanIDs, b.ID)
			}
		}

		if nextCursor == "" || len(books) == 0 || nextCursor == cursor {
			break
		}
		cursor = nextCursor

		t.mu.Lock()
		t.status.Message = fmt.Sprintf("Scanning search index... found %d orphans so far", len(orphanIDs))
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d orphan books to clean up", len(orphanIDs))
	t.status.Progress = 30
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	if len(orphanIDs) == 0 {
		t.mu.Lock()
		t.status.State = "completed"
		t.status.Progress = 100
		t.status.Message = "No orphan books found in search index"
		t.status.EndTime = time.Now()
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)
		return nil
	}

	// 3. Delete orphan books from MeiliSearch
	total := len(orphanIDs)
	deleted := 0
	var deleteErrors []string

	for i, id := range orphanIDs {
		select {
		case <-ctx.Done():
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Cleanup stopped. Deleted %d/%d orphans", deleted, total)
			t.mu.Unlock()
			return ctx.Err()
		default:
		}

		if err := t.searcher.DeleteBook(id); err != nil {
			errMsg := fmt.Sprintf("Failed to delete book %d: %v", id, err)
			log.Printf("Orphan cleanup: %s", errMsg)
			deleteErrors = append(deleteErrors, errMsg)
		} else {
			deleted++
		}

		// Update progress every 10 books or on last item
		if (i+1)%10 == 0 || i == total-1 {
			t.mu.Lock()
			t.status.Progress = 30 + float64(i+1)/float64(total)*70
			t.status.Message = fmt.Sprintf("Deleting orphans: %d/%d (errors: %d)", i+1, total, len(deleteErrors))
			t.mu.Unlock()
			GetManager().BroadcastTaskProgress(t.id)
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.status.Progress = 100
	t.status.EndTime = time.Now()

	if len(deleteErrors) > 0 {
		t.status.State = "completed"
		t.status.Message = fmt.Sprintf("Cleanup completed: %d/%d orphans deleted with %d errors", deleted, total, len(deleteErrors))
		t.status.Error = fmt.Sprintf("%d deletions failed. Last error: %s", len(deleteErrors), deleteErrors[len(deleteErrors)-1])
	} else {
		t.status.State = "completed"
		t.status.Message = fmt.Sprintf("Successfully cleaned up %d orphan books", deleted)
	}

	return nil
}

func (t *CleanupOrphansTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *CleanupOrphansTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}
