package tasks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

// CheckMissingTask checks for books missing from search index
type CheckMissingTask struct {
	id         string
	contentApi *content.Api
	searcher   semantic.Searcher
	status     TaskStatus
	mu         sync.RWMutex
	cancel     context.CancelFunc
}

func NewCheckMissingTask(id string, contentApi *content.Api, searcher semantic.Searcher) *CheckMissingTask {
	return &CheckMissingTask{
		id:         id,
		contentApi: contentApi,
		searcher:   searcher,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeCheckMissing,
			Mode:      TaskModeFull,
			State:     "pending",
			Progress:  0,
			Message:   "Pending check for missing vectors",
			StartTime: time.Now(),
		},
	}
}

func (t *CheckMissingTask) Run() error {
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

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books in Calibre. Fetching search index IDs...", len(calibreIDs))
	t.status.Progress = 10
	t.mu.Unlock()

	// 2. Fetch all IDs from Search Engine
	// We use GetAllWithCursor but only need IDs
	// This might be slow for very large collections, but acceptable for a check task
	var searchIDs []int64
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

		if len(books) == 0 {
			break
		}

		for _, b := range books {
			searchIDs = append(searchIDs, b.ID)
		}

		if nextCursor == "" {
			break
		}

		// Prevent infinite loop
		if nextCursor == cursor {
			break
		}

		cursor = nextCursor

		t.mu.Lock()
		t.status.Message = fmt.Sprintf("Fetched %d IDs from search index...", len(searchIDs))
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Comparing %d Calibre books with %d search index items...", len(calibreIDs), len(searchIDs))
	t.status.Progress = 50
	t.mu.Unlock()

	// 3. Compare
	// Use a map for O(1) lookup
	searchMap := make(map[int64]bool)
	for _, id := range searchIDs {
		searchMap[id] = true
	}

	var missingIDs []int64
	for _, id := range calibreIDs {
		if !searchMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	sort.Slice(missingIDs, func(i, j int) bool {
		return missingIDs[i] < missingIDs[j]
	})

	t.mu.Lock()
	defer t.mu.Unlock()

	t.status.State = "completed"
	t.status.Progress = 100
	t.status.EndTime = time.Now()

	if len(missingIDs) == 0 {
		t.status.Message = "All books have vectors in search index."
	} else {
		// Limit message length
		msg := fmt.Sprintf("Found %d missing vectors. IDs: %v", len(missingIDs), missingIDs)
		if len(msg) > 500 {
			msg = msg[:497] + "..."
		}
		t.status.Message = msg
		// Ideally we should store the full list somewhere or return it in a result field
		// For now, the message gives a summary and some IDs
	}

	return nil
}

func (t *CheckMissingTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *CheckMissingTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}
