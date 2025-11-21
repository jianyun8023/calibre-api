package qdrant

import (
	"context"
	"fmt"
	"strconv"
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

// EnsureIndexes ensures that required payload indexes exist
func (s *Searcher) EnsureIndexes(ctx context.Context) error {
	// Create index for last_modified (datetime type) to support order_by
	if err := s.client.CreatePayloadIndex(ctx, "last_modified", "datetime"); err != nil {
		return fmt.Errorf("failed to create last_modified index: %w", err)
	}

	// Create index for book_id (integer type) for filtering
	if err := s.client.CreatePayloadIndex(ctx, "book_id", "integer"); err != nil {
		return fmt.Errorf("failed to create book_id index: %w", err)
	}

	return nil
}

// GetMaxID retrieves the maximum book ID from Qdrant
func (s *Searcher) GetMaxID() (uint64, error) {
	ctx := context.Background()
	var maxID uint64
	var scrollOffset *uint64
	batchSize := 1000

	// Scroll through all points to find the max ID
	// Qdrant returns points ordered by ID, so the last point will have the max ID
	for {
		// We only need IDs, no payload or vectors
		points, nextOffset, err := s.client.Scroll(ctx, batchSize, scrollOffset, false)
		if err != nil {
			return 0, fmt.Errorf("qdrant scroll failed: %w", err)
		}

		if len(points) > 0 {
			// Update maxID with the ID of the last point in the batch
			lastPoint := points[len(points)-1]
			if lastPoint.ID > maxID {
				maxID = lastPoint.ID
			}
		}

		if nextOffset == nil {
			break
		}
		scrollOffset = nextOffset
	}

	return maxID, nil
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

// HybridSearchCombined performs hybrid search by combining keyword and semantic search results
// Strategy: First use vector search to get semantically relevant candidates,
// then rerank them using keyword matching for better precision
func (s *Searcher) HybridSearchCombined(query string, limit int) ([]semantic.Book, error) {
	// 1. Get a larger candidate set from Semantic Search (cast a wide net)
	// This ensures we capture semantically relevant books
	semanticLimit := limit * 3
	if semanticLimit < 100 {
		semanticLimit = 100
	}
	semanticResults, err := s.Search(query, semanticLimit)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	if len(semanticResults) == 0 {
		return []semantic.Book{}, nil
	}

	// 2. Rerank the semantic results using keyword matching
	// This provides precision on top of semantic recall
	queryLower := strings.ToLower(query)
	keywords := strings.Fields(queryLower) // Split query into keywords

	type scoredBook struct {
		Book          semantic.Book
		SemanticScore float64
		KeywordScore  float64
		FinalScore    float64
	}

	scoredBooks := make([]scoredBook, len(semanticResults))
	for i, result := range semanticResults {
		book := result.Book

		// Calculate keyword matching score
		keywordScore := s.calculateKeywordScore(book, keywords, queryLower)

		// Combine scores:
		// - Semantic score (normalized): 40% weight
		// - Keyword score (0-1): 60% weight
		// This gives more weight to exact keyword matches while maintaining semantic relevance
		normalizedSemanticScore := float64(result.Score) // Convert float32 to float64
		finalScore := (normalizedSemanticScore * 0.4) + (keywordScore * 0.6)

		scoredBooks[i] = scoredBook{
			Book:          book,
			SemanticScore: float64(result.Score),
			KeywordScore:  keywordScore,
			FinalScore:    finalScore,
		}
	}

	// 3. Sort by final score (descending order)
	// Using insertion sort since it's efficient for small arrays
	for i := 1; i < len(scoredBooks); i++ {
		key := scoredBooks[i]
		j := i - 1
		for j >= 0 && scoredBooks[j].FinalScore < key.FinalScore {
			scoredBooks[j+1] = scoredBooks[j]
			j--
		}
		scoredBooks[j+1] = key
	}

	// 4. Return top limit results
	resultCount := len(scoredBooks)
	if resultCount > limit {
		resultCount = limit
	}

	finalBooks := make([]semantic.Book, resultCount)
	for i := 0; i < resultCount; i++ {
		finalBooks[i] = scoredBooks[i].Book
	}

	return finalBooks, nil
}

// calculateKeywordScore calculates a keyword matching score for a book
// Returns a score between 0 and 1 based on keyword matches in title, authors, publisher, tags, and comments
func (s *Searcher) calculateKeywordScore(book semantic.Book, keywords []string, fullQuery string) float64 {
	if len(keywords) == 0 {
		return 0.0
	}

	score := 0.0
	maxScore := 0.0

	titleLower := strings.ToLower(book.Title)
	authorsLower := strings.ToLower(strings.Join(book.Authors, " "))
	publisherLower := strings.ToLower(book.Publisher)
	tagsLower := strings.ToLower(strings.Join(book.Tags, " "))
	commentsLower := strings.ToLower(book.Comments)

	// Exact full query match (highest weight)
	maxScore += 10.0
	if strings.Contains(titleLower, fullQuery) {
		score += 10.0 // Exact match in title
	} else if strings.Contains(authorsLower, fullQuery) {
		score += 8.0 // Exact match in authors
	} else if strings.Contains(publisherLower, fullQuery) {
		score += 6.0 // Exact match in publisher
	} else if strings.Contains(tagsLower, fullQuery) {
		score += 5.0 // Exact match in tags
	} else if strings.Contains(commentsLower, fullQuery) {
		score += 3.0 // Exact match in comments
	}

	// Individual keyword matches (weighted by field importance)
	for _, keyword := range keywords {
		if len(keyword) < 2 { // Skip very short keywords
			continue
		}

		maxScore += 5.0 // Max possible score per keyword

		// Title matches (highest priority)
		if strings.Contains(titleLower, keyword) {
			score += 3.0
		}

		// Author matches
		if strings.Contains(authorsLower, keyword) {
			score += 1.5
		}

		// Publisher matches
		if strings.Contains(publisherLower, keyword) {
			score += 1.0
		}

		// Tags matches
		if strings.Contains(tagsLower, keyword) {
			score += 0.8
		}

		// Comments matches (lowest priority)
		if strings.Contains(commentsLower, keyword) {
			score += 0.3
		}
	}

	// Normalize to 0-1 range
	if maxScore > 0 {
		return score / maxScore
	}
	return 0.0
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
		// Use Qdrant native filter for precise matching (indexed keyword field)
		return s.filterByQdrantMatch(keyword, "publisher", limit, offset)
	case "author":
		// Use Qdrant native filter for precise matching (indexed keyword field)
		return s.filterByQdrantMatch(keyword, "authors", limit, offset)
	case "tags":
		// Use Qdrant native filter for precise matching (indexed keyword field)
		return s.filterByQdrantMatch(keyword, "tags", limit, offset)
	case "isbn":
		// Use Qdrant native filter for exact matching (indexed keyword field)
		return s.filterByQdrantExact(keyword, "isbn", limit, offset)
	case "id":
		id, err := strconv.ParseUint(keyword, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid id: %w", err)
		}
		point, err := s.client.Retrieve(ctx, id)
		if err != nil {
			return nil, 0, fmt.Errorf("retrieve failed: %w", err)
		}
		if point == nil {
			return []semantic.Book{}, 0, nil
		}
		return []semantic.Book{PayloadToBook(point.ID, point.Payload)}, 1, nil
	default:
		// Return all books if no filter type specified, sorted by book_id desc
		return s.getAllBooksSorted(limit, offset)
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

// GetRecent retrieves recent books sorted by book_id descending (newest first)
func (s *Searcher) GetRecent(limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Get total count
	total, err := s.client.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []semantic.Book{}, 0, nil
	}

	// Create OrderBy using book_id payload field (indexed as integer)
	// Order by book_id descending to get newest books first
	orderBy := &OrderBy{
		Key:       "book_id",
		Direction: "desc",
	}

	// For offset-based pagination with order_by, we need to fetch from beginning
	// This is a limitation but acceptable for "recent" endpoint which usually shows first page
	totalToFetch := offset + limit

	// Get ordered points from Qdrant
	points, _, err := s.client.ScrollWithOrder(ctx, totalToFetch, nil, false, orderBy)
	if err != nil {
		return nil, 0, fmt.Errorf("qdrant scroll with order failed: %w", err)
	}

	// Convert to books
	var allBooks []semantic.Book
	for _, point := range points {
		book := PayloadToBook(point.ID, point.Payload)
		allBooks = append(allBooks, book)
	}

	// Apply offset pagination
	start := offset
	end := offset + limit

	if start >= len(allBooks) {
		return []semantic.Book{}, int64(total), nil
	}
	if end > len(allBooks) {
		end = len(allBooks)
	}

	return allBooks[start:end], int64(total), nil
}

// GetAllWithCursor retrieves all books using cursor-based pagination with order_by on book_id
// Requires book_id payload index to be created
func (s *Searcher) GetAllWithCursor(limit int, cursor string) ([]semantic.Book, int64, string, error) {
	ctx := context.Background()

	// Get total count
	total, err := s.client.Count(ctx)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get count: %w", err)
	}

	if total == 0 {
		return []semantic.Book{}, 0, "", nil
	}

	// Create OrderBy using book_id payload field (indexed as integer)
	// Order by book_id descending to get newest books first
	orderBy := &OrderBy{
		Key:       "book_id", // Use indexed payload field
		Direction: "desc",
	}

	// Parse cursor as start_from value (the book_id to start from)
	if cursor != "" {
		parsedID, err := strconv.ParseInt(cursor, 10, 64)
		if err == nil {
			orderBy.StartFrom = parsedID
		}
	}

	// Get ordered points from Qdrant using book_id ordering
	points, _, err := s.client.ScrollWithOrder(ctx, limit, nil, false, orderBy)
	if err != nil {
		return nil, 0, "", fmt.Errorf("qdrant scroll with order failed: %w", err)
	}

	// Convert to books
	books := make([]semantic.Book, len(points))
	for i, point := range points {
		books[i] = PayloadToBook(point.ID, point.Payload)
	}

	// Prepare next cursor (smallest book_id in current page for descending order)
	var nextCursor string
	if len(books) > 0 {
		nextCursor = strconv.FormatInt(books[len(books)-1].ID, 10)
	}

	return books, int64(total), nextCursor, nil
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

// IndexBooks indexes a batch of books by generating embeddings and upserting to Qdrant
func (s *Searcher) IndexBooks(ctx context.Context, books []semantic.Book) error {
	if len(books) == 0 {
		return nil
	}

	// 1. Prepare text for embedding
	texts := make([]string, len(books))
	for i, book := range books {
		// Combine title, authors, and comments for embedding
		// You can adjust this strategy based on what you want to be searchable semantically
		texts[i] = fmt.Sprintf("%s %s %s", book.Title, strings.Join(book.Authors, " "), book.Comments)
	}

	// 2. Generate embeddings
	embeddings, err := s.provider.Embed(texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	if len(embeddings) != len(books) {
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(books), len(embeddings))
	}

	// 3. Create points
	points := make([]Point, len(books))
	for i, book := range books {
		points[i] = Point{
			ID:      uint64(book.ID),
			Vector:  embeddings[i],
			Payload: BookToPayload(book),
		}
	}

	// 4. Upsert to Qdrant
	if err := s.client.UpsertPoints(ctx, points); err != nil {
		return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
	}

	return nil
}

// scrollAndFilterByField scrolls through all books and filters by a specific field
// For ISBN, use exact match; for other fields, use contains match
func (s *Searcher) scrollAndFilterByField(keyword string, fieldName string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()
	keywordLower := strings.ToLower(keyword)

	// Determine if exact match is needed
	exactMatch := (fieldName == "isbn")

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
					if exactMatch {
						// Exact match for ISBN
						match = strings.EqualFold(v, keyword)
					} else {
						// Contains match for other fields
						match = strings.Contains(strings.ToLower(v), keywordLower)
					}
				case []interface{}:
					// For array fields like authors, tags
					for _, item := range v {
						if str, ok := item.(string); ok {
							if exactMatch {
								match = strings.EqualFold(str, keyword)
							} else {
								match = strings.Contains(strings.ToLower(str), keywordLower)
							}
							if match {
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

// filterByQdrantMatch uses Qdrant's native Match filter for keyword fields (contains matching)
// This is much faster than scrollAndFilterByField as it uses indexed keyword fields
func (s *Searcher) filterByQdrantMatch(keyword string, fieldName string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Create Qdrant Match filter for keyword field
	// Match filter performs case-insensitive text matching for keyword fields
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": fieldName,
				"match": map[string]interface{}{
					"text": keyword,
				},
			},
		},
	}

	// Create OrderBy for book_id descending
	orderBy := &OrderBy{
		Key:       "book_id",
		Direction: "desc",
	}

	// For offset-based pagination with filters, we need to fetch all and slice
	// This is acceptable for filtered results which are usually smaller
	totalToFetch := offset + limit
	if totalToFetch > 1000 {
		totalToFetch = 1000 // Cap at 1000 to prevent excessive memory usage
	}

	// Scroll with filter and order
	req := ScrollRequest{
		Limit:       totalToFetch,
		WithPayload: true,
		WithVector:  false,
		Filter:      filter,
		OrderBy:     orderBy,
	}

	points, _, err := s.client.ScrollWithFilter(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("qdrant scroll with filter failed: %w", err)
	}

	// Convert to books
	var books []semantic.Book
	for _, point := range points {
		book := PayloadToBook(point.ID, point.Payload)
		books = append(books, book)
	}

	// Get total count (approximate from fetched results)
	total := int64(len(books))

	// Apply offset pagination
	start := offset
	end := offset + limit

	if start >= len(books) {
		return []semantic.Book{}, total, nil
	}
	if end > len(books) {
		end = len(books)
	}

	return books[start:end], total, nil
}

// filterByQdrantExact uses Qdrant's native Match filter for exact matching (e.g., ISBN)
func (s *Searcher) filterByQdrantExact(keyword string, fieldName string, limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Create Qdrant Match filter with exact value
	filter := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"key": fieldName,
				"match": map[string]interface{}{
					"value": keyword,
				},
			},
		},
	}

	// Create OrderBy for book_id descending
	orderBy := &OrderBy{
		Key:       "book_id",
		Direction: "desc",
	}

	// Scroll with filter and order
	req := ScrollRequest{
		Limit:       limit + offset,
		WithPayload: true,
		WithVector:  false,
		Filter:      filter,
		OrderBy:     orderBy,
	}

	points, _, err := s.client.ScrollWithFilter(ctx, req)
	if err != nil {
		return nil, 0, fmt.Errorf("qdrant scroll with filter failed: %w", err)
	}

	// Convert to books
	var books []semantic.Book
	for _, point := range points {
		book := PayloadToBook(point.ID, point.Payload)
		books = append(books, book)
	}

	total := int64(len(books))

	// Apply offset pagination
	start := offset
	end := offset + limit

	if start >= len(books) {
		return []semantic.Book{}, total, nil
	}
	if end > len(books) {
		end = len(books)
	}

	return books[start:end], total, nil
}

// getAllBooksSorted returns all books sorted by book_id descending using Qdrant's order_by
func (s *Searcher) getAllBooksSorted(limit, offset int) ([]semantic.Book, int64, error) {
	ctx := context.Background()

	// Get total count
	totalCount, err := s.client.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	if totalCount == 0 {
		return []semantic.Book{}, 0, nil
	}

	// Create OrderBy for book_id descending
	orderBy := &OrderBy{
		Key:       "book_id",
		Direction: "desc",
	}

	// For offset-based pagination with order_by, fetch from beginning
	totalToFetch := offset + limit

	// Scroll with order
	points, _, err := s.client.ScrollWithOrder(ctx, totalToFetch, nil, false, orderBy)
	if err != nil {
		return nil, 0, fmt.Errorf("qdrant scroll with order failed: %w", err)
	}

	// Convert to books
	var books []semantic.Book
	for _, point := range points {
		book := PayloadToBook(point.ID, point.Payload)
		books = append(books, book)
	}

	// Apply offset pagination
	start := offset
	end := offset + limit

	if start >= len(books) {
		return []semantic.Book{}, int64(totalCount), nil
	}
	if end > len(books) {
		end = len(books)
	}

	return books[start:end], int64(totalCount), nil
}
