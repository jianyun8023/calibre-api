package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// DeleteBookTask handles asynchronous book deletion from Search Engine
type DeleteBookTask struct {
	id       string
	bookID   int64
	searcher semantic.Searcher
	status   TaskStatus
	mu       sync.RWMutex
	stopChan chan struct{}
}

func NewDeleteBookTask(id string, bookID int64, searcher semantic.Searcher) *DeleteBookTask {
	return &DeleteBookTask{
		id:       id,
		bookID:   bookID,
		searcher: searcher,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeDeleteBook,
			Mode:      TaskModeFull, // Mode doesn't really apply here
			State:     "pending",
			Progress:  0,
			Message:   fmt.Sprintf("Pending deletion of book %d", bookID),
			StartTime: time.Now(),
		},
		stopChan: make(chan struct{}),
	}
}

func (t *DeleteBookTask) Run() error {
	t.mu.Lock()
	t.status.State = "running"
	t.status.Message = fmt.Sprintf("Deleting book %d from search index", t.bookID)
	t.mu.Unlock()

	// Perform deletion
	err := t.searcher.DeleteBook(t.bookID)

	t.mu.Lock()
	defer t.mu.Unlock()

	if err != nil {
		t.status.State = "error"
		t.status.Error = err.Error()
		t.status.Message = fmt.Sprintf("Failed to delete book %d: %v", t.bookID, err)
		return err
	}

	t.status.State = "completed"
	t.status.Progress = 100
	t.status.Message = fmt.Sprintf("Successfully deleted book %d", t.bookID)
	t.status.EndTime = time.Now()
	return nil
}

func (t *DeleteBookTask) Stop() {
	// Deletion is atomic/fast, so stop doesn't do much
	close(t.stopChan)
}

func (t *DeleteBookTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// UpdateMetadataTask handles asynchronous metadata update in Search Engine
type UpdateMetadataTask struct {
	id       string
	book     semantic.Book
	searcher semantic.Searcher
	status   TaskStatus
	mu       sync.RWMutex
	stopChan chan struct{}
}

func NewUpdateMetadataTask(id string, book semantic.Book, searcher semantic.Searcher) *UpdateMetadataTask {
	return &UpdateMetadataTask{
		id:       id,
		book:     book,
		searcher: searcher,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeUpdateMetadata,
			Mode:      TaskModeFull,
			State:     "pending",
			Progress:  0,
			Message:   fmt.Sprintf("Pending metadata update for book %d", book.ID),
			StartTime: time.Now(),
		},
		stopChan: make(chan struct{}),
	}
}

func (t *UpdateMetadataTask) Run() error {
	t.mu.Lock()
	t.status.State = "running"
	t.status.Message = fmt.Sprintf("Updating metadata for book %d in search index", t.book.ID)
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// Perform update (re-index)
	// Note: IndexBooks handles embedding generation internally
	// We need a context, using background since task has its own lifecycle
	ctx := context.Background()
	err := t.searcher.IndexBooks(ctx, []semantic.Book{t.book})

	t.mu.Lock()
	defer t.mu.Unlock()

	if err != nil {
		t.status.State = "error"
		t.status.Error = err.Error()
		t.status.Message = fmt.Sprintf("Failed to update book %d: %v", t.book.ID, err)
		return err
	}

	t.status.State = "completed"
	t.status.Progress = 100
	t.status.Message = fmt.Sprintf("Successfully updated book %d", t.book.ID)
	t.status.EndTime = time.Now()
	return nil
}

func (t *UpdateMetadataTask) Stop() {
	close(t.stopChan)
}

func (t *UpdateMetadataTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}
