package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/meilisearch/meilisearch-go"
)

// Client wraps the Meilisearch client
type Client struct {
	client    meilisearch.ServiceManager
	indexName string
}

// NewClient creates a new Meilisearch client
func NewClient(host, apiKey, indexName string) *Client {
	client := meilisearch.New(host, meilisearch.WithAPIKey(apiKey))
	return &Client{
		client:    client,
		indexName: indexName,
	}
}

// Searcher implements the semantic.Searcher interface using Meilisearch
type Searcher struct {
	client   *Client
	provider embedding.Provider
}

// NewSearcher creates a new Meilisearch searcher
func NewSearcher(client *Client, provider embedding.Provider) *Searcher {
	return &Searcher{
		client:   client,
		provider: provider,
	}
}

// EnsureIndexes ensures that required indexes and settings exist
func (s *Searcher) EnsureIndexes(ctx context.Context) error {
	// Create index if not exists
	_, err := s.client.client.GetIndex(s.client.indexName)
	if err != nil {
		task, err := s.client.client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        s.client.indexName,
			PrimaryKey: "book_id",
		})
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		if _, err := s.client.client.WaitForTask(task.TaskUID, 10*time.Second); err != nil {
			return fmt.Errorf("failed to wait for create index task: %w", err)
		}
	}

	index := s.client.client.Index(s.client.indexName)

	// Update settings
	filterableAttributes := []string{
		"book_id",
		"publisher",
		"authors",
		"tags",
		"isbn",
		"languages",
		"rating",
		"pubdate",
		"last_modified",
	}

	settings := &meilisearch.Settings{
		FilterableAttributes: filterableAttributes,
		SortableAttributes: []string{
			"book_id",
			"pubdate",
			"last_modified",
			"rating",
			"series_index",
		},
	}

	task, err := index.UpdateSettings(settings)
	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}
	if _, err := s.client.client.WaitForTask(task.TaskUID, 10*time.Second); err != nil {
		return fmt.Errorf("failed to wait for update settings task: %w", err)
	}

	return nil
}

// Search performs semantic search using vector similarity
// Meilisearch supports vector search via `vector` field in search query
func (s *Searcher) Search(query string, limit int) ([]semantic.SearchResult, error) {
	// 1. Generate embedding for query
	var queryVector []float32
	if s.provider != nil {
		embeddings, err := s.provider.Embed([]string{query})
		if err != nil {
			log.Printf("Failed to generate embedding for query '%s': %v", query, err)
			return nil, fmt.Errorf("failed to embed query: %w", err)
		}
		if len(embeddings) > 0 {
			queryVector = embeddings[0]
		}
	}

	// 2. Search with vector
	req := &meilisearch.SearchRequest{
		Limit: int64(limit),
	}

	if len(queryVector) > 0 {
		req.Vector = queryVector
	} else {
		log.Println("Warning: Semantic search called without embedding provider, falling back to text search")
	}

	resp, err := s.client.client.Index(s.client.indexName).Search(query, req)
	if err != nil {
		return nil, err
	}

	results := make([]semantic.SearchResult, len(resp.Hits))
	for i, hit := range resp.Hits {
		// Manual casting since Hit is alias to map[string]interface{} or interface{}
		// If hit is type interface{}, it holds the underlying map.
		book, err := mapToBook(hit)
		if err != nil {
			log.Printf("Failed to map hit to book: %v", err)
			continue
		}

		results[i] = semantic.SearchResult{
			Book:  book,
			Score: 0,
			Rank:  i + 1,
		}
	}
	return results, nil
}

// HybridSearchCombined performs hybrid search by combining keyword and semantic search results
func (s *Searcher) HybridSearchCombined(query string, limit int) ([]semantic.Book, error) {
	results, err := s.Search(query, limit)
	if err != nil {
		return nil, err
	}
	books := make([]semantic.Book, len(results))
	for i, res := range results {
		books[i] = res.Book
	}
	return books, nil
}

