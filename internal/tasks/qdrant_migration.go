package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/milvus"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const (
	MigrationBatchSize         = 500 // 每批处理的书籍数量
	MigrationProgressFile      = ".migration_progress.json"
	MigrationProgressSaveEvery = 5000 // 每处理多少条保存一次进度
)

// QdrantMigrationTask handles migration from Milvus to Qdrant
type QdrantMigrationTask struct {
	milvusClient  *milvus.Client
	qdrantClient  *qdrant.Client
	contentClient *content.Api

	status TaskStatus
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Migration progress
	progress MigrationProgress
}

// MigrationProgress tracks migration state
type MigrationProgress struct {
	TotalBooks     int64     `json:"total_books"`
	MigratedBooks  int64     `json:"migrated_books"`
	LastMigratedID int64     `json:"last_migrated_id"`
	StartTime      time.Time `json:"start_time"`
	LastUpdateTime time.Time `json:"last_update_time"`
	Errors         []string  `json:"errors"`
}

// NewQdrantMigrationTask creates a new migration task
func NewQdrantMigrationTask(
	milvusClient *milvus.Client,
	qdrantClient *qdrant.Client,
	contentClient *content.Api,
) *QdrantMigrationTask {
	ctx, cancel := context.WithCancel(context.Background())

	return &QdrantMigrationTask{
		milvusClient:  milvusClient,
		qdrantClient:  qdrantClient,
		contentClient: contentClient,
		ctx:           ctx,
		cancel:        cancel,
		status: TaskStatus{
			Type:      TaskTypeQdrantMigration,
			State:     "pending",
			StartTime: time.Now(),
		},
	}
}

// GetStatus returns the current task status
func (t *QdrantMigrationTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// Stop stops the migration task
func (t *QdrantMigrationTask) Stop() {
	t.cancel()
	t.mu.Lock()
	t.status.State = "stopped"
	t.status.Message = "Migration stopped by user"
	t.mu.Unlock()
}

// Run executes the migration
func (t *QdrantMigrationTask) Run() error {
	t.mu.Lock()
	t.status.State = "running"
	t.status.Message = "Starting migration from Milvus to Qdrant"
	t.mu.Unlock()

	// Load progress if exists
	if err := t.loadProgress(); err != nil {
		log.Printf("No previous progress found, starting fresh migration: %v", err)
		t.progress = MigrationProgress{
			StartTime: time.Now(),
		}
	} else {
		log.Printf("Resuming migration from last migrated ID: %d", t.progress.LastMigratedID)
	}

	// Get total count from Milvus
	totalCount, err := t.milvusClient.GetCollectionStats()
	if err != nil {
		return fmt.Errorf("failed to get Milvus collection stats: %w", err)
	}
	t.progress.TotalBooks = totalCount

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books in Milvus, starting migration...", totalCount)
	t.mu.Unlock()

	// Migrate data in batches
	if err := t.migrateBatches(); err != nil {
		t.mu.Lock()
		t.status.State = "error"
		t.status.Error = err.Error()
		t.mu.Unlock()
		return err
	}

	// Save final progress
	if err := t.saveProgress(); err != nil {
		log.Printf("Warning: failed to save final progress: %v", err)
	}

	t.mu.Lock()
	t.status.State = "completed"
	t.status.Message = fmt.Sprintf("Migration completed: %d books migrated", t.progress.MigratedBooks)
	t.status.EndTime = time.Now()
	t.mu.Unlock()

	// Clean up progress file
	os.Remove(MigrationProgressFile)

	return nil
}

// migrateBatches migrates data in batches
// Strategy: Get book metadata from Calibre first, then supplement with vectors from Milvus
func (t *QdrantMigrationTask) migrateBatches() error {
	batchNum := 0

	for {
		select {
		case <-t.ctx.Done():
			return fmt.Errorf("migration cancelled")
		default:
		}

		batchNum++

		// Get book IDs from Calibre (much faster than fetching individual metadata)
		query := fmt.Sprintf("id:>%d", t.progress.LastMigratedID)
		bookIDs, err := t.contentClient.GetAllBooksIds(query)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get book IDs from Calibre: %v", err)
			t.progress.Errors = append(t.progress.Errors, errMsg)
			return fmt.Errorf("%s", errMsg)
		}

		if len(bookIDs) == 0 {
			break // No more books
		}

		// Process in batches
		for i := 0; i < len(bookIDs); i += MigrationBatchSize {
			select {
			case <-t.ctx.Done():
				return fmt.Errorf("migration cancelled")
			default:
			}

			end := i + MigrationBatchSize
			if end > len(bookIDs) {
				end = len(bookIDs)
			}
			batchIDs := bookIDs[i:end]

			// Fetch metadata from Calibre in batch
			books, err := t.contentClient.GetBookMetaDatas(batchIDs, "library")
			if err != nil {
				log.Printf("Warning: failed to fetch batch metadata: %v", err)
				continue
			}

			// Enrich with cover and file path
			books = content.EnrichBooks(books)

			// Process this batch (fetch vectors from Milvus and write to Qdrant)
			if err := t.processBatchWithBooks(books, batchNum); err != nil {
				errMsg := fmt.Sprintf("failed to process batch %d: %v", batchNum, err)
				t.progress.Errors = append(t.progress.Errors, errMsg)
				log.Printf("Error: %s", errMsg)
				// Don't return error, continue with next batch
			}

			batchNum++

			// Save progress periodically
			if t.progress.MigratedBooks%MigrationProgressSaveEvery == 0 {
				if err := t.saveProgress(); err != nil {
					log.Printf("Warning: failed to save progress: %v", err)
				}
			}

			// Update status
			percentage := float64(t.progress.MigratedBooks) / float64(t.progress.TotalBooks) * 100
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Migrating: %d/%d books (%.1f%%)",
				t.progress.MigratedBooks, t.progress.TotalBooks, percentage)
			t.mu.Unlock()
		}

		// All done if we got less than batch size
		if len(bookIDs) < 10000 { // GetAllBooksIds might return all remaining
			break
		}
	}

	return nil
}

