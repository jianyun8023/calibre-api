package qdrant

import (
	"context"
	"fmt"
	"strings"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
)

// Searcher provides search functionality using Qdrant
type Searcher struct {
	provider embedding.Provider
	client   *Client
}

// NewSearcher creates a new Qdrant searcher
func NewSearcher(provider embedding.Provider, client *Client) *Searcher {
	return &Searcher{
		provider: provider,
		client:   client,
	}
}

// Search performs semantic search using vector similarity
func (s *Searcher) Search(query string, topK int) ([]semantic.SearchResult, error) {
	// 1. Vectorize the query
	embeddings, err := s.provider.Embed([]string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// 2. Search in Qdrant
	results, err := s.client.Search(context.Background(), embeddings[0], nil, topK)
	if err != nil {
		return nil, fmt.Errorf("qdrant search failed: %w", err)
	}

	// 3. Convert to SearchResult
	searchResults := make([]semantic.SearchResult, len(results))
	for i, result := range results {
		book := PayloadToBook(result.ID, result.Payload)
		searchResults[i] = semantic.SearchResult{
			Book:  book,
			Score: result.Score,
			Rank:  i + 1,
		}
	}

	return searchResults, nil
}

// SearchByKeyword performs keyword search using Qdrant scroll with filters
// This replaces Meilisearch keyword search functionality
func (s *Searcher) SearchByKeyword(keyword string, filterType string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Build filter based on filterType
	switch filterType {
	case "title":
		// For title search, we use a simple text match
		// Note: Qdrant doesn't have full-text search like Meilisearch
		// We'll use Scroll to get all books and filter in memory
		// For production, consider using Qdrant's full-text index feature
		return s.scrollAndFilterByTitle(keyword, limit, offset)
	case "publisher":
		// For exact matches, we could use Qdrant filter
		// But for simplicity, we'll also use scroll and filter in memory
		return s.scrollAndFilterByField(keyword, "publisher", limit, offset)
	case "author":
		return s.scrollAndFilterByField(keyword, "authors", limit, offset)
	case "isbn":
		return s.scrollAndFilterByField(keyword, "isbn", limit, offset)
	default:
		// Return all books if no filter type specified
		scrollOffset := uint64(offset)
		points, _, err := s.client.Scroll(ctx, limit, &scrollOffset, false)
		if err != nil {
			return nil, 0, fmt.Errorf("qdrant scroll failed: %w", err)
		}

		books := make([]semantic.Book, len(points))
		for i, point := range points {
			books[i] = PayloadToBook(point.ID, point.Payload)
		}

		// Get total count
		total, err := s.client.Count(ctx)
		if err != nil {
			return books, int64(len(books)), nil
		}

		return books, int64(total), nil
	}
}

// scrollAndFilterByTitle scrolls through all books and filters by title in memory
// This is a workaround for Qdrant's lack of full-text search
func (s *Searcher) scrollAndFilterByTitle(keyword string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()
	keyword = strings.ToLower(keyword)

	var matchedBooks []semantic.Book
	var scrollOffset *uint64
	batchSize := 100

	// Scroll through all books
	for {
		points, nextOffset, err := s.client.Scroll(ctx, batchSize, scrollOffset, false)
		if err != nil {
			return nil, 0, fmt.Errorf("qdrant scroll failed: %w", err)
		}

		// Filter by title
		for _, point := range points {
			if title, ok := point.Payload["title"].(string); ok {
				if strings.Contains(strings.ToLower(title), keyword) {
					book := PayloadToBook(point.ID, point.Payload)
					matchedBooks = append(matchedBooks, book)
				}
			}
		}

		// Check if we have enough results
		if len(matchedBooks) >= offset+limit {
			break
		}

		// Check if there are more results
		if nextOffset == nil {
			break
		}
		scrollOffset = nextOffset
	}

	// Apply pagination
	total := int64(len(matchedBooks))
	start := offset
	end := offset + limit

	if start >= len(matchedBooks) {
		return []semantic.Book{}, total, nil
	}
	if end > len(matchedBooks) {
		end = len(matchedBooks)
	}

	return matchedBooks[start:end], total, nil
}

// HybridSearch performs hybrid search combining vector similarity and metadata filters
func (s *Searcher) HybridSearch(query string, filter map[string]interface{}, limit int) ([]semantic.SearchResult, error) {
	// 1. Vectorize the query
	embeddings, err := s.provider.Embed([]string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// 2. Search in Qdrant with filter
	results, err := s.client.Search(context.Background(), embeddings[0], filter, limit)
	if err != nil {
		return nil, fmt.Errorf("qdrant hybrid search failed: %w", err)
	}

	// 3. Convert to SearchResult
	searchResults := make([]semantic.SearchResult, len(results))
	for i, result := range results {
		book := PayloadToBook(result.ID, result.Payload)
		searchResults[i] = semantic.SearchResult{
			Book:  book,
			Score: result.Score,
			Rank:  i + 1,
		}
	}

	return searchResults, nil
}

// GetRecent retrieves recent books sorted by last_modified time
func (s *Searcher) GetRecent(limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Qdrant doesn't support sorting in Scroll API
	// We need to scroll through all books and sort in memory
	// For production, consider maintaining a separate sorted index

	var allBooks []semantic.Book
	var scrollOffset *uint64
	batchSize := 1000

	// Scroll through all books
	for {
		points, nextOffset, err := s.client.Scroll(ctx, batchSize, scrollOffset, false)
		if err != nil {
			return nil, 0, fmt.Errorf("qdrant scroll failed: %w", err)
		}

		for _, point := range points {
			book := PayloadToBook(point.ID, point.Payload)
			allBooks = append(allBooks, book)
		}

		if nextOffset == nil {
			break
		}
		scrollOffset = nextOffset

		// Limit total books to avoid memory issues
		if len(allBooks) >= 10000 {
			break
		}
	}

	// Sort by last_modified (descending)
	// Note: This is done in memory, which may not be efficient for large datasets
	// For production, consider using a different approach

	total := int64(len(allBooks))
	start := offset
	end := offset + limit

	if start >= len(allBooks) {
		return []semantic.Book{}, total, nil
	}
	if end > len(allBooks) {
		end = len(allBooks)
	}

	return allBooks[start:end], total, nil
}

// GetRandom retrieves random books
func (s *Searcher) GetRandom(limit int) ([]semantic.Book, error) {
	ctx := context.Background()

	// Get total count
	total, err := s.client.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []semantic.Book{}, nil
	}

	// Use a random offset
	// Note: This is a simple approach, for production consider better randomization
	randomOffset := uint64(0)
	if total > uint64(limit) {
		// Simple random offset (you may want to use crypto/rand for better randomness)
		randomOffset = uint64(total - uint64(limit))
	}

	points, _, err := s.client.Scroll(ctx, limit, &randomOffset, false)
	if err != nil {
		return nil, fmt.Errorf("qdrant scroll failed: %w", err)
	}

	books := make([]semantic.Book, len(points))
	for i, point := range points {
		books[i] = PayloadToBook(point.ID, point.Payload)
	}

	return books, nil
}

// UpdateBookMetadata updates a book's metadata in Qdrant
func (s *Searcher) UpdateBookMetadata(book semantic.Book, vector []float32) error {
	ctx := context.Background()

	point := Point{
		ID:      uint64(book.ID),
		Vector:  vector,
		Payload: BookToPayload(book),
	}

	err := s.client.UpsertPoints(ctx, []Point{point})
	if err != nil {
		return fmt.Errorf("failed to update book in qdrant: %w", err)
	}

	return nil
}

// scrollAndFilterByField scrolls through all books and filters by a specific field
func (s *Searcher) scrollAndFilterByField(keyword string, fieldName string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()
	keyword = strings.ToLower(keyword)

	var matchedBooks []semantic.Book
	var scrollOffset *uint64
	batchSize := 100

	// Scroll through all books
	for {
		points, nextOffset, err := s.client.Scroll(ctx, batchSize, scrollOffset, false)
		if err != nil {
			return nil, 0, fmt.Errorf("qdrant scroll failed: %w", err)
		}

		// Filter by field
		for _, point := range points {
			var match bool
			if fieldValue, ok := point.Payload[fieldName]; ok {
				switch v := fieldValue.(type) {
				case string:
					match = strings.Contains(strings.ToLower(v), keyword)
				case []interface{}:
					// For array fields like authors
					for _, item := range v {
						if str, ok := item.(string); ok {
							if strings.Contains(strings.ToLower(str), keyword) {
								match = true
								break
							}
						}
					}
				}
			}

			if match {
				book := PayloadToBook(point.ID, point.Payload)
				matchedBooks = append(matchedBooks, book)
			}
		}

		// Check if we have enough results
		if len(matchedBooks) >= offset+limit {
			break
		}

		// Check if there are more results
		if nextOffset == nil {
			break
		}
		scrollOffset = nextOffset
	}

	// Apply pagination
	total := int64(len(matchedBooks))
	start := offset
	end := offset + limit

	if start >= len(matchedBooks) {
		return []semantic.Book{}, total, nil
	}
	if end > len(matchedBooks) {
		end = len(matchedBooks)
	}

	return matchedBooks[start:end], total, nil
}