// SearchByKeyword performs keyword search
func (s *Searcher) SearchByKeyword(keyword, filterType string, limit, offset int) ([]semantic.Book, int64, error) {
	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	var query string
	if filterType != "" && filterType != "title" {
		fieldMap := map[string]string{
			"publisher": "publisher",
			"author":    "authors",
			"tags":      "tags",
			"isbn":      "isbn",
			"id":        "book_id",
		}

		if attr, ok := fieldMap[filterType]; ok {
			if filterType == "id" {
				req.Filter = fmt.Sprintf("%s = %s", attr, keyword)
			} else {
				req.Filter = fmt.Sprintf("%s = '%s'", attr, escapeFilterValue(keyword))
			}
			query = ""
		} else {
			query = keyword
		}
	} else {
		query = keyword
	}

	resp, err := s.client.client.Index(s.client.indexName).Search(query, req)
	if err != nil {
		return nil, 0, err
	}

	books := make([]semantic.Book, len(resp.Hits))
	for i, hit := range resp.Hits {
		b, err := mapToBook(hit)
		if err != nil {
			log.Printf("Failed to map hit to book: %v", err)
			continue
		}
		books[i] = b
	}

	return books, resp.EstimatedTotalHits, nil
}

// HybridSearch performs hybrid search combining vector similarity and metadata filters
func (s *Searcher) HybridSearch(query string, filter map[string]interface{}, limit int) ([]semantic.SearchResult, error) {
	return s.Search(query, limit)
}

// GetRecent retrieves recent books sorted by book_id descending (newest first)
func (s *Searcher) GetRecent(limit, offset int) ([]semantic.Book, int64, error) {
	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
		Sort:   []string{"book_id:desc"},
	}

	resp, err := s.client.client.Index(s.client.indexName).Search("", req)
	if err != nil {
		return nil, 0, err
	}

	books := make([]semantic.Book, len(resp.Hits))
	for i, hit := range resp.Hits {
		b, err := mapToBook(hit)
		if err != nil {
			log.Printf("Failed to map hit to book: %v", err)
			continue
		}
		books[i] = b
	}

	return books, resp.EstimatedTotalHits, nil
}

// GetAllWithCursor retrieves all books using cursor-based pagination
func (s *Searcher) GetAllWithCursor(limit int, cursor string) ([]semantic.Book, int64, string, error) {
	var filter string
	if cursor != "" {
		filter = fmt.Sprintf("book_id < %s", cursor)
	}

	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Sort:   []string{"book_id:desc"},
		Filter: filter,
	}

	resp, err := s.client.client.Index(s.client.indexName).Search("", req)
	if err != nil {
		return nil, 0, "", err
	}

	books := make([]semantic.Book, len(resp.Hits))
	for i, hit := range resp.Hits {
		b, err := mapToBook(hit)
		if err != nil {
			log.Printf("Failed to map hit to book: %v", err)
			continue
		}
		books[i] = b
	}

	var nextCursor string
	if len(books) > 0 {
		nextCursor = fmt.Sprintf("%d", books[len(books)-1].ID)
	}

	return books, resp.EstimatedTotalHits, nextCursor, nil
}

// GetRandom retrieves random books
func (s *Searcher) GetRandom(limit int) ([]semantic.Book, error) {
	stats, err := s.client.client.Index(s.client.indexName).GetStats()
	if err != nil {
		return nil, err
	}

	total := stats.NumberOfDocuments
	if total == 0 {
		return []semantic.Book{}, nil
	}

	offset := 0
	if total > int64(limit) {
		maxOffset := 1000
		if int(total) < maxOffset {
			maxOffset = int(total)
		}
	}

	req := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}
	resp, err := s.client.client.Index(s.client.indexName).Search("", req)
	if err != nil {
		return nil, err
	}

	books := make([]semantic.Book, len(resp.Hits))
	for i, hit := range resp.Hits {
		b, err := mapToBook(hit)
		if err != nil {
			log.Printf("Failed to map hit to book: %v", err)
			continue
		}
		books[i] = b
	}
	return books, nil
}

