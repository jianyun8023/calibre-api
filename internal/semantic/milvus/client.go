package milvus

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type Client struct {
	conn           client.Client
	collectionName string
	vectorDim      int
}

func NewClient(collectionName string, vectorDim int) *Client {
	return &Client{
		collectionName: collectionName,
		vectorDim:      vectorDim,
	}
}

func (c *Client) Connect(host string, port int) error {
	ctx := context.Background()
	addr := fmt.Sprintf("%s:%d", host, port)

	conn, err := client.NewClient(ctx, client.Config{
		Address: addr,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %w", err)
	}

	c.conn = conn
	return c.InitCollection()
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) InitCollection() error {
	ctx := context.Background()

	has, err := c.conn.HasCollection(ctx, c.collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if has {
		return c.conn.LoadCollection(ctx, c.collectionName, false)
	}

	schema := &entity.Schema{
		CollectionName: c.collectionName,
		Description:    "Calibre books semantic search",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "book_id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: fmt.Sprintf("%d", c.vectorDim),
				},
			},
			{
				Name:     "title",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "512",
				},
			},
			{
				Name:     "authors",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "256",
				},
			},
			{
				Name:     "tags",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "512",
				},
			},
			{
				Name:     "summary",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "2048",
				},
			},
		},
	}

	if err := c.conn.CreateCollection(ctx, schema, entity.DefaultShardNumber); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	index, err := entity.NewIndexHNSW(entity.COSINE, 16, 200)
	if err != nil {
		return fmt.Errorf("failed to create index config: %w", err)
	}

	if err := c.conn.CreateIndex(ctx, c.collectionName, "embedding", index, false); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	if err := c.conn.LoadCollection(ctx, c.collectionName, false); err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	return nil
}

func (c *Client) InsertBooks(books []semantic.BookEmbedding) error {
	if len(books) == 0 {
		return nil
	}

	ctx := context.Background()

	bookIDs := make([]int64, len(books))
	vectors := make([][]float32, len(books))
	titles := make([]string, len(books))
	authors := make([]string, len(books))
	tags := make([]string, len(books))
	summaries := make([]string, len(books))

	for i, book := range books {
		bookIDs[i] = book.Book.ID
		vectors[i] = book.Embedding

		// Handle array fields by joining them
		authorsStr := strings.Join(book.Book.Authors, ", ")
		tagsStr := strings.Join(book.Book.Tags, ", ")

		titles[i] = truncateUTF8ByBytes(cleanUTF8String(book.Book.Title), 512)
		authors[i] = truncateUTF8ByBytes(cleanUTF8String(authorsStr), 256)
		tags[i] = truncateUTF8ByBytes(cleanUTF8String(tagsStr), 512)
		summaries[i] = truncateUTF8ByBytes(cleanUTF8String(book.Book.Comments), 2048)
	}

	data := []entity.Column{
		entity.NewColumnInt64("book_id", bookIDs),
		entity.NewColumnFloatVector("embedding", c.vectorDim, vectors),
		entity.NewColumnVarChar("title", titles),
		entity.NewColumnVarChar("authors", authors),
		entity.NewColumnVarChar("tags", tags),
		entity.NewColumnVarChar("summary", summaries),
	}

	_, err := c.conn.Upsert(ctx, c.collectionName, "", data...)
	if err != nil {
		return fmt.Errorf("failed to upsert data: %w", err)
	}

	return nil
}

