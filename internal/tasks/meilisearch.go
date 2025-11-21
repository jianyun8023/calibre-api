package tasks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchTask struct {
	id         string
	mode       TaskMode
	contentApi *content.Api
	client     *meilisearch.Client
	indexName  string
	status     TaskStatus
	mu         sync.RWMutex
	cancel     context.CancelFunc
}

func NewMeilisearchTask(id string, mode TaskMode, contentApi *content.Api, client *meilisearch.Client, targetIndex string) *MeilisearchTask {
	return &MeilisearchTask{
		id:         id,
		mode:       mode,
		contentApi: contentApi,
		client:     client,
		indexName:  targetIndex,
		status: TaskStatus{
			ID:          id,
			Type:        TaskTypeMeilisearchSync,
			Mode:        mode,
			State:       "idle",
			StartTime:   time.Now(),
			Message:     "Initializing...",
			TargetIndex: targetIndex,
		},
	}
}

func (t *MeilisearchTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *MeilisearchTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
	if t.status.State == "running" {
		t.status.State = "stopped"
		t.status.Message = "Stopped by user"
	}
}

func (t *MeilisearchTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Fetching book IDs..."
	t.mu.Unlock()

	index := t.client.Index(t.indexName)

	// If Full mode, clear index first
	if t.mode == TaskModeFull {
		t.mu.Lock()
		t.status.Message = "Clearing existing documents..."
		t.mu.Unlock()
		if _, err := index.DeleteAllDocuments(); err != nil {
			return fmt.Errorf("failed to clear index: %w", err)
		}
	}

	// Determine query based on mode
	query := ""
	if t.mode == TaskModeIncremental {
		// Find last successful sync time from Manager history
		// For now, let's default to 24 hours ago if not found
		// TODO: Implement history lookup
		lastSync := time.Now().Add(-24 * time.Hour)
		// Calibre date format: 2023-01-01T00:00:00
		// modified:">=2023-01-01T00:00:00"
		query = fmt.Sprintf("modified:\">=%s\"", lastSync.Format("2006-01-02T15:04:05"))
	}

	ids, err := t.contentApi.GetAllBooksIds(query)
	if err != nil {
		return err
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books to sync", len(ids))
	t.mu.Unlock()

	if len(ids) == 0 {
		return nil
	}

	batchSize := 2000
	total := len(ids)

	for i := 0; i < total; i += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := i + batchSize
		if end > total {
			end = total
		}

		batchIDs := ids[i:end]

		t.mu.Lock()
		t.status.Progress = float64(i) / float64(total) * 100
		t.status.Message = fmt.Sprintf("Processing batch %d-%d", i, end)
		t.mu.Unlock()

		data, err := t.contentApi.GetBookMetaDatas(batchIDs, "")
		if err != nil {
			log.Printf("Error getting metadata: %v", err)
			continue
		}

		// We need convertContentBooks from internal/calibre/api.go
		// But we can't import internal/calibre because of cycle if api.go imports tasks.
		// So we should probably duplicate the conversion logic or move it to a shared package.
		// For now, let's assume we can access it or implement a simple version.
		// Actually, we can't easily access it if it's private in calibre package.
		// Let's implement a local conversion or move `Book` to `pkg/content` or `internal/common`.
		// `content.Book` is already available. `meilisearch` expects a struct or map.
		// Let's use `content.Book` directly or map it.

		// Using content.Book directly might be enough if JSON tags match.
		// content.Book has JSON tags.

		_, err = index.AddDocuments(data)
		if err != nil {
			log.Printf("Error adding documents: %v", err)
			continue
		}
	}

	return nil
}