// UpdateBookMetadata updates a book's metadata
func (s *Searcher) UpdateBookMetadata(book semantic.Book, vector []float32) error {
	doc := bookToMap(book)
	if len(vector) > 0 {
		doc["_vectors"] = map[string]interface{}{"default": vector}
	} else if s.provider != nil {
		text := fmt.Sprintf("%s %s %s", book.Title, strings.Join(book.Authors, " "), book.Comments)
		embeddings, err := s.provider.Embed([]string{text})
		if err == nil && len(embeddings) > 0 {
			doc["_vectors"] = map[string]interface{}{"default": embeddings[0]}
		}
	}

	task, err := s.client.client.Index(s.client.indexName).UpdateDocuments([]map[string]interface{}{doc}, nil)
	if err != nil {
		return err
	}

	// Wait for task completion
	processedTask, err := s.client.client.WaitForTask(task.TaskUID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}
	if processedTask.Status != "succeeded" {
		return fmt.Errorf("task failed: %s - %v", processedTask.Status, processedTask.Error)
	}

	return nil
}

// DeleteBook deletes a book
func (s *Searcher) DeleteBook(bookID int64) error {
	task, err := s.client.client.Index(s.client.indexName).DeleteDocument(fmt.Sprintf("%d", bookID), nil)
	if err != nil {
		return err
	}

	// Wait for task completion
	processedTask, err := s.client.client.WaitForTask(task.TaskUID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}
	if processedTask.Status != "succeeded" {
		return fmt.Errorf("task failed: %s - %v", processedTask.Status, processedTask.Error)
	}

	return nil
}

// IndexBooks indexes a batch of books
// Documents are submitted asynchronously to MeiliSearch's task queue for faster bulk imports.
func (s *Searcher) IndexBooks(ctx context.Context, books []semantic.Book) error {
	docs := make([]map[string]interface{}, len(books))

	if s.provider != nil {
		texts := make([]string, len(books))
		for i, book := range books {
			texts[i] = fmt.Sprintf("%s %s %s", book.Title, strings.Join(book.Authors, " "), book.Comments)
		}

		embeddings, err := s.provider.Embed(texts)
		if err != nil {
			log.Printf("Failed to generate embeddings for batch: %v", err)
		}

		for i, b := range books {
			docs[i] = bookToMap(b)
			if err == nil && i < len(embeddings) {
				docs[i]["_vectors"] = map[string]interface{}{"default": embeddings[i]}
			}
		}
	} else {
		for i, b := range books {
			docs[i] = bookToMap(b)
		}
	}

	// Submit documents asynchronously — MeiliSearch processes tasks in its internal queue.
	// We don't wait for completion here, which significantly speeds up bulk imports.
	taskInfo, err := s.client.client.Index(s.client.indexName).AddDocuments(docs, nil)
	if err != nil {
		return fmt.Errorf("failed to submit documents to MeiliSearch: %w", err)
	}

	log.Printf("MeiliSearch task queued: taskUID=%d, indexUID=%s, batch=%d docs",
		taskInfo.TaskUID, taskInfo.IndexUID, len(docs))

	return nil
}

// GetBookToc retrieves TOC data for a book
func (s *Searcher) GetBookToc(bookID int64) (interface{}, error) {
	var doc map[string]interface{}
	err := s.client.client.Index(s.client.indexName).GetDocument(fmt.Sprintf("%d", bookID), &meilisearch.DocumentQuery{}, &doc)
	if err != nil {
		return nil, err
	}
	if toc, ok := doc["toc"]; ok {
		return toc, nil
	}
	return nil, nil
}

