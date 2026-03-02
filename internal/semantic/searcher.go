package semantic

import (
	"context"
)

// Searcher provides search functionality interface
type Searcher interface {
	// Search performs semantic search using vector similarity
	Search(query string, limit int) ([]SearchResult, error)

	// HybridSearchCombined performs hybrid search by combining keyword and semantic search results
	HybridSearchCombined(query string, limit int) ([]Book, error)

	// SearchByKeyword performs keyword search
	SearchByKeyword(keyword, filterType string, limit, offset int) ([]Book, int64, error)

	// HybridSearch performs hybrid search combining vector similarity and metadata filters
	HybridSearch(query string, filter map[string]interface{}, limit int) ([]SearchResult, error)

	// GetRecent retrieves recent books sorted by book_id descending (newest first)
	GetRecent(limit, offset int) ([]Book, int64, error)

	// GetAllWithCursor retrieves all books using cursor-based pagination
	GetAllWithCursor(limit int, cursor string) ([]Book, int64, string, error)

	// GetRandom retrieves random books
	GetRandom(limit int) ([]Book, error)

	// UpdateBookMetadata updates a book's metadata
	UpdateBookMetadata(book Book, vector []float32) error

	// DeleteBook deletes a book
	DeleteBook(bookID int64) error

	// IndexBooks indexes a batch of books
	IndexBooks(ctx context.Context, books []Book) error

	// GetBookToc retrieves TOC data for a book
	GetBookToc(bookID int64) (interface{}, error)

	// UpdateToc updates the TOC field for a book
	UpdateToc(bookID int64, toc interface{}) error

	// EnsureIndexes ensures that required indexes exist
	EnsureIndexes(ctx context.Context) error
}
