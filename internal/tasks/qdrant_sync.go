package tasks

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

type QdrantSyncTask struct {
	id           string
	mode         TaskMode
	contentApi   *content.Api
	qdrantClient *qdrant.Client
	searcher     *qdrant.Searcher
	status       TaskStatus
	mu           sync.RWMutex
	cancel       context.CancelFunc
}

func NewQdrantSyncTask(id string, mode TaskMode, contentApi *content.Api, qdrantClient *qdrant.Client, searcher *qdrant.Searcher) *QdrantSyncTask {
	return &QdrantSyncTask{
		id:           id,
		mode:         mode,
		contentApi:   contentApi,
		qdrantClient: qdrantClient,
		searcher:     searcher,
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeQdrantSync,
			Mode:      mode,
			State:     "idle",
			StartTime: time.Now(),
			Message:   "Initializing...",
		},
	}
}

func (t *QdrantSyncTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *QdrantSyncTask) Stop() {
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

func (t *QdrantSyncTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Starting Qdrant sync..."
	t.mu.Unlock()

	// Determine query based on mode
	query := ""
	if t.mode == TaskModeIncremental {
		// Get max book ID from Qdrant
		// Since Qdrant points have integer IDs (uint64), we can try to find the max ID
		// But Qdrant doesn't support "max" aggregation easily on IDs directly via API unless we scroll/sort
		// For now, let's assume full sync or implement a way to get max ID if needed.
		// Or we can use a simple "scroll" with sort by ID desc limit 1

		// For simplicity in this migration, let's just do full sync if incremental is requested but we can't determine offset easily
		// Or we can implement GetMaxID in searcher

		// Let's try to get max ID using searcher if available
		// Assuming searcher has a method for this or we add one.
		// For now, let's just log that we are doing full sync or implement a basic check

		t.mu.Lock()
		t.status.Message = "Incremental sync requested. Checking Qdrant state..."
		t.mu.Unlock()

		if t.searcher != nil {
			maxID, err := t.searcher.GetMaxID()
			if err != nil {
				log.Printf("Failed to get max ID from Qdrant: %v. Falling back to full sync.", err)
			} else {
				if maxID > 0 {
					query = fmt.Sprintf("id:>%d", maxID)
					t.mu.Lock()
					t.status.Message = fmt.Sprintf("Incremental sync: fetching books with ID > %d", maxID)
					t.mu.Unlock()
				}
			}
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

	batchSize := 100 // Qdrant batch size
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

		// Enrich books
		enrichedBooks := content.EnrichBooks(data)

		// Call Searcher.IndexBooks
		if t.searcher != nil {
			// Convert to semantic books
			var semBooks []semantic.Book
			for _, book := range enrichedBooks {
				semBooks = append(semBooks, semantic.Book{
					ID:           book.ID,
					Title:        book.Title,
					Authors:      book.Authors,
					AuthorSort:   book.AuthorSort,
					Publisher:    book.Publisher,
					Isbn:         book.Isbn,
					Rating:       book.Rating,
					Tags:         book.Tags,
					Languages:    book.Languages,
					Comments:     book.Comments,
					PubDate:      book.PubDate,
					LastModified: book.LastModified,
					SeriesIndex:  book.SeriesIndex,
					Size:         book.Size,
					Identifiers:  book.Identifiers,
					Cover:        book.Cover,
					FilePath:     book.FilePath,
				})
			}

			if err := t.searcher.IndexBooks(ctx, semBooks); err != nil {
				log.Printf("Error indexing batch: %v", err)
				continue
			}
		}
	}

	t.mu.Lock()
	t.status.Progress = 100
	t.status.State = "completed"
	t.status.Message = "Sync completed"
	t.mu.Unlock()

	return nil
}
