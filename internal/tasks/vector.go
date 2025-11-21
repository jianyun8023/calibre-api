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
		// Default to 24 hours ago
		lastSync := time.Now().Add(-24 * time.Hour)
		query = fmt.Sprintf("modified:\">=%s\"", lastSync.Format("2006-01-02T15:04:05"))
	}

	// Use the Index method we just added to Indexer
	err := t.indexer.Index(ctx, query)
	if err != nil {
		return err
	}

	return nil
}
