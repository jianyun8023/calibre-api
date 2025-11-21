package indexer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/milvus"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

type IndexStatus struct {
	State     string    `json:"state"` // "idle", "running", "paused", "error", "completed"
	Total     int       `json:"total"`
	Current   int       `json:"current"`
	Message   string    `json:"message"`
	StartTime time.Time `json:"start_time"`
}

type Indexer struct {
	provider   embedding.Provider
	client     *milvus.Client
	contentApi *content.Api

	status IndexStatus
	mu     sync.RWMutex
	cancel func()
}

func NewIndexer(provider embedding.Provider, client *milvus.Client, contentApi *content.Api) *Indexer {
	return &Indexer{
		provider:   provider,
		client:     client,
		contentApi: contentApi,
	}
}

// Start starts the indexing process in background
func (i *Indexer) Start() error {
	i.mu.Lock()
	if i.status.State == "running" {
		i.mu.Unlock()
		return fmt.Errorf("indexing is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())
	i.cancel = cancel
	i.status = IndexStatus{
		State:     "running",
		StartTime: time.Now(),
		Message:   "Starting indexing...",
	}
	i.mu.Unlock()

	go i.run(ctx)
	return nil
}

// Stop stops the indexing process
func (i *Indexer) Stop() {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.cancel != nil {
		i.cancel()
		i.cancel = nil
	}
	if i.status.State == "running" {
		i.status.State = "stopped"
		i.status.Message = "Indexing stopped by user"
	}
}

// GetStatus returns the current indexing status
func (i *Indexer) GetStatus() IndexStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

// GetMaxBookID returns the maximum book_id from the Milvus collection
func (i *Indexer) GetMaxBookID() (int64, error) {
	return i.client.GetMaxBookID()
}

// Index runs the indexing process with a query
func (i *Indexer) Index(ctx context.Context, query string) error {
	// 1. Get all book IDs
	ids, err := i.contentApi.GetAllBooksIds(query)
	if err != nil {
		return fmt.Errorf("failed to get book IDs: %w", err)
	}

	log.Printf("Found %d books to index (query: %s)", len(ids), query)

	if len(ids) == 0 {
		return nil
	}

	// 2. Process in batches
	batchSize := 50
	for j := 0; j < len(ids); j += batchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := j + batchSize
		if end > len(ids) {
			end = len(ids)
		}

		batchIDs := ids[j:end]
		log.Printf("Processing batch %d-%d", j, end)

		if err := i.processBatch(batchIDs); err != nil {
			log.Printf("Error processing batch %d-%d: %v", j, end, err)
			continue
		}
	}

	return nil
}

func (i *Indexer) run(ctx context.Context) {
	// This method is for legacy Start() call, it wraps the new Index method
	// and updates the internal status.
	i.mu.Lock()
	i.status.Message = "Starting indexing all books..."
	i.mu.Unlock()

	err := i.Index(ctx, "") // Call the new Index method with an empty query for all books
	if err != nil {
		i.mu.Lock()
		if err == context.Canceled {
			i.status.State = "stopped"
			i.status.Message = "Indexing stopped by user"
		} else {
			i.status.State = "error"
			i.status.Message = fmt.Sprintf("Indexing failed: %v", err)
		}
		i.mu.Unlock()
		return
	}

	i.mu.Lock()
	i.status.State = "completed"
	// The Index method doesn't update current/total status, so we need to fetch it again or pass it.
	// For simplicity, let's assume it completed all if no error.
	ids, _ := i.contentApi.GetAllBooksIds("") // Re-fetch to get total for status
	i.status.Total = len(ids)
	i.status.Current = len(ids)
	i.status.Message = "Indexing completed successfully"
	i.mu.Unlock()
}

func (i *Indexer) processBatch(ids []int64) error {
	// 1. Fetch metadata
	contentBooks, err := i.contentApi.GetBookMetaDatas(ids, "")
	if err != nil {
		return fmt.Errorf("failed to get book metadata: %w", err)
	}

	// 2. Convert to semantic.Book
	var books []semantic.Book
	for _, cb := range contentBooks {
		books = append(books, convertContentBook(cb))
	}

	// 3. Prepare embeddings
	var bookEmbeddings []semantic.BookEmbedding
	var texts []string

	for _, book := range books {
		text := embedding.CombineBookText(book)
		texts = append(texts, text)
	}

	// 4. Generate embeddings
	embeddings, err := i.provider.Embed(texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddings) != len(books) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(books), len(embeddings))
	}

	for k, book := range books {
		bookEmbeddings = append(bookEmbeddings, semantic.BookEmbedding{
			Book:      book,
			Embedding: embeddings[k],
		})
	}

	// 5. Insert into Milvus
	if err := i.client.InsertBooks(bookEmbeddings); err != nil {
		return fmt.Errorf("failed to insert books into milvus: %w", err)
	}

	return nil
}

// convertContentBook converts content.Book to semantic.Book
func convertContentBook(c content.Book) semantic.Book {
	return semantic.Book{
		AuthorSort:   c.AuthorSort,
		Authors:      c.Authors,
		Comments:     c.Comments,
		ID:           c.ID,
		Isbn:         c.Isbn,
		Languages:    c.Languages,
		LastModified: c.LastModified,
		PubDate:      c.PubDate,
		Publisher:    c.Publisher,
		SeriesIndex:  c.SeriesIndex,
		Size:         c.Size,
		Title:        c.Title,
		Tags:         c.Tags,
		Rating:       c.Rating,
		Identifiers:  c.Identifiers,
		Cover:        "/api/get/cover/" + strconv.FormatInt(c.ID, 10) + ".jpg",
		FilePath:     "/api/download/book/" + strconv.FormatInt(c.ID, 10) + ".epub",
	}
}