func (c *Client) Search(queryVector []float32, topK int) ([]semantic.SearchResult, error) {
	ctx := context.Background()

	searchVec := entity.FloatVector(queryVector)

	searchParams, err := entity.NewIndexHNSWSearchParam(32)
	if err != nil {
		return nil, fmt.Errorf("failed to create search params: %w", err)
	}

	results, err := c.conn.Search(
		ctx,
		c.collectionName,
		[]string{},
		"",
		[]string{"book_id", "title", "authors", "tags", "summary"},
		[]entity.Vector{searchVec},
		"embedding",
		entity.COSINE,
		topK,
		searchParams,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return []semantic.SearchResult{}, nil
	}

	var searchResults []semantic.SearchResult
	result := results[0]
	ids := result.IDs.(*entity.ColumnInt64).Data()

	for i := 0; i < len(ids); i++ {
		book := semantic.Book{
			ID: ids[i],
		}

		if titleCol := result.Fields.GetColumn("title"); titleCol != nil {
			if col, ok := titleCol.(*entity.ColumnVarChar); ok && i < col.Len() {
				book.Title = col.Data()[i]
			}
		}
		if authorsCol := result.Fields.GetColumn("authors"); authorsCol != nil {
			if col, ok := authorsCol.(*entity.ColumnVarChar); ok && i < col.Len() {
				// Split back to array
				book.Authors = strings.Split(col.Data()[i], ", ")
			}
		}
		if tagsCol := result.Fields.GetColumn("tags"); tagsCol != nil {
			if col, ok := tagsCol.(*entity.ColumnVarChar); ok && i < col.Len() {
				// Split back to array
				book.Tags = strings.Split(col.Data()[i], ", ")
			}
		}
		if summaryCol := result.Fields.GetColumn("summary"); summaryCol != nil {
			if col, ok := summaryCol.(*entity.ColumnVarChar); ok && i < col.Len() {
				book.Comments = col.Data()[i]
			}
		}

		var distance float32
		if len(result.Scores) > i {
			distance = float32(result.Scores[i])
		}

		var similarity float64
		if distance <= 2.0 {
			similarity = 1.0 - (float64(distance) / 2.0)
		} else {
			similarity = 0.0
		}
		if similarity < 0 {
			similarity = 0
		}
		if similarity > 1 {
			similarity = 1
		}

		searchResults = append(searchResults, semantic.SearchResult{
			Book:  book,
			Score: float32(similarity),
			Rank:  i + 1,
		})
	}

	return searchResults, nil
}

func (c *Client) GetCollectionStats() (int64, error) {
	ctx := context.Background()
	expr := "book_id >= 0"
	results, err := c.conn.Query(ctx, c.collectionName, []string{}, expr, []string{"book_id"})
	if err != nil {
		has, err := c.conn.HasCollection(ctx, c.collectionName)
		if err != nil || !has {
			return 0, fmt.Errorf("collection not found or inaccessible")
		}
		return 0, nil
	}

	if len(results) > 0 {
		return int64(results[0].Len()), nil
	}

	return 0, nil
}

// GetMaxBookID returns the maximum book_id in the collection
func (c *Client) GetMaxBookID() (int64, error) {
	ctx := context.Background()

	// Check if collection exists and is loaded
	has, err := c.conn.HasCollection(ctx, c.collectionName)
	if err != nil || !has {
		return 0, fmt.Errorf("collection not found or inaccessible")
	}

	// Query all book_ids
	expr := "book_id >= 0"
	results, err := c.conn.Query(ctx, c.collectionName, []string{}, expr, []string{"book_id"})
	if err != nil {
		return 0, fmt.Errorf("failed to query book_ids: %w", err)
	}

	if len(results) == 0 {
		return 0, nil // No data in collection
	}

	// Get book_id column
	bookIDColumn := results[0].(*entity.ColumnInt64)
	if bookIDColumn.Len() == 0 {
		return 0, nil
	}

	// Find maximum book_id
	maxID := int64(0)
	for _, id := range bookIDColumn.Data() {
		if id > maxID {
			maxID = id
		}
	}

	return maxID, nil
}

// Helper functions from textutil.go

func truncateUTF8ByBytes(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}

	// If string is too long, cut it
	s = s[:maxBytes]

	// Check if we cut in the middle of a UTF-8 character
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if r != utf8.RuneError {
			return s
		}
		// Remove invalid byte at the end
		s = s[:len(s)-size]
	}
	return s
}

func cleanUTF8String(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	v := make([]rune, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				continue
			}
		}
		v = append(v, r)
	}
	return string(v)
}