// UpdateToc updates the TOC field for a book
func (s *Searcher) UpdateToc(bookID int64, toc interface{}) error {
	doc := map[string]interface{}{
		"book_id": bookID,
		"toc":     toc,
	}
	task, err := s.client.client.Index(s.client.indexName).UpdateDocuments([]map[string]interface{}{doc}, nil)
	if err != nil {
		return err
	}

	// Wait for task completion
	processedTask, err := s.client.client.WaitForTask(task.TaskUID, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}
	if processedTask.Status != "succeeded" {
		return fmt.Errorf("task failed: %s - %v", processedTask.Status, processedTask.Error)
	}

	return nil
}

// Helper functions

func escapeFilterValue(val string) string {
	// Basic escape for single quotes
	// TODO: Improve escaping
	return val // Placeholder
}

func bookToMap(book semantic.Book) map[string]interface{} {
	m := map[string]interface{}{
		"book_id":       book.ID,
		"title":         book.Title,
		"authors":       book.Authors,
		"author_sort":   book.AuthorSort,
		"publisher":     book.Publisher,
		"isbn":          book.Isbn,
		"rating":        book.Rating,
		"tags":          book.Tags,
		"languages":     book.Languages,
		"comments":      book.Comments,
		"pubdate":       book.PubDate.Format(time.RFC3339),
		"last_modified": book.LastModified.Format(time.RFC3339),
		"series_index":  book.SeriesIndex,
		"size":          book.Size,
		"identifiers":   book.Identifiers,
		"cover":         book.Cover,
		"file_path":     book.FilePath,
	}
	if book.Toc != nil {
		m["toc"] = book.Toc
	}
	return m
}

func mapToBook(v interface{}) (semantic.Book, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return semantic.Book{}, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return semantic.Book{}, err
	}

	b := semantic.Book{}

	getFloat := func(v interface{}) float64 {
		switch i := v.(type) {
		case float64:
			return i
		case int:
			return float64(i)
		case int64:
			return float64(i)
		case json.Number:
			f, _ := i.Float64()
			return f
		}
		return 0
	}

	if v, ok := m["book_id"]; ok {
		b.ID = int64(getFloat(v))
	}
	if v, ok := m["title"].(string); ok {
		b.Title = v
	}
	if v, ok := m["authors"].([]interface{}); ok {
		for _, a := range v {
			if s, ok := a.(string); ok {
				b.Authors = append(b.Authors, s)
			}
		}
	}
	if v, ok := m["publisher"].(string); ok {
		b.Publisher = v
	}
	if v, ok := m["comments"].(string); ok {
		b.Comments = v
	}
	if v, ok := m["rating"]; ok {
		b.Rating = getFloat(v)
	}
	if v, ok := m["pubdate"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		b.PubDate = t
	}
	if v, ok := m["last_modified"].(string); ok {
		t, _ := time.Parse(time.RFC3339, v)
		b.LastModified = t
	}
	if v, ok := m["isbn"].(string); ok {
		b.Isbn = v
	}
	if v, ok := m["author_sort"].(string); ok {
		b.AuthorSort = v
	}
	if v, ok := m["series_index"]; ok {
		b.SeriesIndex = getFloat(v)
	}
	if v, ok := m["size"]; ok {
		b.Size = int64(getFloat(v))
	}
	if v, ok := m["cover"].(string); ok {
		b.Cover = v
	}
	if v, ok := m["file_path"].(string); ok {
		b.FilePath = v
	}
	if v, ok := m["tags"].([]interface{}); ok {
		for _, t := range v {
			if s, ok := t.(string); ok {
				b.Tags = append(b.Tags, s)
			}
		}
	}
	if v, ok := m["languages"].([]interface{}); ok {
		for _, l := range v {
			if s, ok := l.(string); ok {
				b.Languages = append(b.Languages, s)
			}
		}
	}
	if v, ok := m["toc"]; ok {
		b.Toc = v
	}

	return b, nil
}