// processBatchWithBooks processes a batch of books with metadata from Calibre
func (t *QdrantMigrationTask) processBatchWithBooks(books []content.Book, batchNum int) error {
	if len(books) == 0 {
		return nil
	}

	// Extract book IDs
	bookIDs := make([]int64, len(books))
	for i, book := range books {
		bookIDs[i] = book.ID
	}

	// Build Milvus query expression for these book IDs
	// e.g., "book_id in [1,2,3,4,5]"
	idStrs := make([]string, len(bookIDs))
	for i, id := range bookIDs {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	expr := fmt.Sprintf("book_id in [%s]", strings.Join(idStrs, ","))

	// Query vectors from Milvus
	results, err := t.milvusClient.QueryBatch(expr, len(bookIDs), 0)
	if err != nil {
		log.Printf("Warning: failed to query vectors from Milvus: %v", err)
		// Continue without vectors - we'll skip books without vectors
		results = []entity.Column{}
	}

	// Build a map of book_id -> vector
	vectorMap := make(map[int64][]float32)
	if len(results) > 0 {
		var idColumn *entity.ColumnInt64
		var vectorColumn *entity.ColumnFloatVector

		for _, col := range results {
			switch c := col.(type) {
			case *entity.ColumnInt64:
				if c.Name() == "book_id" {
					idColumn = c
				}
			case *entity.ColumnFloatVector:
				if c.Name() == "embedding" {
					vectorColumn = c
				}
			}
		}

		if idColumn != nil && vectorColumn != nil {
			ids := idColumn.Data()
			vectors := vectorColumn.Data()

			if len(ids) == len(vectors) {
				for i, id := range ids {
					vectorMap[id] = vectors[i]
				}
			}
		}
	}

	// Create Qdrant points for books that have vectors
	points := make([]qdrant.Point, 0, len(books))
	skippedCount := 0

	for _, book := range books {
		vector, hasVector := vectorMap[book.ID]
		if !hasVector {
			log.Printf("Warning: book ID %d has no vector in Milvus, skipping", book.ID)
			skippedCount++
			continue
		}

		// Convert content.Book to semantic.Book
		semanticBook := semantic.Book{
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
		}

		point := qdrant.Point{
			ID:      uint64(book.ID),
			Vector:  vector,
			Payload: qdrant.BookToPayload(semanticBook),
		}
		points = append(points, point)
	}

	if skippedCount > 0 {
		log.Printf("Skipped %d books without vectors in batch %d", skippedCount, batchNum)
	}

	// Insert into Qdrant
	if len(points) > 0 {
		if err := t.qdrantClient.UpsertPoints(t.ctx, points); err != nil {
			return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
		}

		// Update progress
		t.progress.MigratedBooks += int64(len(points))
		if len(books) > 0 {
			// Update last migrated ID to the last book in this batch
			t.progress.LastMigratedID = books[len(books)-1].ID
		}
		t.progress.LastUpdateTime = time.Now()

		log.Printf("Batch %d: migrated %d books (skipped %d without vectors)",
			batchNum, len(points), skippedCount)
	}

	return nil
}

// processBatch processes a batch of results from Milvus (legacy method, kept for reference)
func (t *QdrantMigrationTask) processBatch(results []entity.Column, batchNum int) error {
	// Extract book IDs and vectors
	var bookIDs []int64
	var vectors [][]float32

	// Find the book_id column
	var idColumn *entity.ColumnInt64
	var vectorColumn *entity.ColumnFloatVector

	for _, col := range results {
		switch c := col.(type) {
		case *entity.ColumnInt64:
			if c.Name() == "book_id" {
				idColumn = c
			}
		case *entity.ColumnFloatVector:
			if c.Name() == "embedding" {
				vectorColumn = c
			}
		}
	}

	if idColumn == nil || vectorColumn == nil {
		return fmt.Errorf("missing book_id or embedding column in batch %d", batchNum)
	}

	bookIDs = idColumn.Data()

	// Extract vectors - ColumnFloatVector.Data() returns [][]float32
	vectors = vectorColumn.Data()

	if len(bookIDs) != len(vectors) {
		return fmt.Errorf("mismatch between book IDs (%d) and vectors (%d) in batch %d",
			len(bookIDs), len(vectors), batchNum)
	}

	// Fetch metadata from Calibre API
	books, err := t.fetchBooksMetadata(bookIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata: %w", err)
	}

	// Create Qdrant points
	points := make([]qdrant.Point, 0, len(bookIDs))
	for i, bookID := range bookIDs {
		// Find corresponding book metadata
		var book semantic.Book
		found := false
		for _, b := range books {
			if b.ID == bookID {
				book = b
				found = true
				break
			}
		}

		if !found {
			log.Printf("Warning: metadata not found for book ID %d, skipping", bookID)
			continue
		}

		// Create point
		point := qdrant.Point{
			ID:      uint64(bookID),
			Vector:  vectors[i],
			Payload: qdrant.BookToPayload(book),
		}
		points = append(points, point)
	}

	// Insert into Qdrant
	if len(points) > 0 {
		if err := t.qdrantClient.UpsertPoints(t.ctx, points); err != nil {
			return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
		}

		// Update progress
		t.progress.MigratedBooks += int64(len(points))
		if len(bookIDs) > 0 {
			t.progress.LastMigratedID = bookIDs[len(bookIDs)-1]
		}
		t.progress.LastUpdateTime = time.Now()
	}

	return nil
}

// fetchBooksMetadata fetches book metadata from Calibre API
func (t *QdrantMigrationTask) fetchBooksMetadata(bookIDs []int64) ([]semantic.Book, error) {
	// Build filter for these book IDs
	// This is a simplified version - you may need to adjust based on your API
	books := make([]semantic.Book, 0, len(bookIDs))

	for _, id := range bookIDs {
		// Fetch individual book metadata
		book, err := t.contentClient.GetBookDetail(id)
		if err != nil {
			log.Printf("Warning: failed to fetch book %d: %v", id, err)
			continue
		}

		// Convert to semantic.Book
		semanticBook := semantic.Book{
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
		}
		books = append(books, semanticBook)
	}

	return books, nil
}

// saveProgress saves migration progress to file
func (t *QdrantMigrationTask) saveProgress() error {
	data, err := json.MarshalIndent(t.progress, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	if err := os.WriteFile(MigrationProgressFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %w", err)
	}

	return nil
}

// loadProgress loads migration progress from file
func (t *QdrantMigrationTask) loadProgress() error {
	data, err := os.ReadFile(MigrationProgressFile)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &t.progress); err != nil {
		return fmt.Errorf("failed to unmarshal progress: %w", err)
	}

	return nil
}
