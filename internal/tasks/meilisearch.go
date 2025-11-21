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
		// Get max book ID from Meilisearch as the baseline for incremental sync
		maxID, err := t.getMaxBookID(index)
		if err != nil {
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Warning: Could not get max book ID, using full sync instead: %v", err)
			t.mu.Unlock()
			log.Printf("Failed to get max book ID from Meilisearch: %v", err)
			// Fall back to empty query (full sync) if we can't get max ID
			query = ""
		} else if maxID > 0 {
			// Query for books with ID greater than the max ID in Meilisearch
			query = fmt.Sprintf("id:>%d", maxID)
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Incremental sync: indexing books with id > %d", maxID)
			t.mu.Unlock()
		} else {
			// No books in Meilisearch yet, do full sync
			t.mu.Lock()
			t.status.Message = "No existing data in Meilisearch, performing full sync"
			t.mu.Unlock()
			query = ""
		}
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

		// Enrich books with Cover and FilePath fields
		enrichedBooks := content.EnrichBooks(data)

		_, err = index.AddDocuments(enrichedBooks)
		if err != nil {
			log.Printf("Error adding documents: %v", err)
			continue
		}
	}

	return nil
}

// getMaxBookID gets the maximum book ID from Meilisearch index
func (t *MeilisearchTask) getMaxBookID(index *meilisearch.Index) (int64, error) {
	// Use Meilisearch search API with sort by id descending and limit 1
	searchReq := &meilisearch.SearchRequest{
		Limit: 1,
		Sort:  []string{"id:desc"},
	}

	result, err := index.Search("", searchReq)
	if err != nil {
		return 0, fmt.Errorf("failed to search for max ID: %w", err)
	}

	// Check if we got any results
	if len(result.Hits) == 0 {
		return 0, nil // No documents in index
	}

	// Extract ID from the first (and only) hit
	hit := result.Hits[0].(map[string]interface{})
	idValue, ok := hit["id"]
	if !ok {
		return 0, fmt.Errorf("id field not found in document")
	}

	// Convert to int64
	var maxID int64
	switch v := idValue.(type) {
	case float64:
		maxID = int64(v)
	case int64:
		maxID = v
	case int:
		maxID = int64(v)
	default:
		return 0, fmt.Errorf("unexpected id type: %T", idValue)
	}

	return maxID, nil
}
