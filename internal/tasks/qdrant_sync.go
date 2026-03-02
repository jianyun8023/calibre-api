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

type SearchSyncTask struct {
	id         string
	mode       TaskMode
	contentApi *content.Api
	searcher   semantic.Searcher
	status     TaskStatus
	mu         sync.RWMutex
	cancel     context.CancelFunc
	errors     []string // 记录同步过程中的错误
}

func NewSearchSyncTask(id string, mode TaskMode, contentApi *content.Api, searcher semantic.Searcher) *SearchSyncTask {
	return &SearchSyncTask{
		id:         id,
		mode:       mode,
		contentApi: contentApi,
		searcher:   searcher,
		errors:     make([]string, 0),
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeQdrantSync, // Keep type for now or rename to TaskTypeSearchSync if frontend updated
			Mode:      mode,
			State:     "idle",
			StartTime: time.Now(),
			Message:   "Initializing...",
		},
	}
}

func (t *SearchSyncTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *SearchSyncTask) Stop() {
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

func (t *SearchSyncTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Starting search index sync..."
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	var ids []int64

	if t.mode == TaskModeIncremental && t.searcher != nil {
		// 增量同步：比较 Calibre 和 Search Engine 的 ID 差异，同步缺失的书籍
		var err error
		ids, err = t.findMissingBooks(ctx)
		if err != nil {
			log.Printf("Failed to find missing books: %v. Falling back to full sync.", err)
			// 回退到全量同步
			ids, err = t.contentApi.GetAllBooksIds("")
			if err != nil {
				return fmt.Errorf("failed to get all book IDs: %w", err)
			}
		}
	} else {
		// 全量同步
		var err error
		ids, err = t.contentApi.GetAllBooksIds("")
		if err != nil {
			return fmt.Errorf("failed to get all book IDs: %w", err)
		}
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books to sync", len(ids))
	t.mu.Unlock()

	if len(ids) == 0 {
		t.mu.Lock()
		t.status.Progress = 100
		t.status.State = "completed"
		t.status.Message = "No books to sync"
		t.mu.Unlock()
		return nil
	}

	batchSize := 100 // Batch size
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
				errMsg := fmt.Sprintf("Batch %d-%d failed: %v", i, end, err)
				log.Printf("Error indexing batch: %v", err)
				t.mu.Lock()
				t.errors = append(t.errors, errMsg)
				t.status.Message = fmt.Sprintf("Processing batch %d-%d (errors: %d)", i, end, len(t.errors))
				t.mu.Unlock()
				// 广播任务进度更新
				GetManager().BroadcastTaskProgress(t.id)
				continue
			}

			// 成功处理一批，更新进度
			GetManager().BroadcastTaskProgress(t.id)
		}
	}

	t.mu.Lock()
	t.status.Progress = 100
	if len(t.errors) > 0 {
		t.status.State = "completed"
		t.status.Message = fmt.Sprintf("Sync completed with %d errors", len(t.errors))
		// 将错误信息放入 Error 字段，方便前端显示
		t.status.Error = fmt.Sprintf("%d batches failed. Last error: %s", len(t.errors), t.errors[len(t.errors)-1])
	} else {
		t.status.State = "completed"
		t.status.Message = "Sync completed successfully"
	}
	t.mu.Unlock()
	// 完成时最后广播一次
	GetManager().BroadcastTaskProgress(t.id)

	return nil
}

// findMissingBooks 比较 Calibre 和 Search Engine 的 ID，返回缺失的书籍 ID
func (t *SearchSyncTask) findMissingBooks(ctx context.Context) ([]int64, error) {
	t.mu.Lock()
	t.status.Message = "Incremental sync: fetching Calibre book IDs..."
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 1. 获取 Calibre 中所有书籍 ID
	calibreIDs, err := t.contentApi.GetAllBooksIds("")
	if err != nil {
		return nil, fmt.Errorf("failed to get calibre IDs: %w", err)
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books in Calibre. Fetching search index IDs...", len(calibreIDs))
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 2. 获取 Search Engine 中所有书籍 ID
	searchIDMap := make(map[int64]bool)
	cursor := ""
	limit := 1000

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		books, _, nextCursor, err := t.searcher.GetAllWithCursor(limit, cursor)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch from search engine: %w", err)
		}

		addedNew := false
		for _, b := range books {
			if !searchIDMap[b.ID] {
				searchIDMap[b.ID] = true
				addedNew = true
			}
		}

		if nextCursor == "" || len(books) == 0 {
			break
		}

		// 防止死循环：如果 cursor 没有变化，或者没有读到新数据（说明一直在读重复数据）
		if nextCursor == cursor || !addedNew {
			break
		}
		cursor = nextCursor
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Comparing %d Calibre books with %d search index items...", len(calibreIDs), len(searchIDMap))
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 3. 找出缺失的 ID
	var missingIDs []int64
	for _, id := range calibreIDs {
		if !searchIDMap[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	log.Printf("Incremental sync: found %d missing books out of %d total", len(missingIDs), len(calibreIDs))

	return missingIDs, nil
}
