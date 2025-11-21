package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic/indexer"
)

type VectorTask struct {
	id      string
	mode    TaskMode
	indexer *indexer.Indexer
	status  TaskStatus
	mu      sync.RWMutex
	cancel  context.CancelFunc
}

func NewVectorTask(id string, mode TaskMode, idx *indexer.Indexer) *VectorTask {
	return &VectorTask{
		id:      id,
		mode:    mode,
		indexer: idx,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeVectorSync,
			Mode:      mode,
			State:     "idle",
			StartTime: time.Now(),
			Message:   "Initializing...",
		},
	}
}

func (t *VectorTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *VectorTask) Stop() {
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

func (t *VectorTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Starting vector indexing..."
	t.mu.Unlock()

	// Determine query based on mode
	query := ""
	if t.mode == TaskModeIncremental {
		// Get max book_id from Milvus as the baseline for incremental sync
		maxID, err := t.indexer.GetMaxBookID()
		if err != nil {
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Warning: Could not get max book_id, using full sync instead: %v", err)
			t.mu.Unlock()
			// Fall back to empty query (full sync) if we can't get max ID
			query = ""
		} else if maxID > 0 {
			// Query for books with ID greater than the max ID in Milvus
			query = fmt.Sprintf("id:>%d", maxID)
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Incremental sync: indexing books with id > %d", maxID)
			t.mu.Unlock()
		} else {
			// No books in Milvus yet, do full sync
			t.mu.Lock()
			t.status.Message = "No existing data in vector database, performing full sync"
			t.mu.Unlock()
			query = ""
		}
	}

	// Use the Index method we just added to Indexer
	err := t.indexer.Index(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
